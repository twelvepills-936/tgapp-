package httphandler

import (
	"net/http"
	"sync"
)

// TelegramWebhookSlot registers the webhook route before the bot client is ready.
type TelegramWebhookSlot struct {
	mu sync.RWMutex
	h  *TelegramWebhookHandler
}

func NewTelegramWebhookSlot() *TelegramWebhookSlot {
	return &TelegramWebhookSlot{h: NewTelegramWebhookHandler(nil)}
}

func (s *TelegramWebhookSlot) Set(bot TelegramUpdateProcessor) {
	s.mu.Lock()
	s.h = NewTelegramWebhookHandler(bot)
	s.mu.Unlock()
}

func (s *TelegramWebhookSlot) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	h := s.h
	s.mu.RUnlock()
	h.ServeHTTP(w, r)
}
