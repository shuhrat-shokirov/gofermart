//nolint:errcheck,gocritic,nolintlint,forcetypeassert
package postgresql

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPostgresql_GetUserBalance(t *testing.T) {
	const (
		login    = "login"
		amount   = 100
		withdraw = 50
	)

	t.Run("successful get user balance", func(t *testing.T) {
		mockPool := new(MockPool)
		mockRow := new(MockRow)
		mockPool.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).Return(mockRow)
		mockRow.On("Scan", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			*(args.Get(0).(*int)) = amount
			*(args.Get(1).(*int)) = withdraw
		}).Return(nil)

		postgres := &Postgresql{pool: mockPool}
		_, err := postgres.GetUserBalance(context.TODO(), login)

		assert.NoError(t, err)
		mockPool.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})

	t.Run("failed get user balance", func(t *testing.T) {
		mockPool := new(MockPool)
		mockRow := new(MockRow)
		mockPool.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).Return(mockRow)
		mockRow.On("Scan", mock.Anything, mock.Anything).Return(errors.New("scan error"))

		postgres := &Postgresql{pool: mockPool}
		_, err := postgres.GetUserBalance(context.TODO(), login)

		assert.Error(t, err)
		assert.EqualError(t, err, "can't scan: operation failed after 3 retries: scan error")
		mockPool.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})
}
