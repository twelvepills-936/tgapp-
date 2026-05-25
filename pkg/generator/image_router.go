package generator

import (
	"context"
	"fmt"

	"gitlab16.skiftrade.kz/templates/go/pkg/config"
)

// ImageRouter dispatches image generation by model slug.
type ImageRouter struct {
	nanoBanana ImageGenerator
	aliceArt   ImageGenerator
}

// NewImageRouter wires image providers from configuration.
func NewImageRouter(yandex config.ConfigYandex, gemini config.ConfigGemini) ImageGenerator {
	return &ImageRouter{
		nanoBanana: NewNanoBananaClient(gemini),
		aliceArt:   NewAliceAIArtClient(yandex),
	}
}

func (r *ImageRouter) GenerateImage(ctx context.Context, model string, in ImageGenerateInput) (ImageResult, error) {
	slug, err := NormalizeImageModel(model)
	if err != nil {
		return ImageResult{}, err
	}

	switch slug {
	case ModelNanoBanana:
		if r.nanoBanana == nil {
			return ImageResult{}, newProviderError("nano-banana", "WAVESPEED_API_KEY or GEMINI_API_KEY is not configured")
		}
		return r.nanoBanana.GenerateImage(ctx, slug, in)
	case ModelAliceAIArt:
		if r.aliceArt == nil {
			return ImageResult{}, newProviderError("alice-ai-art", "YANDEX_GPT_API_KEY and YANDEX_GPT_FOLDER_ID are not configured")
		}
		return r.aliceArt.GenerateImage(ctx, slug, in)
	default:
		return ImageResult{}, fmt.Errorf("%w: %q", ErrUnsupportedModel, model)
	}
}
