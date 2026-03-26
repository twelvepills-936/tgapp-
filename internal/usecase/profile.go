package usecase

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"github.com/jackc/pgx/v5"
	repoModels "gitlab16.skiftrade.kz/templates/go/internal/repository/models"
	ucModels "gitlab16.skiftrade.kz/templates/go/internal/usecase/models"
)

func (uc *useCase) RegisterByTelegram(ctx context.Context, input ucModels.RegisterByTelegramInput) (output ucModels.RegisterByTelegramOutput, err error) {
	// Validate input
	if err := input.Validate(); err != nil {
		return ucModels.RegisterByTelegramOutput{}, err
	}

	// decode init_data_raw (base64) and parse user json param similar to Node parseAuthToken path
	decoded, decodeErr := base64.StdEncoding.DecodeString(input.InitDataRaw)
	if decodeErr != nil {
		return output, fmt.Errorf("failed to decode init data: %w", decodeErr)
	}
	params, parseErr := url.ParseQuery(string(decoded))
	if parseErr != nil {
		return output, fmt.Errorf("failed to parse init data: %w", parseErr)
	}
	userStr := params.Get("user")
	if userStr == "" {
		return output, errors.New("user not found in init data")
	}
	// we don't need full user struct now; minimally extract fields from json
	// to keep scope tight, parse a subset via a lightweight map
	tg := struct {
		ID           int64  `json:"id"`
		FirstName    string `json:"first_name"`
		Username     string `json:"username"`
		PhotoURL     string `json:"photo_url"`
		LanguageCode string `json:"language_code"`
	}{}
	if unmarshalErr := json.Unmarshal([]byte(userStr), &tg); unmarshalErr != nil {
		return output, fmt.Errorf("failed to unmarshal user data: %w", unmarshalErr)
	}

	tx, txErr := uc.repo.DBBeginTransaction(ctx)
	if txErr != nil {
		return output, txErr
	}
	defer func() {
		if err != nil && tx != nil {
			// Use background context for rollback to avoid context cancellation issues
			_ = tx.Rollback(context.Background())
		}
	}()

	// if exists, return already registered
	if _, checkErr := uc.repo.GetProfileByTelegramID(ctx, tx, intToString(tg.ID)); checkErr == nil {
		err = ucModels.ErrProfileAlreadyRegistered
		return output, err
	} else if !errors.Is(checkErr, pgx.ErrNoRows) {
		err = checkErr
		return output, err
	}

	pid, createErr := uc.repo.CreateProfile(ctx, tx, repoModels.Profile{
		Name:             tg.FirstName,
		TelegramID:       intToString(tg.ID),
		Avatar:           tg.PhotoURL,
		Location:         tg.LanguageCode,
		Role:             "",
		Description:      "",
		TelegramInitData: input.InitDataRaw,
		Username:         tg.Username,
		Verified:         false,
	})
	if createErr != nil {
		err = createErr
		return output, err
	}

	if _, walletErr := uc.repo.CreateWalletForUser(ctx, tx, pid); walletErr != nil {
		err = walletErr
		return output, err
	}

	if input.StartParam != "" {
		// find referrer by telegram_id
		ref, refErr := uc.repo.GetProfileByTelegramID(ctx, tx, input.StartParam)
		if refErr == nil && ref.ID != 0 {
			if addRefErr := uc.repo.AddReferral(ctx, tx, ref.ID, pid); addRefErr != nil {
				err = addRefErr
				return output, err
			}
		}
	}

	if commitErr := tx.Commit(ctx); commitErr != nil {
		err = commitErr
		return output, err
	}

	return ucModels.RegisterByTelegramOutput{ProfileID: pid}, nil
}

func (uc *useCase) GetUserByTelegramID(ctx context.Context, telegramID string) (ucModels.GetProfileOutput, error) {
	// Validate telegram_id
	if telegramID == "" {
		return ucModels.GetProfileOutput{}, fmt.Errorf("telegram_id is required")
	}

	if len(telegramID) > 100 {
		return ucModels.GetProfileOutput{}, fmt.Errorf("telegram_id too long")
	}

	p, err := uc.repo.GetProfileByTelegramID(ctx, nil, telegramID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ucModels.GetProfileOutput{}, ucModels.ErrProfileNotFound
		}
		return ucModels.GetProfileOutput{}, err
	}
	return ucModels.GetProfileOutput{Data: ucModels.ProfileUser{
		ID:         p.ID,
		Name:       p.Name,
		TelegramID: p.TelegramID,
		Avatar:     p.Avatar,
		Username:   p.Username,
		Verified:   p.Verified,
	}}, nil
}

// helpers
func intToString(v int64) string { return fmt.Sprintf("%d", v) }
