package user

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/romanp1989/gophermart/internal/config"
	"github.com/romanp1989/gophermart/internal/cookies"
	"github.com/romanp1989/gophermart/internal/domain"
	"github.com/romanp1989/gophermart/internal/user"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"io"
	"net/http"
)

type Handler struct {
	service *user.Service
	logger  *zap.Logger
}

func NewUserHandler(userService *user.Service, logger *zap.Logger) *Handler {
	return &Handler{
		service: userService,
		logger:  logger,
	}
}

func (h *Handler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	requestReg := new(registrationRequest)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "некорректный формат запроса", http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(body, requestReg); err != nil {
		http.Error(w, "некорректный формат запроса", http.StatusBadRequest)
		return
	}

	if !h.service.ValidateLogin(requestReg.Login) {
		http.Error(w, "неверный формат логина. логин может содержать только буквы латинского алфавита и цифры. длина логина от  до 16 символов", http.StatusBadRequest)
		return
	}

	if !h.service.ValidatePassword(requestReg.Password) {
		http.Error(w, "неверный формат пароля. "+
			"пароль должен содержать хотя бы 1 букву латинского алфавита в верхнем регистре, 1 букву в нижнем регистре, "+
			"1 цифру, 1 специальный символ. длина пароля от 8 до 32 символов", http.StatusBadRequest)
		return
	}

	userReg := domain.User{
		Login:    requestReg.Login,
		Password: requestReg.Password,
	}

	createdUser, err := h.service.CreateUser(r.Context(), userReg)
	if err != nil {
		if errors.Is(err, domain.ErrLoginExists) {
			http.Error(w, "логин уже занят", http.StatusConflict)
			return
		}

		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	_, err = cookies.NewCookie(w, createdUser.ID, config.Options.FlagSecretKey)
	if err != nil {
		h.logger.Sugar().Error(err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)

}

func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	requestReg := new(registrationRequest)

	ctx, cancel := context.WithTimeout(r.Context(), config.Options.FlagTimeoutContext)
	defer cancel()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "некорректный формат запроса", http.StatusBadRequest)
		return
	}

	if err = json.Unmarshal(body, requestReg); err != nil {
		http.Error(w, "некорректный формат запроса", http.StatusBadRequest)
		return
	}

	userReg := &domain.User{
		Login:    requestReg.Login,
		Password: requestReg.Password,
	}

	userReg, err = h.service.Authorization(ctx, userReg)
	if err != nil {
		h.logger.Sugar().Error(err)

		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, err = cookies.NewCookie(w, userReg.ID, config.Options.FlagSecretKey)
	if err != nil {
		h.logger.Sugar().Error(err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
}
