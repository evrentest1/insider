package message

import (
	"errors"

	"github.com/evrentest1/insider/internal/business/types/deliverystatus"
)

var ErrMessageNotFound = errors.New("messages not found")

type Message struct {
	ID          string
	PhoneNumber string
	Content     string
	Status      deliverystatus.DeliveryStatus
}
