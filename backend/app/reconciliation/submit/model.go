package submit

type Request struct {
	Items []struct {
		SKU       string  `json:"sku" binding:"required"`
		ActualQty float64 `json:"actualQty" binding:"required,min=0"`
	} `json:"items" binding:"required"`
	Note string `json:"note"`
}
