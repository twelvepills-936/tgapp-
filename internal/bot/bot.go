package bot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	api           *tgbotapi.BotAPI
	webhookSecret string
}

// New creates Telegram bot when TELEGRAM_BOT_TOKEN is configured.
func New(cfg Config) (*Bot, error) {
	if cfg.Token == "" {
		return &Bot{}, nil
	}
	api, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		return nil, err
	}
	return &Bot{
		api:           api,
		webhookSecret: cfg.WebhookSecret,
	}, nil
}

// WebhookSecret returns the expected X-Telegram-Bot-Api-Secret-Token value.
func (b *Bot) WebhookSecret() string {
	if b == nil {
		return ""
	}
	return b.webhookSecret
}

// Enabled reports whether the bot API client is configured.
func (b *Bot) Enabled() bool {
	return b != nil && b.api != nil
}

// RegisterWebhook registers the bot webhook URL with Telegram.
func (b *Bot) RegisterWebhook(ctx context.Context, webhookURL string) error {
	if !b.Enabled() || webhookURL == "" {
		return nil
	}

	payload := map[string]any{
		"url":              webhookURL,
		"allowed_updates":  []string{"message", "callback_query"},
		"max_connections":  40,
		"drop_pending_updates": true,
	}
	if b.webhookSecret != "" {
		payload["secret_token"] = b.webhookSecret
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/setWebhook", b.api.Token)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result struct {
		OK          bool   `json:"ok"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}
	if !result.OK {
		if result.Description == "" {
			result.Description = resp.Status
		}
		return fmt.Errorf("setWebhook: %s", result.Description)
	}

	slog.InfoContext(ctx, "telegram webhook registered", slog.String("url", webhookURL))
	return nil
}

// StartPolling handles updates via long polling (local development).
func (b *Bot) StartPolling(ctx context.Context) {
	if !b.Enabled() {
		return
	}
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30
	updates := b.api.GetUpdatesChan(u)
	for {
		select {
		case <-ctx.Done():
			return
		case upd, ok := <-updates:
			if !ok {
				return
			}
			b.HandleUpdate(ctx, upd)
		}
	}
}

// HandleUpdate processes a single Telegram update.
func (b *Bot) HandleUpdate(ctx context.Context, upd tgbotapi.Update) {
	if !b.Enabled() {
		return
	}
	if upd.Message != nil && upd.Message.IsCommand() && upd.Message.Command() == "start" {
		name := upd.Message.From.FirstName
		msg := tgbotapi.NewMessage(upd.Message.Chat.ID, "Добро пожаловать в CyberMate, "+name+"💚\n– Получайте рекламные задания...")
		if _, err := b.api.Send(msg); err != nil {
			slog.ErrorContext(ctx, "failed to send telegram message", slog.Any("error", err))
		}
	}
}

// SendMessage sends a text with optional inline buttons.
func (b *Bot) SendMessage(telegramID int64, text string, buttons []tgbotapi.InlineKeyboardButton) error {
	if !b.Enabled() {
		return nil
	}
	msg := tgbotapi.NewMessage(telegramID, text)
	if len(buttons) > 0 {
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(buttons)
	}
	_, err := b.api.Send(msg)
	return err
}
