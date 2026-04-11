package usecase

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"time"

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
	if err := validateTelegramID(telegramID); err != nil {
		return ucModels.GetProfileOutput{}, err
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

func (uc *useCase) GetWalletByTelegramID(ctx context.Context, telegramID string) (ucModels.GetWalletOutput, error) {
	if err := validateTelegramID(telegramID); err != nil {
		return ucModels.GetWalletOutput{}, err
	}

	wallet, err := uc.repo.GetWalletByTelegramID(ctx, nil, telegramID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ucModels.GetWalletOutput{}, ucModels.ErrProfileNotFound
		}
		return ucModels.GetWalletOutput{}, err
	}

	transactions, err := uc.repo.ListWalletTransactionsByTelegramID(ctx, nil, telegramID, 20)
	if err != nil {
		return ucModels.GetWalletOutput{}, err
	}

	items := make([]ucModels.WalletTransactionView, 0, len(transactions))
	for _, item := range transactions {
		items = append(items, ucModels.WalletTransactionView{
			ID:          item.ID,
			Date:        item.Date.Format("2006-01-02 15:04"),
			Type:        item.Type,
			Amount:      item.Amount,
			Status:      item.Status,
			Description: item.Description,
		})
	}

	return ucModels.GetWalletOutput{
		Wallet: ucModels.WalletView{
			ID:               wallet.ID,
			ProfileID:        wallet.ProfileID,
			Balance:          wallet.Balance,
			TotalEarned:      wallet.TotalEarned,
			BalanceAvailable: wallet.BalanceAvailable,
		},
		Transactions: items,
	}, nil
}

func (uc *useCase) GetReferralsByTelegramID(ctx context.Context, telegramID string) (ucModels.GetReferralsOutput, error) {
	if err := validateTelegramID(telegramID); err != nil {
		return ucModels.GetReferralsOutput{}, err
	}

	if _, err := uc.repo.GetProfileByTelegramID(ctx, nil, telegramID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ucModels.GetReferralsOutput{}, ucModels.ErrProfileNotFound
		}
		return ucModels.GetReferralsOutput{}, err
	}

	items, err := uc.repo.ListReferralsByTelegramID(ctx, nil, telegramID)
	if err != nil {
		return ucModels.GetReferralsOutput{}, err
	}

	result := make([]ucModels.ReferralView, 0, len(items))
	for _, item := range items {
		result = append(result, ucModels.ReferralView{
			ID:                  item.ID,
			TelegramID:          item.TelegramID,
			Name:                item.Name,
			Username:            item.Username,
			CompletedTasksCount: item.CompletedTasksCount,
			Earnings:            item.Earnings,
		})
	}

	return ucModels.GetReferralsOutput{Items: result}, nil
}

func (uc *useCase) SavePromptHistory(ctx context.Context, input ucModels.SavePromptHistoryInput) (ucModels.SavePromptHistoryOutput, error) {
	if err := input.Validate(); err != nil {
		return ucModels.SavePromptHistoryOutput{}, err
	}

	profile, err := uc.repo.GetProfileByTelegramID(ctx, nil, input.TelegramID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ucModels.SavePromptHistoryOutput{}, ucModels.ErrProfileNotFound
		}
		return ucModels.SavePromptHistoryOutput{}, err
	}

	category := input.Category
	if category == "" {
		category = "general"
	}

	item := repoModels.PromptHistory{
		ProfileID:  profile.ID,
		TelegramID: input.TelegramID,
		Prompt:     input.Prompt,
		Category:   category,
	}

	id, err := uc.repo.CreatePromptHistory(ctx, nil, item)
	if err != nil {
		return ucModels.SavePromptHistoryOutput{}, err
	}

	item.ID = id
	item.CreatedAt = time.Now()

	return ucModels.SavePromptHistoryOutput{Item: mapPromptHistoryItem(item)}, nil
}

func (uc *useCase) GetPromptHistoryByTelegramID(ctx context.Context, telegramID string) (ucModels.GetPromptHistoryOutput, error) {
	if err := validateTelegramID(telegramID); err != nil {
		return ucModels.GetPromptHistoryOutput{}, err
	}

	if _, err := uc.repo.GetProfileByTelegramID(ctx, nil, telegramID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ucModels.GetPromptHistoryOutput{}, ucModels.ErrProfileNotFound
		}
		return ucModels.GetPromptHistoryOutput{}, err
	}

	items, err := uc.repo.ListPromptHistoryByTelegramID(ctx, nil, telegramID, 50)
	if err != nil {
		return ucModels.GetPromptHistoryOutput{}, err
	}

	result := make([]ucModels.PromptHistoryItem, 0, len(items))
	for _, item := range items {
		result = append(result, mapPromptHistoryItem(item))
	}

	return ucModels.GetPromptHistoryOutput{Items: result}, nil
}

func validateTelegramID(telegramID string) error {
	if telegramID == "" {
		return fmt.Errorf("telegram_id is required")
	}

	if len(telegramID) > 100 {
		return fmt.Errorf("telegram_id too long")
	}

	return nil
}

func mapPromptHistoryItem(item repoModels.PromptHistory) ucModels.PromptHistoryItem {
	return ucModels.PromptHistoryItem{
		ID:        item.ID,
		Prompt:    item.Prompt,
		Category:  item.Category,
		CreatedAt: item.CreatedAt.Format("2006-01-02 15:04"),
	}
}

// helpers
func intToString(v int64) string { return fmt.Sprintf("%d", v) }
