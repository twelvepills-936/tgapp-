package usecase

import (
	"context"
	"encoding/base64"
	"errors"
	"net/url"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"gitlab16.skiftrade.kz/templates/go/internal"
	repoModels "gitlab16.skiftrade.kz/templates/go/internal/repository/models"
	ucModels "gitlab16.skiftrade.kz/templates/go/internal/usecase/models"
)

type fakeTx struct{}

func (f *fakeTx) Begin(ctx context.Context) (pgx.Tx, error) { return nil, nil }
func (f *fakeTx) Commit(ctx context.Context) error          { return nil }
func (f *fakeTx) Rollback(ctx context.Context) error        { return nil }
func (f *fakeTx) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	return 0, nil
}
func (f *fakeTx) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults { return nil }
func (f *fakeTx) LargeObjects() pgx.LargeObjects                               { return pgx.LargeObjects{} }
func (f *fakeTx) Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error) {
	return nil, nil
}
func (f *fakeTx) Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}
func (f *fakeTx) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return nil, nil
}
func (f *fakeTx) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row { return nil }
func (f *fakeTx) QueryFunc(ctx context.Context, sql string, args []interface{}, scans []interface{}, fn func(pgx.Row) error) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}
func (f *fakeTx) Conn() *pgx.Conn { return nil }

type fakeRepoProfile struct {
	exists       map[string]repoModels.Profile
	wallets      map[string]repoModels.Wallet
	referrals    map[string][]repoModels.Referral
	transactions map[string][]repoModels.WalletTransaction
	prompts      map[string][]repoModels.PromptHistory
	nextPromptID int64
}

func (f *fakeRepoProfile) DBBeginTransaction(ctx context.Context) (pgx.Tx, error) {
	return &fakeTx{}, nil
}
func (f *fakeRepoProfile) CreateProfile(ctx context.Context, tx pgx.Tx, p repoModels.Profile) (int64, error) {
	f.exists[p.TelegramID] = p
	return 1, nil
}
func (f *fakeRepoProfile) GetProfileByTelegramID(ctx context.Context, tx pgx.Tx, telegramID string) (repoModels.Profile, error) {
	if p, ok := f.exists[telegramID]; ok {
		return p, nil
	}
	return repoModels.Profile{}, pgx.ErrNoRows
}
func (f *fakeRepoProfile) CreateWalletForUser(ctx context.Context, tx pgx.Tx, profileID int64) (int64, error) {
	return 1, nil
}
func (f *fakeRepoProfile) AddReferral(ctx context.Context, tx pgx.Tx, referrerProfileID int64, refereeProfileID int64) error {
	return nil
}
func (f *fakeRepoProfile) GetWalletByTelegramID(ctx context.Context, tx pgx.Tx, telegramID string) (repoModels.Wallet, error) {
	if w, ok := f.wallets[telegramID]; ok {
		return w, nil
	}
	if _, ok := f.exists[telegramID]; ok {
		return repoModels.Wallet{ProfileID: 1}, nil
	}
	return repoModels.Wallet{}, pgx.ErrNoRows
}
func (f *fakeRepoProfile) ListWalletTransactionsByTelegramID(ctx context.Context, tx pgx.Tx, telegramID string, limit int32) ([]repoModels.WalletTransaction, error) {
	if items, ok := f.transactions[telegramID]; ok {
		return items, nil
	}
	return []repoModels.WalletTransaction{}, nil
}
func (f *fakeRepoProfile) ListReferralsByTelegramID(ctx context.Context, tx pgx.Tx, telegramID string) ([]repoModels.Referral, error) {
	if items, ok := f.referrals[telegramID]; ok {
		return items, nil
	}
	return []repoModels.Referral{}, nil
}
func (f *fakeRepoProfile) CreatePromptHistory(ctx context.Context, tx pgx.Tx, item repoModels.PromptHistory) (int64, error) {
	f.nextPromptID++
	item.ID = f.nextPromptID
	f.prompts[item.TelegramID] = append([]repoModels.PromptHistory{item}, f.prompts[item.TelegramID]...)
	return item.ID, nil
}
func (f *fakeRepoProfile) ListPromptHistoryByTelegramID(ctx context.Context, tx pgx.Tx, telegramID string, limit int32) ([]repoModels.PromptHistory, error) {
	items := f.prompts[telegramID]
	if len(items) > int(limit) {
		return items[:limit], nil
	}
	return items, nil
}

var _ internal.Repository = (*fakeRepoProfile)(nil)

func TestRegisterByTelegram_CreatesProfile(t *testing.T) {
	repo := &fakeRepoProfile{exists: map[string]repoModels.Profile{}, wallets: map[string]repoModels.Wallet{}, referrals: map[string][]repoModels.Referral{}, transactions: map[string][]repoModels.WalletTransaction{}, prompts: map[string][]repoModels.PromptHistory{}}
	uc := NewUseCase(repo)

	values := url.Values{}
	values.Set("user", `{"id":123,"first_name":"Ivan","username":"ivan","photo_url":"","language_code":"ru"}`)
	initRaw := base64.StdEncoding.EncodeToString([]byte(values.Encode()))

	out, err := uc.RegisterByTelegram(context.Background(), ucModels.RegisterByTelegramInput{InitDataRaw: initRaw})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.ProfileID == 0 {
		t.Fatalf("expected profile id > 0")
	}
}

