package reconciliation

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"stock-management/backend/app"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) List(c *gin.Context) {
	list, err := h.svc.List(c.Request.Context())
	if err != nil {
		app.JSONErrorDefault(c, err)
		return
	}
	app.JSONSuccess(c, list)
}

func (h *Handler) Submit(c *gin.Context) {
	session := c.MustGet(string(app.UserContextKey)).(*app.UserSession)

	var req SubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		app.JSONErrorBadRequest(c, err)
		return
	}

	recon, err := h.svc.Submit(c.Request.Context(), session, req)
	if err != nil {
		app.JSONErrorDefault(c, err)
		return
	}
	c.JSON(http.StatusCreated, recon)
}

func (h *Handler) Approve(c *gin.Context) {
	session := c.MustGet(string(app.UserContextKey)).(*app.UserSession)
	id := c.Param("id")

	var req ApproveRequest
	c.ShouldBindJSON(&req)

	if err := h.svc.Approve(c.Request.Context(), session, id, req); err != nil {
		app.JSONErrorBadRequest(c, err)
		return
	}
	app.JSONSuccess(c, gin.H{"message": "Reconciliation approved"})
}
