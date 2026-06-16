package adjust

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
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

func (s *Storage) Execute(ctx context.Context, session *app.UserSession, req Request) error {
	objID, err := primitive.ObjectIDFromHex(req.StockID)
	if err != nil {
		return errors.New("invalid stock ID format")
	}

	var original app.StockLevel
	if err = s.col.FindOne(ctx, bson.M{"_id": objID}).Decode(&original); err != nil {
		return errors.New("stock record not found")
	}

	qtyChange := req.StockOnHand - original.StockOnHand
	availableStock := req.StockOnHand - original.CommittedStock

	update := bson.M{"$set": bson.M{
		"stockOnHand":    req.StockOnHand,
		"availableStock": availableStock,
		"reorderPoint":   req.ReorderPoint,
		"location":       req.Location,
	}}

	if _, err = s.col.UpdateOne(ctx, bson.M{"_id": objID}, update); err != nil {
		return err
	}

	app.LogAudit(ctx, session.Username, "Adjust", "Stock", objID.Hex(), original.SKU, qtyChange, req.Note)
	return nil
}
