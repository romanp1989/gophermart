package order

import (
	"context"
	"errors"
	"github.com/romanp1989/gophermart/internal/domain"
)

type ValidateError error

var (
	ErrIssetOrder         ValidateError = errors.New("номер заказа уже был загружен")
	ErrIssetOrderNotOwner ValidateError = errors.New("номер заказа уже загружен другим пользователем")
	ErrInvalidFormat      ValidateError = errors.New("некорректный формат номера")
	ErrNotFoundOrder      ValidateError = errors.New("не найден указанный заказ")
)

type Validator struct {
	orderStorage orderStorage
}

func NewValidator(storage orderStorage) *Validator {
	return &Validator{orderStorage: storage}
}

func (v *Validator) Validate(ctx context.Context, number string, userID domain.UserID) ValidateError {
	if len(number) < 1 {
		return ErrInvalidFormat
	}

	if !luhnAlgorithm(number) {
		return ErrInvalidFormat
	}

	order, err := v.orderStorage.LoadOrder(ctx, number)
	if err != nil {
		return err
	}

	if order.UserID != userID {
		return ErrIssetOrderNotOwner
	}

	return nil
}

func luhnAlgorithm(cardNumber string) bool {
	// this function implements the luhn algorithm
	// it takes as argument a cardnumber of type string
	// and it returns a boolean (true or false) if the
	// card number is valid or not

	// initialise a variable to keep track of the total sum of digits
	total := 0
	// Initialize a flag to track whether the current digit is the second digit from the right.
	isSecondDigit := false

	// iterate through the card number digits in reverse order
	for i := len(cardNumber) - 1; i >= 0; i-- {
		// conver the digit character to an integer
		digit := int(cardNumber[i] - '0')

		if isSecondDigit {
			// double the digit for each second digit from the right
			digit *= 2
			if digit > 9 {
				// If doubling the digit results in a two-digit number,
				//subtract 9 to get the sum of digits.
				digit -= 9
			}
		}

		// Add the current digit to the total sum
		total += digit

		//Toggle the flag for the next iteration.
		isSecondDigit = !isSecondDigit
	}

	// return whether the total sum is divisible by 10
	// making it a valid luhn number
	return total%10 == 0
}
