package deletion

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"stock-management/backend/app"
)

type Storage struct {
	prodCol  *mongo.Collection
	stockCol *mongo.Collection
}

func NewStorage(prodCol, stockCol *mongo.Collection) *Storage {
	return &Storage{prodCol: prodCol, stockCol: stockCol}
}

func (s *Storage) Execute(ctx context.Context, session *app.UserSession, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid product ID format")
	}

	var p app.Product
	if err = s.prodCol.FindOneAndDelete(ctx, bson.M{"_id": objID}).Decode(&p); err != nil {
		return errors.New("product not found")
	}

	_, _ = s.stockCol.DeleteOne(ctx, bson.M{"productId": objID})
	app.LogAudit(ctx, session.Username, "Delete", "Product", objID.Hex(), p.SKU, 0, "Deleted product profile and associated stock level")
	return nil
}
