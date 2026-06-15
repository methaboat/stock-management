package auth

import (
	"context"
	"net/http"
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

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		app.JSONErrorBadRequest(c, err)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.service.Authenticate(ctx, req)
	if err != nil {
		app.JSONErrorUnauthorized(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}
