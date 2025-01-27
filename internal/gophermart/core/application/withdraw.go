package application

import (
	"context"
	"fmt"

	"gofermart/internal/gophermart/core/model"
)

func (a *Application) UserWithdraw(ctx context.Context, login string, request model.WithdrawRequest) error {
	balance, err := a.repo.GetUserBalance(ctx, login)
	if err != nil {
		return fmt.Errorf("can't get user balance: %w", err)
	}

	if balance.Amount < convertToPence(request.Sum) {
		return ErrInsufficientFunds
	}

	if !isValidOrderID(request.Order) {
		return fmt.Errorf("invalid order id: %w", ErrInvalidOrderID)
	}

	if err := a.repo.UserWithdraw(ctx, login, model.Withdraw{
		Amount:  convertToPence(request.Sum),
		OrderID: request.Order,
	}); err != nil {
		return fmt.Errorf("can't withdraw: %w", err)
	}

	return nil
}

func (a *Application) UserWithdrawals(ctx context.Context, login string) ([]model.WithdrawResponse, error) {
	withdrawals, err := a.repo.GetUserWithdrawals(ctx, login)
	if err != nil {
		return nil, fmt.Errorf("can't get user withdrawals: %w", err)
	}

	list := make([]model.WithdrawResponse, 0, len(withdrawals))
	for _, w := range withdrawals {
		list = append(list, model.WithdrawResponse{
			Order:       w.OrderID,
			Sum:         convertToPounds(w.Amount),
			ProcessedAt: w.CreatedAt,
		})
	}

	if len(list) == 0 {
		return nil, ErrNotFound
	}

	return list, nil
}
