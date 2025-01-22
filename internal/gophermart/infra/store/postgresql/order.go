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
			if err := rows.Scan(&order.Number, &order.Status, &order.Accrual, &order.UploadedAt); err != nil {
				return fmt.Errorf("can't scan: %w", err)
			}

			result = append(result, order)
		}

		return nil
	})
}
