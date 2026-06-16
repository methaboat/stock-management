package submit

import (
	"context"
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
	prodCol  *mongo.Collection
}

func NewStorage(reconCol, stockCol, prodCol *mongo.Collection) *Storage {
	return &Storage{reconCol: reconCol, stockCol: stockCol, prodCol: prodCol}
}

func (s *Storage) Execute(ctx context.Context, session *app.UserSession, req Request) (*reconlist.Reconciliation, error) {
	actualMap := make(map[string]float64)
	for _, item := range req.Items {
		actualMap[item.SKU] = item.ActualQty
	}

	cursor, err := s.stockCol.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var stocks []app.StockLevel
	if err = cursor.All(ctx, &stocks); err != nil {
		return nil, err
	}

	var products []app.Product
	if pCursor, err := s.prodCol.Find(ctx, bson.M{}); err == nil {
		pCursor.All(ctx, &products)
		pCursor.Close(ctx)
	}
	nameMap := make(map[string]string)
	for _, p := range products {
		nameMap[p.SKU] = p.Name
	}

	var items []reconlist.ReconciliationItem
	for _, stock := range stocks {
		actual := actualMap[stock.SKU]
		items = append(items, reconlist.ReconciliationItem{
			ProductID:   stock.ProductID,
			SKU:         stock.SKU,
			ProductName: nameMap[stock.SKU],
			ExpectedQty: stock.StockOnHand,
			ActualQty:   actual,
			Difference:  actual - stock.StockOnHand,
		})
	}

	recon := &reconlist.Reconciliation{
		ID:          primitive.NewObjectID(),
		Date:        time.Now(),
		Status:      "Draft",
		Items:       items,
		SubmittedBy: session.Username,
		Note:        req.Note,
		CreatedAt:   time.Now(),
	}

	if _, err = s.reconCol.InsertOne(ctx, recon); err != nil {
		return nil, err
	}

	app.LogAudit(ctx, session.Username, "Reconcile", "Stock", recon.ID.Hex(), "", 0, "Submitted end-of-day reconciliation")
	return recon, nil
}
