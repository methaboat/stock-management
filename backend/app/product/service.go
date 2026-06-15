package product

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
	List(ctx context.Context) ([]Response, error)
	Create(ctx context.Context, session *app.UserSession, req CreateRequest) (*Response, error)
	Update(ctx context.Context, session *app.UserSession, id string, req UpdateRequest) error
	Delete(ctx context.Context, session *app.UserSession, id string) error
}

type productService struct {
	prodCol  *mongo.Collection
	stockCol *mongo.Collection
}

func NewService(prodCol, stockCol *mongo.Collection) Service {
	return &productService{prodCol: prodCol, stockCol: stockCol}
}

func (s *productService) List(ctx context.Context) ([]Response, error) {
	cursor, err := s.prodCol.Find(ctx, bson.M{})
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

func (s *productService) Create(ctx context.Context, session *app.UserSession, req CreateRequest) (*Response, error) {
	// SKU unique check
	var existing app.Product
	err := s.prodCol.FindOne(ctx, bson.M{"sku": req.SKU}).Decode(&existing)
	if err == nil {
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

	_, err = s.prodCol.InsertOne(ctx, p)
	if err != nil {
		return nil, err
	}

	// Initialize associated stock level
	stock := app.StockLevel{
		ID:             primitive.NewObjectID(),
		ProductID:      p.ID,
		SKU:            p.SKU,
		StockOnHand:    0,
		CommittedStock: 0,
		AvailableStock: 0,
		ReorderPoint:   10,
		Location: app.Location{
			Warehouse: "Main Warehouse",
			Aisle:     "A1",
			Bin:       "01",
		},
	}
	_, _ = s.stockCol.InsertOne(ctx, stock)

	app.LogAudit(ctx, session.Username, "Create", "Product", p.ID.Hex(), p.SKU, 0, "Created product profile and initialized stock")

	return &p, nil
}

func (s *productService) Update(ctx context.Context, session *app.UserSession, id string, req UpdateRequest) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid product ID format")
	}

	update := bson.M{
		"$set": bson.M{
			"name":         req.Name,
			"barcode":      req.Barcode,
			"category":     req.Category,
			"brand":        req.Brand,
			"costPrice":    req.CostPrice,
			"sellingPrice": req.SellingPrice,
			"supplierIds":  req.SupplierIDs,
			"updatedAt":    time.Now(),
		},
	}

	var updated app.Product
	err = s.prodCol.FindOneAndUpdate(ctx, bson.M{"_id": objID}, update).Decode(&updated)
	if err != nil {
		return errors.New("product not found")
	}

	app.LogAudit(ctx, session.Username, "Update", "Product", objID.Hex(), updated.SKU, 0, "Updated product fields")
	return nil
}

func (s *productService) Delete(ctx context.Context, session *app.UserSession, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid product ID format")
	}

	var p app.Product
	err = s.prodCol.FindOneAndDelete(ctx, bson.M{"_id": objID}).Decode(&p)
	if err != nil {
		return errors.New("product not found")
	}

	// Delete associated stock level
	_, _ = s.stockCol.DeleteOne(ctx, bson.M{"productId": objID})

	app.LogAudit(ctx, session.Username, "Delete", "Product", objID.Hex(), p.SKU, 0, "Deleted product profile and associated stock level")
	return nil
}
