package send

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

func (s *Storage) Execute(ctx context.Context, session *app.UserSession, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid PO ID format")
	}

	var po app.PurchaseOrder
	if err = s.col.FindOne(ctx, bson.M{"_id": objID}).Decode(&po); err != nil {
		return errors.New("PO not found")
	}
	if po.Status != "Draft" {
		return errors.New("can only send POs in Draft status")
	}

	_, err = s.col.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": bson.M{
		"status":    "Sent",
		"updatedAt": time.Now(),
	}})
	if err != nil {
		return err
	}

	app.LogAudit(ctx, session.Username, "Send", "PO", objID.Hex(), "", 0, "Sent purchase order "+po.PONumber+" to supplier")
	return nil
}
