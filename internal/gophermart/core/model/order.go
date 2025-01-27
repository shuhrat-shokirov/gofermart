package model

import "time"

type OrderRequest struct {
	ID     string
	Status string
}

type OrderResponse struct {
	UploadedAt time.Time `json:"uploaded_at"`
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    float64   `json:"accrual,omitempty"`
}

type Order struct {
	CreatedAt time.Time
	OrderID   string
	Status    string
	Amount    int
}

const (
	OrderStatusNew        = "NEW"
	OrderStatusInProgress = "PROCESSING"
	OrderStatusDone       = "PROCESSED"
	OrderStatusFailed     = "INVALID"
)
