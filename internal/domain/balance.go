package domain

import (
	"errors"
	"time"
)

type BalanceType int

var (
	ErrBalanceInsufficientFunds = errors.New("на счету недостаточно средств")
)

var (
	BalanceTypeAdded     BalanceType = 0
	BalanceTypeWithdrawn BalanceType = 1
)

type Balance struct {
	ID          int64
	CreatedAt   time.Time
	OrderNumber string
	UserID      UserID
	Sum         float64
	Type        BalanceType
}

type BalanceSum struct {
	UserID UserID
	Type   BalanceType
	Sum    float64
}

type UserSum struct {
	Current   float64
	Withdrawn float64
}
