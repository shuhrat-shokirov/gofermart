package application

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"gofermart/internal/gophermart/core/client"
	"gofermart/internal/gophermart/core/model"
)

func (a *Application) handleOrders(ctx context.Context) {
	orders, err := a.repo.GetPendingOrders(ctx)
	if err != nil {
		a.logger.Errorf("can't get orders: %v", err)
		return
	}

	if len(orders) == 0 {
		return
	}

	jobs := make(chan model.Order, len(orders))
	results := make(chan error, len(orders))

	var (
		wg         sync.WaitGroup
		maxWorkers = 5
	)

	for range maxWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for order := range jobs {
				if err := a.processOrder(ctx, order); err != nil {
					results <- err
					if errors.Is(err, client.ErrTooManyRequests) {
						break
					}
				}
			}
		}()
	}

	for _, order := range orders {
		jobs <- order
	}
	close(jobs)

	wg.Wait()
	close(results)

	for err := range results {
		if err != nil && !errors.Is(err, client.ErrTooManyRequests) {
			a.logger.Errorf("can't process order: %v", err)
		}
	}
}

func (a *Application) processOrder(ctx context.Context, order model.Order) error {
	select {
	case <-ctx.Done():
		return nil
	default:
	}

	resp, err := a.client.SendOrder(ctx, order.OrderID)
	if err != nil {
		if errors.Is(err, client.ErrTooManyRequests) {
			return fmt.Errorf("can't send order %s: %w", order.OrderID, err)
		}
		if errors.Is(err, ErrNotFound) {
			return nil
		}
		return fmt.Errorf("can't send order %s: %w", order.OrderID, err)
	}

	var amount int
	if resp.Accrual != nil {
		amount = convertToPence(*resp.Accrual)
	}

	if err := a.repo.SetBalance(ctx, order.OrderID, resp.Status, amount); err != nil {
		return fmt.Errorf("can't save balance %s: %w", order.OrderID, err)
	}

	return nil
}
