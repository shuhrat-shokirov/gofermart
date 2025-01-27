package model

type UserBalance struct {
	Amount   int
	Withdraw int
}

type UserBalanceResponse struct {
	Current  float64 `json:"current"`
	Withdraw float64 `json:"withdraw"`
}
