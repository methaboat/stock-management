package reports

import (
	"context"
	"sort"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"stock-management/backend/app"
)

type Service interface {
	GetValuation(ctx context.Context) (*ValuationReport, error)
}

type reportsService struct {
	prodCol  *mongo.Collection
	stockCol *mongo.Collection
	poCol    *mongo.Collection
}

func NewService(prodCol, stockCol, poCol *mongo.Collection) Service {
	return &reportsService{
		prodCol:  prodCol,
		stockCol: stockCol,
		poCol:    poCol,
	}
}

// CalculateValuation computes FIFO, LIFO, and Average Cost valuations for a single product
func CalculateValuation(baseCost float64, qtyOnHand float64, layers []CostLayer) (fifo, lifo, avg float64) {
	if qtyOnHand <= 0 {
		return 0, 0, 0
	}

	var totalQtyReceived float64
	var totalCostReceived float64
	for _, l := range layers {
		totalQtyReceived += l.Quantity
		totalCostReceived += l.Quantity * l.Cost
	}

	// 1. Calculate FIFO (newest shipments value the remaining stock)
	fifo = 0.0
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

	// 2. Calculate LIFO (oldest shipments value the remaining stock)
	lifo = 0.0
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

	// 3. Calculate Average Cost
	if totalQtyReceived > 0 {
		weightedAvg := totalCostReceived / totalQtyReceived
		avg = qtyOnHand * weightedAvg
	} else {
		avg = qtyOnHand * baseCost
	}

	return fifo, lifo, avg
}

func (s *reportsService) GetValuation(ctx context.Context) (*ValuationReport, error) {
	// 1. Fetch products
	pCursor, err := s.prodCol.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer pCursor.Close(ctx)

	var products []app.Product
	if err = pCursor.All(ctx, &products); err != nil {
		return nil, err
	}

	// 2. Fetch stock levels
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

	// 3. Fetch received POs
	poCursor, err := s.poCol.Find(ctx, bson.M{"status": "Received"})
	var purchaseOrders []app.PurchaseOrder = []app.PurchaseOrder{}
	if err == nil {
		poCursor.All(ctx, &purchaseOrders)
		poCursor.Close(ctx)
	}

	// Sort POs chronologically (oldest first)
	sort.Slice(purchaseOrders, func(i, j int) bool {
		return purchaseOrders[i].CreatedAt.Before(purchaseOrders[j].CreatedAt)
	})

	var report ValuationReport
	report.Items = []ValuationEntry{}

	// 4. Calculate
	for _, p := range products {
		qtyOnHand := stockMap[p.ID]
		
		var layers []CostLayer = []CostLayer{}
		for _, po := range purchaseOrders {
			for _, item := range po.Items {
				if item.ProductID == p.ID && item.ReceivedQuantity > 0 {
					layers = append(layers, CostLayer{
						Quantity: item.ReceivedQuantity,
						Cost:     item.UnitCost,
					})
				}
			}
		}

		fifoVal, lifoVal, avgCostVal := CalculateValuation(p.CostPrice, qtyOnHand, layers)

		entry := ValuationEntry{
			SKU:            p.SKU,
			ProductName:    p.Name,
			StockOnHand:    qtyOnHand,
			BaseCost:       p.CostPrice,
			FIFOValuation:  fifoVal,
			LIFOValuation:  lifoVal,
			AvgCostVal:     avgCostVal,
		}

		report.Items = append(report.Items, entry)
		report.TotalFIFO += fifoVal
		report.TotalLIFO += lifoVal
		report.TotalAverageCost += avgCostVal
	}

	return &report, nil
}
