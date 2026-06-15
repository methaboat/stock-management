package audit

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

func (h *Handler) List(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	list, err := h.service.List(ctx)
	if err != nil {
		app.JSONErrorDefault(c, err)
		return
	}
	app.JSONSuccess(c, list)
}
