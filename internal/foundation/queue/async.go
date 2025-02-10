package queue

import (
	"github.com/hibiken/asynq"
	"github.com/sirupsen/logrus"
)

type Queue struct {
	Server    *asynq.Server
	Client    *asynq.Client
	Scheduler *asynq.Scheduler
}

func New(redisAddress string, log *logrus.Logger) *Queue {
	redisConnOpt := asynq.RedisClientOpt{
		Addr: redisAddress,
	}
	schedulerOpts := &asynq.SchedulerOpts{
		Logger: log,
	}

	srv := asynq.NewServer(redisConnOpt, asynq.Config{})
	c := asynq.NewClient(redisConnOpt)
	schedule := asynq.NewScheduler(redisConnOpt, schedulerOpts)
	return &Queue{
		Server:    srv,
		Client:    c,
		Scheduler: schedule,
	}
}
