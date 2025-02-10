package messagerepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/evrentest1/insider/internal/business/domain/message"
	db "github.com/evrentest1/insider/internal/business/domain/message/stores/db/postgres"
	"github.com/evrentest1/insider/internal/business/types/deliverystatus"
)

type Repository struct {
	queries *db.Queries
}

func New(pg *sql.DB) *Repository {
	return &Repository{
		queries: db.New(pg),
	}
}

func (r *Repository) GetMessagesByStatusAndLimit(ctx context.Context, status deliverystatus.DeliveryStatus, limit int32) ([]message.Message, error) {
	msgs, err := r.queries.GetMessagesByStatusWithLimit(ctx, db.GetMessagesByStatusWithLimitParams{
		Status: db.FromDomainStatus(status),
		Limit:  limit,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, message.ErrMessageNotFound
		}
		return nil, fmt.Errorf("get messages by status and limit: %w", err)
	}

	return db.ToDomainMessagesFromGetMessagesByStatusWithLimitRow(msgs), nil
}

func (r *Repository) GetMessagesByStatus(ctx context.Context, status deliverystatus.DeliveryStatus) ([]message.Message, error) {
	msgs, err := r.queries.GetAllMessagesByStatus(ctx, db.FromDomainStatus(status))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, message.ErrMessageNotFound
		}
		return nil, fmt.Errorf("get messages by status: %w", err)
	}

	return db.ToDomainMessagesFromGetAllMessagesByStatusRow(msgs), nil
}

func (r *Repository) UpdateMessageStatus(ctx context.Context, status deliverystatus.DeliveryStatus, id string) error {
	mid, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return fmt.Errorf("parse id: %w", err)
	}

	err = r.queries.UpdateMessageStatus(ctx, db.UpdateMessageStatusParams{
		Status: db.FromDomainStatus(status),
		ID:     mid,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return message.ErrMessageNotFound
		}
		return fmt.Errorf("update message status: %w", err)
	}

	return nil
}

func (r *Repository) UpdateMessageID(ctx context.Context, id, externalID string) error {
	mid, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return fmt.Errorf("parse id: %w", err)
	}

	err = r.queries.UpdateMessageId(ctx, db.UpdateMessageIdParams{
		Status: db.MessageSendingStatusSuccess,
		ID:     mid,
		MessageID: sql.NullString{
			String: externalID,
			Valid:  true,
		},
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return message.ErrMessageNotFound
		}
		return fmt.Errorf("update message id: %w", err)
	}

	return nil
}
