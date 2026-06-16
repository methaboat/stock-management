package create

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"stock-management/backend/app"
)

type Storage struct {
	soCol    *mongo.Collection
	stockCol *mongo.Collection
}

func NewStorage(soCol, stockCol *mongo.Collection) *Storage {
	return &Storage{soCol: soCol, stockCol: stockCol}
}

func (s *Storage) Execute(ctx context.Context, session *app.UserSession, req Request) (*Response, error) {
	for _, item := range req.Items {
		var stock app.StockLevel
		if err := s.stockCol.FindOne(ctx, bson.M{"productId": item.ProductID}).Decode(&stock); err != nil {
			return nil, fmt.Errorf("stock record not found for SKU %s", item.SKU)
		}
		if stock.AvailableStock < item.Quantity {
			return nil, fmt.Errorf("insufficient stock for SKU %s (Available: %v, Requested: %v)", item.SKU, stock.AvailableStock, item.Quantity)
		}
	}

	for _, item := range req.Items {
		_, err := s.stockCol.UpdateOne(ctx, bson.M{"productId": item.ProductID}, bson.M{
			"$inc": bson.M{"availableStock": -item.Quantity, "committedStock": item.Quantity},
		})
		if err != nil {
			return nil, err
		}
		app.LogAudit(ctx, session.Username, "Reserve", "Stock", item.ProductID.Hex(), item.SKU, -item.Quantity, "Reserved stock for customer purchase")
	}

	so := app.SalesOrder{
		ID:           primitive.NewObjectID(),
		SONumber:     fmt.Sprintf("SO-%d", time.Now().UnixNano()/100000),
		CustomerName: req.CustomerName,
		Status:       "Pending",
		Items:        req.Items,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	var total float64
	for i, item := range so.Items {
		total += item.Quantity * item.UnitPrice
		so.Items[i].ShippedQuantity = 0
	}
	so.TotalAmount = total

	if _, err := s.soCol.InsertOne(ctx, so); err != nil {
		return nil, err
	}

	app.LogAudit(ctx, session.Username, "Create", "SO", so.ID.Hex(), "", 0, "Created sales order "+so.SONumber)
	return &so, nil
}
