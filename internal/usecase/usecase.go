package usecase

import (
	"gitlab16.skiftrade.kz/templates/go/internal"
	"gitlab16.skiftrade.kz/templates/go/pkg/config"
	"gitlab16.skiftrade.kz/templates/go/pkg/generator"
)

// UseCaseOptions tunes use case behaviour.
type UseCaseOptions struct {
	SkipRegistrationCheck bool
	SkipAIWalletCheck     bool
}

// useCase implements internal.UseCase.
type useCase struct {
	repo                  internal.Repository
	modelRouter           generator.Generator
	imageGenerator        generator.ImageGenerator
	skipRegistrationCheck bool
	skipAIWalletCheck     bool
}

// NewUseCase wires repository layer into business logic.
func NewUseCase(repo internal.Repository, modelRouter generator.Generator, imageGenerator generator.ImageGenerator, opts UseCaseOptions) internal.UseCase {
	if modelRouter == nil {
		modelRouter = generator.NewModelRouter(config.ConfigYandex{}, config.ConfigGemini{}, config.ConfigAI{})
	}
	return &useCase{
		repo:                  repo,
		modelRouter:           modelRouter,
		imageGenerator:        imageGenerator,
		skipRegistrationCheck: opts.SkipRegistrationCheck,
		skipAIWalletCheck:     opts.SkipAIWalletCheck,
	}
}
