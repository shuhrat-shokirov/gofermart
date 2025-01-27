package store

import (
	"testing"

	"github.com/stretchr/testify/require"

	"gofermart/internal/gophermart/infra/store/memory"
	"gofermart/internal/gophermart/infra/store/postgresql"
)

func TestNewStore(t *testing.T) {

	t.Run("empty dns", func(t *testing.T) {
		store, err := NewStore(Config{
			Postgresql: &postgresql.Config{},
		})
		require.Error(t, err)
		require.Nil(t, store)
	})

	t.Run("successful memory store", func(t *testing.T) {
		store, err := NewStore(Config{
			Memory: &memory.Config{},
		})
		require.NoError(t, err)
		require.NotNil(t, store)
	})

	t.Run("failed store config", func(t *testing.T) {
		store, err := NewStore(Config{})
		require.Error(t, err)
		require.Nil(t, store)
	})
}
