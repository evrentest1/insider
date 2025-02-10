package handlers

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
)

type HealthChecker interface {
	IsOK(ctx context.Context) bool
}
type HealthHandler struct {
	db    HealthChecker
	cache HealthChecker
}

func NewHealthHandler(db, c HealthChecker) *HealthHandler {
	return &HealthHandler{
		db:    db,
		cache: c,
	}
}

// handler godoc
//
//	@Summary		Health Info
//	@Description	Health information
//	@Tags			health
//	@Accept			json
//	@Produce		json
//	@Success		200
//	@Failure		500	{object}	echo.HTTPError	"DB is not healthy"
//	@Failure		500	{object}	echo.HTTPError	"Cache is not healthy"
//	@Router			/v1/health [get]
func (h *HealthHandler) handler(c echo.Context) error {
	if ok := h.db.IsOK(c.Request().Context()); !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "DB is not healthy")
	}

	if ok := h.cache.IsOK(c.Request().Context()); !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "Cache is not healthy")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"DB":    "OK",
		"Cache": "OK",
	})
}
