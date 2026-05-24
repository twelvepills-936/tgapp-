package repository

import "errors"

// ErrInsufficientBalance is returned when wallet balance is too low.
var ErrInsufficientBalance = errors.New("insufficient balance")
