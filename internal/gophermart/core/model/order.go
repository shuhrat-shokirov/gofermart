package model

import "time"

type OrderRequest struct {
	ID     string
	Login  string
	Status string
}

type Order struct {
	UploadedAt time.Time `json:"uploaded_at"`
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    int       `json:"accrual,omitempty"`
}
