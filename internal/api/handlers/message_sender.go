package handlers

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
)

type ServiceController interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context)
}

type MessageSender struct {
	messageSenderServiceController ServiceController
}

func NewMessageSenderServiceHandler(sc ServiceController) *MessageSender {
	return &MessageSender{
		messageSenderServiceController: sc,
	}
}

// start godoc
//
//	@Summary		Start auto message sending
//	@Description	Start auto message sending
//	@Tags			auto message sending
//	@Accept			json
//	@Produce		json
//	@Success		200
//	@Failure		500	{object}	echo.HTTPError	"Failed to start auto message sending"
//	@Router			/v1/service/message-sender/start [post]
func (h *MessageSender) start(c echo.Context) error {
	err := h.messageSenderServiceController.Start(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to start auto message sending").SetInternal(err)
	}

	return c.NoContent(http.StatusOK)
}

// stop godoc
//
//	@Summary		Stop auto message sending
//	@Description	Stop auto message sending
//	@Tags			auto message sending
//	@Accept			json
//	@Produce		json
//	@Success		200
//	@Router			/v1/service/message-sender/stop [post]
func (h *MessageSender) stop(c echo.Context) error {
	h.messageSenderServiceController.Stop(c.Request().Context())

	return c.NoContent(http.StatusOK)
}
