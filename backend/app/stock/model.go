package stock

import (
	"stock-management/backend/app"
)

type AdjustRequest struct {
	StockID      string       `json:"stockId" binding:"required"`
	StockOnHand  float64      `json:"stockOnHand" binding:"required,min=0"`
	ReorderPoint float64      `json:"reorderPoint" binding:"required"`
	Location     app.Location `json:"location"`
	Note         string       `json:"note" binding:"required"`
}

type ScanRequest struct {
	Barcode  string  `json:"barcode" binding:"required"`
	Status   string  `json:"status" binding:"required,oneof=in out"`
	Quantity float64 `json:"quantity" binding:"required,gt=0"`
	Note     string  `json:"note"`
}

type ScanResponse struct {
	Product     app.Product    `json:"product"`
	Stock       app.StockLevel `json:"stock"`
	QuantityChanged float64    `json:"quantityChanged"`
}

type Response = app.StockLevel
