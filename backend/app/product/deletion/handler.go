package deletion

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
	session := c.MustGet(string(app.UserContextKey)).(*app.UserSession)
	id := c.Param("id")
	if id == "" {
		app.JSONErrorBadRequest(c, nil)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if err := h.storage.Execute(ctx, session, id); err != nil {
		app.JSONErrorDefault(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
