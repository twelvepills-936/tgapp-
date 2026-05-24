package generator

import (
	"errors"
	"strings"
)

// ErrContentPolicy is returned when the provider refuses the prompt (safety / moderation).
var ErrContentPolicy = errors.New("content policy")

// IsContentPolicy reports provider refusals (e.g. Alice AI ART moderation).
func IsContentPolicy(err error) bool {
	if errors.Is(err, ErrContentPolicy) {
		return true
	}
	var pe *ProviderError
	if !errors.As(err, &pe) {
		return false
	}
	m := strings.ToLower(pe.Message)
	return strings.Contains(m, "не могу сгенерировать") ||
		strings.Contains(m, "не могу создать") ||
		strings.Contains(m, "другую тему") ||
		strings.Contains(m, "cannot generate") ||
		strings.Contains(m, "content policy") ||
		strings.Contains(m, "safety")
}
