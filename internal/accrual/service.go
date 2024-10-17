package accrual

import (
	"encoding/json"
	"github.com/romanp1989/gophermart/internal/config"
	"github.com/romanp1989/gophermart/internal/domain"
	"go.uber.org/zap"
	"io"
	"net/http"
	"sync"
	"time"
)

type accrualStorage interface {
	GetNewOrdersToSend() ([]domain.Order, error)
	UpdateOrder(order *domain.AccrualResponse, userID domain.UserID) error
	Update(order domain.Order) error
	AddBalance(o *domain.Balance) error
}

type Service struct {
	storage   accrualStorage
	log       *zap.Logger
	waitGroup sync.WaitGroup
	closeChan chan struct{}
}

func NewService(accrualStore accrualStorage, log *zap.Logger) *Service {
	return &Service{
		storage: accrualStore,
		log:     log,
	}
}

// OrderStatusChecker Обновление бонусов по заказам через сервис Accrual
func (s *Service) OrderStatusChecker(duration time.Duration) {
	ticker := time.NewTicker(duration)
	go func() {
		for {
			select {
			case <-ticker.C:
				s.waitGroup.Add(1)

				err := s.doProcess()
				if err != nil {
					s.log.Error("ошибка обновления бонусов по заказу", zap.Error(err))
				}
				s.waitGroup.Done()
				ticker.Reset(2 * time.Second)
			case <-s.closeChan:
				ticker.Stop()
				return
			}
		}
	}()
}

// Сохранение баллов лояльности по заказам
func (s *Service) doProcess() error {
	newOrders, err := s.storage.GetNewOrdersToSend()
	if err != nil {
		return err
	}

	defer func() {
		if len(newOrders) > 0 {
			for _, o := range newOrders {
				o.Status = domain.OrderStatusNew
				if err = s.storage.Update(o); err != nil {
					s.log.With(zap.Error(err)).Error("ошибка обновления статуса заказа")
				}
			}
		}
	}()

	for i, order := range newOrders {
		accrualResp, err := s.getWithdrawalFromAccrual(order.Number)
		if len(accrualResp.Order) != 0 || err == nil {
			if err = s.storage.UpdateOrder(accrualResp, order.UserID); err == nil {
				balance := &domain.Balance{
					OrderNumber: order.Number,
					UserID:      order.UserID,
					Sum:         accrualResp.Accrual,
					Type:        domain.BalanceTypeAdded,
					CreatedAt:   time.Now(),
				}
				if err = s.storage.AddBalance(balance); err != nil {
					continue
				}
				newOrders = append(newOrders[:i], newOrders[i+1:]...)
			}
		}
	}

	return nil
}

// Получение информации о баллах лояльности из сервиса Accrual
func (s *Service) getWithdrawalFromAccrual(orderNumber string) (*domain.AccrualResponse, error) {
	var (
		accrualResp *domain.AccrualResponse
		resp        *http.Response
		err         error
	)

	host := config.Options.FlagAccrualAddress + "/api/orders/" + orderNumber
	resp, err = http.Get(host)
	if err != nil || resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusNoContent {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(body, &accrualResp); err != nil {
		return nil, err
	}
	return accrualResp, nil
}

func (s *Service) Close() {
	close(s.closeChan)
	s.waitGroup.Wait()
}
