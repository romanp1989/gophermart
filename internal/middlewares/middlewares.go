package middlewares

import (
	"github.com/romanp1989/gophermart/internal/cookies"
	"net/http"
)

type Middlewares struct {
	key string
}

func New(key string) *Middlewares {
	return &Middlewares{key: key}
}

func (m *Middlewares) AuthMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("Token")
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		userID, err := cookies.GetUserID(cookie.Value)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		ctx := cookies.Context(r.Context(), *userID)
		r = r.WithContext(ctx)

		h.ServeHTTP(w, r.WithContext(ctx))
	})

}
