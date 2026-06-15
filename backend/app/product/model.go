package product

import (
	"stock-management/backend/app"
)

type CreateRequest struct {
	SKU          string   `json:"sku" binding:"required"`
	Name         string   `json:"name" binding:"required"`
	Barcode      string   `json:"barcode"`
	Category     string   `json:"category"`
	Brand        string   `json:"brand"`
	CostPrice    float64  `json:"costPrice" binding:"required,gt=0"`
	SellingPrice float64  `json:"sellingPrice" binding:"required,gt=0"`
	SupplierIDs  []string `json:"supplierIds"`
}

type UpdateRequest struct {
	Name         string   `json:"name" binding:"required"`
	Barcode      string   `json:"barcode"`
	Category     string   `json:"category"`
	Brand        string   `json:"brand"`
	CostPrice    float64  `json:"costPrice" binding:"required,gt=0"`
	SellingPrice float64  `json:"sellingPrice" binding:"required,gt=0"`
	SupplierIDs  []string `json:"supplierIds"`
}

type Response = app.Product
