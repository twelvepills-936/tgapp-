package main

import (
	"context"
	"log/slog"

	"gitlab16.skiftrade.kz/templates/go/internal/bot"
	"gitlab16.skiftrade.kz/templates/go/internal/repository"
	repoModels "gitlab16.skiftrade.kz/templates/go/internal/repository/models"
	"gitlab16.skiftrade.kz/templates/go/internal/service"
	"gitlab16.skiftrade.kz/templates/go/internal/usecase"
	api "gitlab16.skiftrade.kz/templates/go/pkg/api"
	"gitlab16.skiftrade.kz/templates/go/pkg/app"
	"gitlab16.skiftrade.kz/templates/go/pkg/config"
	"gitlab16.skiftrade.kz/templates/go/pkg/logger"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	addConfig := config.LoadConfig()

	cfg := app.LoadConfigFromEnv()

	application, err := app.New(ctx, cfg)
	if err != nil {
		panic(err)
	}

	// Telegram bot: start polling in background if token is configured.
	b, err := bot.New()
	if err != nil {
		slog.WarnContext(ctx, "failed to init bot", logger.ErrorAttr(err))
	} else {
		go b.StartPolling(ctx)
	}

	pool, err := repository.NewPostgres(ctx, repoModels.ConfigPostgres(addConfig.Postgres))
	if err != nil {
		slog.ErrorContext(ctx, "failed to init postgres", logger.ErrorAttr(err))
		return
	}
	defer pool.Close()
	repo := repository.NewRepository(pool)

	uc := usecase.NewUseCase(repo)
	svc := service.NewService(uc)

	// Register gRPC services BEFORE starting the server
	api.RegisterCyberMateServer(application.GrpcServer, svc)

	err = application.Init(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to init app", logger.ErrorAttr(err))
		return
	}

	err = api.RegisterCyberMateHandler(ctx, application.ServeMux, application.GrpcConn)
	if err != nil {
		slog.ErrorContext(ctx, "failed to register cybermate handler", logger.ErrorAttr(err))
		return
	}

	err = application.Run(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to run app", logger.ErrorAttr(err))
		return
	}
}
