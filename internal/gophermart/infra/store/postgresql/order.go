//nolint:gocritic,goconst,nolintlint
package postgresql

import (
	"context"
	"fmt"
	"time"

	"gofermart/internal/gophermart/core/model"
)

func (p *Postgresql) SaveOrder(ctx context.Context, login string, request model.OrderRequest) error {
	query := `INSERT INTO orders (login, order_id, status, created_at) VALUES ($1, $2, $3, $4);`

	return retry(func() error {
		_, err := p.pool.Exec(ctx, query, login, request.ID, request.Status, time.Now())
		if err != nil {
			return fmt.Errorf("can't exec: %w", err)
		}

		return nil
	})
}

func (p *Postgresql) SetBalance(ctx context.Context, orderID, status string, amount int) error {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("can't begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
			return
		}
		_ = tx.Commit(ctx)
	}()

	queryOrder := `update orders set status = $1, amount = $2, updated_at = now() where order_id = $3 returning login;`

	var userLogin string
	err = retry(func() error {
		return tx.QueryRow(ctx, queryOrder, status, amount, orderID).Scan(&userLogin)
	})
	if err != nil {
		return fmt.Errorf("can't query: %w", err)
	}

	queryBalance := `update balance set amount = amount + $1, updated_at = now() where login = $2;`
	err = retry(func() error {
		_, err := tx.Exec(ctx, queryBalance, amount, userLogin)
		if err != nil {
			return fmt.Errorf("can't exec: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("can't exec: %w", err)
	}

	return nil
}

func (p *Postgresql) GetOrderLogin(ctx context.Context, orderID string) (string, error) {
	query := `SELECT login FROM orders WHERE order_id = $1;`

	var login string
	row := p.pool.QueryRow(ctx, query, orderID)

	if err := retry(func() error {
		return row.Scan(&login)
	}); err != nil {
		return "", fmt.Errorf("can't scan: %w", err)
	}

	return login, nil
}

func (p *Postgresql) GetUserOrders(ctx context.Context, login string) ([]model.Order, error) {
	query := `SELECT order_id, status, amount, created_at FROM orders WHERE login = $1 order by created_at desc;`

	result := make([]model.Order, 0)
	return result, retry(func() error {
		rows, err := p.pool.Query(ctx, query, login)
		if err != nil {
			return fmt.Errorf("can't query: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var order model.Order
			if err := rows.Scan(&order.OrderID, &order.Status, &order.Amount, &order.CreatedAt); err != nil {
				return fmt.Errorf("can't scan: %w", err)
			}

			result = append(result, order)
		}

		return nil
	})
}

func (p *Postgresql) GetPendingOrders(ctx context.Context) ([]model.Order, error) {
	query := `SELECT order_id, status, amount 
	FROM orders 
		WHERE status = any ($1)
		ORDER BY created_at limit 10;`

	result := make([]model.Order, 0)
	return result, retry(func() error {
		rows, err := p.pool.Query(ctx, query, []string{model.OrderStatusInProgress, model.OrderStatusNew})
		if err != nil {
			return fmt.Errorf("can't query: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var order model.Order
			if err := rows.Scan(&order.OrderID, &order.Status, &order.Amount); err != nil {
				return fmt.Errorf("can't scan: %w", err)
			}

			result = append(result, order)
		}

		return nil
	})
}
