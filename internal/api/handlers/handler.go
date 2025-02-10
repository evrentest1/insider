package handlers

import (
	"github.com/evrentest1/insider/internal/app/domain/message"
	messagerepository "github.com/evrentest1/insider/internal/business/domain/message/stores/db"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	swagger "github.com/swaggo/echo-swagger"
)

type Handler struct {
	log                            *logrus.Logger
	echo                           *echo.Echo
	dbChecker                      HealthChecker
	cacheChecker                   HealthChecker
	messageGetter                  MessageGetter
	messageSenderServiceController ServiceController
}

func New(l *logrus.Logger, e *echo.Echo, dbChecker, cacheChecker HealthChecker, repository *messagerepository.Repository, messageSenderServiceController ServiceController) *Handler {
	return &Handler{
		log:                            l,
		echo:                           e,
		dbChecker:                      dbChecker,
		cacheChecker:                   cacheChecker,
		messageGetter:                  message.NewService(repository),
		messageSenderServiceController: messageSenderServiceController,
	}
}

// RegisterRoutes registers all routes for the API
//
//	@title		API
//	@version	1.0
//
//	@host		localhost:8080
//	@schemes	http
func (h *Handler) RegisterRoutes() {
	health := NewHealthHandler(h.dbChecker, h.cacheChecker)
	messages := NewMessageHandler(h.messageGetter)
	messageSender := NewMessageSenderServiceHandler(h.messageSenderServiceController)

	h.echo.POST("/v1/service/message-sender/start", messageSender.start)
	h.echo.POST("/v1/service/message-sender/stop", messageSender.stop)
	h.echo.GET("/v1/messages/sent", messages.handler)
	h.echo.GET("/v1/health", health.handler)
	h.echo.GET("/swagger/*", swagger.WrapHandler)
}
