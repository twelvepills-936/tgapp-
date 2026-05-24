package generator

import "errors"

// ErrProvider indicates an upstream AI provider failure.
var ErrProvider = errors.New("provider error")

// ProviderError carries a human-readable message from Gemini/OpenAI/etc.
type ProviderError struct {
	Provider string
	Message  string
}

func (e *ProviderError) Error() string {
	return e.Provider + ": " + e.Message
}

func (e *ProviderError) Unwrap() error {
	return ErrProvider
}

func newProviderError(provider, message string) error {
	return &ProviderError{Provider: provider, Message: message}
}
