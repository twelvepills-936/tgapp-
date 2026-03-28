package models

import (
	"encoding/base64"
	"errors"
	"fmt"
)

var (
	ErrProfileNotFound          = errors.New("ErrProfileNotFound")
	ErrProfileAlreadyRegistered = errors.New("ErrProfileAlreadyRegistered")
	ErrInvalidInput             = errors.New("ErrInvalidInput")
)

type RegisterByTelegramInput struct {
	InitDataRaw string
	StartParam  string
}

// Validate checks the input data
func (i *RegisterByTelegramInput) Validate() error {
	if i.InitDataRaw == "" {
		return fmt.Errorf("%w: init_data_raw is required", ErrInvalidInput)
	}

	if len(i.InitDataRaw) > 10000 {
		return fmt.Errorf("%w: init_data_raw too long", ErrInvalidInput)
	}

	// Проверка на корректность base64
	if _, err := base64.StdEncoding.DecodeString(i.InitDataRaw); err != nil {
		return fmt.Errorf("%w: init_data_raw is not valid base64", ErrInvalidInput)
	}

	return nil
}

type RegisterByTelegramOutput struct {
    ProfileID int64
}

type GetProfileOutput struct {
    Data ProfileUser
}

type ProfileUser struct {
    ID         int64
    Name       string
    TelegramID string
    Avatar     string
    Username   string
    Verified   bool
}


