package usecase

import (
    "context"
    "errors"
    "testing"

    "github.com/jackc/pgx/v5"
    "gitlab16.skiftrade.kz/templates/go/internal"
    repoModels "gitlab16.skiftrade.kz/templates/go/internal/repository/models"
    ucModels "gitlab16.skiftrade.kz/templates/go/internal/usecase/models"
)

type fakeRepo struct{}

func (f *fakeRepo) DBBeginTransaction(ctx context.Context) (pgx.Tx, error) { return nil, nil }
func (f *fakeRepo) ReadUser(ctx context.Context, id int64, _ pgx.Tx) (repoModels.User, error) {
    if id == 42 { return repoModels.User{ID: 42, Name: "John", Surname: "Doe"}, nil }
    return repoModels.User{}, repoModels.ErrUserIsNotFound
}
// Facebase methods (unused in this test)
func (f *fakeRepo) CreateProfile(ctx context.Context, tx pgx.Tx, p repoModels.Profile) (int64, error) { return 0, nil }
func (f *fakeRepo) GetProfileByTelegramID(ctx context.Context, tx pgx.Tx, telegramID string) (repoModels.Profile, error) { return repoModels.Profile{}, errors.New("not impl") }
func (f *fakeRepo) CreateWalletForUser(ctx context.Context, tx pgx.Tx, profileID int64) (int64, error) { return 0, nil }
func (f *fakeRepo) AddReferral(ctx context.Context, tx pgx.Tx, referrerProfileID int64, refereeProfileID int64) error { return nil }

func TestGetUser_OK(t *testing.T) {
    var _ internal.Repository = (*fakeRepo)(nil)
    uc := NewUseCase(&fakeRepo{})
    out, err := uc.GetUser(context.Background(), ucModels.GetUserInput{UserID: 42})
    if err != nil { t.Fatalf("unexpected error: %v", err) }
    if out.Data.ID != 42 || out.Data.Name != "John" || out.Data.Surname != "Doe" {
        t.Fatalf("unexpected output: %+v", out)
    }
}

func TestGetUser_NotFound(t *testing.T) {
    uc := NewUseCase(&fakeRepo{})
    _, err := uc.GetUser(context.Background(), ucModels.GetUserInput{UserID: 99})
    if !errors.Is(err, ucModels.ErrUserIsNotFound) {
        t.Fatalf("expected ErrUserIsNotFound, got %v", err)
    }
}


