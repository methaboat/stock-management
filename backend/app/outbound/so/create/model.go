package create

import "stock-management/backend/app"

type Request struct {
	CustomerName string       `json:"customerName" binding:"required"`
	Items        []app.SOItem `json:"items" binding:"required,dive"`
}

type Response = app.SalesOrder
