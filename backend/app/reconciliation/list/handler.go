package list

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
	list, err := h.storage.Execute(c.Request.Context())
	if err != nil {
		app.JSONErrorDefault(c, err)
		return
	}
	app.JSONSuccess(c, list)
}
