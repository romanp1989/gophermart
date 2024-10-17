package cookies

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/romanp1989/gophermart/internal/config"
	"github.com/romanp1989/gophermart/internal/domain"
	"net/http"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID domain.UserID `json:"user_id,omitempty"`
}

func NewCookie(w http.ResponseWriter, userID domain.UserID, secretKey string) (string, error) {
	token, err := createToken(userID, secretKey)
	if err != nil {
		return "", err
	}

	cookie := &http.Cookie{
		Name:  "Token",
		Value: token,
		Path:  "/",
	}

	http.SetCookie(w, cookie)

	return token, nil
}

func createToken(userID domain.UserID, secretKey string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{},
		UserID:           userID,
	})

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func Validation(tokenString string, secretKey string) bool {
	token, err := jwt.Parse(tokenString,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(secretKey), nil
		})

	if err != nil || !token.Valid {
		return false
	}
	return true
}

func GetUserID(tokenString string) (*domain.UserID, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("неизвестный алгоритм подписи: %v", t.Header["alg"])
			}
			return []byte(config.Options.FlagSecretKey), nil
		})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, err
	}

	fmt.Println("токен валидный", claims.UserID)
	return &claims.UserID, nil
}
