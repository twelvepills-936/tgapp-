package repository

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"strconv"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"gitlab16.skiftrade.kz/templates/go/internal"
	"gitlab16.skiftrade.kz/templates/go/internal/repository/models"
	"gitlab16.skiftrade.kz/templates/go/pkg/logger"
)

type Repository struct {
	db      *pgxpool.Pool
	builder sq.StatementBuilderType
}

// Queryable interface for executing queries on both pool and transaction
type Queryable interface {
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
}

// NewRepository creates a postgres-backed repository.
func NewRepository(db *pgxpool.Pool) internal.Repository {
	return &Repository{
		db:      db,
		builder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Repository) DBBeginTransaction(ctx context.Context) (pgx.Tx, error) {
	return r.db.Begin(ctx)
}

// getQueryable returns the appropriate queryable interface (transaction or pool)
func (r *Repository) getQueryable(tx pgx.Tx) Queryable {
	if tx != nil {
		return tx
	}
	return r.db
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// NewPostgres initializes pgx pool with telemetry.
func NewPostgres(ctx context.Context, c models.ConfigPostgres) (*pgxpool.Pool, error) {
	dsn := getDSN(c)

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		slog.ErrorContext(ctx, "failed to open PostgresDB connection", logger.ErrorAttr(err))
		return nil, fmt.Errorf("parse config: %w", err)
	}

	cfg.ConnConfig.Tracer = otelpgx.NewTracer()

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	// Проверка коннекта
	if err = pool.Ping(ctx); err != nil {
		slog.ErrorContext(ctx, "failed to ping for PostgresDB", logger.ErrorAttr(err))
		return nil, err
	}

	if err = otelpgx.RecordStats(pool); err != nil {
		return nil, fmt.Errorf("unable to record database stats: %w", err)
	}

	return pool, nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func getDSN(cfg models.ConfigPostgres) string {
	if cfg.DatabaseURL != "" {
		return mergeDSNPoolParams(cfg.DatabaseURL, cfg)
	}

	u := url.URL{
		Scheme: "postgresql",
		User:   url.UserPassword(cfg.User, cfg.Pass),
		Host:   fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Path:   cfg.DBName,
	}

	q := u.Query()
	q.Set("sslmode", cfg.SSLMode)
	q.Set("pool_max_conns", strconv.FormatInt(cfg.PoolMaxConns, 10))
	q.Set("pool_min_conns", strconv.FormatInt(cfg.PoolMinConns, 10))
	q.Set("pool_max_conn_lifetime", cfg.PoolMaxConnLifeTime.String())
	q.Set("pool_max_conn_idle_time", cfg.PoolMaxConnIdleTime.String())
	q.Set("pool_health_check_period", cfg.PoolHealthCheckPeriod.String())

	if cfg.SSLRootCert != "" {
		q.Set("sslrootcert", cfg.SSLRootCert)
	}

	u.RawQuery = q.Encode()

	return u.String()
}

func mergeDSNPoolParams(dsn string, cfg models.ConfigPostgres) string {
	parsed, err := url.Parse(dsn)
	if err != nil {
		return dsn
	}

	q := parsed.Query()
	if q.Get("sslmode") == "" {
		ssl := cfg.SSLMode
		if ssl == "" {
			ssl = "require"
			if strings.Contains(parsed.Host, ".railway.internal") {
				ssl = "disable"
			}
		}
		q.Set("sslmode", ssl)
	}
	q.Set("pool_max_conns", strconv.FormatInt(cfg.PoolMaxConns, 10))
	q.Set("pool_min_conns", strconv.FormatInt(cfg.PoolMinConns, 10))
	q.Set("pool_max_conn_lifetime", cfg.PoolMaxConnLifeTime.String())
	q.Set("pool_max_conn_idle_time", cfg.PoolMaxConnIdleTime.String())
	q.Set("pool_health_check_period", cfg.PoolHealthCheckPeriod.String())
	if cfg.SSLRootCert != "" {
		q.Set("sslrootcert", cfg.SSLRootCert)
	}
	parsed.RawQuery = q.Encode()
	return parsed.String()
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
