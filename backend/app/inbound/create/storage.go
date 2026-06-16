package create

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"stock-management/backend/app"
)

type Storage struct {
	col *mongo.Collection
}

func NewStorage(col *mongo.Collection) *Storage {
	return &Storage{col: col}
}

func (s *Storage) Execute(ctx context.Context, session *app.UserSession, req Request) (*Response, error) {
	po := app.PurchaseOrder{
		ID:           primitive.NewObjectID(),
		PONumber:     fmt.Sprintf("PO-%d", time.Now().UnixNano()/100000),
		SupplierName: req.SupplierName,
		Status:       "Draft",
		Items:        req.Items,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	var total float64
	for i, item := range po.Items {
		total += item.Quantity * item.UnitCost
		po.Items[i].ReceivedQuantity = 0
	}
	po.TotalAmount = total

	if _, err := s.col.InsertOne(ctx, po); err != nil {
		return nil, err
	}

	app.LogAudit(ctx, session.Username, "Create", "PO", po.ID.Hex(), "", 0, "Created purchase order "+po.PONumber)
	return &po, nil
}
