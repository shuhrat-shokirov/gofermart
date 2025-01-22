package postgresql

import (
	"context"
	"fmt"
)

func (p *Postgresql) CreateUser(ctx context.Context, login, password string) error {
	query := `INSERT INTO users (login, password) VALUES ($1, $2);`

	return retry(func() error {
		_, err := p.pool.Exec(ctx, query, login, password)
		if err != nil {
			return fmt.Errorf("can't exec: %w", err)
		}

		return nil
	})
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
