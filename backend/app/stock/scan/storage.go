package scan

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"stock-management/backend/app"
)

type Storage struct {
	stockCol *mongo.Collection
	prodCol  *mongo.Collection
}

func NewStorage(stockCol, prodCol *mongo.Collection) *Storage {
	return &Storage{stockCol: stockCol, prodCol: prodCol}
}

func (s *Storage) Execute(ctx context.Context, session *app.UserSession, req Request) (*Response, error) {
	var product app.Product
	if err := s.prodCol.FindOne(ctx, bson.M{"$or": []bson.M{
		{"barcode": req.Barcode},
		{"sku": req.Barcode},
	}}).Decode(&product); err != nil {
		return nil, errors.New("product not found for barcode: " + req.Barcode)
	}

	var stock app.StockLevel
	if err := s.stockCol.FindOne(ctx, bson.M{"productId": product.ID}).Decode(&stock); err != nil {
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
	_, err := s.stockCol.UpdateOne(ctx,
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

	return &Response{Product: product, Stock: stock, QuantityChanged: qtyDelta}, nil
}
