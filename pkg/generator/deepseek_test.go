package generator

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab16.skiftrade.kz/templates/go/pkg/config"
)

func TestDeepSeekClient_Generate_OK(t *testing.T) {
	var gotAuth, gotProject string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotProject = r.Header.Get("OpenAI-Project")
		if r.URL.Path != "/responses" {
			t.Fatalf("path = %s", r.URL.Path)
		}

		var req yandexResponsesRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req.Model != "gpt://folder/deepseek-v32/latest" {
			t.Fatalf("model = %q", req.Model)
		}
		if req.Input == "" {
			t.Fatal("expected input")
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(yandexResponsesResponse{
			Output: []struct {
				Content []struct {
					Text string `json:"text"`
				} `json:"content"`
			}{
				{Content: []struct {
					Text string `json:"text"`
				}{{Text: "DeepSeek ответ"}}},
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

	res, err := gen.Generate(context.Background(), TextGenerateInput{
		Prompt:   "Привет",
		Category: "text",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Text != "DeepSeek ответ" {
		t.Fatalf("text = %q", res.Text)
	}
	if gotAuth != "Api-Key key" || gotProject != "folder" {
		t.Fatalf("headers: auth=%q project=%q", gotAuth, gotProject)
	}
}

func TestNormalizeModel_DeepSeekAliases(t *testing.T) {
	m, err := NormalizeModel("deepseek-v32/latest")
	if err != nil || m != ModelDeepSeek {
		t.Fatalf("got %q err=%v", m, err)
	}
}
