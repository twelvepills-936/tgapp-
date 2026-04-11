package models

import "time"

type Wallet struct {
	ID               int64
	ProfileID        int64
	Balance          int64 // store as cents
	TotalEarned      int64
	BalanceAvailable int64
}

type WalletTransaction struct {
	ID          int64
	WalletID    int64
	Date        time.Time
	Type        string
	Amount      int64
	Status      string
	Description string
	Details     string
}

type WithdrawMethod struct {
	ID       int64
	WalletID int64
	Type     string
	Details  string
}

type Referral struct {
	ID                  int64
	TelegramID          string
	Name                string
	Username            string
	CompletedTasksCount int64
	Earnings            int64
}

type PromptHistory struct {
	ID         int64
	ProfileID  int64
	TelegramID string
	Prompt     string
	Category   string
	CreatedAt  time.Time
}
