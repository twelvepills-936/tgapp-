package usecase

import (
	"context"
	"os"
	"strings"

	"gitlab16.skiftrade.kz/templates/go/pkg/generator"
)

const imageDirectorSystem = "Ты арт-директор для генерации изображений. " +
	"По запросу пользователя составь одно связное ТЗ на русском языке (5–8 предложений): " +
	"сюжет и главный объект, композиция и ракурс, источник света и атмосфера, палитра, художественный стиль, настроение. " +
	"Пиши только текст для нейросети-рисовальщика, без markdown, списков и вступлений."

// enrichImagePrompt expands a draft prompt via YandexGPT (art director). Falls back to draft on error.
func (uc *useCase) enrichImagePrompt(ctx context.Context, draft string) string {
	if uc.modelRouter == nil {
		return draft
	}
	if strings.EqualFold(strings.TrimSpace(os.Getenv("SKIP_IMAGE_PROMPT_ENRICH")), "true") {
		return draft
	}

	draft = strings.TrimSpace(draft)
	if draft == "" {
		return draft
	}

	result, err := uc.modelRouter.Generate(ctx, generator.ModelYandexGPT, generator.TextGenerateInput{
		Prompt:   imageDirectorSystem + "\n\nЗапрос:\n" + draft,
		Category: "image",
	})
	if err != nil {
		return draft
	}

	expanded := strings.TrimSpace(result.Text)
	if expanded == "" || len(expanded) < 40 {
		return draft
	}
	return expanded
}
