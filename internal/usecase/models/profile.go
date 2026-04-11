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

type GetWalletOutput struct {
	Wallet       WalletView
	Transactions []WalletTransactionView
}

type WalletView struct {
	ID               int64
	ProfileID        int64
	Balance          int64
	TotalEarned      int64
	BalanceAvailable int64
}

type WalletTransactionView struct {
	ID          int64
	Date        string
	Type        string
	Amount      int64
	Status      string
	Description string
}

type GetReferralsOutput struct {
	Items []ReferralView
}

type ReferralView struct {
	ID                  int64
	TelegramID          string
	Name                string
	Username            string
	CompletedTasksCount int64
	Earnings            int64
}

type SavePromptHistoryInput struct {
	TelegramID string
	Prompt     string
	Category   string
}

func (i *SavePromptHistoryInput) Validate() error {
	if i.TelegramID == "" {
		return fmt.Errorf("%w: telegram_id is required", ErrInvalidInput)
	}
	if i.Prompt == "" {
		return fmt.Errorf("%w: prompt is required", ErrInvalidInput)
	}
	if len(i.Prompt) > 4000 {
		return fmt.Errorf("%w: prompt too long", ErrInvalidInput)
	}
	if len(i.Category) > 100 {
		return fmt.Errorf("%w: category too long", ErrInvalidInput)
	}
	return nil
}

type SavePromptHistoryOutput struct {
	Item PromptHistoryItem
}

type GetPromptHistoryOutput struct {
	Items []PromptHistoryItem
}

type PromptHistoryItem struct {
	ID        int64
	Prompt    string
	Category  string
	CreatedAt string
}
