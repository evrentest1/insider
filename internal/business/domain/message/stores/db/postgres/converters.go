package db

import (
	"fmt"

	domain "github.com/evrentest1/insider/internal/business/domain/message"
	"github.com/evrentest1/insider/internal/business/types/deliverystatus"
)

func ToDomainMessagesFromGetMessagesByStatusWithLimitRow(msgs []GetMessagesByStatusWithLimitRow) []domain.Message {
	messages := make([]domain.Message, len(msgs))
	for i := range msgs {
		messages[i] = domain.Message{
			ID:          fmt.Sprintf("%d", msgs[i].ID),
			PhoneNumber: msgs[i].PhoneNumber,
			Content:     msgs[i].Content,
		}
	}
	return messages
}

func ToDomainMessagesFromGetAllMessagesByStatusRow(msgs []GetAllMessagesByStatusRow) []domain.Message {
	messages := make([]domain.Message, len(msgs))
	for i := range msgs {
		messages[i] = domain.Message{
			PhoneNumber: msgs[i].PhoneNumber,
			Content:     msgs[i].Content,
		}
	}
	return messages
}

func FromDomainStatus(status deliverystatus.DeliveryStatus) MessageSendingStatus {
	switch status {
	case deliverystatus.Pending:
		return MessageSendingStatusPending
	case deliverystatus.Success:
		return MessageSendingStatusSuccess
	case deliverystatus.Failed:
		return MessageSendingStatusFailed
	case deliverystatus.InProgress:
		return MessageSendingStatusInProgress
	}
	return MessageSendingStatusPending
}
