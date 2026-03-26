package bot

import (
	"context"
	"log/slog"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	api *tgbotapi.BotAPI
}

// New creates Telegram bot if TELEGRAM_BOT_TOKEN is configured.
func New() (*Bot, error) {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		return &Bot{}, nil
	}
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	return &Bot{api: api}, nil
}

// StartPolling starts handling /start command.
func (b *Bot) StartPolling(ctx context.Context) {
	if b.api == nil {
		return
	}
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30
	updates := b.api.GetUpdatesChan(u)
	for {
		select {
		case <-ctx.Done():
			return
		case upd := <-updates:
			if upd.Message != nil && upd.Message.IsCommand() && upd.Message.Command() == "start" {
				name := upd.Message.From.FirstName
				msg := tgbotapi.NewMessage(upd.Message.Chat.ID, "Добро пожаловать в Facebase, "+name+"💚\n– Получайте рекламные задания...")
				if _, err := b.api.Send(msg); err != nil {
					slog.ErrorContext(ctx, "failed to send telegram message", slog.Any("error", err))
				}
			}
		}
	}
}

// SendMessage sends a text with optional inline buttons.
func (b *Bot) SendMessage(telegramID int64, text string, buttons []tgbotapi.InlineKeyboardButton) error {
	if b.api == nil {
		return nil
	}
	msg := tgbotapi.NewMessage(telegramID, text)
	if len(buttons) > 0 {
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(buttons)
	}
	_, err := b.api.Send(msg)
	return err
}


