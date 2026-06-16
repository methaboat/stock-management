package approve

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"stock-management/backend/app"
	reconlist "stock-management/backend/app/reconciliation/list"
)

type Storage struct {
	reconCol *mongo.Collection
	stockCol *mongo.Collection
}

func NewStorage(reconCol, stockCol *mongo.Collection) *Storage {
	return &Storage{reconCol: reconCol, stockCol: stockCol}
}

func (s *Storage) Execute(ctx context.Context, session *app.UserSession, id string, req Request) error {
	if session.Role != "Admin" && session.Role != "Manager" {
		return errors.New("only Admin or Manager can approve reconciliations")
	}

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid reconciliation ID")
	}

	var recon reconlist.Reconciliation
	if err = s.reconCol.FindOne(ctx, bson.M{"_id": objID}).Decode(&recon); err != nil {
		return errors.New("reconciliation not found")
	}
	if recon.Status == "Approved" {
		return errors.New("already approved")
	}

	now := time.Now()
	for _, item := range recon.Items {
		if item.Difference != 0 {
			_, _ = s.stockCol.UpdateOne(ctx,
				bson.M{"productId": item.ProductID},
				bson.M{"$set": bson.M{"stockOnHand": item.ActualQty, "availableStock": item.ActualQty}},
			)
		}
	}

	note := req.Note
	if note == "" {
		note = "Approved reconciliation"
	}

	_, err = s.reconCol.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": bson.M{
		"status":     "Approved",
		"approvedBy": session.Username,
		"approvedAt": now,
		"note":       note,
	}})
	if err != nil {
		return err
	}

	app.LogAudit(ctx, session.Username, "Approve", "Reconciliation", objID.Hex(), "", 0, note)
	return nil
}
