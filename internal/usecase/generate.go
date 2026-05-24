package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	repo "gitlab16.skiftrade.kz/templates/go/internal/repository"
	ucModels "gitlab16.skiftrade.kz/templates/go/internal/usecase/models"
	"gitlab16.skiftrade.kz/templates/go/pkg/generator"
)

func (uc *useCase) GenerateText(ctx context.Context, input ucModels.GenerateTextInput) (ucModels.GenerateTextOutput, error) {
	requireTelegramID := !uc.skipRegistrationCheck
	if uc.skipRegistrationCheck && input.TelegramID == "" {
		input.TelegramID = "dev"
	}
	input = normalizeGenerateTextInput(input)
	if err := input.Validate(requireTelegramID); err != nil {
		return ucModels.GenerateTextOutput{}, err
	}

	model, err := generator.NormalizeModel(input.Model)
	if err != nil {
		return ucModels.GenerateTextOutput{}, mapGeneratorError(err)
	}

	var profileID int64
	if !uc.skipRegistrationCheck {
		profile, err := uc.repo.GetProfileByTelegramID(ctx, nil, input.TelegramID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ucModels.GenerateTextOutput{}, ucModels.ErrProfileNotFound
			}
			return ucModels.GenerateTextOutput{}, err
		}
		profileID = profile.ID

		wallet, err := uc.repo.GetWalletByTelegramID(ctx, nil, input.TelegramID)
		if err != nil {
			return ucModels.GenerateTextOutput{}, err
		}
		if wallet.BalanceAvailable < generator.TokenCostFor(model) {
			return ucModels.GenerateTextOutput{}, ucModels.ErrInsufficientBalance
		}
		_ = wallet
	} else if p, err := uc.repo.GetProfileByTelegramID(ctx, nil, input.TelegramID); err == nil {
		profileID = p.ID
	}

	category := input.Category
	if category == "" {
		category = "general"
	}

	result, err := uc.modelRouter.Generate(ctx, model, generator.TextGenerateInput{
		Prompt:   input.Prompt,
		Category: category,
		Messages: toGeneratorMessages(input.Messages),
	})
	if err != nil {
		return ucModels.GenerateTextOutput{}, mapGeneratorError(err)
	}

	tokensUsed := result.TokensUsed
	if tokensUsed == 0 {
		tokensUsed = generator.TokenCostFor(model)
	}

	if profileID != 0 && !uc.skipRegistrationCheck {
		cost := generator.TokenCostFor(model)
		desc := fmt.Sprintf("AI generation (%s)", model)
		if err := uc.repo.DeductWalletBalance(ctx, nil, profileID, cost, desc); err != nil {
			if errors.Is(err, repo.ErrInsufficientBalance) {
				return ucModels.GenerateTextOutput{}, ucModels.ErrInsufficientBalance
			}
			return ucModels.GenerateTextOutput{}, err
		}
	}

	if profileID != 0 {
		_, _ = uc.SavePromptHistory(ctx, ucModels.SavePromptHistoryInput{
			TelegramID: input.TelegramID,
			Prompt:     input.Prompt,
			Category:   category,
			Model:      model,
		})
	}

	return ucModels.GenerateTextOutput{
		Text:       result.Text,
		Model:      model,
		TokensUsed: tokensUsed,
	}, nil
}

func normalizeGenerateTextInput(in ucModels.GenerateTextInput) ucModels.GenerateTextInput {
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

func toGeneratorMessages(in []ucModels.ChatMessageInput) []generator.ChatMessage {
	if len(in) == 0 {
		return nil
	}
	out := make([]generator.ChatMessage, 0, len(in))
	for _, m := range in {
		out = append(out, generator.ChatMessage{Role: m.Role, Content: m.Content})
	}
	return out
}

func mapGeneratorError(err error) error {
	if errors.Is(err, generator.ErrUnsupportedModel) {
		return fmt.Errorf("%w: %s", ucModels.ErrInvalidInput, err.Error())
	}
	return err
}
