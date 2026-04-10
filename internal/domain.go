package internal

import (
	"context"

	"github.com/jackc/pgx/v5"
	repoModels "gitlab16.skiftrade.kz/templates/go/internal/repository/models"
	ucModels "gitlab16.skiftrade.kz/templates/go/internal/usecase/models"
)

type Repository interface {
	DBBeginTransaction(ctx context.Context) (pgx.Tx, error)

	CreateProfile(ctx context.Context, tx pgx.Tx, p repoModels.Profile) (int64, error)
	GetProfileByTelegramID(ctx context.Context, tx pgx.Tx, telegramID string) (repoModels.Profile, error)
	CreateWalletForUser(ctx context.Context, tx pgx.Tx, profileID int64) (int64, error)
	AddReferral(ctx context.Context, tx pgx.Tx, referrerProfileID int64, refereeProfileID int64) error
}

type UseCase interface {
	RegisterByTelegram(ctx context.Context, input ucModels.RegisterByTelegramInput) (ucModels.RegisterByTelegramOutput, error)
	GetUserByTelegramID(ctx context.Context, telegramID string) (ucModels.GetProfileOutput, error)
}
