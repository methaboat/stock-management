package create

import "stock-management/backend/app"

type Request struct {
	SupplierName string       `json:"supplierName" binding:"required"`
	Items        []app.POItem `json:"items" binding:"required,dive"`
}

type Response = app.PurchaseOrder
