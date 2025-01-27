//nolint:wrapcheck,gocritic,nolintlint,errcheck,forcetypeassert
package postgresql

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"gofermart/internal/gophermart/core/model"
	"gofermart/internal/gophermart/core/repositories"
)

func TestPostgresql_SaveOrder(t *testing.T) {
	login := uuid.NewString()

	t.Run("successful save order", func(t *testing.T) {
		mockPool := new(MockPool)

		mockPool.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(pgconn.NewCommandTag("INSERT 1"), nil)

		postgres := &Postgresql{pool: mockPool}

		err := postgres.SaveOrder(context.TODO(), login, model.OrderRequest{
			ID:     uuid.NewString(),
			Status: model.OrderStatusNew,
		})

		assert.NoError(t, err)

		mockPool.AssertExpectations(t)
	})

	t.Run("failed save order", func(t *testing.T) {
		mockPool := new(MockPool)

		mockPool.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(pgconn.CommandTag{}, errors.New("exec error"))

		postgres := &Postgresql{pool: mockPool}

		err := postgres.SaveOrder(context.TODO(), login, model.OrderRequest{
			ID:     uuid.NewString(),
			Status: model.OrderStatusNew,
		})

		assert.Error(t, err)
		assert.EqualError(t, err, "operation failed after 3 retries: can't exec: exec error")

		mockPool.AssertExpectations(t)
	})

	t.Run("failed save order with duplicate error", func(t *testing.T) {
		mockPool := new(MockPool)

		mockPool.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(pgconn.CommandTag{}, repositories.ErrDuplicate)

		postgres := &Postgresql{pool: mockPool}

		err := postgres.SaveOrder(context.TODO(), login, model.OrderRequest{
			ID:     uuid.NewString(),
			Status: model.OrderStatusNew,
		})

		assert.Error(t, err)
		assert.ErrorIs(t, err, repositories.ErrDuplicate)

		mockPool.AssertExpectations(t)
	})
}

func TestPostgresql_SetBalance(t *testing.T) {
	const (
		amount = 100
		login  = "user_login"
	)
	t.Run("successful set balance", func(t *testing.T) {
		mockPool := new(MockPool)
		mockTx := new(MockTx)
		mockRow := new(MockRow)

		mockPool.On("Begin", mock.Anything).Return(mockTx, nil)
		mockTx.On("QueryRow", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockRow)
		mockRow.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
			*(args.Get(0).(*string)) = login
		}).Return(nil)
		mockTx.On("Exec", mock.Anything, mock.Anything, mock.Anything).
			Return(pgconn.NewCommandTag("UPDATE 1"), nil)
		mockTx.On("Commit", mock.Anything).Return(nil)

		postgres := &Postgresql{pool: mockPool}

		err := postgres.SetBalance(context.TODO(), uuid.NewString(), model.OrderStatusNew, amount)

		assert.NoError(t, err)
		mockPool.AssertExpectations(t)
		mockTx.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})

	t.Run("failed QueryRow", func(t *testing.T) {
		mockPool := new(MockPool)
		mockTx := new(MockTx)
		mockRow := new(MockRow)

		mockPool.On("Begin", mock.Anything).Return(mockTx, nil)
		mockTx.On("QueryRow", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockRow)
		mockRow.On("Scan", mock.Anything).Return(errors.New("query row error"))
		mockTx.On("Rollback", mock.Anything).Return(nil)

		postgres := &Postgresql{pool: mockPool}

		err := postgres.SetBalance(context.TODO(), uuid.NewString(), model.OrderStatusNew, amount)

		assert.Error(t, err)
		assert.EqualError(t, err, "can't query: query row error")
		mockPool.AssertExpectations(t)
		mockTx.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})

	t.Run("failed Exec", func(t *testing.T) {
		mockPool := new(MockPool)
		mockTx := new(MockTx)
		mockRow := new(MockRow)

		mockPool.On("Begin", mock.Anything).Return(mockTx, nil)
		mockTx.On("QueryRow", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockRow)
		mockRow.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
			*(args.Get(0).(*string)) = login
		}).Return(nil)
		mockTx.On("Exec", mock.Anything, mock.Anything, mock.Anything).
			Return(pgconn.CommandTag{}, errors.New("exec error"))
		mockTx.On("Rollback", mock.Anything).Return(nil)

		postgres := &Postgresql{pool: mockPool}

		err := postgres.SetBalance(context.TODO(), uuid.NewString(), model.OrderStatusNew, amount)

		assert.Error(t, err)
		assert.EqualError(t, err, "can't exec: exec error")
		mockPool.AssertExpectations(t)
		mockTx.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})
}

func TestPostgresql_GetOrderLogin(t *testing.T) {
	t.Run("successful get order login", func(t *testing.T) {
		mockPool := new(MockPool)
		mockRow := new(MockRow)
		mockPool.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).Return(mockRow)
		mockRow.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
			*(args.Get(0).(*string)) = "login"
		}).Return(nil)

		postgres := &Postgresql{pool: mockPool}
		_, err := postgres.GetOrderLogin(context.TODO(), "order_id")

		assert.NoError(t, err)
		mockPool.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})

	t.Run("failed get order login", func(t *testing.T) {
		mockPool := new(MockPool)
		mockRow := new(MockRow)
		mockPool.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).Return(mockRow)
		mockRow.On("Scan", mock.Anything).Return(errors.New("scan error"))

		postgres := &Postgresql{pool: mockPool}
		_, err := postgres.GetOrderLogin(context.TODO(), "order_id")

		assert.Error(t, err)
		assert.EqualError(t, err, "can't scan: operation failed after 3 retries: scan error")
		mockPool.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})

	t.Run("not found order login", func(t *testing.T) {
		mockPool := new(MockPool)
		mockRow := new(MockRow)

		mockPool.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).Return(mockRow)
		mockRow.On("Scan", mock.Anything).Return(pgx.ErrNoRows)

		postgres := &Postgresql{pool: mockPool}

		_, err := postgres.GetOrderLogin(context.TODO(), "order_id")

		assert.Error(t, err)
		assert.ErrorIs(t, err, repositories.ErrNotFound)
		mockPool.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})
}
