package po

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
	Send(ctx context.Context, session *app.UserSession, id string) error
	Receive(ctx context.Context, session *app.UserSession, id string, req ReceiveRequest) error
}

type poService struct {
	poCol    *mongo.Collection
	stockCol *mongo.Collection
}

func NewService(poCol, stockCol *mongo.Collection) Service {
	return &poService{poCol: poCol, stockCol: stockCol}
}

func (s *poService) List(ctx context.Context) ([]Response, error) {
	cursor, err := s.poCol.Find(ctx, bson.M{})
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

func (s *poService) Create(ctx context.Context, session *app.UserSession, req CreateRequest) (*Response, error) {
	po := app.PurchaseOrder{
		ID:           primitive.NewObjectID(),
		PONumber:     fmt.Sprintf("PO-%d", time.Now().UnixNano()/100000),
		SupplierName: req.SupplierName,
		Status:       "Draft",
		Items:        req.Items,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	var total float64
	for i, item := range po.Items {
		total += item.Quantity * item.UnitCost
		po.Items[i].ReceivedQuantity = 0
	}
	po.TotalAmount = total

	_, err := s.poCol.InsertOne(ctx, po)
	if err != nil {
		return nil, err
	}

	app.LogAudit(ctx, session.Username, "Create", "PO", po.ID.Hex(), "", 0, "Created purchase order "+po.PONumber)
	return &po, nil
}

func (s *poService) Send(ctx context.Context, session *app.UserSession, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid PO ID format")
	}

	var po app.PurchaseOrder
	err = s.poCol.FindOne(ctx, bson.M{"_id": objID}).Decode(&po)
	if err != nil {
		return errors.New("PO not found")
	}

	if po.Status != "Draft" {
		return errors.New("can only send POs in Draft status")
	}

	update := bson.M{
		"$set": bson.M{
			"status":    "Sent",
			"updatedAt": time.Now(),
		},
	}

	_, err = s.poCol.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return err
	}

	app.LogAudit(ctx, session.Username, "Send", "PO", objID.Hex(), "", 0, "Sent purchase order "+po.PONumber+" to supplier")
	return nil
}

func (s *poService) Receive(ctx context.Context, session *app.UserSession, id string, req ReceiveRequest) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid PO ID format")
	}

	var po app.PurchaseOrder
	err = s.poCol.FindOne(ctx, bson.M{"_id": objID}).Decode(&po)
	if err != nil {
		return errors.New("PO not found")
	}

	if po.Status != "Sent" {
		return errors.New("can only receive goods for POs in Sent status")
	}

	allReceived := true
	for i, item := range po.Items {
		newReceive, ok := req.ReceivedQty[item.SKU]
		if ok && newReceive > 0 {
			po.Items[i].ReceivedQuantity += newReceive

			// Increment stock counts
			stockUpdate := bson.M{
				"$inc": bson.M{
					"stockOnHand":    newReceive,
					"availableStock": newReceive,
				},
			}
			_, _ = s.stockCol.UpdateOne(ctx, bson.M{"productId": item.ProductID}, stockUpdate)

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

	update := bson.M{
		"$set": bson.M{
			"items":     po.Items,
			"status":    newStatus,
			"updatedAt": time.Now(),
		},
	}

	_, err = s.poCol.UpdateOne(ctx, bson.M{"_id": objID}, update)
	return err
}
