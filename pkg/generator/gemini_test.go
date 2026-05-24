package generator

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gitlab16.skiftrade.kz/templates/go/pkg/config"
)

func TestGeminiClient_Generate_OK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-goog-api-key") != "secret" {
			t.Fatalf("missing api key header")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"candidates":[{"content":{"parts":[{"text":"Готово!"}]}}],"usageMetadata":{"totalTokenCount":42}}`))
	}))
	defer srv.Close()

	gen := newGeminiClient(config.ConfigGemini{
		APIKey:  "secret",
		Model:   "gemini-2.0-flash",
		BaseURL: srv.URL,
	})

	res, err := gen.Generate(context.Background(), TextGenerateInput{Prompt: "Напиши пост", Category: "text"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Text != "Готово!" || res.TokensUsed != 42 {
		t.Fatalf("unexpected result: %+v", res)
	}
}

func TestNormalizeModel(t *testing.T) {
	m, err := NormalizeModel("gemini-flash")
	if err != nil || m != ModelGeminiFlash {
		t.Fatalf("got %q err=%v", m, err)
	}
	_, err = NormalizeModel("unknown")
	if !strings.Contains(err.Error(), "unsupported") {
		t.Fatalf("expected unsupported model error, got %v", err)
	}
}
