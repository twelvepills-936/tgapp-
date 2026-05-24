package httphandler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab16.skiftrade.kz/templates/go/internal"
	ucModels "gitlab16.skiftrade.kz/templates/go/internal/usecase/models"
	"gitlab16.skiftrade.kz/templates/go/pkg/generator"
)

type fakeGenerateUC struct{}

func (fakeGenerateUC) RegisterByTelegram(context.Context, ucModels.RegisterByTelegramInput) (ucModels.RegisterByTelegramOutput, error) {
	return ucModels.RegisterByTelegramOutput{}, nil
}
func (fakeGenerateUC) GetUserByTelegramID(context.Context, string) (ucModels.GetProfileOutput, error) {
	return ucModels.GetProfileOutput{}, nil
}
func (fakeGenerateUC) GetWalletByTelegramID(context.Context, string) (ucModels.GetWalletOutput, error) {
	return ucModels.GetWalletOutput{}, nil
}
func (fakeGenerateUC) GetReferralsByTelegramID(context.Context, string) (ucModels.GetReferralsOutput, error) {
	return ucModels.GetReferralsOutput{}, nil
}
func (fakeGenerateUC) SavePromptHistory(context.Context, ucModels.SavePromptHistoryInput) (ucModels.SavePromptHistoryOutput, error) {
	return ucModels.SavePromptHistoryOutput{}, nil
}
func (fakeGenerateUC) GetPromptHistoryByTelegramID(context.Context, string) (ucModels.GetPromptHistoryOutput, error) {
	return ucModels.GetPromptHistoryOutput{}, nil
}
func (fakeGenerateUC) GenerateImage(context.Context, ucModels.GenerateImageInput) (ucModels.GenerateImageOutput, error) {
	return ucModels.GenerateImageOutput{}, nil
}
func (fakeGenerateUC) GenerateText(_ context.Context, input ucModels.GenerateTextInput) (ucModels.GenerateTextOutput, error) {
	if input.Prompt == "" {
		return ucModels.GenerateTextOutput{}, ucModels.ErrInvalidInput
	}
	model := input.Model
	if model == "" {
		model = generator.DefaultModel
	}
	return ucModels.GenerateTextOutput{Text: "ok", Model: model, TokensUsed: 10}, nil
}

var _ internal.UseCase = fakeGenerateUC{}

func TestGenerateTextHandler_OK(t *testing.T) {
	body, _ := json.Marshal(map[string]string{
		"telegramId": "123",
		"prompt":     "Hello",
		"model":      "yandexgpt",
	})
	req := httptest.NewRequest(http.MethodPost, generateTextPath, bytes.NewReader(body))
	rec := httptest.NewRecorder()

	NewGenerateTextHandler(fakeGenerateUC{}).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: %d body: %s", rec.Code, rec.Body.String())
	}

	var resp generateTextResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Data.Text != "ok" || resp.Data.Model != "yandexgpt" || resp.Data.TokensUsed != 10 {
		t.Fatalf("unexpected response: %+v", resp.Data)
	}
}

func TestGenerateTextHandler_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, generateTextPath, bytes.NewReader([]byte("{")))
	rec := httptest.NewRecorder()
	NewGenerateTextHandler(fakeGenerateUC{}).ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: %d", rec.Code)
	}
}
