package message

import (
	"context"
	"errors"
	"fmt"

	"github.com/evrentest1/insider/internal/business/domain/message"
	messagerepository "github.com/evrentest1/insider/internal/business/domain/message/stores/db"
	"github.com/evrentest1/insider/internal/business/types/deliverystatus"
)

type Service struct {
	repository *messagerepository.Repository
}

func NewService(r *messagerepository.Repository) *Service {
	return &Service{
		repository: r,
	}
}

func (s *Service) GetSentMessages(ctx context.Context) ([]Message, error) {
	messages, err := s.repository.GetMessagesByStatus(ctx, deliverystatus.Success)
	if err != nil {
		if errors.Is(err, message.ErrMessageNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("get messages by status and limit: %w", err)
	}

	return ToMessages(messages), nil
}
