package postgresql

import (
	"context"
	"fmt"

	"gofermart/internal/gophermart/core/model"
	"gofermart/internal/gophermart/core/repositories"
)

func (p *Postgresql) UserWithdraw(ctx context.Context, login string, request model.Withdraw) error {
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

	queryGetBalance := `select amount from balance where login = $1;`

	var amount int
	err = retry(func() error {
		return tx.QueryRow(ctx, queryGetBalance, login).Scan(&amount)
	})
	if err != nil {
		return fmt.Errorf("can't query: %w", err)
	}

	if amount < request.Amount {
		return fmt.Errorf("insufficient funds: %w", repositories.ErrInsufficientFunds)
	}

	queryBalance := `update balance set amount = amount - $1, withdraw = withdraw + $1 
               where login = $2;`

	err = retry(func() error {
		_, err := tx.Exec(ctx, queryBalance, request.Amount, login)
		if err != nil {
			return fmt.Errorf("can't exec: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("can't exec: %w", err)
	}

	queryWithdraw := `insert into withdraw (login, amount, order_id) values ($1, $2, $3);`

	err = retry(func() error {
		_, err := tx.Exec(ctx, queryWithdraw, login, request.Amount, request.OrderID)
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

func (p *Postgresql) GetUserWithdrawals(ctx context.Context, login string) ([]model.Withdraw, error) {
	query := `SELECT amount, order_id, created_at FROM withdraw WHERE login = $1;`

	rows, err := p.pool.Query(ctx, query, login)
	if err != nil {
		return nil, fmt.Errorf("can't query: %w", err)
	}
	defer rows.Close()

	var withdrawals []model.Withdraw
	for rows.Next() {
		var w model.Withdraw
		if err := rows.Scan(&w.Amount, &w.OrderID, &w.CreatedAt); err != nil {
			return nil, fmt.Errorf("can't scan: %w", err)
		}

		withdrawals = append(withdrawals, w)
	}

	return withdrawals, nil
}
