package so

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"stock-management/backend/app"
)

type Service interface {
	List(ctx context.Context) ([]Response, error)
	Create(ctx context.Context, session *app.UserSession, req CreateRequest) (*Response, error)
	Ship(ctx context.Context, session *app.UserSession, id string) error
}

type soService struct {
	soCol    *mongo.Collection
	stockCol *mongo.Collection
}

func NewService(soCol, stockCol *mongo.Collection) Service {
	return &soService{soCol: soCol, stockCol: stockCol}
}

func (s *soService) List(ctx context.Context) ([]Response, error) {
	cursor, err := s.soCol.Find(ctx, bson.M{})
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

func (s *soService) Create(ctx context.Context, session *app.UserSession, req CreateRequest) (*Response, error) {
	// 1. Verify availability of stock first
	for _, item := range req.Items {
		var stock app.StockLevel
		err := s.stockCol.FindOne(ctx, bson.M{"productId": item.ProductID}).Decode(&stock)
		if err != nil {
			return nil, fmt.Errorf("stock record not found for SKU %s", item.SKU)
		}

		if stock.AvailableStock < item.Quantity {
			return nil, fmt.Errorf("insufficient stock for SKU %s (Available: %v, Requested: %v)", item.SKU, stock.AvailableStock, item.Quantity)
		}
	}

	// 2. Reserve stock in Database (Decrement AvailableStock, Increment CommittedStock)
	for _, item := range req.Items {
		stockUpdate := bson.M{
			"$inc": bson.M{
				"availableStock": -item.Quantity,
				"committedStock": item.Quantity,
			},
		}
		_, err := s.stockCol.UpdateOne(ctx, bson.M{"productId": item.ProductID}, stockUpdate)
		if err != nil {
			return nil, err
		}

		app.LogAudit(ctx, session.Username, "Reserve", "Stock", item.ProductID.Hex(), item.SKU, -item.Quantity, "Reserved stock for customer purchase")
	}

	// 3. Save Sales Order
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

	_, err := s.soCol.InsertOne(ctx, so)
	if err != nil {
		return nil, err
	}

	app.LogAudit(ctx, session.Username, "Create", "SO", so.ID.Hex(), "", 0, "Created sales order "+so.SONumber)
	return &so, nil
}

func (s *soService) Ship(ctx context.Context, session *app.UserSession, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid SO ID format")
	}

	var so app.SalesOrder
	err = s.soCol.FindOne(ctx, bson.M{"_id": objID}).Decode(&so)
	if err != nil {
		return errors.New("sales order not found")
	}

	if so.Status != "Pending" && so.Status != "Processing" {
		return errors.New("can only ship orders in Pending or Processing state")
	}

	// Deduct stock counts from shelves
	for i, item := range so.Items {
		so.Items[i].ShippedQuantity = item.Quantity

		stockUpdate := bson.M{
			"$inc": bson.M{
				"stockOnHand":    -item.Quantity,
				"committedStock": -item.Quantity,
			},
		}
		_, _ = s.stockCol.UpdateOne(ctx, bson.M{"productId": item.ProductID}, stockUpdate)

		app.LogAudit(ctx, session.Username, "Ship", "Stock", item.ProductID.Hex(), item.SKU, -item.Quantity, fmt.Sprintf("Picked, packed, and shipped for %s", so.SONumber))
	}

	update := bson.M{
		"$set": bson.M{
			"status":    "Shipped",
			"items":     so.Items,
			"updatedAt": time.Now(),
		},
	}

	_, err = s.soCol.UpdateOne(ctx, bson.M{"_id": objID}, update)
	return err
}
