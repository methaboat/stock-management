package ship

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

type Storage struct {
	soCol    *mongo.Collection
	stockCol *mongo.Collection
}

func NewStorage(soCol, stockCol *mongo.Collection) *Storage {
	return &Storage{soCol: soCol, stockCol: stockCol}
}

func (s *Storage) Execute(ctx context.Context, session *app.UserSession, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid SO ID format")
	}

	var so app.SalesOrder
	if err = s.soCol.FindOne(ctx, bson.M{"_id": objID}).Decode(&so); err != nil {
		return errors.New("sales order not found")
	}
	if so.Status != "Pending" && so.Status != "Processing" {
		return errors.New("can only ship orders in Pending or Processing state")
	}

	for i, item := range so.Items {
		so.Items[i].ShippedQuantity = item.Quantity
		_, _ = s.stockCol.UpdateOne(ctx, bson.M{"productId": item.ProductID}, bson.M{
			"$inc": bson.M{"stockOnHand": -item.Quantity, "committedStock": -item.Quantity},
		})
		app.LogAudit(ctx, session.Username, "Ship", "Stock", item.ProductID.Hex(), item.SKU, -item.Quantity, fmt.Sprintf("Picked, packed, and shipped for %s", so.SONumber))
	}

	_, err = s.soCol.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": bson.M{
		"status":    "Shipped",
		"items":     so.Items,
		"updatedAt": time.Now(),
	}})
	return err
}
