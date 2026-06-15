package reports

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"stock-management/backend/app"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Valuation(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	report, err := h.service.GetValuation(ctx)
	if err != nil {
		app.JSONErrorDefault(c, err)
		return
	}
	app.JSONSuccess(c, report)
}
