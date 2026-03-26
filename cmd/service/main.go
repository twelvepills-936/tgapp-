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
	ctx, c := context.WithCancel(context.Background())
	defer c()

	cfg := app.LoadConfigFromEnv()

	application, err := app.New(ctx, cfg)
	if err != nil {
		panic(err)
	}

	addConfig := config.LoadConfig()

	// Telegram bot initialization
	// Bot is created but polling is not started here to avoid blocking the service.
	// If needed, run bot.StartPolling(ctx) in a separate goroutine.
	_, err = bot.New()
	if err != nil {
		slog.WarnContext(ctx, "failed to init bot", logger.ErrorAttr(err))
		// Continue without bot - it's not critical for service startup
	}

	pool, err := repository.NewPostgres(ctx, repoModels.ConfigPostgres(addConfig.Postgres))
	if err != nil {
		slog.ErrorContext(ctx, "failed to init postgres", logger.ErrorAttr(err))
		return
	}
	defer pool.Close()
	repo := repository.NewRepository(pool)

	// Create single instances of usecase and service
	uc := usecase.NewUseCase(repo)
	svc := service.NewService(uc)

	// Register gRPC services BEFORE starting the server
	api.RegisterUsersServer(application.GrpcServer, svc)
	api.RegisterFacebaseServer(application.GrpcServer, svc)

	err = application.Init(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to init app", logger.ErrorAttr(err))
		return
	}
	err = api.RegisterUsersHandler(ctx, application.ServeMux, application.GrpcConn)
	if err != nil {
		slog.ErrorContext(ctx, "failed to register users handler", logger.ErrorAttr(err))
		return
	}

	err = api.RegisterFacebaseHandler(ctx, application.ServeMux, application.GrpcConn)
	if err != nil {
		slog.ErrorContext(ctx, "failed to register facebase handler", logger.ErrorAttr(err))
		return
	}

	err = application.Run(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to run app", logger.ErrorAttr(err))
		return
	}
}
