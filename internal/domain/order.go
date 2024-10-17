package domain

import (
	"errors"
	"time"
)

type Status string

var (
	OrderStatusNew        Status = "NEW"
	OrderStatusProcessing Status = "PROCESSING"
	OrderStatusInvalid    Status = "INVALID"
	OrderStatusProcessed  Status = "PROCESSED"
)

type OrderNumberError error

var (
	ErrInvalidFormat OrderNumberError = errors.New("некорректный формат номера")
)

type Order struct {
	ID        int64
	CreatedAt time.Time
	Number    string
	Status    Status
	UserID    UserID
}

type OrderWithBalance struct {
	Order
	Balance float64
}
