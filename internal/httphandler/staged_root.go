package httphandler

import (
	"net/http"
	"sync"
)

// StagedRoot serves GET /health immediately while the full handler tree is still wiring up.
type StagedRoot struct {
	mu    sync.RWMutex
	ready http.Handler
}

func NewStagedRoot() *StagedRoot {
	return &StagedRoot{}
}

func (s *StagedRoot) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet && r.URL.Path == "/health" {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		return
	}

	s.mu.RLock()
	h := s.ready
	s.mu.RUnlock()
	if h == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "starting"})
		return
	}
	h.ServeHTTP(w, r)
}

func (s *StagedRoot) SetReady(h http.Handler) {
	s.mu.Lock()
	s.ready = h
	s.mu.Unlock()
}
