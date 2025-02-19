package model

type UserBalance struct {
	Amount   int
	Withdraw int
}

type UserBalanceResponse struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}
