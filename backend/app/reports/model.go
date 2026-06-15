package reports

type ValuationEntry struct {
	SKU            string  `json:"sku"`
	ProductName    string  `json:"productName"`
	StockOnHand    float64 `json:"stockOnHand"`
	BaseCost       float64 `json:"baseCost"`
	FIFOValuation  float64 `json:"fifoValuation"`
	LIFOValuation  float64 `json:"lifoValuation"`
	AvgCostVal     float64 `json:"avgCostValuation"`
}

type ValuationReport struct {
	Items            []ValuationEntry `json:"items"`
	TotalFIFO        float64          `json:"totalFifo"`
	TotalLIFO        float64          `json:"totalLifo"`
	TotalAverageCost float64          `json:"totalAverageCost"`
}

type CostLayer struct {
	Quantity float64
	Cost     float64
}
