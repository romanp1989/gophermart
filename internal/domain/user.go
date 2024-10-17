package domain

import "errors"

type UserID int64

var ErrLoginExists = errors.New("данные логин уже используется")

type User struct {
	ID       UserID
	Login    string
	Password string
}
