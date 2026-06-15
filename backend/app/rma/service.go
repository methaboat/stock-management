package rma

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"stock-management/backend/app"
)

type Service interface {
	List(ctx context.Context) ([]Response, error)
	Create(ctx context.Context, session *app.UserSession, req CreateRequest) (*Response, error)
}

type rmaService struct {
	rmaCol   *mongo.Collection
	stockCol *mongo.Collection
}

func NewService(rmaCol, stockCol *mongo.Collection) Service {
	return &rmaService{rmaCol: rmaCol, stockCol: stockCol}
}

func (s *rmaService) List(ctx context.Context) ([]Response, error) {
	cursor, err := s.rmaCol.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var list []Response = []Response{}
	if err = cursor.All(ctx, &list); err != nil {
		return nil, err
	}
	return list, nil
}

func (s *rmaService) Create(ctx context.Context, session *app.UserSession, req CreateRequest) (*Response, error) {
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

	_, err := s.rmaCol.InsertOne(ctx, rma)
	if err != nil {
		return nil, err
	}

	if rma.Status == "Restocked" {
		restockUpdate := bson.M{
			"$inc": bson.M{
				"stockOnHand":    rma.Quantity,
				"availableStock": rma.Quantity,
			},
		}
		_, _ = s.stockCol.UpdateOne(ctx, bson.M{"productId": rma.ProductID}, restockUpdate)

		app.LogAudit(ctx, session.Username, "Return", "Stock", rma.ProductID.Hex(), rma.SKU, rma.Quantity, "Returned and restocked into warehouse shelves")
	} else {
		app.LogAudit(ctx, session.Username, "Return", "Stock", rma.ProductID.Hex(), rma.SKU, 0, "Returned merchandise logged as Damaged (not restocked)")
	}

	return &rma, nil
}
