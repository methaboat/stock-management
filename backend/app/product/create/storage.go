package create

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
	prodCol  *mongo.Collection
	stockCol *mongo.Collection
}

func NewStorage(prodCol, stockCol *mongo.Collection) *Storage {
	return &Storage{prodCol: prodCol, stockCol: stockCol}
}

func (s *Storage) Execute(ctx context.Context, session *app.UserSession, req Request) (*Response, error) {
	var existing app.Product
	if err := s.prodCol.FindOne(ctx, bson.M{"sku": req.SKU}).Decode(&existing); err == nil {
		return nil, errors.New("SKU already exists in catalog")
	}

	p := app.Product{
		ID:           primitive.NewObjectID(),
		SKU:          req.SKU,
		Name:         req.Name,
		Barcode:      req.Barcode,
		Category:     req.Category,
		Brand:        req.Brand,
		CostPrice:    req.CostPrice,
		SellingPrice: req.SellingPrice,
		SupplierIDs:  req.SupplierIDs,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if _, err := s.prodCol.InsertOne(ctx, p); err != nil {
		return nil, err
	}

	stock := app.StockLevel{
		ID:             primitive.NewObjectID(),
		ProductID:      p.ID,
		SKU:            p.SKU,
		StockOnHand:    0,
		CommittedStock: 0,
		AvailableStock: 0,
		ReorderPoint:   10,
		Location:       app.Location{Warehouse: "Main Warehouse", Aisle: "A1", Bin: "01"},
	}
	_, _ = s.stockCol.InsertOne(ctx, stock)

	app.LogAudit(ctx, session.Username, "Create", "Product", p.ID.Hex(), p.SKU, 0, "Created product profile and initialized stock")
	return &p, nil
}
