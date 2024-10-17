package order

import (
	"context"
	"errors"
	"github.com/romanp1989/gophermart/internal/domain"
	"go.uber.org/zap"
	"sort"
	"time"
)

type orderStorage interface {
	LoadOrder(ctx context.Context, orderNumber string) (*domain.Order, error)
	CreateOrder(ctx context.Context, order *domain.Order) (int64, error)
	LoadOrdersWithBalance(ctx context.Context, userID domain.UserID) ([]domain.OrderWithBalance, error)
	LoadOrdersToProcess(ctx context.Context) ([]domain.Order, error)
}

type Service struct {
	storage   orderStorage
	validator *Validator
	log       *zap.Logger
}

func NewService(orderStore orderStorage, validator *Validator, log *zap.Logger) *Service {
	return &Service{
		storage:   orderStore,
		validator: validator,
		log:       log,
	}
}

func (s *Service) CreateOrder(ctx context.Context, orderNumber string, userID domain.UserID) (*domain.Order, error) {
	err := s.validator.Validate(ctx, orderNumber, userID)
	if err != nil && !errors.Is(err, ErrNotFoundOrder) {
		return nil, err
	}
	if err == nil {
		return nil, ErrIssetOrder
	}

	order := &domain.Order{
		ID:        0,
		CreatedAt: time.Now(),
		Number:    orderNumber,
		Status:    domain.OrderStatusNew,
		UserID:    userID,
	}

	id, err := s.storage.CreateOrder(ctx, order)
	if err != nil {
		return nil, err
	}

	order.ID = id

	return order, err
}

func (s *Service) GetUserOrders(ctx context.Context, userID domain.UserID) ([]domain.OrderWithBalance, error) {
	orders, err := s.storage.LoadOrdersWithBalance(ctx, userID)
	if err != nil {
		return nil, err
	}

	sort.SliceStable(orders, func(i, j int) bool {
		return orders[i].CreatedAt.Before(orders[j].CreatedAt)
	})

	return orders, nil
}
