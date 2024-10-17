package balance

import (
	"context"
	"errors"
	"github.com/romanp1989/gophermart/internal/domain"
	"github.com/romanp1989/gophermart/internal/order"
	"go.uber.org/zap"
)

type balanceStorage interface {
	GetUserWithdrawals(ctx context.Context, userID domain.UserID) ([]*domain.Balance, error)
	GetUserBalance(ctx context.Context, userID domain.UserID) ([]*domain.BalanceSum, error)
	Withdraw(ctx context.Context, userID *domain.UserID, number string, sum float64) (*domain.Balance, error)
}

type Service struct {
	storage   balanceStorage
	validator *order.Validator
	log       *zap.Logger
}

func NewService(balanceStore balanceStorage, validator *order.Validator, log *zap.Logger) *Service {
	return &Service{
		storage:   balanceStore,
		validator: validator,
		log:       log,
	}
}

func (s *Service) LoadSum(ctx context.Context, userID domain.UserID) (*domain.UserSum, error) {
	sums, err := s.storage.GetUserBalance(ctx, userID)
	if err != nil {
		return nil, err
	}

	sumRes := &domain.UserSum{}
	for _, sum := range sums {
		switch sum.Type {
		case domain.BalanceTypeAdded:
			sumRes.Current = sum.Sum
		case domain.BalanceTypeWithdrawn:
			sumRes.Withdrawn = sum.Sum
		}
	}

	if sumRes.Current > 0 {
		sumRes.Current = sumRes.Current - sumRes.Withdrawn
	}

	return sumRes, nil
}

func (s *Service) Withdraw(ctx context.Context, userID *domain.UserID, orderNumber string, sum float64) (*domain.Balance, error) {
	err := s.validator.Validate(ctx, orderNumber, *userID)
	if err != nil && !errors.Is(err, order.ErrNotFoundOrder) {
		return nil, err
	}

	balance, err := s.storage.Withdraw(ctx, userID, orderNumber, sum)
	if err != nil {
		return nil, err
	}

	return balance, nil
}

func (s *Service) LoadWithdrawals(ctx context.Context, userID domain.UserID) ([]*domain.Balance, error) {
	withdrawals, err := s.storage.GetUserWithdrawals(ctx, userID)
	if err != nil {
		return nil, err
	}
	return withdrawals, nil
}
