package valuation

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"stock-management/backend/app"
)

type Handler struct {
	storage *Storage
}

func NewHandler(s *Storage) *Handler {
	return &Handler{storage: s}
}

func (h *Handler) Handle(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	report, err := h.storage.Execute(ctx)
	if err != nil {
		app.JSONErrorDefault(c, err)
		return
	}
	app.JSONSuccess(c, report)
}
