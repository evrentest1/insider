package task

import domain "github.com/evrentest1/insider/internal/business/domain/message"

type Message struct {
	ID          string `json:"id,omitempty"`
	Content     string `json:"content"`
	PhoneNumber string `json:"to"`
}

func FromDomainMessages(msg []domain.Message) []Message {
	messages := make([]Message, len(msg))
	for i := range msg {
		messages[i] = Message{
			ID:          msg[i].ID,
			Content:     msg[i].Content,
			PhoneNumber: msg[i].PhoneNumber,
		}
	}
	return messages
}
