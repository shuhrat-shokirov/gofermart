package memory

import (
	"context"
	"sort"
	"sync"
	"time"

	"gofermart/internal/gophermart/core/model"
	"gofermart/internal/gophermart/core/repositories"
)

type Config struct {
}

type Memory struct {
	mu      *sync.Mutex
	orderMu *sync.Mutex
	users   map[string]string
	orders  map[string]Order
}

type Order struct {
	CreatedAt time.Time
	Login     string
	Status    string
	Amount    int
}

func New() (*Memory, error) {
	return &Memory{
		mu:      &sync.Mutex{},
		orderMu: &sync.Mutex{},
		users:   make(map[string]string),
		orders:  make(map[string]Order),
	}, nil
}

func (s *Memory) Ping(ctx context.Context) error {
	return nil
}

func (s *Memory) Close() error {
	return nil
}

func (s *Memory) CreateUser(ctx context.Context, login, password string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.users[login]; ok {
		return repositories.ErrDuplicate
	}

	s.users[login] = password

	return nil
}

func (s *Memory) GetUserPassword(ctx context.Context, login string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	password, ok := s.users[login]
	if !ok {
		return "", repositories.ErrNotFound
	}

	return password, nil
}

func (s *Memory) SaveOrder(ctx context.Context, login string, order model.OrderRequest) error {
	s.orderMu.Lock()
	defer s.orderMu.Unlock()

	if _, ok := s.orders[order.ID]; ok {
		return repositories.ErrDuplicate
	}

	s.orders[order.ID] = Order{
		Login:     login,
		Status:    order.Status,
		CreatedAt: time.Now(),
	}

	return nil
}

func (s *Memory) GetOrderLogin(ctx context.Context, orderID string) (string, error) {
	s.orderMu.Lock()
	defer s.orderMu.Unlock()

	order, ok := s.orders[orderID]
	if !ok {
		return "", repositories.ErrNotFound
	}

	return order.Login, nil
}

func (s *Memory) GetUserOrders(ctx context.Context, login string) ([]model.Order, error) {
	s.orderMu.Lock()
	defer s.orderMu.Unlock()

	var orders []model.Order
	for id, order := range s.orders {
		if order.Login == login {
			orders = append(orders, model.Order{
				Number:     id,
				Status:     order.Status,
				Accrual:    order.Amount,
				UploadedAt: order.CreatedAt,
			})
		}
	}

	sort.Slice(orders, func(i, j int) bool {
		return orders[i].UploadedAt.Before(orders[j].UploadedAt)
	})

	return orders, nil
}
