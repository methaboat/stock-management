package so

import (
	"stock-management/backend/app"
)

type CreateRequest struct {
	CustomerName string       `json:"customerName" binding:"required"`
	Items        []app.SOItem `json:"items" binding:"required,dive"`
}

type Response = app.SalesOrder
