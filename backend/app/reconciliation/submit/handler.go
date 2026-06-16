package submit

import (
	"net/http"

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

	recon, err := h.storage.Execute(c.Request.Context(), session, req)
	if err != nil {
		app.JSONErrorDefault(c, err)
		return
	}
	c.JSON(http.StatusCreated, recon)
}
