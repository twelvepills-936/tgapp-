package models

import "errors"

var (
	ErrUserIsNotFound = errors.New("ErrUserIsNotFound")
)

type User struct {
	ID      int64
	Name    string
	Surname string
}
