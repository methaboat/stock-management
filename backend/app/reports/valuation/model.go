package valuation

type Entry struct {
	SKU           string  `json:"sku"`
	ProductName   string  `json:"productName"`
	StockOnHand   float64 `json:"stockOnHand"`
	BaseCost      float64 `json:"baseCost"`
	FIFOValuation float64 `json:"fifoValuation"`
	LIFOValuation float64 `json:"lifoValuation"`
	AvgCostVal    float64 `json:"avgCostValuation"`
}

type Report struct {
	Items            []Entry `json:"items"`
	TotalFIFO        float64 `json:"totalFifo"`
	TotalLIFO        float64 `json:"totalLifo"`
	TotalAverageCost float64 `json:"totalAverageCost"`
}

type costLayer struct {
	Quantity float64
	Cost     float64
}
