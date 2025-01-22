package postgresql

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"gofermart/internal/gophermart/core/repositories"
)

type Config struct {
	Dsn string
}

type Postgresql struct {
	pool *pgxpool.Pool
}

func New(dsn string) (*Postgresql, error) {
	ctx := context.TODO()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("can't create pool: %w", err)
	}

	err = pool.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("can't ping: %w", err)
	}

	if err = createTables(ctx, pool); err != nil {
		return nil, fmt.Errorf("can't create tables: %w", err)
	}

	return &Postgresql{pool: pool}, nil
}

func (p *Postgresql) Ping(ctx context.Context) error {
	err := p.pool.Ping(ctx)
	if err != nil {
		return fmt.Errorf("can't ping: %w", err)
	}

	return nil
}

func (p *Postgresql) Close() error {
	p.pool.Close()
	return nil
}

func retry(operation func() error) error {
	const maxRetries = 3
	retryIntervals := []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}

	var lastErr error
	for i := 0; i <= maxRetries; i++ {
		err := operation()
		if err == nil {
			return nil
		}

		if isDuplicateError(err) {
			return repositories.ErrDuplicate
		}

		if errors.Is(err, pgx.ErrNoRows) {
			return repositories.ErrNotFound
		}

		if !isRetrievableError(err) {
			return fmt.Errorf("operation failed after %d retries: %w", maxRetries, err)
		}

		lastErr = err
		if i < maxRetries {
			time.Sleep(retryIntervals[i])
			continue
		}
	}

	return fmt.Errorf("operation failed after %d retries: %w", maxRetries, lastErr)
}

func isRetrievableError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgerrcode.ConnectionException, pgerrcode.ConnectionDoesNotExist, pgerrcode.ConnectionFailure:
			return true
		}
	}
	return false
}

func isDuplicateError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == pgerrcode.UniqueViolation
	}
	return false
}
