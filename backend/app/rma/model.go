package rma

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"stock-management/backend/app"
)

type CreateRequest struct {
	SalesOrderID primitive.ObjectID `json:"salesOrderId" binding:"required"`
	ProductID    primitive.ObjectID `json:"productId" binding:"required"`
	SKU          string             `json:"sku" binding:"required"`
	Quantity     float64            `json:"quantity" binding:"required,gt=0"`
	Reason       string             `json:"reason"`
	Status       string             `json:"status" binding:"required,oneof=Restocked Damaged"`
}

type Response = app.ReturnsRMA
