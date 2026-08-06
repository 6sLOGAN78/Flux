package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

type HealthHandler struct {
	dbPool *pgxpool.Pool
}

func NewHealthHandler(dbPool *pgxpool.Pool) *HealthHandler {
	return &HealthHandler{dbPool: dbPool}
}

func (h *HealthHandler) HealthCheck(c echo.Context) error {
	dbState := "connected"
	if h.dbPool == nil {
		dbState = "disconnected"
	} else {
		ctx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
		defer cancel()
		if err := h.dbPool.Ping(ctx); err != nil {
			dbState = "disconnected"
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":   "ok",
		"database": dbState,
	})
}
