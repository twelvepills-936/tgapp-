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
func (f *fakeTx) Commit(ctx context.Context) error { return nil }
func (f *fakeTx) Rollback(ctx context.Context) error { return nil }
func (f *fakeTx) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) { return 0, nil }
func (f *fakeTx) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults { return nil }
func (f *fakeTx) LargeObjects() pgx.LargeObjects { return pgx.LargeObjects{} }
func (f *fakeTx) Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error) { return nil, nil }
func (f *fakeTx) Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) { return pgconn.CommandTag{}, nil }
func (f *fakeTx) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) { return nil, nil }
func (f *fakeTx) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row { return nil }
func (f *fakeTx) QueryFunc(ctx context.Context, sql string, args []interface{}, scans []interface{}, fn func(pgx.Row) error) (pgconn.CommandTag, error) { return pgconn.CommandTag{}, nil }
func (f *fakeTx) Conn() *pgx.Conn { return nil }

type fakeRepoProfile struct{
    exists map[string]repoModels.Profile
}

func (f *fakeRepoProfile) DBBeginTransaction(ctx context.Context) (pgx.Tx, error) { return &fakeTx{}, nil }
func (f *fakeRepoProfile) ReadUser(ctx context.Context, id int64, dbTx pgx.Tx) (repoModels.User, error) { return repoModels.User{}, nil }
func (f *fakeRepoProfile) CreateProfile(ctx context.Context, tx pgx.Tx, p repoModels.Profile) (int64, error) { f.exists[p.TelegramID] = p; return 1, nil }
func (f *fakeRepoProfile) GetProfileByTelegramID(ctx context.Context, tx pgx.Tx, telegramID string) (repoModels.Profile, error) {
    if p, ok := f.exists[telegramID]; ok { return p, nil }
    return repoModels.Profile{}, pgx.ErrNoRows
}
func (f *fakeRepoProfile) CreateWalletForUser(ctx context.Context, tx pgx.Tx, profileID int64) (int64, error) { return 1, nil }
func (f *fakeRepoProfile) AddReferral(ctx context.Context, tx pgx.Tx, referrerProfileID int64, refereeProfileID int64) error { return nil }

func TestRegisterByTelegram_CreatesProfile(t *testing.T) {
    var _ internal.Repository = (*fakeRepoProfile)(nil)
    repo := &fakeRepoProfile{exists: map[string]repoModels.Profile{}}
    uc := NewUseCase(repo)

    values := url.Values{}
    values.Set("user", `{"id":123,"first_name":"Ivan","username":"ivan","photo_url":"","language_code":"ru"}`)
    initRaw := base64.StdEncoding.EncodeToString([]byte(values.Encode()))

    out, err := uc.RegisterByTelegram(context.Background(), ucModels.RegisterByTelegramInput{InitDataRaw: initRaw})
    if err != nil { t.Fatalf("unexpected error: %v", err) }
    if out.ProfileID == 0 { t.Fatalf("expected profile id > 0") }
}

func TestRegisterByTelegram_AlreadyExists(t *testing.T) {
    repo := &fakeRepoProfile{exists: map[string]repoModels.Profile{"123":{TelegramID:"123"}}}
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
    repo := &fakeRepoProfile{exists: map[string]repoModels.Profile{}}
    uc := NewUseCase(repo)
    _, err := uc.GetUserByTelegramID(context.Background(), "not-exists")
    if !errors.Is(err, ucModels.ErrProfileNotFound) {
        t.Fatalf("expected ErrProfileNotFound, got %v", err)
    }
}


