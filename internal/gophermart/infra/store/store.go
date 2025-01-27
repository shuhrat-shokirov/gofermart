package store

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"gofermart/internal/gophermart/core/model"
	"gofermart/internal/gophermart/infra/store/memory"
	"gofermart/internal/gophermart/infra/store/postgresql"
)

type Config struct {
	Memory     *memory.Config
	Postgresql *postgresql.Config
	Logger     zap.SugaredLogger
}

type Store interface {
	Ping(ctx context.Context) error
	Close() error

	CreateUser(ctx context.Context, login, password string) error
	GetUserPassword(ctx context.Context, login string) (string, error)

	SaveOrder(ctx context.Context, login string, request model.OrderRequest) error
	SetBalance(ctx context.Context, orderID, status string, amount int) error

	GetOrderLogin(ctx context.Context, orderID string) (string, error)
	GetUserOrders(ctx context.Context, login string) ([]model.Order, error)
	GetPendingOrders(ctx context.Context) ([]model.Order, error)

	GetUserBalance(ctx context.Context, login string) (model.UserBalance, error)

	UserWithdraw(ctx context.Context, login string, request model.Withdraw) error
	GetUserWithdrawals(ctx context.Context, login string) ([]model.Withdraw, error)
}

func NewStore(conf Config) (Store, error) {
	switch {
	case conf.Postgresql != nil:
		store, err := postgresql.New(conf.Postgresql.Dsn, zap.SugaredLogger{})
		if err != nil {
			return nil, fmt.Errorf("can't create postgresql store: %w", err)
		}

		return store, nil
	case conf.Memory != nil:
		store, err := memory.New()
		if err != nil {
			return nil, fmt.Errorf("can't create memory store: %w", err)
		}

		return store, nil
	default:
		return nil, errors.New("store config is not provided")
	}
}
