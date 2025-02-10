package postgresql

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"gofermart/internal/gophermart/core/repositories"
)

func TestPostgresql_Ping(t *testing.T) {
	t.Run("successful ping", func(t *testing.T) {
		mockPool := new(MockPool)
		mockPool.On("Ping", mock.Anything).Return(nil)

		postgres := &Postgresql{pool: mockPool}
		err := postgres.Ping(context.TODO())

		assert.NoError(t, err)
		mockPool.AssertExpectations(t)
	})

	t.Run("failed ping", func(t *testing.T) {
		mockPool := new(MockPool)
		mockPool.On("Ping", mock.Anything).Return(errors.New("connection error"))

		postgres := &Postgresql{pool: mockPool}
		err := postgres.Ping(context.TODO())

		assert.Error(t, err)
		assert.EqualError(t, err, "can't ping: connection error")
		mockPool.AssertExpectations(t)
	})
}

func TestPostgresql_Close(t *testing.T) {
	t.Run("successful close", func(t *testing.T) {
		mockPool := new(MockPool)
		mockPool.On("Close")

		postgres := &Postgresql{pool: mockPool}
		err := postgres.Close()

		assert.NoError(t, err)
		mockPool.AssertExpectations(t)
	})
}

func Test_retry(t *testing.T) {
	t.Run("successful operation", func(t *testing.T) {
		err := retry(func() error {
			return nil
		})

		assert.NoError(t, err)
	})

	t.Run("failed operation", func(t *testing.T) {
		err := retry(func() error {
			return errors.New("error")
		})

		assert.Error(t, err)
		assert.EqualError(t, err, "operation failed after 3 retries: error")
	})

	t.Run("failed operation with duplicate error", func(t *testing.T) {
		err := retry(func() error {
			pgError := pgconn.PgError{}
			pgError.Code = pgerrcode.UniqueViolation
			return &pgError
		})

		assert.Error(t, err)
		require.ErrorIs(t, err, repositories.ErrDuplicate)
	})

	t.Run("failed operation with no rows error", func(t *testing.T) {
		err := retry(func() error {
			return pgx.ErrNoRows
		})

		assert.Error(t, err)
		require.ErrorIs(t, err, repositories.ErrNotFound)
	})

	t.Run("failed operation with non-retriable error", func(t *testing.T) {
		err := retry(func() error {
			return errors.New("non-retriable error")
		})

		assert.Error(t, err)
		assert.EqualError(t, err, "operation failed after 3 retries: non-retriable error")
	})

	t.Run("failed operation with retriable error", func(t *testing.T) {
		err := retry(func() error {
			pgError := pgconn.PgError{}
			pgError.Code = pgerrcode.ConnectionException
			return &pgError
		})

		assert.Error(t, err)
		assert.EqualError(t, err,
			fmt.Sprintf("operation failed after 3 retries: :  (SQLSTATE %s)", pgerrcode.ConnectionException))
	})
}
