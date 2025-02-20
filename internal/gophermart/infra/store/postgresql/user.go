package postgresql

import (
	"context"
	"fmt"
)

func (p *Postgresql) CreateUser(ctx context.Context, login, password string) error {
	query := `INSERT INTO users (login, password) VALUES ($1, $2);`

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

	_, err = tx.Exec(ctx, query, login, password)
	if err != nil {
		return fmt.Errorf("can't create user: %w", err)
	}

	createBalanceQuery := `INSERT INTO balance (login) VALUES ($1);`
	_, err = tx.Exec(ctx, createBalanceQuery, login)
	if err != nil {
		return fmt.Errorf("can't add balance : %w", err)
	}

	return nil
}

func (p *Postgresql) GetUserPassword(ctx context.Context, login string) (string, error) {
	query := `SELECT password FROM users WHERE login = $1;`

	var password string
	row := p.pool.QueryRow(ctx, query, login)

	if err := retry(func() error {
		return row.Scan(&password)
	}); err != nil {
		return "", fmt.Errorf("can't scan: %w", err)
	}

	return password, nil
}
