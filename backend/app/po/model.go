package po

import (
	"stock-management/backend/app"
)

type CreateRequest struct {
	SupplierName string       `json:"supplierName" binding:"required"`
	Items        []app.POItem `json:"items" binding:"required,dive"`
}

type ReceiveRequest struct {
	ReceivedQty map[string]float64 `json:"receivedQty" binding:"required"`
}

type Response = app.PurchaseOrder
