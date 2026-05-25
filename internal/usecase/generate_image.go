package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	repo "gitlab16.skiftrade.kz/templates/go/internal/repository"
	ucModels "gitlab16.skiftrade.kz/templates/go/internal/usecase/models"
	"gitlab16.skiftrade.kz/templates/go/pkg/generator"
)

func (uc *useCase) GenerateImage(ctx context.Context, input ucModels.GenerateImageInput) (ucModels.GenerateImageOutput, error) {
	if uc.imageGenerator == nil {
		return ucModels.GenerateImageOutput{}, generator.ErrImageGeneratorUnavailable
	}

	requireTelegramID := !uc.skipRegistrationCheck
	if uc.skipRegistrationCheck && input.TelegramID == "" {
		input.TelegramID = "dev"
	}
	input = normalizeGenerateImageInput(input)
	if err := input.Validate(requireTelegramID); err != nil {
		return ucModels.GenerateImageOutput{}, err
	}

	model, err := generator.NormalizeImageModel(input.Model)
	if err != nil {
		return ucModels.GenerateImageOutput{}, mapGeneratorError(err)
	}

	var profileID int64
	if !uc.skipRegistrationCheck {
		profile, err := uc.repo.GetProfileByTelegramID(ctx, nil, input.TelegramID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ucModels.GenerateImageOutput{}, ucModels.ErrProfileNotFound
			}
			return ucModels.GenerateImageOutput{}, err
		}
		profileID = profile.ID

		if !uc.skipAIWalletCheck {
			wallet, err := uc.repo.GetWalletByTelegramID(ctx, nil, input.TelegramID)
			if err != nil {
				return ucModels.GenerateImageOutput{}, err
			}
			if wallet.BalanceAvailable < generator.ImageTokenCostFor(model) {
				return ucModels.GenerateImageOutput{}, ucModels.ErrInsufficientBalance
			}
		}
	} else if p, err := uc.repo.GetProfileByTelegramID(ctx, nil, input.TelegramID); err == nil {
		profileID = p.ID
	}

	category := input.Category
	if category == "" {
		category = "image"
	}

	imageIn := generator.ImageGenerateInput{
		Prompt:   input.Prompt,
		Category: category,
		Messages: toGeneratorMessages(input.Messages),
	}
	draftPrompt := generator.BuildImagePrompt(imageIn)
	finalPrompt := uc.enrichImagePrompt(ctx, draftPrompt)

	result, err := uc.imageGenerator.GenerateImage(ctx, model, generator.ImageGenerateInput{
		Prompt:   finalPrompt,
		Category: category,
	})
	if err != nil {
		return ucModels.GenerateImageOutput{}, mapGeneratorError(err)
	}

	tokensUsed := result.TokensUsed
	if tokensUsed == 0 {
		tokensUsed = generator.ImageTokenCostFor(model)
	}

	imageURL, err := imageURLForBytes(ctx, input.TelegramID, result.ImageBytes, result.MimeType)
	if err != nil {
		return ucModels.GenerateImageOutput{}, err
	}

	if profileID != 0 && !uc.skipRegistrationCheck && !uc.skipAIWalletCheck {
		cost := generator.ImageTokenCostFor(model)
		desc := fmt.Sprintf("AI image generation (%s)", model)
		if err := uc.repo.DeductWalletBalance(ctx, nil, profileID, cost, desc); err != nil {
			if errors.Is(err, repo.ErrInsufficientBalance) {
				return ucModels.GenerateImageOutput{}, ucModels.ErrInsufficientBalance
			}
			return ucModels.GenerateImageOutput{}, err
		}
	}

	if profileID != 0 {
		_, _ = uc.SavePromptHistory(ctx, ucModels.SavePromptHistoryInput{
			TelegramID: input.TelegramID,
			Prompt:     input.Prompt,
			Response:   "[изображение]",
			Category:   category,
			Model:      model,
			SessionID:  strings.TrimSpace(input.SessionID),
		})
	}

	return ucModels.GenerateImageOutput{
		ImageURL:   imageURL,
		Model:      model,
		TokensUsed: tokensUsed,
	}, nil
}

func normalizeGenerateImageInput(in ucModels.GenerateImageInput) ucModels.GenerateImageInput {
	prepared := generator.PrepareTextInput(generator.TextGenerateInput{
		Prompt:   in.Prompt,
		Category: in.Category,
		Messages: toGeneratorMessages(in.Messages),
	})
	in.Prompt = prepared.Prompt
	if len(prepared.Messages) == 0 {
		in.Messages = nil
		return in
	}
	in.Messages = make([]ucModels.ChatMessageInput, len(prepared.Messages))
	for i, m := range prepared.Messages {
		in.Messages[i] = ucModels.ChatMessageInput{Role: m.Role, Content: m.Content}
	}
	return in
}
