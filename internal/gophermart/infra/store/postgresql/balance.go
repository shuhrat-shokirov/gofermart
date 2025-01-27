package postgresql

import (
	"context"
	"fmt"

	"gofermart/internal/gophermart/core/model"
)

//nolint:gocritic,goconst,nolintlint
func (p *Postgresql) GetUserBalance(ctx context.Context, login string) (model.UserBalance, error) {
	query := `SELECT amount, withdraw FROM balance WHERE login = $1;`

	var balance model.UserBalance
	row := p.pool.QueryRow(ctx, query, login)

	if err := retry(func() error {
		return row.Scan(&balance.Amount, &balance.Withdraw)
	}); err != nil {
		return model.UserBalance{}, fmt.Errorf("can't scan: %w", err)
	}

	return balance, nil
}
