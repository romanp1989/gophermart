package balance

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/romanp1989/gophermart/internal/balance"
	"github.com/romanp1989/gophermart/internal/config"
	"github.com/romanp1989/gophermart/internal/cookies"
	"github.com/romanp1989/gophermart/internal/domain"
	"github.com/romanp1989/gophermart/internal/order"
	"go.uber.org/zap"
	"io"
	"net/http"
	"time"
)

type Handler struct {
	balanceService *balance.Service
	logger         *zap.Logger
}

type balanceResponse struct {
	Current float64 `json:"current"`
	Sum     float64 `json:"withdrawn"`
}

type withdrawRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

type balanceListResponse struct {
	Order       string  `json:"order"`
	Sum         float64 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}

func NewBalanceHandler(balanceService *balance.Service, logger *zap.Logger) *Handler {
	return &Handler{
		balanceService: balanceService,
		logger:         logger,
	}
}

func (h *Handler) GetBalanceHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), config.Options.FlagTimeoutContext)
	defer cancel()

	userID, ok := cookies.UIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	balanceSum, err := h.balanceService.LoadSum(ctx, *userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := balanceResponse{
		Current: balanceSum.Current,
		Sum:     balanceSum.Withdrawn,
	}

	b, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(b)
}

func (h *Handler) WithdrawHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), config.Options.FlagTimeoutContext)
	defer cancel()

	userID, ok := cookies.UIDFromContext(ctx)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	req := withdrawRequest{}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err = json.Unmarshal(body, &req); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = h.balanceService.Withdraw(ctx, userID, req.Order, req.Sum)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidFormat) {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		} else if errors.Is(err, domain.ErrBalanceInsufficientFunds) {
			w.WriteHeader(http.StatusPaymentRequired)
			return
		} else if errors.Is(err, order.ErrNotFoundOrder) {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetWithdrawHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), config.Options.FlagTimeoutContext)
	defer cancel()

	userID, ok := cookies.UIDFromContext(ctx)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	withdrawals, err := h.balanceService.LoadWithdrawals(ctx, *userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(withdrawals) < 1 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	resp := make([]*balanceListResponse, 0, len(withdrawals))
	for _, withdrawal := range withdrawals {
		resp = append(resp, &balanceListResponse{
			Order:       withdrawal.OrderNumber,
			Sum:         withdrawal.Sum,
			ProcessedAt: withdrawal.CreatedAt.Format(time.RFC3339),
		})
	}

	b, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(b)
}
