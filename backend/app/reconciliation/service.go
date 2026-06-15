package reconciliation

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"stock-management/backend/app"
)

type Service interface {
	List(ctx context.Context) ([]Reconciliation, error)
	Submit(ctx context.Context, session *app.UserSession, req SubmitRequest) (*Reconciliation, error)
	Approve(ctx context.Context, session *app.UserSession, id string, req ApproveRequest) error
}

type reconciliationService struct {
	reconCol *mongo.Collection
	stockCol *mongo.Collection
	prodCol  *mongo.Collection
}

func NewService(reconCol, stockCol, prodCol *mongo.Collection) Service {
	return &reconciliationService{reconCol: reconCol, stockCol: stockCol, prodCol: prodCol}
}

func (s *reconciliationService) List(ctx context.Context) ([]Reconciliation, error) {
	cursor, err := s.reconCol.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var list []Reconciliation
	if err = cursor.All(ctx, &list); err != nil {
		return nil, err
	}
	if list == nil {
		list = []Reconciliation{}
	}
	return list, nil
}

func (s *reconciliationService) Submit(ctx context.Context, session *app.UserSession, req SubmitRequest) (*Reconciliation, error) {
	// Build a map of SKU -> actual qty from the request
	actualMap := make(map[string]float64)
	for _, item := range req.Items {
		actualMap[item.SKU] = item.ActualQty
	}

	// Fetch all stock levels to get expected quantities
	cursor, err := s.stockCol.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var stocks []app.StockLevel
	if err = cursor.All(ctx, &stocks); err != nil {
		return nil, err
	}

	// Fetch product names
	prodCursor, err := s.prodCol.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer prodCursor.Close(ctx)

	var products []app.Product
	prodCursor.All(ctx, &products)

	nameMap := make(map[string]string)
	for _, p := range products {
		nameMap[p.SKU] = p.Name
	}

	// Build reconciliation items
	var items []ReconciliationItem
	for _, stock := range stocks {
		actual, ok := actualMap[stock.SKU]
		if !ok {
			actual = 0
		}
		items = append(items, ReconciliationItem{
			ProductID:   stock.ProductID,
			SKU:         stock.SKU,
			ProductName: nameMap[stock.SKU],
			ExpectedQty: stock.StockOnHand,
			ActualQty:   actual,
			Difference:  actual - stock.StockOnHand,
		})
	}

	recon := &Reconciliation{
		ID:          primitive.NewObjectID(),
		Date:        time.Now(),
		Status:      "Draft",
		Items:       items,
		SubmittedBy: session.Username,
		Note:        req.Note,
		CreatedAt:   time.Now(),
	}

	_, err = s.reconCol.InsertOne(ctx, recon)
	if err != nil {
		return nil, err
	}

	app.LogAudit(ctx, session.Username, "Reconcile", "Stock", recon.ID.Hex(), "", 0, "Submitted end-of-day reconciliation")
	return recon, nil
}

func (s *reconciliationService) Approve(ctx context.Context, session *app.UserSession, id string, req ApproveRequest) error {
	if session.Role != "Admin" && session.Role != "Manager" {
		return errors.New("only Admin or Manager can approve reconciliations")
	}

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid reconciliation ID")
	}

	var recon Reconciliation
	if err = s.reconCol.FindOne(ctx, bson.M{"_id": objID}).Decode(&recon); err != nil {
		return errors.New("reconciliation not found")
	}
	if recon.Status == "Approved" {
		return errors.New("already approved")
	}

	now := time.Now()

	// Apply actual quantities to stock
	for _, item := range recon.Items {
		if item.Difference != 0 {
			_, _ = s.stockCol.UpdateOne(ctx,
				bson.M{"productId": item.ProductID},
				bson.M{"$set": bson.M{
					"stockOnHand":    item.ActualQty,
					"availableStock": item.ActualQty,
				}},
			)
		}
	}

	note := req.Note
	if note == "" {
		note = "Approved reconciliation"
	}

	_, err = s.reconCol.UpdateOne(ctx,
		bson.M{"_id": objID},
		bson.M{"$set": bson.M{
			"status":     "Approved",
			"approvedBy": session.Username,
			"approvedAt": now,
			"note":       note,
		}},
	)
	if err != nil {
		return err
	}

	app.LogAudit(ctx, session.Username, "Approve", "Reconciliation", objID.Hex(), "", 0, note)
	return nil
}
