package application

import (
	"context"
	"errors"

	"gofermart/internal/gophermart/core/client"
	"gofermart/internal/gophermart/core/model"
)

func (a *Application) handleOrders(ctx context.Context) {
	a.logger.Info("worker started")

	orders, err := a.repo.GetPendingOrders(ctx)
	if err != nil {
		a.logger.Errorf("can't get orders: %v", err)
		return
	}

	if len(orders) == 0 {
		return
	}

	for _, order := range orders {
		resp, err := a.client.SendOrder(ctx, order.OrderID)
		if err != nil {
			if errors.Is(err, client.ErrTooManyRequests) {
				return
			}

			if errors.Is(err, client.ErrNotFound) {
				continue
			}

			a.logger.Errorf("can't send order %s: %v", order.OrderID, err)
			continue
		}

		if resp.Status != model.OrderStatusDone {
			if err := a.repo.UpdateOrder(ctx, order.OrderID, resp.Status); err != nil {
				a.logger.Errorf("can't save order status %s: %v", order.OrderID, err)
			}

			continue
		}

		var amount int
		if resp.Accrual != nil {
			amount = convertToPence(*resp.Accrual)
		}

		if err := a.repo.SetBalance(ctx, order.OrderID, resp.Status, amount); err != nil {
			a.logger.Errorf("can't save balance %s: %v", order.OrderID, err)
		}
	}
}
