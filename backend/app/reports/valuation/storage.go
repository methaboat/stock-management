package valuation

import (
	"context"
	"sort"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"stock-management/backend/app"
)

type Storage struct {
	prodCol  *mongo.Collection
	stockCol *mongo.Collection
	poCol    *mongo.Collection
}

func NewStorage(prodCol, stockCol, poCol *mongo.Collection) *Storage {
	return &Storage{prodCol: prodCol, stockCol: stockCol, poCol: poCol}
}

func (s *Storage) Execute(ctx context.Context) (*Report, error) {
	pCursor, err := s.prodCol.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer pCursor.Close(ctx)
	var products []app.Product
	if err = pCursor.All(ctx, &products); err != nil {
		return nil, err
	}

	sCursor, err := s.stockCol.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer sCursor.Close(ctx)
	var stockList []app.StockLevel
	if err = sCursor.All(ctx, &stockList); err != nil {
		return nil, err
	}

	stockMap := make(map[primitive.ObjectID]float64)
	for _, st := range stockList {
		stockMap[st.ProductID] = st.StockOnHand
	}

	var purchaseOrders []app.PurchaseOrder
	if poCursor, err := s.poCol.Find(ctx, bson.M{"status": "Received"}); err == nil {
		poCursor.All(ctx, &purchaseOrders)
		poCursor.Close(ctx)
	}
	sort.Slice(purchaseOrders, func(i, j int) bool {
		return purchaseOrders[i].CreatedAt.Before(purchaseOrders[j].CreatedAt)
	})

	report := &Report{Items: []Entry{}}
	for _, p := range products {
		qtyOnHand := stockMap[p.ID]

		var layers []costLayer
		for _, po := range purchaseOrders {
			for _, item := range po.Items {
				if item.ProductID == p.ID && item.ReceivedQuantity > 0 {
					layers = append(layers, costLayer{Quantity: item.ReceivedQuantity, Cost: item.UnitCost})
				}
			}
		}

		fifoVal, lifoVal, avgVal := calculateValuation(p.CostPrice, qtyOnHand, layers)
		report.Items = append(report.Items, Entry{
			SKU:           p.SKU,
			ProductName:   p.Name,
			StockOnHand:   qtyOnHand,
			BaseCost:      p.CostPrice,
			FIFOValuation: fifoVal,
			LIFOValuation: lifoVal,
			AvgCostVal:    avgVal,
		})
		report.TotalFIFO += fifoVal
		report.TotalLIFO += lifoVal
		report.TotalAverageCost += avgVal
	}
	return report, nil
}

func calculateValuation(baseCost, qtyOnHand float64, layers []costLayer) (fifo, lifo, avg float64) {
	if qtyOnHand <= 0 {
		return 0, 0, 0
	}

	var totalQty, totalCost float64
	for _, l := range layers {
		totalQty += l.Quantity
		totalCost += l.Quantity * l.Cost
	}

	remainingFIFO := qtyOnHand
	for i := len(layers) - 1; i >= 0 && remainingFIFO > 0; i-- {
		take := layers[i].Quantity
		if take > remainingFIFO {
			take = remainingFIFO
		}
		fifo += take * layers[i].Cost
		remainingFIFO -= take
	}
	if remainingFIFO > 0 {
		fifo += remainingFIFO * baseCost
	}

	remainingLIFO := qtyOnHand
	for i := 0; i < len(layers) && remainingLIFO > 0; i++ {
		take := layers[i].Quantity
		if take > remainingLIFO {
			take = remainingLIFO
		}
		lifo += take * layers[i].Cost
		remainingLIFO -= take
	}
	if remainingLIFO > 0 {
		lifo += remainingLIFO * baseCost
	}

	if totalQty > 0 {
		avg = qtyOnHand * (totalCost / totalQty)
	} else {
		avg = qtyOnHand * baseCost
	}
	return fifo, lifo, avg
}
