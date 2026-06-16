package login

import (
	"context"
	"net/http"
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
	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		app.JSONErrorBadRequest(c, err)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.storage.Authenticate(ctx, req)
	if err != nil {
		app.JSONErrorUnauthorized(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}
