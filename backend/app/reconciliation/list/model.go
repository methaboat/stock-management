package list

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ReconciliationItem struct {
	ProductID   primitive.ObjectID `bson:"productId" json:"productId"`
	SKU         string             `bson:"sku" json:"sku"`
	ProductName string             `bson:"productName" json:"productName"`
	ExpectedQty float64            `bson:"expectedQty" json:"expectedQty"`
	ActualQty   float64            `bson:"actualQty" json:"actualQty"`
	Difference  float64            `bson:"difference" json:"difference"`
}

type Reconciliation struct {
	ID          primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Date        time.Time            `bson:"date" json:"date"`
	Status      string               `bson:"status" json:"status"`
	Items       []ReconciliationItem `bson:"items" json:"items"`
	SubmittedBy string               `bson:"submittedBy" json:"submittedBy"`
	ApprovedBy  string               `bson:"approvedBy,omitempty" json:"approvedBy,omitempty"`
	ApprovedAt  *time.Time           `bson:"approvedAt,omitempty" json:"approvedAt,omitempty"`
	Note        string               `bson:"note" json:"note"`
	CreatedAt   time.Time            `bson:"createdAt" json:"createdAt"`
}
