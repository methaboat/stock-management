package scan

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
	session := c.MustGet(string(app.UserContextKey)).(*app.UserSession)
	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		app.JSONErrorBadRequest(c, err)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	result, err := h.storage.Execute(ctx, session, req)
	if err != nil {
		app.JSONErrorBadRequest(c, err)
		return
	}
	app.JSONSuccess(c, result)
}
