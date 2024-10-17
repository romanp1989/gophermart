package order

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/romanp1989/gophermart/internal/config"
	"github.com/romanp1989/gophermart/internal/cookies"
	"github.com/romanp1989/gophermart/internal/order"
	"go.uber.org/zap"
	"io"
	"net/http"
	"time"
)

type Handler struct {
	service *order.Service
	logger  *zap.Logger
}

type orderListResponse struct {
	Number     string  `json:"number"`
	Status     string  `json:"status"`
	Accrual    float64 `json:"accrual,omitempty"`
	UploadedAt string  `json:"uploaded_at"`
}

func NewOrderHandler(orderService *order.Service, logger *zap.Logger) *Handler {
	return &Handler{
		service: orderService,
		logger:  logger,
	}
}

func (h *Handler) CreateOrderHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), config.Options.FlagTimeoutContext)
	defer cancel()

	userID, ok := cookies.UIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	numberOrder, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, err = h.service.CreateOrder(ctx, string(numberOrder), *userID)

	if err != nil {
		if errors.Is(err, order.ErrIssetOrder) {
			w.WriteHeader(http.StatusOK)
			return
		} else if errors.Is(err, order.ErrIssetOrderNotOwner) {
			w.WriteHeader(http.StatusConflict)
			return
		} else if errors.Is(err, order.ErrInvalidFormat) {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) ListOrdersHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), config.Options.FlagTimeoutContext)
	defer cancel()

	userID, ok := cookies.UIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Get orders of user
	orders, err := h.service.GetUserOrders(ctx, *userID)
	if err != nil {
		h.logger.Sugar().Error(err)

		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := make([]*orderListResponse, 0, len(orders))
	for _, o := range orders {
		resp = append(resp, &orderListResponse{
			Number:     o.Number,
			Status:     string(o.Status),
			Accrual:    o.Balance,
			UploadedAt: o.CreatedAt.Format(time.RFC3339),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	ordersResp, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(ordersResp)

}
