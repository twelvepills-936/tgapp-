package generator

import (
	"errors"
	"fmt"
	"strings"
)

// ErrUnsupportedModel is returned for unknown model slugs.
var ErrUnsupportedModel = errors.New("unsupported model")

const (
	ModelYandexGPT   = "yandexgpt"
	ModelGeminiFlash = "gemini-flash"
	ModelOpenAI      = "openai"
)

// DefaultModel is used when the client omits model.
const DefaultModel = ModelYandexGPT

// TokenCost is the in-app token price per generation request.
var TokenCost = map[string]int64{
	ModelYandexGPT:   1,
	ModelGeminiFlash: 1,
	ModelOpenAI:      2,
}

// NormalizeModel validates and normalizes the model slug.
func NormalizeModel(model string) (string, error) {
	m := strings.TrimSpace(strings.ToLower(model))
	if m == "" {
		return DefaultModel, nil
	}
	switch m {
	case ModelYandexGPT, ModelGeminiFlash, ModelOpenAI:
		return m, nil
	default:
		return "", fmt.Errorf("%w: %q", ErrUnsupportedModel, model)
	}
}

// TokenCostFor returns billing cost for a model slug.
func TokenCostFor(model string) int64 {
	if c, ok := TokenCost[model]; ok {
		return c
	}
	return 1
}
