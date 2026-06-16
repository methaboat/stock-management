package create

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"stock-management/backend/app"
)

type Storage struct {
	rmaCol   *mongo.Collection
	stockCol *mongo.Collection
}

func NewStorage(rmaCol, stockCol *mongo.Collection) *Storage {
	return &Storage{rmaCol: rmaCol, stockCol: stockCol}
}

func (s *Storage) Execute(ctx context.Context, session *app.UserSession, req Request) (*Response, error) {
	rma := app.ReturnsRMA{
		ID:           primitive.NewObjectID(),
		SalesOrderID: req.SalesOrderID,
		ProductID:    req.ProductID,
		SKU:          req.SKU,
		Quantity:     req.Quantity,
		Reason:       req.Reason,
		Status:       req.Status,
		CreatedAt:    time.Now(),
	}

	if _, err := s.rmaCol.InsertOne(ctx, rma); err != nil {
		return nil, err
	}

	if rma.Status == "Restocked" {
		_, _ = s.stockCol.UpdateOne(ctx, bson.M{"productId": rma.ProductID}, bson.M{
			"$inc": bson.M{"stockOnHand": rma.Quantity, "availableStock": rma.Quantity},
		})
		app.LogAudit(ctx, session.Username, "Return", "Stock", rma.ProductID.Hex(), rma.SKU, rma.Quantity, "Returned and restocked into warehouse shelves")
	} else {
		app.LogAudit(ctx, session.Username, "Return", "Stock", rma.ProductID.Hex(), rma.SKU, 0, "Returned merchandise logged as Damaged (not restocked)")
	}

	return &rma, nil
}
