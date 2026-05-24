package usecase

import (
	"strings"
	"testing"

	ucModels "gitlab16.skiftrade.kz/templates/go/internal/usecase/models"
)

func TestNormalizeGenerateTextInput_longHistoryPassesValidate(t *testing.T) {
	longAssistant := strings.Repeat("a", 8000)
	in := normalizeGenerateTextInput(ucModels.GenerateTextInput{
		Prompt: "Напиши схему HTTPS",
		Messages: []ucModels.ChatMessageInput{
			{Role: "user", Content: "первый вопрос"},
			{Role: "assistant", Content: longAssistant},
		},
	})
	if err := in.Validate(false); err != nil {
		t.Fatalf("validate after normalize: %v", err)
	}
	if len(in.Messages[1].Content) >= len(longAssistant) {
		t.Fatal("expected assistant message to be trimmed")
	}
}
