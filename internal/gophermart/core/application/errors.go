package application

import "errors"

var (
	ErrUserNotFound  = errors.New("user not found")
	ErrUserExists    = errors.New("user exists")
	ErrIncorrectPass = errors.New("incorrect password")

	ErrOrderAlreadyExists       = errors.New("order already exists")
	ErrOrderExistsOnAnotherUser = errors.New("order exists on another user")
	ErrInvalidOrderID           = errors.New("invalid order id")
)
