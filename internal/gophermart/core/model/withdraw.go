package model

import "time"

type WithdrawRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

type Withdraw struct {
	CreatedAt time.Time
	OrderID   string
	Amount    int
}

type WithdrawResponse struct {
	ProcessedAt time.Time `json:"processed_at"`
	Order       string    `json:"order"`
	Sum         float64   `json:"sum"`
}
