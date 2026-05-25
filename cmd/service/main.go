package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"gitlab16.skiftrade.kz/templates/go/internal/bot"
	"gitlab16.skiftrade.kz/templates/go/internal/httphandler"
	"gitlab16.skiftrade.kz/templates/go/internal/migrate"
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

	cfg := app.Config{
		GRPCPort:     addConfig.App.GRPCPort,
		HTTPPort:     addConfig.App.HTTPPort,
		ReadTimeout:  addConfig.Server.ReadTimeout,
		WriteTimeout: addConfig.Server.WriteTimeout,
		IdleTimeout:  addConfig.Server.IdleTimeout,
		CORSOrigins:  addConfig.CORS.AllowedOrigins,
	}

	migrationsDir := migrate.ResolveDir()
	_, migrationsStatErr := os.Stat(migrationsDir)
	if p := os.Getenv("PORT"); p != "" && os.Getenv("APP_HTTP_PORT") != "" && p != os.Getenv("APP_HTTP_PORT") {
		slog.WarnContext(ctx, "APP_HTTP_PORT is ignored when PORT is set (use PORT on Railway)",
			slog.String("port", p),
			slog.String("app_http_port", os.Getenv("APP_HTTP_PORT")),
		)
	}
	slog.InfoContext(ctx, "starting CyberMate backend",
		slog.Int("http_port", cfg.HTTPPort),
		slog.String("port_env", os.Getenv("PORT")),
		slog.Int("grpc_port", cfg.GRPCPort),
		slog.String("environment", addConfig.App.Environment),
		slog.Bool("database_url_set", addConfig.Postgres.DatabaseURL != ""),
		slog.String("migrations_dir", migrationsDir),
		slog.Bool("migrations_dir_ok", migrationsStatErr == nil),
	)

	application, err := app.New(ctx, cfg)
	if err != nil {
		slog.ErrorContext(ctx, "failed to create app", logger.ErrorAttr(err))
		os.Exit(1)
	}

	staged := httphandler.NewStagedRoot()
	application.SetHTTPRootHandler(staged)

	if err := application.Init(ctx); err != nil {
		slog.ErrorContext(ctx, "failed to init app", logger.ErrorAttr(err))
		os.Exit(1)
	}

	listenErr := make(chan error, 1)
	go func() {
		if err := application.Run(ctx); err != nil {
			listenErr <- err
		}
	}()

	listenCtx, listenCancel := context.WithTimeout(ctx, 15*time.Second)
	if err := application.WaitHTTPListening(listenCtx); err != nil {
		slog.ErrorContext(ctx, "HTTP server did not bind in time", logger.ErrorAttr(err))
		os.Exit(1)
	}
	listenCancel()
	slog.InfoContext(ctx, "HTTP port open, continuing startup")

	pool, err := repository.NewPostgres(ctx, repoModels.ConfigPostgres(addConfig.Postgres))
	if err != nil {
		slog.ErrorContext(ctx, "failed to init postgres", logger.ErrorAttr(err))
		os.Exit(1)
	}
	defer pool.Close()

	if err := migrate.Apply(ctx, pool, migrationsDir); err != nil {
		slog.ErrorContext(ctx, "failed to apply database migrations",
			logger.ErrorAttr(err),
			slog.String("dir", migrationsDir),
		)
		os.Exit(1)
	}
	slog.InfoContext(ctx, "database migrations applied (HTTP server keeps running)",
		slog.String("dir", migrationsDir),
	)

	repo := repository.NewRepository(pool)

	modelRouter := generator.NewModelRouter(addConfig.Yandex, addConfig.Gemini, addConfig.AI)
	imageGen := generator.NewImageRouter(addConfig.Yandex, addConfig.Gemini)
	uc := usecase.NewUseCase(repo, modelRouter, imageGen, usecase.UseCaseOptions{
		SkipRegistrationCheck: config.SkipRegistrationCheck(),
		SkipAIWalletCheck:     config.SkipAIWalletCheck(),
	})
	svc := service.NewService(uc)

	api.RegisterCyberMateServer(application.GrpcServer, svc)

	if err := api.RegisterCyberMateHandler(ctx, application.ServeMux, application.GrpcConn); err != nil {
		slog.ErrorContext(ctx, "failed to register cybermate handler", logger.ErrorAttr(err))
		os.Exit(1)
	}

	rootMux := http.NewServeMux()
	httphandler.NewProfileRESTHandler(uc).RegisterRoutes(rootMux)
	rootMux.HandleFunc("POST /v1/generate/text", httphandler.NewGenerateTextHandler(uc).ServeHTTP)
	rootMux.HandleFunc("POST /v1/generate/image", httphandler.NewGenerateImageHandler(uc).ServeHTTP)
	tgWebhook := httphandler.NewTelegramWebhookSlot()
	rootMux.Handle("/v1/telegram/webhook", tgWebhook)
	rootMux.Handle("/", application.ServeMux)
	staged.SetReady(httphandler.NormalizePath(rootMux))

	slog.InfoContext(ctx, "API routes ready")

	tgCfg := bot.LoadConfig()
	go initTelegramBot(ctx, tgCfg, tgWebhook)

	select {
	case err := <-listenErr:
		if err != nil {
			slog.ErrorContext(ctx, "HTTP server stopped", logger.ErrorAttr(err))
			os.Exit(1)
		}
	}
}

func initTelegramBot(ctx context.Context, tgCfg bot.Config, slot *httphandler.TelegramWebhookSlot) {
	tgBot, err := bot.New(tgCfg)
	if err != nil {
		slog.WarnContext(ctx, "failed to init bot", logger.ErrorAttr(err))
		return
	}
	slot.Set(tgBot)

	if !tgBot.Enabled() {
		return
	}
	if tgCfg.UseWebhook() {
		if err := tgBot.RegisterWebhook(ctx, tgCfg.WebhookURL); err != nil {
			slog.ErrorContext(ctx, "failed to register telegram webhook", logger.ErrorAttr(err))
		}
		return
	}
	go tgBot.StartPolling(ctx)
}
