package model

type ClientResponse struct {
	Accrual *float64 `json:"accrual"`
	OrderID string   `json:"order"`
	Status  string   `json:"status"`
}
