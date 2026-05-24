package generator

import (
	"context"
	"fmt"
	"strings"

	"gitlab16.skiftrade.kz/templates/go/pkg/config"
)

// ModelRouter dispatches generation to a provider by model slug.
type ModelRouter struct {
	yandex   *yandexClient
	deepseek *deepSeekClient
	gemini   *geminiClient
	openai   *openAIClient
}

// NewModelRouter wires providers from configuration (nil if not configured).
func NewModelRouter(yandex config.ConfigYandex, gemini config.ConfigGemini, openAI config.ConfigAI) *ModelRouter {
	r := &ModelRouter{}
	if yandex.APIKey != "" && yandex.FolderID != "" {
		r.yandex = newYandexClient(yandex)
		if strings.TrimSpace(yandex.DeepSeekModel) != "" {
			r.deepseek = newDeepSeekClient(yandex)
		}
	}
	if gemini.APIKey != "" {
		r.gemini = newGeminiClient(gemini)
	}
	if openAI.APIKey != "" {
		r.openai = newOpenAIClient(openAI)
	}
	return r
}

// Generate runs the selected model provider.
func (r *ModelRouter) Generate(ctx context.Context, model string, in TextGenerateInput) (Result, error) {
	slug, err := NormalizeModel(model)
	if err != nil {
		return Result{}, err
	}

	switch slug {
	case ModelYandexGPT:
		if r.yandex == nil {
			return Result{}, newProviderError("yandexgpt", "YANDEX_GPT_API_KEY and YANDEX_GPT_FOLDER_ID are not configured")
		}
		return r.yandex.Generate(ctx, in)
	case ModelDeepSeek:
		if r.deepseek == nil {
			return Result{}, newProviderError("deepseek", "YANDEX_GPT_API_KEY, YANDEX_GPT_FOLDER_ID and YANDEX_DEEPSEEK_MODEL are not configured")
		}
		return r.deepseek.Generate(ctx, in)
	case ModelGeminiFlash:
		if r.gemini == nil {
			return Result{}, newProviderError("gemini", "GEMINI_API_KEY is not configured")
		}
		return r.gemini.Generate(ctx, in)
	case ModelOpenAI:
		if r.openai == nil {
			return Result{}, newProviderError("openai", "OPENAI_API_KEY is not configured")
		}
		return r.openai.Generate(ctx, in)
	default:
		return Result{}, fmt.Errorf("unsupported model %q", slug)
	}
}
