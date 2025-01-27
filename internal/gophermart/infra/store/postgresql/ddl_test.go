package postgresql

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_createTables(t *testing.T) {
	t.Run("successful create tables", func(t *testing.T) {
		mockPool := new(MockPool)
		mockTx := new(MockTx)

		mockPool.On("Begin", mock.Anything).Return(mockTx, nil)
		mockTx.On("Exec", mock.Anything, mock.Anything, mock.Anything).Return(pgconn.CommandTag{}, nil)
		mockTx.On("Commit", mock.Anything).Return(nil)

		err := createTables(context.TODO(), mockPool)
		require.NoError(t, err)

		mockPool.AssertExpectations(t)
		mockTx.AssertExpectations(t)
	})

	t.Run("failed create tables", func(t *testing.T) {
		mockPool := new(MockPool)
		mockTx := new(MockTx)

		mockPool.On("Begin", mock.Anything).Return(mockTx, nil)
		mockTx.On("Exec", mock.Anything, mock.Anything, mock.Anything).Return(pgconn.CommandTag{}, nil)
		mockTx.On("Commit", mock.Anything).Return(errors.New("commit error"))

		err := createTables(context.TODO(), mockPool)

		require.Error(t, err)
		require.EqualError(t, err, "err committing transaction: commit error")

		mockPool.AssertExpectations(t)
		mockTx.AssertExpectations(t)
	})
}
