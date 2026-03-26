package usecase

import (
	"gitlab16.skiftrade.kz/templates/go/internal"
)

// useCase implements internal.UseCase.
type useCase struct {
	repo internal.Repository
}

// NewUseCase wires repository layer into business logic.
func NewUseCase(
	repo internal.Repository) internal.UseCase {
	return &useCase{
		repo: repo,
	}
}
