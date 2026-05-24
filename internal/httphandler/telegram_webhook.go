package httphandler

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const telegramWebhookPath = "/v1/telegram/webhook"

// TelegramUpdateProcessor handles incoming Telegram webhook updates.
type TelegramUpdateProcessor interface {
	Enabled() bool
	WebhookSecret() string
	HandleUpdate(ctx context.Context, upd tgbotapi.Update)
}

// TelegramWebhookHandler serves POST /v1/telegram/webhook.
type TelegramWebhookHandler struct {
	bot TelegramUpdateProcessor
}

func NewTelegramWebhookHandler(b TelegramUpdateProcessor) *TelegramWebhookHandler {
	return &TelegramWebhookHandler{bot: b}
}

func (h *TelegramWebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet, http.MethodHead:
		// Health check / mistaken Mini App URL — do not return 405 plain text.
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true,"endpoint":"telegram-webhook","usage":"POST only"}`))
		return
	case http.MethodPost:
		// continue below
	default:
		w.Header().Set("Allow", "POST, GET, HEAD")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if h.bot == nil || !h.bot.Enabled() {
		http.Error(w, "telegram bot is not configured", http.StatusServiceUnavailable)
		return
	}

	secret := h.bot.WebhookSecret()
	if secret != "" {
		got := strings.TrimSpace(r.Header.Get("X-Telegram-Bot-Api-Secret-Token"))
		if got != secret {
			slog.WarnContext(r.Context(), "telegram webhook: invalid secret token")
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
	}

	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	var upd tgbotapi.Update
	if err := json.Unmarshal(body, &upd); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	h.bot.HandleUpdate(r.Context(), upd)
	w.WriteHeader(http.StatusOK)
}
