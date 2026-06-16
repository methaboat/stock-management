package adjust

import "stock-management/backend/app"

type Request struct {
	StockID      string       `json:"stockId" binding:"required"`
	StockOnHand  float64      `json:"stockOnHand" binding:"required,min=0"`
	ReorderPoint float64      `json:"reorderPoint" binding:"required"`
	Location     app.Location `json:"location"`
	Note         string       `json:"note" binding:"required"`
}
