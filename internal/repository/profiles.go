package repository

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"
	repoModels "gitlab16.skiftrade.kz/templates/go/internal/repository/models"
)

// CreateProfile inserts a new profile and returns its id.
func (r *Repository) CreateProfile(ctx context.Context, tx pgx.Tx, p repoModels.Profile) (int64, error) {
	const q = `
INSERT INTO profiles(name, telegram_id, avatar, location, role, description, telegram_init_data, username, verified)
VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)
RETURNING id`

	var id int64
	qry := r.getQueryable(tx)
	err := qry.QueryRow(ctx, q, p.Name, p.TelegramID, p.Avatar, p.Location, p.Role, p.Description, p.TelegramInitData, p.Username, p.Verified).Scan(&id)
	return id, err
}

// GetProfileByTelegramID returns profile by telegram id or pgx.ErrNoRows.
func (r *Repository) GetProfileByTelegramID(ctx context.Context, tx pgx.Tx, telegramID string) (repoModels.Profile, error) {
	const q = `
SELECT id, name, telegram_id, avatar, location, role, description, telegram_init_data, username, verified, created_at, updated_at
FROM profiles WHERE telegram_id = $1 LIMIT 1`

	var p repoModels.Profile
	qry := r.getQueryable(tx)
	err := qry.QueryRow(ctx, q, telegramID).Scan(&p.ID, &p.Name, &p.TelegramID, &p.Avatar, &p.Location, &p.Role, &p.Description, &p.TelegramInitData, &p.Username, &p.Verified, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return p, pgx.ErrNoRows
		}
		slog.ErrorContext(ctx, "failed to get profile", slog.Any("error", err), slog.String("telegram_id", telegramID))
		return p, err
	}
	return p, nil
}

// CreateWalletForUser creates wallet for profile.
func (r *Repository) CreateWalletForUser(ctx context.Context, tx pgx.Tx, profileID int64) (int64, error) {
	const q = `INSERT INTO wallets(profile_id, balance, total_earned, balance_available) VALUES($1,0,0,0) RETURNING id`
	var id int64
	qry := r.getQueryable(tx)
	err := qry.QueryRow(ctx, q, profileID).Scan(&id)
	return id, err
}

// AddReferral links referrer to referee with zeroed stats.
func (r *Repository) AddReferral(ctx context.Context, tx pgx.Tx, referrerProfileID int64, refereeProfileID int64) error {
	const q = `INSERT INTO referrals(referrer_profile_id, referee_profile_id, completed_tasks_count, earnings) VALUES($1,$2,0,0)`
	qry := r.getQueryable(tx)
	_, err := qry.Exec(ctx, q, referrerProfileID, refereeProfileID)
	return err
}

// GetWalletByTelegramID returns wallet data for a profile by telegram id.
func (r *Repository) GetWalletByTelegramID(ctx context.Context, tx pgx.Tx, telegramID string) (repoModels.Wallet, error) {
	const q = `
SELECT COALESCE(w.id, 0), p.id, COALESCE(w.balance, 0), COALESCE(w.total_earned, 0), COALESCE(w.balance_available, 0)
FROM profiles p
LEFT JOIN wallets w ON w.profile_id = p.id
WHERE p.telegram_id = $1
LIMIT 1`

	var w repoModels.Wallet
	qry := r.getQueryable(tx)
	err := qry.QueryRow(ctx, q, telegramID).Scan(&w.ID, &w.ProfileID, &w.Balance, &w.TotalEarned, &w.BalanceAvailable)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return w, pgx.ErrNoRows
		}
		slog.ErrorContext(ctx, "failed to get wallet", slog.Any("error", err), slog.String("telegram_id", telegramID))
		return w, err
	}
	return w, nil
}

