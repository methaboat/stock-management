package app

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Product struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	SKU          string             `bson:"sku" json:"sku"`
	Name         string             `bson:"name" json:"name"`
	Barcode      string             `bson:"barcode" json:"barcode"`
	Category     string             `bson:"category" json:"category"`
	Brand        string             `bson:"brand" json:"brand"`
	CostPrice    float64            `bson:"costPrice" json:"costPrice"`
	SellingPrice float64            `bson:"sellingPrice" json:"sellingPrice"`
	SupplierIDs  []string           `bson:"supplierIds" json:"supplierIds"`
	CreatedAt    time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt    time.Time          `bson:"updatedAt" json:"updatedAt"`
}

type StockLevel struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ProductID      primitive.ObjectID `bson:"productId" json:"productId"`
	SKU            string             `bson:"sku" json:"sku"`
	StockOnHand    float64            `bson:"stockOnHand" json:"stockOnHand"`
	CommittedStock float64            `bson:"committedStock" json:"committedStock"`
	AvailableStock float64            `bson:"availableStock" json:"availableStock"`
	ReorderPoint   float64            `bson:"reorderPoint" json:"reorderPoint"`
	Location       Location           `bson:"location" json:"location"`
}

type POItem struct {
	ProductID        primitive.ObjectID `bson:"productId" json:"productId"`
	SKU              string             `bson:"sku" json:"sku"`
	Quantity         float64            `bson:"quantity" json:"quantity"`
	UnitCost         float64            `bson:"unitCost" json:"unitCost"`
	ReceivedQuantity float64            `bson:"receivedQuantity" json:"receivedQuantity"`
}

type PurchaseOrder struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	PONumber     string             `bson:"poNumber" json:"poNumber"`
	SupplierName string             `bson:"supplierName" json:"supplierName"`
	Status       string             `bson:"status" json:"status"`
	Items        []POItem           `bson:"items" json:"items"`
	TotalAmount  float64            `bson:"totalAmount" json:"totalAmount"`
	CreatedAt    time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt    time.Time          `bson:"updatedAt" json:"updatedAt"`
}

type SOItem struct {
	ProductID       primitive.ObjectID `bson:"productId" json:"productId"`
	SKU             string             `bson:"sku" json:"sku"`
	Quantity        float64            `bson:"quantity" json:"quantity"`
	UnitPrice       float64            `bson:"unitPrice" json:"unitPrice"`
	ShippedQuantity float64            `bson:"shippedQuantity" json:"shippedQuantity"`
}

type SalesOrder struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	SONumber     string             `bson:"soNumber" json:"soNumber"`
	CustomerName string             `bson:"customerName" json:"customerName"`
	Status       string             `bson:"status" json:"status"`
	Items        []SOItem           `bson:"items" json:"items"`
	TotalAmount  float64            `bson:"totalAmount" json:"totalAmount"`
	CreatedAt    time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt    time.Time          `bson:"updatedAt" json:"updatedAt"`
}

type ReturnsRMA struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	SalesOrderID primitive.ObjectID `bson:"salesOrderId" json:"salesOrderId"`
	ProductID    primitive.ObjectID `bson:"productId" json:"productId"`
	SKU          string             `bson:"sku" json:"sku"`
	Quantity     float64            `bson:"quantity" json:"quantity"`
	Reason       string             `bson:"reason" json:"reason"`
	Status       string             `bson:"status" json:"status"`
	CreatedAt    time.Time          `bson:"createdAt" json:"createdAt"`
}
