package models

import "errors"

var (
	ErrUserIsNotFound  = errors.New("ErrUserIsNotFound")
	ErrUserIDIsInvalid = errors.New("ErrUserIDIsInvalid")
)

type GetUserInput struct {
	UserID int64
}

func (i *GetUserInput) Validate() error {
	if i.UserID <= 0 {
		return ErrUserIDIsInvalid
	}
	return nil
}

type GetUserOutput struct {
	Data User
}

type User struct {
	ID      int64
	Name    string
	Surname string
}
