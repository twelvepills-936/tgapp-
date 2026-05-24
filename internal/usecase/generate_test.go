package usecase

import (
	"context"
	"errors"
	"testing"

	repoModels "gitlab16.skiftrade.kz/templates/go/internal/repository/models"
	ucModels "gitlab16.skiftrade.kz/templates/go/internal/usecase/models"
	"gitlab16.skiftrade.kz/templates/go/pkg/generator"
)

func TestGenerateText_Yandex_OK(t *testing.T) {
	repo := &fakeRepoProfile{
		exists:  map[string]repoModels.Profile{"123": {ID: 1, TelegramID: "123"}},
		prompts: map[string][]repoModels.PromptHistory{},
	}
	uc := NewUseCase(repo, &fakeModelRouter{text: "hello", model: generator.ModelYandexGPT}, nil, UseCaseOptions{SkipRegistrationCheck: true})

	out, err := uc.GenerateText(context.Background(), ucModels.GenerateTextInput{
		Prompt:   "test",
		Category: "text",
		Model:    generator.ModelYandexGPT,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Text != "hello" || out.Model != generator.ModelYandexGPT {
		t.Fatalf("unexpected output: %+v", out)
	}
}

type fakeModelRouter struct {
	text  string
	model string
}

func (f *fakeModelRouter) Generate(_ context.Context, model string, _ generator.TextGenerateInput) (generator.Result, error) {
	if f.model != "" && model != f.model {
		return generator.Result{}, generator.ErrUnsupportedModel
	}
	return generator.Result{Text: f.text, TokensUsed: 10}, nil
}

func TestGenerateText_ProfileNotFound(t *testing.T) {
	repo := &fakeRepoProfile{exists: map[string]repoModels.Profile{}}
	uc := NewUseCase(repo, &fakeModelRouter{text: "x", model: generator.ModelYandexGPT}, nil, UseCaseOptions{SkipRegistrationCheck: false})

	_, err := uc.GenerateText(context.Background(), ucModels.GenerateTextInput{
		TelegramID: "999",
		Prompt:     "test",
		Model:      generator.ModelYandexGPT,
	})
	if !errors.Is(err, ucModels.ErrProfileNotFound) {
		t.Fatalf("expected ErrProfileNotFound, got %v", err)
	}
}

func TestGenerateText_UnsupportedModel(t *testing.T) {
	uc := NewUseCase(&fakeRepoProfile{}, &fakeModelRouter{}, nil, UseCaseOptions{SkipRegistrationCheck: true})
	_, err := uc.GenerateText(context.Background(), ucModels.GenerateTextInput{
		Prompt: "x",
		Model:  "bad-model",
	})
	if !errors.Is(err, ucModels.ErrInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
}

func TestGenerateText_DefaultModel(t *testing.T) {
	uc := NewUseCase(&fakeRepoProfile{}, &fakeModelRouter{text: "ok"}, nil, UseCaseOptions{SkipRegistrationCheck: true})
	out, err := uc.GenerateText(context.Background(), ucModels.GenerateTextInput{Prompt: "hi"})
	if err != nil || out.Model != generator.DefaultModel {
		t.Fatalf("got %+v err=%v", out, err)
	}
}
