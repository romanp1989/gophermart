package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/romanp1989/gophermart/internal/api/balance"
	"github.com/romanp1989/gophermart/internal/api/order"
	"github.com/romanp1989/gophermart/internal/api/user"
	"github.com/romanp1989/gophermart/internal/middlewares"
)

func NewRoutes(u *user.Handler, o *order.Handler, b *balance.Handler, middlewares *middlewares.Middlewares) *chi.Mux {
	r := chi.NewRouter()

	r.Post("/api/user/register", u.RegisterHandler)
	r.Post("/api/user/login", u.LoginHandler)
	r.Post("/api/user/orders", middlewares.AuthMiddleware(o.CreateOrderHandler))
	r.Get("/api/user/orders", middlewares.AuthMiddleware(o.ListOrdersHandler))
	r.Get("/api/user/balance", middlewares.AuthMiddleware(b.GetBalanceHandler))
	r.Post("/api/user/balance/withdraw", middlewares.AuthMiddleware(b.WithdrawHandler))
	r.Get("/api/user/withdrawals", middlewares.AuthMiddleware(b.GetWithdrawHandler))
	return r
}
