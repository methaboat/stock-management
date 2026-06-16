package update

import (
	"context"
	"errors"
	"time"

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

func (s *Storage) Execute(ctx context.Context, session *app.UserSession, id string, req Request) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid product ID format")
	}

	update := bson.M{"$set": bson.M{
		"name":         req.Name,
		"barcode":      req.Barcode,
		"category":     req.Category,
		"brand":        req.Brand,
		"costPrice":    req.CostPrice,
		"sellingPrice": req.SellingPrice,
		"supplierIds":  req.SupplierIDs,
		"updatedAt":    time.Now(),
	}}

	var updated app.Product
	if err = s.col.FindOneAndUpdate(ctx, bson.M{"_id": objID}, update).Decode(&updated); err != nil {
		return errors.New("product not found")
	}

	app.LogAudit(ctx, session.Username, "Update", "Product", objID.Hex(), updated.SKU, 0, "Updated product fields")
	return nil
}
