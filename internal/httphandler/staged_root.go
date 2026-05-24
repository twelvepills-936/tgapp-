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

func isHealthRequest(r *http.Request) bool {
	switch r.Method {
	case http.MethodGet, http.MethodHead:
	default:
		return false
	}
	path := r.URL.Path
	return path == "/health" || path == "/health/"
}

func (s *StagedRoot) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if isHealthRequest(r) {
		if r.Method == http.MethodHead {
			w.WriteHeader(http.StatusOK)
			return
		}
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
