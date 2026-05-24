package generator

import "context"

// Generator generates text for a selected model slug.
type Generator interface {
	Generate(ctx context.Context, model string, in TextGenerateInput) (Result, error)
}
