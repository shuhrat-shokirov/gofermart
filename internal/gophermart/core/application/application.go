package application

import (
	"context"

	"gofermart/internal/gophermart/core/model"
)

type Repo interface {
	CreateUser(ctx context.Context, login, password string) error
	GetUserPassword(ctx context.Context, login string) (string, error)

	SaveOrder(ctx context.Context, login string, request model.OrderRequest) error
	GetOrderLogin(ctx context.Context, orderID string) (string, error)
	GetUserOrders(ctx context.Context, login string) ([]model.Order, error)
}

type Application struct {
	repo   Repo
	secret string
}

func NewApplication(secret string, repo Repo) *Application {
	return &Application{
		repo:   repo,
		secret: secret,
	}
}
