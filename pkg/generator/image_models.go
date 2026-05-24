package generator

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

const (
	// ModelNanoBanana is the app slug for Gemini native image generation.
	ModelNanoBanana = "nano-banana"
	// ModelAliceAIArt is the app slug for Yandex Alice AI ART image generation.
	ModelAliceAIArt = "alice-ai-art"
)

// DefaultImageModel is used when the client omits model on image requests.
const DefaultImageModel = ModelNanoBanana

// ImageTokenCost is the in-app token price per image generation request.
var ImageTokenCost = map[string]int64{
	ModelNanoBanana:  3,
	ModelAliceAIArt: 4,
}

// NormalizeImageModel validates and normalizes the image model slug.
func NormalizeImageModel(model string) (string, error) {
	m := strings.TrimSpace(strings.ToLower(model))
	if m == "" {
		return DefaultImageModel, nil
	}
	switch m {
	case ModelNanoBanana, ModelAliceAIArt:
		return m, nil
	case "aliceai-image-art-3.0", "aliceai-image-art-3.0/latest":
		return ModelAliceAIArt, nil
	default:
		return "", fmt.Errorf("%w: %q", ErrUnsupportedModel, model)
	}
}

// ImageTokenCostFor returns billing cost for an image model slug.
func ImageTokenCostFor(model string) int64 {
	if c, ok := ImageTokenCost[model]; ok {
		return c
	}
	return 3
}

// ImageGenerator generates images from a text prompt (optional multi-turn context in Messages).
type ImageGenerator interface {
	GenerateImage(ctx context.Context, model string, in ImageGenerateInput) (ImageResult, error)
}

// ErrImageGeneratorUnavailable is returned when image generation is not configured.
var ErrImageGeneratorUnavailable = errors.New("image generator is not configured")