// ListWalletTransactionsByTelegramID returns recent wallet transactions for a profile.
func (r *Repository) ListWalletTransactionsByTelegramID(ctx context.Context, tx pgx.Tx, telegramID string, limit int32) ([]repoModels.WalletTransaction, error) {
	const q = `
SELECT wt.id, wt.wallet_id, wt.date, wt.type, wt.amount, wt.status, COALESCE(wt.description, ''), COALESCE(wt.details, '')
FROM profiles p
JOIN wallets w ON w.profile_id = p.id
JOIN wallet_transactions wt ON wt.wallet_id = w.id
WHERE p.telegram_id = $1
ORDER BY wt.date DESC, wt.id DESC
LIMIT $2`

	qry := r.getQueryable(tx)
	rows, err := qry.Query(ctx, q, telegramID, limit)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list wallet transactions", slog.Any("error", err), slog.String("telegram_id", telegramID))
		return nil, err
	}
	defer rows.Close()

	items := make([]repoModels.WalletTransaction, 0)
	for rows.Next() {
		var item repoModels.WalletTransaction
		if scanErr := rows.Scan(&item.ID, &item.WalletID, &item.Date, &item.Type, &item.Amount, &item.Status, &item.Description, &item.Details); scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return items, nil
}

// ListReferralsByTelegramID returns referrals made by the specified profile.
func (r *Repository) ListReferralsByTelegramID(ctx context.Context, tx pgx.Tx, telegramID string) ([]repoModels.Referral, error) {
	const q = `
SELECT r.id, referee.telegram_id, referee.name, COALESCE(referee.username, ''), r.completed_tasks_count, r.earnings
FROM profiles owner
JOIN referrals r ON r.referrer_profile_id = owner.id
JOIN profiles referee ON referee.id = r.referee_profile_id
WHERE owner.telegram_id = $1
ORDER BY r.id DESC`

	qry := r.getQueryable(tx)
	rows, err := qry.Query(ctx, q, telegramID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list referrals", slog.Any("error", err), slog.String("telegram_id", telegramID))
		return nil, err
	}
	defer rows.Close()

	items := make([]repoModels.Referral, 0)
	for rows.Next() {
		var item repoModels.Referral
		if scanErr := rows.Scan(&item.ID, &item.TelegramID, &item.Name, &item.Username, &item.CompletedTasksCount, &item.Earnings); scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return items, nil
}

// CreatePromptHistory saves a prompt history record for a profile.
func (r *Repository) CreatePromptHistory(ctx context.Context, tx pgx.Tx, item repoModels.PromptHistory) (int64, error) {
	const q = `
INSERT INTO prompt_history(profile_id, prompt, category)
VALUES($1, $2, $3)
RETURNING id`

	qry := r.getQueryable(tx)
	var id int64
	err := qry.QueryRow(ctx, q, item.ProfileID, item.Prompt, item.Category).Scan(&id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to create prompt history", slog.Any("error", err), slog.Int64("profile_id", item.ProfileID))
		return 0, err
	}
	return id, nil
}

// ListPromptHistoryByTelegramID returns recent saved prompts for a profile.
func (r *Repository) ListPromptHistoryByTelegramID(ctx context.Context, tx pgx.Tx, telegramID string, limit int32) ([]repoModels.PromptHistory, error) {
	const q = `
SELECT ph.id, ph.profile_id, p.telegram_id, ph.prompt, COALESCE(ph.category, ''), ph.created_at
FROM profiles p
JOIN prompt_history ph ON ph.profile_id = p.id
WHERE p.telegram_id = $1
ORDER BY ph.created_at DESC, ph.id DESC
LIMIT $2`

	qry := r.getQueryable(tx)
	rows, err := qry.Query(ctx, q, telegramID, limit)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list prompt history", slog.Any("error", err), slog.String("telegram_id", telegramID))
		return nil, err
	}
	defer rows.Close()

	items := make([]repoModels.PromptHistory, 0)
	for rows.Next() {
		var item repoModels.PromptHistory
		if scanErr := rows.Scan(&item.ID, &item.ProfileID, &item.TelegramID, &item.Prompt, &item.Category, &item.CreatedAt); scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return items, nil
}
