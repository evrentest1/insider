package task

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	messagerepository "github.com/evrentest1/insider/internal/business/domain/message/stores/db"
	"github.com/evrentest1/insider/internal/business/types/deliverystatus"
	"github.com/evrentest1/insider/internal/config"
	"github.com/evrentest1/insider/internal/foundation/httpkit"
	"github.com/evrentest1/insider/internal/foundation/queue"

	"github.com/hibiken/asynq"
	"github.com/sirupsen/logrus"
)

const (
	typeMessageFetcherTask = "task:message:fetcher"
	typeMessageSenderTask  = "task:message:sender"

	limit         = 2
	retryDuration = 30 * time.Second
)

var (
	errFailedToStartAsynqServer        = fmt.Errorf("failed to start asynq server after 30 seconds")
	errContextCancelledStopAsynqServer = fmt.Errorf("context canceled: stopping asynq server startup")
)

type Service struct {
	redisAddress      string
	scheduler         *asynq.Scheduler
	repository        *messagerepository.Repository
	server            *asynq.Server
	client            *asynq.Client
	log               *logrus.Logger
	httpRequester     httpkit.RequesterService
	configWebHookSite config.WebHookSite
}

func NewService(redisAddress string, log *logrus.Logger, r *messagerepository.Repository, httpRequester httpkit.RequesterService, configWebHookSite config.WebHookSite) *Service {
	s := &Service{
		redisAddress:      redisAddress,
		repository:        r,
		log:               log,
		httpRequester:     httpRequester,
		configWebHookSite: configWebHookSite,
	}

	q := queue.New(redisAddress, log)
	s.server = q.Server
	s.client = q.Client
	s.scheduler = q.Scheduler
	return s
}

func (s *Service) Start(ctx context.Context) error {
	if err := s.startQueueServer(ctx); err != nil {
		return fmt.Errorf("start queue server: %w", err)
	}

	if err := s.messageFetcherHandler(ctx, nil); err != nil {
		return fmt.Errorf("message fetcher handler: %w", err)
	}

	if _, err := s.scheduler.Register("*/2 * * * *", s.newMessageFetcherTask()); err != nil {
		return fmt.Errorf("register a task: %w", err)
	}

	if err := s.scheduler.Start(); err != nil {
		if err.Error() != "asynq: the scheduler is already running" {
			return fmt.Errorf("start scheduler: %w", err)
		}
	}

	return nil
}

func (s *Service) Stop(_ context.Context) {
	s.scheduler.Shutdown()
	s.server.Shutdown()
}

func (s *Service) newMessageFetcherTask() *asynq.Task {
	return asynq.NewTask(typeMessageFetcherTask, nil)
}

func (s *Service) messageFetcherHandler(ctx context.Context, _ *asynq.Task) error {
	messages, err := s.repository.GetMessagesByStatusAndLimit(ctx, deliverystatus.Pending, limit)
	if err != nil {
		return fmt.Errorf("get messages by status and limit: %w", err)
	}

	for _, msg := range FromDomainMessages(messages) {
		task, err := s.newMessageSenderTask(msg)
		if err != nil {
			s.log.Error("new message sender task: %w", err)
		}
		if _, err := s.client.Enqueue(task, asynq.TaskID(msg.ID)); err != nil {
			s.log.Error("enqueue task: %w", err)
		}

		if err := s.repository.UpdateMessageStatus(ctx, deliverystatus.InProgress, msg.ID); err != nil {
			s.log.Error("update message status: %w", err)
		}
	}

	return nil
}

func (s *Service) newMessageSenderTask(message Message) (*asynq.Task, error) {
	payload, err := json.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}
	return asynq.NewTask(typeMessageSenderTask, payload), nil
}

func (s *Service) messageSenderHandler(ct context.Context, t *asynq.Task) error {
	var msg Message
	if err := json.Unmarshal(t.Payload(), &msg); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	type Resp struct {
		Message   string `json:"message"`
		MessageID string `json:"messageId"`
	}
	var resp Resp
	_, _, _, err := s.httpRequester.Post(ct, s.configWebHookSite.URL, map[string]string{"x-ins-auth-key": s.configWebHookSite.Key}, msg, &resp)
	if err != nil {
		if err := s.repository.UpdateMessageStatus(ct, deliverystatus.Failed, msg.ID); err != nil {
			return fmt.Errorf("update message status: %w", err)
		}
		return fmt.Errorf("post request: %w", err)
	}
	// TODO: cache the response

	if err := s.repository.UpdateMessageID(ct, msg.ID, resp.MessageID); err != nil {
		return fmt.Errorf("update message status: %w", err)
	}

	return nil
}

func (s *Service) startQueueServer(ctx context.Context) error {
	mux := asynq.NewServeMux()
	mux.HandleFunc(typeMessageFetcherTask, s.messageFetcherHandler)
	mux.HandleFunc(typeMessageSenderTask, s.messageSenderHandler)

	startTime := time.Now()

	for {
		// Stop retrying if 30 seconds have passed
		if time.Since(startTime) > retryDuration {
			return errFailedToStartAsynqServer
		}

		// Stop retrying if the context is canceled
		select {
		case <-ctx.Done():
			return errContextCancelledStopAsynqServer
		default:
		}

		err := s.server.Start(mux)
		if err != nil {
			if err.Error() == "asynq: Server closed" {
				s.log.Warn("Asynq server closed, attempting restart...")

				// Reinitialize underlying components
				svc := NewService(s.redisAddress, s.log, s.repository, s.httpRequester, s.configWebHookSite)
				s.server = svc.server
				s.scheduler = svc.scheduler
				s.client = svc.client

				// Continue retrying immediately without sleep
				continue
			}

			if err.Error() == "asynq: the server is already running" {
				s.log.Warn("Asynq server is already running, skipping restart.")
				return nil
			}

			return fmt.Errorf("start worker: %w", err)
		}

		// Server started successfully
		break
	}

	return nil
}
