package receive

type Request struct {
	ReceivedQty map[string]float64 `json:"receivedQty" binding:"required"`
}
