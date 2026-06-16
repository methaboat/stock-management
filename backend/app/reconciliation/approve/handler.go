package approve

import (
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

	var req Request
	c.ShouldBindJSON(&req)

	if err := h.storage.Execute(c.Request.Context(), session, id, req); err != nil {
		app.JSONErrorBadRequest(c, err)
		return
	}
	app.JSONSuccess(c, gin.H{"message": "Reconciliation approved"})
}
