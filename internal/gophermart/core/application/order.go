package application

import (
	"context"
	"errors"
	"fmt"

	"gofermart/internal/gophermart/core/model"
	"gofermart/internal/gophermart/core/repositories"
)

const (
	orderStatusNew        = "NEW"
	orderStatusInProgress = "PROCESSING"
	orderStatusDone       = "PROCESSED"
	orderStatusFailed     = "INVALID"
)

func (a *Application) UserOrder(ctx context.Context, userLogin, orderID string) error {
	if !isValidOrderID(orderID) {
		return fmt.Errorf("invalid order id: %w", ErrInvalidOrderID)
	}

	if err := a.repo.SaveOrder(ctx, userLogin, model.OrderRequest{
		ID:     orderID,
		Login:  userLogin,
		Status: orderStatusNew,
	}); err != nil {
		if errors.Is(err, repositories.ErrDuplicate) {
			login, err := a.repo.GetOrderLogin(ctx, orderID)
			if err != nil {
				return fmt.Errorf("can't get order login: %w", err)
			}

			if login != userLogin {
				return fmt.Errorf("order already exists for another user: %w", ErrOrderExistsOnAnotherUser)
			}

			return fmt.Errorf("order already exists: %w", ErrOrderAlreadyExists)
		}

		return fmt.Errorf("can't save order: %w", err)
	}

	return nil
}

//nolint:mnd,gocritic,nolintlint
func isValidOrderID(orderID string) bool {
	var sum int
	alt := false
	for i := len(orderID) - 1; i >= 0; i-- {
		n := int(orderID[i] - '0')
		if alt {
			n *= 2
			if n > 9 {
				n -= 9
			}
		}
		sum += n
		alt = !alt
	}
	return sum%10 == 0
}

func (a *Application) UserOrders(ctx context.Context, userLogin string) ([]model.Order, error) {
	orders, err := a.repo.GetUserOrders(ctx, userLogin)
	if err != nil {
		return nil, fmt.Errorf("can't get user orders: %w", err)
	}

	return orders, nil
}
