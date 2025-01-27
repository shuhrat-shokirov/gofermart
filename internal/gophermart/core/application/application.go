package application

import (
	"context"

	"go.uber.org/zap"

	"gofermart/internal/gophermart/core/model"
)

type Repo interface {
	CreateUser(ctx context.Context, login, password string) error
	GetUserPassword(ctx context.Context, login string) (string, error)

	SaveOrder(ctx context.Context, login string, request model.OrderRequest) error
	UpdateOrder(ctx context.Context, orderID string, status string) error
	SetBalance(ctx context.Context, orderID, status string, amount int) error

	GetOrderLogin(ctx context.Context, orderID string) (string, error)
	GetUserOrders(ctx context.Context, login string) ([]model.Order, error)
	GetPendingOrders(ctx context.Context) ([]model.Order, error)

	GetUserBalance(ctx context.Context, login string) (model.UserBalance, error)

	UserWithdraw(ctx context.Context, login string, request model.Withdraw) error
	GetUserWithdrawals(ctx context.Context, login string) ([]model.Withdraw, error)
}

type Client interface {
	SendOrder(ctx context.Context, orderID string) (model.ClientResponse, error)
}

type Application struct {
	repo   Repo
	client Client
	logger zap.SugaredLogger
	secret string
}

type Config struct {
	Repo   Repo
	Client Client
	Logger zap.SugaredLogger
	Secret string
}

func NewApplication(conf Config) *Application {
	return &Application{
		repo:   conf.Repo,
		secret: conf.Secret,
		client: conf.Client,
		logger: conf.Logger,
	}
}

func (a *Application) RunWorker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			a.handleOrders(ctx)
		}
	}
}
