package httphandler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type fakeTelegramBot struct {
	secret  string
	enabled bool
	called  bool
}

func (f *fakeTelegramBot) Enabled() bool              { return f.enabled }
func (f *fakeTelegramBot) WebhookSecret() string      { return f.secret }
func (f *fakeTelegramBot) HandleUpdate(context.Context, tgbotapi.Update) {
	f.called = true
}

func TestTelegramWebhookHandler_OK(t *testing.T) {
	fake := &fakeTelegramBot{enabled: true, secret: "s3cret"}
	h := &TelegramWebhookHandler{bot: fake}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, telegramWebhookPath, bytes.NewReader([]byte(`{"update_id":1}`)))
	req.Header.Set("X-Telegram-Bot-Api-Secret-Token", "s3cret")
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if !fake.called {
		t.Fatal("expected HandleUpdate to be called")
	}
}

func TestTelegramWebhookHandler_ForbiddenOnBadSecret(t *testing.T) {
	fake := &fakeTelegramBot{enabled: true, secret: "s3cret"}
	h := &TelegramWebhookHandler{bot: fake}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, telegramWebhookPath, bytes.NewReader([]byte(`{}`)))
	req.Header.Set("X-Telegram-Bot-Api-Secret-Token", "wrong")
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestTelegramWebhookHandler_ServiceUnavailable(t *testing.T) {
	h := NewTelegramWebhookHandler(nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, telegramWebhookPath, bytes.NewReader([]byte(`{}`)))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestTelegramWebhookHandler_GET_OK(t *testing.T) {
	h := NewTelegramWebhookHandler(nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, telegramWebhookPath, nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestTelegramWebhookHandler_MethodNotAllowed(t *testing.T) {
	h := NewTelegramWebhookHandler(nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, telegramWebhookPath, nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d", rec.Code)
	}
}
