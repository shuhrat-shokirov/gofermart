package postgresql

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func createTables(ctx context.Context, pool *pgxpool.Pool) error {
	userTable := `
	CREATE TABLE IF NOT EXISTS users (
	    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	    login VARCHAR(255) NOT NULL,
	    password VARCHAR(255) NOT NULL,
	    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
	    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
	    CONSTRAINT users_login_key UNIQUE (login)
);`

	orderTable := `
	CREATE TABLE IF NOT EXISTS orders (
	    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	    login VARCHAR(255) NOT NULL,
	    order_id VARCHAR(255) NOT NULL,
	    status VARCHAR(255) NOT NULL default 'NEW',
	    amount bigint NOT NULL default 0,
	    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
	    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
	    CONSTRAINT orders_id_key UNIQUE (order_id)
);`

	balanceTable := `
	CREATE TABLE IF NOT EXISTS balance (
	    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	    login VARCHAR(255) NOT NULL,
	    amount bigint NOT NULL default 0 CHECK (amount >= 0),
	    withdraw bigint NOT NULL default 0,
	    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
	    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
	    CONSTRAINT balance_login_key UNIQUE (login)
);`

	withdrawTable := `
	CREATE TABLE IF NOT EXISTS withdraw (
	    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	    login VARCHAR(255) NOT NULL,
	    amount bigint NOT NULL default 0 CHECK (amount > 0),
	    order_id VARCHAR(255) NOT NULL,
	    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
	    CONSTRAINT withdraw_order_id_key UNIQUE (order_id)
);`

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("err starting transaction: %w", err)
	}

	if _, err := tx.Exec(ctx, userTable); err != nil {
		_ = tx.Rollback(ctx)
		return fmt.Errorf("err creating users table: %w", err)
	}

	if _, err := tx.Exec(ctx, orderTable); err != nil {
		_ = tx.Rollback(ctx)
		return fmt.Errorf("err creating orders table: %w", err)
	}

	if _, err := tx.Exec(ctx, balanceTable); err != nil {
		_ = tx.Rollback(ctx)
		return fmt.Errorf("err creating balance table: %w", err)
	}

	if _, err := tx.Exec(ctx, withdrawTable); err != nil {
		_ = tx.Rollback(ctx)
		return fmt.Errorf("err creating withdraw table: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("err committing transaction: %w", err)
	}

	return nil
}
