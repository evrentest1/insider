package message

import domain "github.com/evrentest1/insider/internal/business/domain/message"

type Message struct {
	Content     string `json:"content"`
	PhoneNumber string `json:"phoneNumber"`
}

func ToMessages(msgs []domain.Message) []Message {
	messages := make([]Message, len(msgs))
	for i := range msgs {
		messages[i] = Message{
			Content:     msgs[i].Content,
			PhoneNumber: msgs[i].PhoneNumber,
		}
	}
	return messages
}
