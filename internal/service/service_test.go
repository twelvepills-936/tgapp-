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

func (f *fakeUC) GetUser(ctx context.Context, input ucModels.GetUserInput) (ucModels.GetUserOutput, error) {
    if input.UserID == 7 { return ucModels.GetUserOutput{Data: ucModels.User{ID:7, Name:"N", Surname:"S"}}, nil }
    return ucModels.GetUserOutput{}, ucModels.ErrUserIsNotFound
}
func (f *fakeUC) RegisterByTelegram(ctx context.Context, input ucModels.RegisterByTelegramInput) (ucModels.RegisterByTelegramOutput, error) {
    if input.InitDataRaw == "dup" { return ucModels.RegisterByTelegramOutput{}, ucModels.ErrProfileAlreadyRegistered }
    if input.InitDataRaw == "" { return ucModels.RegisterByTelegramOutput{}, errors.New("bad") }
    return ucModels.RegisterByTelegramOutput{ProfileID: 1}, nil
}
func (f *fakeUC) GetUserByTelegramID(ctx context.Context, telegramID string) (ucModels.GetProfileOutput, error) {
    if telegramID == "x" { return ucModels.GetProfileOutput{}, ucModels.ErrProfileNotFound }
    return ucModels.GetProfileOutput{Data: ucModels.ProfileUser{ID:1, Name:"A"}}, nil
}

var _ internal.UseCase = (*fakeUC)(nil)

func TestService_GetUser_OK(t *testing.T) {
    s := NewService(&fakeUC{})
    out, err := s.GetUser(context.Background(), &api.GetUserRequest{UserId: 7})
    if err != nil { t.Fatalf("unexpected error: %v", err) }
    if out.GetData().GetId() != 7 { t.Fatalf("unexpected id: %d", out.GetData().GetId()) }
}

func TestService_RegisterByTelegram_AlreadyExists(t *testing.T) {
    s := NewService(&fakeUC{})
    _, err := s.RegisterByTelegram(context.Background(), &api.RegisterByTelegramRequest{InitDataRaw: "dup"})
    if err == nil { t.Fatalf("expected error") }
}

func TestService_GetUserByTelegramId_NotFound(t *testing.T) {
    s := NewService(&fakeUC{})
    _, err := s.GetUserByTelegramId(context.Background(), &api.GetUserByTelegramIdRequest{TelegramId: "x"})
    if err == nil { t.Fatalf("expected not found") }
}


