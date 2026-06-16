package scan

import "stock-management/backend/app"

type Request struct {
	Barcode  string  `json:"barcode" binding:"required"`
	Status   string  `json:"status" binding:"required,oneof=in out"`
	Quantity float64 `json:"quantity" binding:"required,gt=0"`
	Note     string  `json:"note"`
}

type Response struct {
	Product         app.Product    `json:"product"`
	Stock           app.StockLevel `json:"stock"`
	QuantityChanged float64        `json:"quantityChanged"`
}
