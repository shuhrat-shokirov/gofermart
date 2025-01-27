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
	mu          *sync.Mutex
	orderMu     *sync.Mutex
	userBMu     *sync.Mutex
	withdrawMu  *sync.Mutex
	users       map[string]string
	orders      map[string]Order
	userBalance map[string]UserBalance
	withdraws   map[string]Withdraw
}

type Order struct {
	CreatedAt time.Time
	Login     string
	Status    string
	Amount    int
}

type UserBalance struct {
	Amount   int
	Withdraw int
}

type Withdraw struct {
	CreatedAt time.Time
	OrderID   string
	Amount    int
}

func New() (*Memory, error) {
	return &Memory{
		mu:          &sync.Mutex{},
		orderMu:     &sync.Mutex{},
		userBMu:     &sync.Mutex{},
		users:       make(map[string]string),
		orders:      make(map[string]Order),
		userBalance: make(map[string]UserBalance),
		withdraws:   make(map[string]Withdraw),
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
				OrderID:   id,
				Status:    order.Status,
				Amount:    order.Amount,
				CreatedAt: order.CreatedAt,
			})
		}
	}

	sort.Slice(orders, func(i, j int) bool {
		return orders[i].CreatedAt.Before(orders[j].CreatedAt)
	})

	return orders, nil
}

func (s *Memory) GetPendingOrders(ctx context.Context) ([]model.Order, error) {
	s.orderMu.Lock()
	defer s.orderMu.Unlock()

	var orders []model.Order
	for id, order := range s.orders {
		if order.Status == model.OrderStatusInProgress || order.Status == model.OrderStatusNew {
			orders = append(orders, model.Order{
				OrderID: id,
				Status:  order.Status,
				Amount:  order.Amount,
			})
		}
	}

	return orders, nil
}

func (s *Memory) SetBalance(ctx context.Context, orderID, status string, amount int) error {
	s.userBMu.Lock()
	defer s.userBMu.Unlock()

	s.orderMu.Lock()
	defer s.orderMu.Unlock()

	order, ok := s.orders[orderID]
	if !ok {
		return repositories.ErrNotFound
	}

	s.userBalance[order.Login] = UserBalance{
		Amount:   s.userBalance[order.Login].Amount + amount,
		Withdraw: s.userBalance[order.Login].Withdraw,
	}

	order.Amount = amount
	order.Status = status
	s.orders[orderID] = order

	return nil
}

func (s *Memory) GetUserBalance(ctx context.Context, login string) (model.UserBalance, error) {
	s.userBMu.Lock()
	defer s.userBMu.Unlock()

	balance, ok := s.userBalance[login]
	if !ok {
		return model.UserBalance{}, repositories.ErrNotFound
	}

	return model.UserBalance{
		Amount:   balance.Amount,
		Withdraw: balance.Withdraw,
	}, nil
}

func (s *Memory) UserWithdraw(ctx context.Context, login string, request model.Withdraw) error {
	s.userBMu.Lock()
	defer s.userBMu.Unlock()

	s.withdrawMu.Lock()
	defer s.withdrawMu.Unlock()

	balance, ok := s.userBalance[login]
	if !ok {
		return repositories.ErrNotFound
	}

	if balance.Amount < request.Amount {
		return repositories.ErrInsufficientFunds
	}

	balance.Amount -= request.Amount
	balance.Withdraw += request.Amount
	s.userBalance[login] = balance

	if _, ok := s.withdraws[request.OrderID]; ok {
		return repositories.ErrDuplicate
	}

	s.withdraws[request.OrderID] = Withdraw{
		OrderID:   request.OrderID,
		Amount:    request.Amount,
		CreatedAt: time.Now(),
	}

	return nil
}

func (s *Memory) GetUserWithdrawals(ctx context.Context, login string) ([]model.Withdraw, error) {
	s.withdrawMu.Lock()
	defer s.withdrawMu.Unlock()

	var withdrawals []model.Withdraw
	for _, withdraw := range s.withdraws {
		if withdraw.OrderID == login {
			withdrawals = append(withdrawals, model.Withdraw{
				OrderID:   withdraw.OrderID,
				Amount:    withdraw.Amount,
				CreatedAt: withdraw.CreatedAt,
			})
		}
	}

	sort.Slice(withdrawals, func(i, j int) bool {
		return withdrawals[i].CreatedAt.Before(withdrawals[j].CreatedAt)
	})

	return withdrawals, nil
}
