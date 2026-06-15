package stock

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"stock-management/backend/app"
)

type Service interface {
	List(ctx context.Context) ([]Response, error)
	Adjust(ctx context.Context, session *app.UserSession, req AdjustRequest) error
	LowAlerts(ctx context.Context) ([]Response, error)
	Scan(ctx context.Context, session *app.UserSession, req ScanRequest) (*ScanResponse, error)
}

type stockService struct {
	stockCol *mongo.Collection
	prodCol  *mongo.Collection
}

func NewService(stockCol, prodCol *mongo.Collection) Service {
	return &stockService{stockCol: stockCol, prodCol: prodCol}
}

func (s *stockService) List(ctx context.Context) ([]Response, error) {
	cursor, err := s.stockCol.Find(ctx, bson.M{})
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

func (s *stockService) Adjust(ctx context.Context, session *app.UserSession, req AdjustRequest) error {
	objID, err := primitive.ObjectIDFromHex(req.StockID)
	if err != nil {
		return errors.New("invalid stock ID format")
	}

	var original app.StockLevel
	err = s.stockCol.FindOne(ctx, bson.M{"_id": objID}).Decode(&original)
	if err != nil {
		return errors.New("stock record not found")
	}

	qtyChange := req.StockOnHand - original.StockOnHand
	availableStock := req.StockOnHand - original.CommittedStock

	update := bson.M{
		"$set": bson.M{
			"stockOnHand":    req.StockOnHand,
			"availableStock": availableStock,
			"reorderPoint":   req.ReorderPoint,
			"location":       req.Location,
		},
	}

	_, err = s.stockCol.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return err
	}

	app.LogAudit(ctx, session.Username, "Adjust", "Stock", objID.Hex(), original.SKU, qtyChange, req.Note)
	return nil
}

func (s *stockService) LowAlerts(ctx context.Context) ([]Response, error) {
	filter := bson.M{
		"$expr": bson.M{
			"$lte": bson.A{"$availableStock", "$reorderPoint"},
		},
	}

	cursor, err := s.stockCol.Find(ctx, filter)
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

func (s *stockService) Scan(ctx context.Context, session *app.UserSession, req ScanRequest) (*ScanResponse, error) {
	// Look up product by barcode or SKU
	var product app.Product
	err := s.prodCol.FindOne(ctx, bson.M{"$or": []bson.M{
		{"barcode": req.Barcode},
		{"sku": req.Barcode},
	}}).Decode(&product)
	if err != nil {
		return nil, errors.New("product not found for barcode: " + req.Barcode)
	}

	var stock app.StockLevel
	if err = s.stockCol.FindOne(ctx, bson.M{"productId": product.ID}).Decode(&stock); err != nil {
		return nil, errors.New("stock record not found for product")
	}

	var qtyDelta float64
	if req.Status == "in" {
		qtyDelta = req.Quantity
	} else {
		if stock.StockOnHand < req.Quantity {
			return nil, errors.New("insufficient stock: only " + fmt.Sprintf("%.0f", stock.StockOnHand) + " available")
		}
		qtyDelta = -req.Quantity
	}

	newQty := stock.StockOnHand + qtyDelta
	_, err = s.stockCol.UpdateOne(ctx,
		bson.M{"productId": product.ID},
		bson.M{"$set": bson.M{
			"stockOnHand":    newQty,
			"availableStock": newQty - stock.CommittedStock,
		}},
	)
	if err != nil {
		return nil, err
	}

	note := req.Note
	if note == "" {
		note = "Barcode scan " + req.Status
	}
	app.LogAudit(ctx, session.Username, "Scan-"+req.Status, "Stock", stock.ID.Hex(), product.SKU, qtyDelta, note)

	stock.StockOnHand = newQty
	stock.AvailableStock = newQty - stock.CommittedStock

	return &ScanResponse{
		Product:         product,
		Stock:           stock,
		QuantityChanged: qtyDelta,
	}, nil
}
