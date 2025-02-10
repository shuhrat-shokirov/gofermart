package memory

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gofermart/internal/gophermart/core/model"
	"gofermart/internal/gophermart/core/repositories"
)

func TestNew(t *testing.T) {
	memory, err := New()
	require.NoError(t, err)

	require.NotNil(t, memory.mu)
	require.NotNil(t, memory.orderMu)
	require.NotNil(t, memory.userBMu)
	require.NotNil(t, memory.withdrawMu)
	require.NotNil(t, memory.users)
	require.NotNil(t, memory.orders)
	require.NotNil(t, memory.userBalance)
	require.NotNil(t, memory.withdraws)

	require.Empty(t, memory.users)
	require.Empty(t, memory.orders)
	require.Empty(t, memory.userBalance)
	require.Empty(t, memory.withdraws)
}

func TestMemory_PingAndClose(t *testing.T) {
	memory, err := New()
	require.NoError(t, err)

	ctx := context.Background()

	err = memory.Ping(ctx)
	require.NoError(t, err)

	err = memory.Close()
	require.NoError(t, err)
}

func TestMemory_OtherFunc(t *testing.T) {
	ctx := context.Background()

	uuidNew, err := uuid.NewV7()
	require.NoError(t, err)

	login := uuidNew.String()

	otherLogin := uuid.NewString()

	passUUID, err := uuid.NewV7()
	require.NoError(t, err)

	pass := passUUID.String()

	t.Run("CreateUser", func(t *testing.T) {
		memory, err := New()
		require.NoError(t, err)

		err = memory.CreateUser(ctx, login, pass)
		require.NoError(t, err)

		err = memory.CreateUser(ctx, login, pass)
		require.Error(t, err)
		assert.Error(t, repositories.ErrDuplicate, err.Error())

		assert.NotEmpty(t, memory.users[login])
		assert.NotNil(t, memory.userBalance[login])
	})

	t.Run("GetUserPassword", func(t *testing.T) {
		memory, err := New()
		require.NoError(t, err)

		err = memory.CreateUser(ctx, login, pass)
		require.NoError(t, err)

		password, err := memory.GetUserPassword(ctx, login)
		require.NoError(t, err)
		require.Equal(t, pass, password)

		_, err = memory.GetUserPassword(ctx, otherLogin)
		require.Error(t, err)
		assert.Error(t, repositories.ErrNotFound, err.Error())
	})

	t.Run("SaveOrder", func(t *testing.T) {
		memory, err := New()
		require.NoError(t, err)

		order := model.OrderRequest{
			ID:     uuid.NewString(),
			Status: model.OrderStatusNew,
		}

		err = memory.SaveOrder(ctx, login, order)
		require.NoError(t, err)

		err = memory.SaveOrder(ctx, login, order)
		require.Error(t, err)
		assert.Error(t, repositories.ErrDuplicate, err.Error())

		assert.Equal(t, memory.orders[order.ID].Login, login)
	})

	t.Run("GetOrderLogin", func(t *testing.T) {
		memory, err := New()
		require.NoError(t, err)

		order := model.OrderRequest{
			ID:     uuid.NewString(),
			Status: model.OrderStatusNew,
		}

		err = memory.SaveOrder(ctx, login, order)
		require.NoError(t, err)

		loginOrder, err := memory.GetOrderLogin(ctx, order.ID)
		require.NoError(t, err)
		require.Equal(t, login, loginOrder)

		_, err = memory.GetOrderLogin(ctx, uuid.NewString())
		require.Error(t, err)
		assert.Error(t, repositories.ErrNotFound, err.Error())
	})

	t.Run("GetUserOrders", func(t *testing.T) {
		memory, err := New()
		require.NoError(t, err)

		order := model.OrderRequest{
			ID:     uuid.NewString(),
			Status: model.OrderStatusNew,
		}

		err = memory.SaveOrder(ctx, login, order)
		require.NoError(t, err)

		orders, err := memory.GetUserOrders(ctx, login)
		require.NoError(t, err)
		require.Len(t, orders, 1)

		orders, err = memory.GetUserOrders(ctx, otherLogin)
		require.NoError(t, err)
		require.Empty(t, orders)
	})

	t.Run("GetPendingOrders", func(t *testing.T) {
		memory, err := New()
		require.NoError(t, err)

		orders, err := memory.GetPendingOrders(ctx)
		require.NoError(t, err)
		require.Empty(t, orders)

		err = memory.SaveOrder(ctx, login, model.OrderRequest{
			ID:     uuid.NewString(),
			Status: model.OrderStatusNew,
		})
		require.NoError(t, err)

		orders, err = memory.GetPendingOrders(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, orders)
		assert.Len(t, orders, 1)

		err = memory.SaveOrder(ctx, login, model.OrderRequest{
			ID:     uuid.NewString(),
			Status: model.OrderStatusDone,
		})
		require.NoError(t, err)
		assert.Len(t, orders, 1)
	})

	t.Run("SetBalance", func(t *testing.T) {
		memory, err := New()
		require.NoError(t, err)

		orderID := uuid.NewString()

		amount := 100

		err = memory.SetBalance(ctx, orderID, model.OrderStatusDone, amount)
		require.Error(t, err)
		assert.Error(t, repositories.ErrNotFound, err.Error())

		err = memory.CreateUser(ctx, login, pass)
		require.NoError(t, err)

		err = memory.SaveOrder(ctx, login, model.OrderRequest{
			ID:     orderID,
			Status: model.OrderStatusNew,
		})
		require.NoError(t, err)

		err = memory.SetBalance(ctx, orderID, model.OrderStatusDone, amount)
		require.NoError(t, err)

		balance, err := memory.GetUserBalance(ctx, login)
		require.NoError(t, err)
		require.Equal(t, amount, balance.Amount)

		newOrderID := uuid.NewString()

		err = memory.SaveOrder(ctx, login, model.OrderRequest{
			ID:     newOrderID,
			Status: model.OrderStatusNew,
		})
		require.NoError(t, err)

		err = memory.SetBalance(ctx, newOrderID, model.OrderStatusDone, amount)
		require.NoError(t, err)

		balance, err = memory.GetUserBalance(ctx, login)
		require.NoError(t, err)
		require.Equal(t, amount+amount, balance.Amount)
	})

	t.Run("GetUserBalance", func(t *testing.T) {
		memory, err := New()
		require.NoError(t, err)

		_, err = memory.GetUserBalance(ctx, login)
		require.Error(t, err)
		assert.Error(t, repositories.ErrNotFound, err.Error())

		err = memory.CreateUser(ctx, login, pass)
		require.NoError(t, err)

		balance, err := memory.GetUserBalance(ctx, login)
		require.NoError(t, err)
		require.Equal(t, 0, balance.Amount)
		require.Equal(t, 0, balance.Withdraw)

		amount := 100

		err = memory.SetBalance(ctx, uuid.NewString(), model.OrderStatusDone, amount)
		require.Error(t, err)
		assert.Error(t, repositories.ErrNotFound, err.Error())

		orderID := uuid.NewString()

		err = memory.SaveOrder(ctx, login, model.OrderRequest{
			ID:     orderID,
			Status: model.OrderStatusNew,
		})
		require.NoError(t, err)

		err = memory.SetBalance(ctx, orderID, model.OrderStatusDone, amount)
		require.NoError(t, err)

		balance, err = memory.GetUserBalance(ctx, login)
		require.NoError(t, err)
		require.Equal(t, amount, balance.Amount)
		require.Equal(t, 0, balance.Withdraw)
	})

	t.Run("UserWithdraw", func(t *testing.T) {
		memory, err := New()
		require.NoError(t, err)

		amount := 100
		orderID := uuid.NewString()

		err = memory.UserWithdraw(ctx, login, model.Withdraw{
			Amount:  amount,
			OrderID: orderID,
		})
		require.Error(t, err)
		assert.Error(t, repositories.ErrNotFound, err.Error())

		err = memory.CreateUser(ctx, login, pass)
		require.NoError(t, err)

		err = memory.UserWithdraw(ctx, login, model.Withdraw{
			Amount:  amount,
			OrderID: orderID,
		})
		require.Error(t, err)
		assert.Error(t, repositories.ErrInsufficientFunds, err.Error())

		err = memory.SaveOrder(ctx, login, model.OrderRequest{
			ID:     orderID,
			Status: model.OrderStatusNew,
		})
		require.NoError(t, err)

		err = memory.SetBalance(ctx, orderID, model.OrderStatusDone, amount)
		require.NoError(t, err)

		err = memory.UserWithdraw(ctx, login, model.Withdraw{
			Amount:  amount,
			OrderID: orderID,
		})
		require.NoError(t, err)

		balance, err := memory.GetUserBalance(ctx, login)
		require.NoError(t, err)
		require.Equal(t, 0, balance.Amount)
		require.Equal(t, amount, balance.Withdraw)

		err = memory.UserWithdraw(ctx, login, model.Withdraw{
			Amount:  amount,
			OrderID: orderID,
		})
		require.Error(t, err)
		assert.Error(t, repositories.ErrDuplicate, err.Error())
	})

	t.Run("GetUserWithdrawals", func(t *testing.T) {
		memory, err := New()
		require.NoError(t, err)

		list, err := memory.GetUserWithdrawals(ctx, login)
		require.NoError(t, err)
		assert.Nil(t, list)

		err = memory.CreateUser(ctx, login, pass)
		require.NoError(t, err)

		_, err = memory.GetUserWithdrawals(ctx, login)
		require.NoError(t, err)

		amount := 100
		orderID := uuid.NewString()

		err = memory.SaveOrder(ctx, login, model.OrderRequest{
			ID:     orderID,
			Status: model.OrderStatusNew,
		})
		require.NoError(t, err)

		err = memory.SetBalance(ctx, orderID, model.OrderStatusDone, amount)
		require.NoError(t, err)

		err = memory.UserWithdraw(ctx, login, model.Withdraw{
			Amount:  amount,
			OrderID: orderID,
		})
		require.NoError(t, err)

		withdrawals, err := memory.GetUserWithdrawals(ctx, login)
		require.NoError(t, err)
		require.Len(t, withdrawals, 1)
	})
}
