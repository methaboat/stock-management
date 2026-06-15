package rma

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

func (h *Handler) Create(c *gin.Context) {
	session := c.MustGet(string(app.UserContextKey)).(*app.UserSession)
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		app.JSONErrorBadRequest(c, err)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.service.Create(ctx, session, req)
	if err != nil {
		app.JSONErrorDefault(c, err)
		return
	}

	c.JSON(http.StatusCreated, resp)
}
