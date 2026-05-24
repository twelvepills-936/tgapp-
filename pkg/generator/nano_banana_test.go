package generator

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab16.skiftrade.kz/templates/go/pkg/config"
)

func TestNanoBananaClient_GenerateImage(t *testing.T) {
	t.Parallel()

	const pngB64 = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg=="

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-goog-api-key") != "test-key" {
			t.Errorf("unexpected api key: %q", r.Header.Get("x-goog-api-key"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"candidates":[{"content":{"parts":[{"inlineData":{"mimeType":"image/png","data":"` + pngB64 + `"}}]}}],
			"usageMetadata":{"totalTokenCount":42}
		}`))
	}))
	defer srv.Close()

	gen := NewNanoBananaClient(config.ConfigGemini{
		APIKey:     "test-key",
		ImageModel: "gemini-2.5-flash-image",
		BaseURL:    srv.URL,
	})

	res, err := gen.GenerateImage(context.Background(), ModelNanoBanana, ImageGenerateInput{Prompt: "a red circle", Category: "image"})
	if err != nil {
		t.Fatalf("GenerateImage: %v", err)
	}
	if len(res.ImageBytes) == 0 {
		t.Fatal("expected image bytes")
	}
	if res.MimeType != "image/png" {
		t.Fatalf("mimeType = %q, want image/png", res.MimeType)
	}
	if res.TokensUsed != 42 {
		t.Fatalf("tokens = %d, want 42", res.TokensUsed)
	}
	if _, err := base64.StdEncoding.DecodeString(pngB64); err != nil {
		t.Fatalf("fixture png invalid: %v", err)
	}
}

func TestNewNanoBananaClient_NoAPIKey(t *testing.T) {
	t.Parallel()
	if NewNanoBananaClient(config.ConfigGemini{}) != nil {
		t.Fatal("expected nil client without API key")
	}
}
