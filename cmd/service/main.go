package main

import (
	"context"
	"log/slog"

	"net/http"

	"gitlab16.skiftrade.kz/templates/go/internal/bot"
	"gitlab16.skiftrade.kz/templates/go/internal/httphandler"
	"gitlab16.skiftrade.kz/templates/go/internal/repository"
	repoModels "gitlab16.skiftrade.kz/templates/go/internal/repository/models"
	"gitlab16.skiftrade.kz/templates/go/internal/service"
	"gitlab16.skiftrade.kz/templates/go/internal/usecase"
	api "gitlab16.skiftrade.kz/templates/go/pkg/api"
	"gitlab16.skiftrade.kz/templates/go/pkg/app"
	"gitlab16.skiftrade.kz/templates/go/pkg/config"
	"gitlab16.skiftrade.kz/templates/go/pkg/generator"
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

	tgCfg := bot.LoadConfig()
	tgBot, err := bot.New(tgCfg)
	if err != nil {
		slog.WarnContext(ctx, "failed to init bot", logger.ErrorAttr(err))
	} else if tgBot.Enabled() {
		if tgCfg.UseWebhook() {
			if err := tgBot.RegisterWebhook(ctx, tgCfg.WebhookURL); err != nil {
				slog.ErrorContext(ctx, "failed to register telegram webhook", logger.ErrorAttr(err))
			}
		} else {
			go tgBot.StartPolling(ctx)
		}
	}

	pool, err := repository.NewPostgres(ctx, repoModels.ConfigPostgres(addConfig.Postgres))
	if err != nil {
		slog.ErrorContext(ctx, "failed to init postgres", logger.ErrorAttr(err))
		return
	}
	defer pool.Close()
	repo := repository.NewRepository(pool)

	modelRouter := generator.NewModelRouter(addConfig.Yandex, addConfig.Gemini, addConfig.AI)
	imageGen := generator.NewNanoBananaClient(addConfig.Gemini)
	uc := usecase.NewUseCase(repo, modelRouter, imageGen, usecase.UseCaseOptions{
		SkipRegistrationCheck: config.SkipRegistrationCheck(),
	})
	svc := service.NewService(uc)

	// Register gRPC services BEFORE starting the server
	api.RegisterCyberMateServer(application.GrpcServer, svc)

	err = api.RegisterCyberMateHandler(ctx, application.ServeMux, application.GrpcConn)
	if err != nil {
		slog.ErrorContext(ctx, "failed to register cybermate handler", logger.ErrorAttr(err))
		return
	}

	rootMux := http.NewServeMux()
	rootMux.Handle("/v1/generate/text", httphandler.NewGenerateTextHandler(uc))
	rootMux.Handle("/v1/generate/image", httphandler.NewGenerateImageHandler(uc))
	rootMux.Handle("/v1/telegram/webhook", httphandler.NewTelegramWebhookHandler(tgBot))
	rootMux.Handle("/", application.ServeMux)
	application.SetHTTPRootHandler(httphandler.NormalizePath(rootMux))

	err = application.Init(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to init app", logger.ErrorAttr(err))
		return
	}

	err = application.Run(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to run app", logger.ErrorAttr(err))
		return
	}
}
