package service

import (
	"context"
	"errors"
	"testing"

	"gitlab16.skiftrade.kz/templates/go/internal"
	ucModels "gitlab16.skiftrade.kz/templates/go/internal/usecase/models"
	api "gitlab16.skiftrade.kz/templates/go/pkg/api"
)

type fakeUC struct{}

func (f *fakeUC) RegisterByTelegram(ctx context.Context, input ucModels.RegisterByTelegramInput) (ucModels.RegisterByTelegramOutput, error) {
	if input.InitDataRaw == "dup" {
		return ucModels.RegisterByTelegramOutput{}, ucModels.ErrProfileAlreadyRegistered
	}
	if input.InitDataRaw == "" {
		return ucModels.RegisterByTelegramOutput{}, errors.New("bad")
	}
	return ucModels.RegisterByTelegramOutput{ProfileID: 1}, nil
}

func (f *fakeUC) GetUserByTelegramID(ctx context.Context, telegramID string) (ucModels.GetProfileOutput, error) {
	if telegramID == "x" {
		return ucModels.GetProfileOutput{}, ucModels.ErrProfileNotFound
	}
	return ucModels.GetProfileOutput{Data: ucModels.ProfileUser{ID: 1, Name: "A"}}, nil
}

func (f *fakeUC) GetWalletByTelegramID(ctx context.Context, telegramID string) (ucModels.GetWalletOutput, error) {
	if telegramID == "x" {
		return ucModels.GetWalletOutput{}, ucModels.ErrProfileNotFound
	}
	return ucModels.GetWalletOutput{
		Wallet:       ucModels.WalletView{ID: 10, Balance: 1500, TotalEarned: 3000, BalanceAvailable: 1200},
		Transactions: []ucModels.WalletTransactionView{{ID: 1, Type: "referral", Amount: 1500, Status: "completed", Description: "Reward"}},
	}, nil
}

func (f *fakeUC) GetReferralsByTelegramID(ctx context.Context, telegramID string) (ucModels.GetReferralsOutput, error) {
	if telegramID == "x" {
		return ucModels.GetReferralsOutput{}, ucModels.ErrProfileNotFound
	}
	return ucModels.GetReferralsOutput{Items: []ucModels.ReferralView{{ID: 1, Name: "Friend", Username: "friend", Earnings: 500}}}, nil
}

func (f *fakeUC) SavePromptHistory(ctx context.Context, input ucModels.SavePromptHistoryInput) (ucModels.SavePromptHistoryOutput, error) {
	if input.Prompt == "" {
		return ucModels.SavePromptHistoryOutput{}, ucModels.ErrInvalidInput
	}
	return ucModels.SavePromptHistoryOutput{Item: ucModels.PromptHistoryItem{ID: 1, Prompt: input.Prompt, Category: input.Category, CreatedAt: "2026-04-11 12:00"}}, nil
}

func (f *fakeUC) GetPromptHistoryByTelegramID(ctx context.Context, telegramID string) (ucModels.GetPromptHistoryOutput, error) {
	if telegramID == "x" {
		return ucModels.GetPromptHistoryOutput{}, ucModels.ErrProfileNotFound
	}
	return ucModels.GetPromptHistoryOutput{Items: []ucModels.PromptHistoryItem{{ID: 1, Prompt: "Write a Telegram bio", Category: "profile"}}}, nil
}

var _ internal.UseCase = (*fakeUC)(nil)

func TestService_RegisterByTelegram_AlreadyExists(t *testing.T) {
	s := NewService(&fakeUC{})
	_, err := s.RegisterByTelegram(context.Background(), &api.RegisterByTelegramRequest{InitDataRaw: "dup"})
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestService_GetUserByTelegramId_NotFound(t *testing.T) {
	s := NewService(&fakeUC{})
	_, err := s.GetUserByTelegramId(context.Background(), &api.GetUserByTelegramIdRequest{TelegramId: "x"})
	if err == nil {
		t.Fatalf("expected not found")
	}
}

func TestService_GetUserByTelegramId_OK(t *testing.T) {
	s := NewService(&fakeUC{})
	out, err := s.GetUserByTelegramId(context.Background(), &api.GetUserByTelegramIdRequest{TelegramId: "123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.GetData().GetId() != 1 {
		t.Fatalf("unexpected id: %d", out.GetData().GetId())
	}
}

func TestService_GetWalletByTelegramId_OK(t *testing.T) {
	s := NewService(&fakeUC{})
	out, err := s.GetWalletByTelegramId(context.Background(), &api.GetWalletByTelegramIdRequest{TelegramId: "123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.GetWallet().GetBalance() != 1500 {
		t.Fatalf("unexpected balance: %d", out.GetWallet().GetBalance())
	}
}

func TestService_GetReferralsByTelegramId_OK(t *testing.T) {
	s := NewService(&fakeUC{})
	out, err := s.GetReferralsByTelegramId(context.Background(), &api.GetReferralsByTelegramIdRequest{TelegramId: "123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.GetItems()) != 1 {
		t.Fatalf("unexpected referrals count: %d", len(out.GetItems()))
	}
}

func TestService_CreatePromptHistory_OK(t *testing.T) {
	s := NewService(&fakeUC{})
	out, err := s.CreatePromptHistory(context.Background(), &api.CreatePromptHistoryRequest{TelegramId: "123", Prompt: "Hello", Category: "general"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.GetItem().GetPrompt() != "Hello" {
		t.Fatalf("unexpected prompt: %s", out.GetItem().GetPrompt())
	}
}

func TestService_GetPromptHistoryByTelegramId_OK(t *testing.T) {
	s := NewService(&fakeUC{})
	out, err := s.GetPromptHistoryByTelegramId(context.Background(), &api.GetPromptHistoryByTelegramIdRequest{TelegramId: "123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.GetItems()) != 1 {
		t.Fatalf("unexpected history count: %d", len(out.GetItems()))
	}
}
