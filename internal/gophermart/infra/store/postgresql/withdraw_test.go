//nolint:wrapcheck,gocritic,nolintlint,errcheck,forcetypeassert
package postgresql

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"gofermart/internal/gophermart/core/model"
)

func TestPostgresql_UserWithdraw(t *testing.T) {
	t.Run("successful withdrawal", func(t *testing.T) {
		mockPool := new(MockPool)
		mockTx := new(MockTx)
		mockRow := new(MockRow)

		mockPool.On("Begin", mock.Anything).Return(mockTx, nil)
		mockTx.On("QueryRow", mock.Anything, "select amount from balance where login = $1;",
			mock.Anything).Return(mockRow)
		mockRow.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
			*(args.Get(0).(*int)) = 100
		}).Return(nil)
		mockTx.On("Exec", mock.Anything, `update balance set amount = amount - $1, withdraw = withdraw + $1 
               where login = $2;`, mock.Anything, mock.Anything).Return(pgconn.NewCommandTag("UPDATE 1"), nil)

		mockTx.On("Exec", mock.Anything,
			"insert into withdraw (login, amount, order_id) values ($1, $2, $3);",
			mock.Anything, mock.Anything, mock.Anything).Return(pgconn.NewCommandTag("INSERT 1"), nil)

		mockTx.On("Commit", mock.Anything).Return(nil)

		postgres := &Postgresql{pool: mockPool}
		request := model.Withdraw{
			Amount:  50,
			OrderID: "order123",
		}

		err := postgres.UserWithdraw(context.TODO(), "testuser", request)

		assert.NoError(t, err)

		mockPool.AssertExpectations(t)
		mockTx.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})

	t.Run("insufficient funds", func(t *testing.T) {
		mockPool := new(MockPool)
		mockTx := new(MockTx)
		mockRow := new(MockRow)

		mockPool.On("Begin", mock.Anything).Return(mockTx, nil)

		mockTx.On("QueryRow", mock.Anything, "select amount from balance where login = $1;",
			mock.Anything).Return(mockRow)
		mockRow.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
			*(args.Get(0).(*int)) = 100
		}).Return(nil)

		mockTx.On("Commit", mock.Anything).Return(nil)

		postgres := &Postgresql{pool: mockPool}
		request := model.Withdraw{
			Amount:  150,
			OrderID: "order123",
		}

		err := postgres.UserWithdraw(context.TODO(), "testuser", request)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient funds")

		mockPool.AssertExpectations(t)
		mockTx.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})

	t.Run("failed query balance", func(t *testing.T) {
		mockPool := new(MockPool)
		mockTx := new(MockTx)
		mockRow := new(MockRow)

		mockPool.On("Begin", mock.Anything).Return(mockTx, nil)

		mockTx.On("QueryRow", mock.Anything, "select amount from balance where login = $1;",
			mock.Anything).Return(mockRow)
		mockRow.On("Scan", mock.Anything).Return(assert.AnError)

		mockTx.On("Rollback", mock.Anything).Return(nil)

		postgres := &Postgresql{pool: mockPool}
		request := model.Withdraw{
			Amount:  50,
			OrderID: "order123",
		}

		err := postgres.UserWithdraw(context.TODO(), "testuser", request)

		assert.Error(t, err)
		assert.EqualError(t, err, "can't query: assert.AnError general error for testing")

		mockPool.AssertExpectations(t)
		mockTx.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})

	t.Run("failed update balance", func(t *testing.T) {
		mockPool := new(MockPool)
		mockTx := new(MockTx)
		mockRow := new(MockRow)

		mockPool.On("Begin", mock.Anything).Return(mockTx, nil)

		mockTx.On("QueryRow", mock.Anything, "select amount from balance where login = $1;",
			mock.Anything).Return(mockRow)
		mockRow.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
			*(args.Get(0).(*int)) = 100
		}).Return(nil)

		mockTx.On("Exec", mock.Anything, `update balance set amount = amount - $1, withdraw = withdraw + $1 
               where login = $2;`, mock.Anything).
			Return(pgconn.NewCommandTag("UPDATE 0"), assert.AnError)

		mockTx.On("Rollback", mock.Anything).Return(nil)

		postgres := &Postgresql{pool: mockPool}
		request := model.Withdraw{
			Amount:  50,
			OrderID: "order123",
		}

		err := postgres.UserWithdraw(context.TODO(), "testuser", request)

		assert.Error(t, err)
		assert.EqualError(t, err, "can't exec: assert.AnError general error for testing")

		mockPool.AssertExpectations(t)
		mockTx.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})

	t.Run("failed insert withdraw", func(t *testing.T) {
		mockPool := new(MockPool)
		mockTx := new(MockTx)
		mockRow := new(MockRow)

		mockPool.On("Begin", mock.Anything).Return(mockTx, nil)

		mockTx.On("QueryRow", mock.Anything, "select amount from balance where login = $1;",
			mock.Anything).Return(mockRow)
		mockRow.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
			*(args.Get(0).(*int)) = 100
		}).Return(nil)

		mockTx.On("Exec", mock.Anything, `update balance set amount = amount - $1, withdraw = withdraw + $1 
               where login = $2;`, mock.Anything).Return(pgconn.NewCommandTag("UPDATE 1"), nil)

		mockTx.On("Exec", mock.Anything,
			"insert into withdraw (login, amount, order_id) values ($1, $2, $3);", mock.Anything).
			Return(pgconn.NewCommandTag("UPDATE 0"), assert.AnError)

		mockTx.On("Rollback", mock.Anything).Return(nil)

		postgres := &Postgresql{pool: mockPool}
		request := model.Withdraw{
			Amount:  50,
			OrderID: "order123",
		}

		err := postgres.UserWithdraw(context.TODO(), "testuser", request)

		assert.Error(t, err)
		assert.EqualError(t, err, "can't exec: assert.AnError general error for testing")

		mockPool.AssertExpectations(t)
		mockTx.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})
}
