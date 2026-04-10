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
