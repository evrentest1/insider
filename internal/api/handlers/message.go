package handlers

import (
	"context"
	"net/http"

	"github.com/evrentest1/insider/internal/app/domain/message"

	"github.com/labstack/echo/v4"
)

type MessageGetter interface {
	GetSentMessages(ctx context.Context) ([]message.Message, error)
}
type MessageHandler struct {
	messageGetter MessageGetter
}

func NewMessageHandler(m MessageGetter) *MessageHandler {
	return &MessageHandler{
		messageGetter: m,
	}
}

// handler godoc
//
//	@Summary		Get sent messages
//	@Description	Get sent messages
//	@Tags			messages
//	@Accept			json
//	@Produce		json
//	@Success		200
//	@Failure		404	{object}	echo.HTTPError	"No messages found"
//	@Failure		500	{object}	echo.HTTPError	"Failed to get messages"
//	@Router			/v1/messages/sent [get] []message.Message
func (h *MessageHandler) handler(c echo.Context) error {
	msgs, err := h.messageGetter.GetSentMessages(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get messages").SetInternal(err)
	}

	if len(msgs) == 0 {
		return echo.NewHTTPError(http.StatusNotFound, "No messages found")
	}

	return c.JSON(http.StatusOK, msgs)
}
