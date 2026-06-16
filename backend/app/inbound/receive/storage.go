package receive

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
	poCol    *mongo.Collection
	stockCol *mongo.Collection
}

func NewStorage(poCol, stockCol *mongo.Collection) *Storage {
	return &Storage{poCol: poCol, stockCol: stockCol}
}

func (s *Storage) Execute(ctx context.Context, session *app.UserSession, id string, req Request) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid PO ID format")
	}

	var po app.PurchaseOrder
	if err = s.poCol.FindOne(ctx, bson.M{"_id": objID}).Decode(&po); err != nil {
		return errors.New("PO not found")
	}
	if po.Status != "Sent" {
		return errors.New("can only receive goods for POs in Sent status")
	}

	allReceived := true
	for i, item := range po.Items {
		if newReceive, ok := req.ReceivedQty[item.SKU]; ok && newReceive > 0 {
			po.Items[i].ReceivedQuantity += newReceive
			_, _ = s.stockCol.UpdateOne(ctx, bson.M{"productId": item.ProductID}, bson.M{
				"$inc": bson.M{"stockOnHand": newReceive, "availableStock": newReceive},
			})
			app.LogAudit(ctx, session.Username, "Receive", "Stock", item.ProductID.Hex(), item.SKU, newReceive, fmt.Sprintf("Received shipment for PO %s", po.PONumber))
		}
		if po.Items[i].ReceivedQuantity < item.Quantity {
			allReceived = false
		}
	}

	newStatus := "Sent"
	if allReceived {
		newStatus = "Received"
	}

	_, err = s.poCol.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": bson.M{
		"items":     po.Items,
		"status":    newStatus,
		"updatedAt": time.Now(),
	}})
	return err
}
