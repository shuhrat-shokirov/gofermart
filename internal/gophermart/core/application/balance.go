package application

import (
	"context"
	"fmt"

	"gofermart/internal/gophermart/core/model"
)

func (a *Application) UserBalance(ctx context.Context, login string) (model.UserBalanceResponse, error) {
	balance, err := a.repo.GetUserBalance(ctx, login)
	if err != nil {
		return model.UserBalanceResponse{}, fmt.Errorf("can't get user balance: %w", err)
	}

	return model.UserBalanceResponse{
		Current:   convertToPounds(balance.Amount),
		Withdrawn: convertToPounds(balance.Withdraw),
	}, nil
}
