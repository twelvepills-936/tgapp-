package main

import (
	"context"
	"log/slog"
	"os"

	"gitlab16.skiftrade.kz/templates/go/internal/migrate"
	"gitlab16.skiftrade.kz/templates/go/internal/repository"
	repoModels "gitlab16.skiftrade.kz/templates/go/internal/repository/models"
	"gitlab16.skiftrade.kz/templates/go/pkg/config"
	"gitlab16.skiftrade.kz/templates/go/pkg/logger"
)

func main() {
	ctx := context.Background()
	cfg := config.LoadConfig()

	pool, err := repository.NewPostgres(ctx, repoModels.ConfigPostgres(cfg.Postgres))
	if err != nil {
		slog.ErrorContext(ctx, "postgres connect failed", logger.ErrorAttr(err))
		os.Exit(1)
	}
	defer pool.Close()

	dir := migrate.ResolveDir()
	if err := migrate.Apply(ctx, pool, dir); err != nil {
		slog.ErrorContext(ctx, "migrate failed", logger.ErrorAttr(err), slog.String("dir", dir))
		os.Exit(1)
	}

	slog.InfoContext(ctx, "migrations complete", slog.String("dir", dir))
}
