package generator

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gitlab16.skiftrade.kz/templates/go/pkg/config"
)

func TestDeepSeekClient_Generate_withContext(t *testing.T) {
	var gotInput string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req yandexResponsesRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		gotInput = req.Input
		_ = json.NewEncoder(w).Encode(yandexResponsesResponse{
			Output: []struct {
				Content []struct {
					Text string `json:"text"`
				} `json:"content"`
			}{
				{Content: []struct {
					Text string `json:"text"`
				}{{Text: "ok"}}},
			},
		})
	}))
	defer srv.Close()

	gen := newDeepSeekClient(config.ConfigYandex{
		APIKey:           "key",
		FolderID:         "folder",
		DeepSeekModel:    "deepseek-v32/latest",
		ResponsesBaseURL: srv.URL,
	})

	_, err := gen.Generate(context.Background(), TextGenerateInput{
		Prompt: "продолжи",
		Messages: []ChatMessage{
			{Role: "user", Content: "привет"},
			{Role: "assistant", Content: "здравствуйте"},
		},
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if !strings.Contains(gotInput, "привет") || !strings.Contains(gotInput, "здравствуйте") {
		t.Fatalf("input missing context: %q", gotInput)
	}
}
