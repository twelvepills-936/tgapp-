package generator

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gitlab16.skiftrade.kz/templates/go/pkg/config"
)

func TestYandexClient_Generate_OK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"result":{"alternatives":[{"message":{"text":"Привет!"}}],"usage":{"totalTokens":"5"}}}`))
	}))
	defer srv.Close()

	gen := newYandexClient(config.ConfigYandex{
		APIKey:   "key",
		FolderID: "folder",
		ModelURI: "gpt://folder/yandexgpt/latest",
		BaseURL:  srv.URL,
	})

	res, err := gen.Generate(context.Background(), TextGenerateInput{Prompt: "hi", Category: "text"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Text != "Привет!" {
		t.Fatalf("unexpected text: %q", res.Text)
	}
}

func TestParseYandexAPIError_String(t *testing.T) {
	msg := parseYandexAPIError([]byte(`{"error":"Permission denied"}`))
	if msg != "Permission denied" {
		t.Fatalf("got %q", msg)
	}
}

func TestYandexClient_Generate_WithHistory(t *testing.T) {
	var gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(buf)
		gotBody = string(buf)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"result":{"alternatives":[{"message":{"text":"Продолжаю тему."}}],"usage":{"totalTokens":"12"}}}`))
	}))
	defer srv.Close()

	gen := newYandexClient(config.ConfigYandex{
		APIKey:   "key",
		FolderID: "folder",
		ModelURI: "gpt://folder/yandexgpt/latest",
		BaseURL:  srv.URL,
	})

	_, err := gen.Generate(context.Background(), TextGenerateInput{
		Category: "text",
		Messages: []ChatMessage{
			{Role: "user", Content: "Пиши пост про кофе"},
			{Role: "assistant", Content: "Кофе — напиток..."},
		},
		Prompt: "Сделай короче",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(gotBody, "Пиши пост про кофе") || !strings.Contains(gotBody, "Сделай короче") {
		t.Fatalf("history not sent to yandex: %s", gotBody)
	}
}

func TestParseYandexAPIError_Object(t *testing.T) {
	msg := parseYandexAPIError([]byte(`{"error":{"message":"Invalid model","code":3}}`))
	if msg != "Invalid model" {
		t.Fatalf("got %q", msg)
	}
}
