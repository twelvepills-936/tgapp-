package usecase

import (
	"context"
	"errors"

	repoModels "gitlab16.skiftrade.kz/templates/go/internal/repository/models"
	"gitlab16.skiftrade.kz/templates/go/internal/usecase/models"
)

func (uc *useCase) GetUser(ctx context.Context, input models.GetUserInput) (output models.GetUserOutput, err error) {
	err = input.Validate()
	if err != nil {
		return output, err
	}

	user, err := uc.repo.ReadUser(ctx, input.UserID, nil)
	if err != nil {
		if errors.Is(err, repoModels.ErrUserIsNotFound) {
			return output, models.ErrUserIsNotFound
		}
		return output, err
	}

	return models.GetUserOutput{Data: models.User{
		ID:      user.ID,
		Name:    user.Name,
		Surname: user.Surname,
	}}, nil
}