func TestRegisterByTelegram_AlreadyExists(t *testing.T) {
	repo := &fakeRepoProfile{exists: map[string]repoModels.Profile{"123": {TelegramID: "123"}}, wallets: map[string]repoModels.Wallet{}, referrals: map[string][]repoModels.Referral{}, transactions: map[string][]repoModels.WalletTransaction{}, prompts: map[string][]repoModels.PromptHistory{}}
	uc := NewUseCase(repo)

	values := url.Values{}
	values.Set("user", `{"id":123,"first_name":"Ivan","username":"ivan"}`)
	initRaw := base64.StdEncoding.EncodeToString([]byte(values.Encode()))

	_, err := uc.RegisterByTelegram(context.Background(), ucModels.RegisterByTelegramInput{InitDataRaw: initRaw})
	if !errors.Is(err, ucModels.ErrProfileAlreadyRegistered) {
		t.Fatalf("expected ErrProfileAlreadyRegistered, got %v", err)
	}
}

func TestGetUserByTelegramID_NotFound(t *testing.T) {
	repo := &fakeRepoProfile{exists: map[string]repoModels.Profile{}, wallets: map[string]repoModels.Wallet{}, referrals: map[string][]repoModels.Referral{}, transactions: map[string][]repoModels.WalletTransaction{}, prompts: map[string][]repoModels.PromptHistory{}}
	uc := NewUseCase(repo)
	_, err := uc.GetUserByTelegramID(context.Background(), "not-exists")
	if !errors.Is(err, ucModels.ErrProfileNotFound) {
		t.Fatalf("expected ErrProfileNotFound, got %v", err)
	}
}

func TestGetWalletByTelegramID_OK(t *testing.T) {
	repo := &fakeRepoProfile{
		exists:    map[string]repoModels.Profile{"123": {TelegramID: "123"}},
		wallets:   map[string]repoModels.Wallet{"123": {ID: 10, ProfileID: 1, Balance: 500, TotalEarned: 800, BalanceAvailable: 300}},
		referrals: map[string][]repoModels.Referral{},
		transactions: map[string][]repoModels.WalletTransaction{
			"123": {{ID: 1, Type: "deposit", Amount: 500, Status: "completed", Description: "Top up"}},
		},
		prompts: map[string][]repoModels.PromptHistory{},
	}
	uc := NewUseCase(repo)

	out, err := uc.GetWalletByTelegramID(context.Background(), "123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Wallet.Balance != 500 {
		t.Fatalf("unexpected balance: %d", out.Wallet.Balance)
	}
	if len(out.Transactions) != 1 {
		t.Fatalf("unexpected transactions count: %d", len(out.Transactions))
	}
}

func TestGetReferralsByTelegramID_OK(t *testing.T) {
	repo := &fakeRepoProfile{
		exists:  map[string]repoModels.Profile{"123": {TelegramID: "123"}},
		wallets: map[string]repoModels.Wallet{},
		referrals: map[string][]repoModels.Referral{
			"123": {{ID: 1, TelegramID: "456", Name: "Friend", Username: "friend", Earnings: 200}},
		},
		transactions: map[string][]repoModels.WalletTransaction{},
		prompts:      map[string][]repoModels.PromptHistory{},
	}
	uc := NewUseCase(repo)

	out, err := uc.GetReferralsByTelegramID(context.Background(), "123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Items) != 1 {
		t.Fatalf("unexpected referrals count: %d", len(out.Items))
	}
}

func TestSavePromptHistory_AndGetHistory_OK(t *testing.T) {
	repo := &fakeRepoProfile{
		exists:       map[string]repoModels.Profile{"123": {ID: 1, TelegramID: "123"}},
		wallets:      map[string]repoModels.Wallet{},
		referrals:    map[string][]repoModels.Referral{},
		transactions: map[string][]repoModels.WalletTransaction{},
		prompts:      map[string][]repoModels.PromptHistory{},
	}
	uc := NewUseCase(repo)

	saved, err := uc.SavePromptHistory(context.Background(), ucModels.SavePromptHistoryInput{
		TelegramID: "123",
		Prompt:     "Write a launch announcement",
		Category:   "marketing",
	})
	if err != nil {
		t.Fatalf("unexpected save error: %v", err)
	}
	if saved.Item.ID == 0 {
		t.Fatalf("expected saved prompt id > 0")
	}

	history, err := uc.GetPromptHistoryByTelegramID(context.Background(), "123")
	if err != nil {
		t.Fatalf("unexpected history error: %v", err)
	}
	if len(history.Items) != 1 {
		t.Fatalf("unexpected history count: %d", len(history.Items))
	}
	if history.Items[0].Prompt != "Write a launch announcement" {
		t.Fatalf("unexpected prompt: %s", history.Items[0].Prompt)
	}
}
