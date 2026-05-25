package generator

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gitlab16.skiftrade.kz/templates/go/pkg/config"
)

func TestNanoBananaWaveSpeed_GenerateImage(t *testing.T) {
	t.Parallel()

	const pngB64 = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg=="

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("unexpected auth: %q", r.Header.Get("Authorization"))
		}
		if !strings.Contains(r.URL.Path, "nano-banana-pro/text-to-image") {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"code":200,
			"message":"success",
			"data":{"status":"completed","outputs":["` + pngB64 + `"]}
		}`))
	}))
	defer srv.Close()

	gen := NewNanoBananaClient(config.ConfigGemini{
		APIKey:             "test-key",
		ImageModel:         "google/nano-banana-pro",
		BaseURL:            "https://llm.wavespeed.ai/v1",
		WaveSpeedImageBase: srv.URL,
	})

	res, err := gen.GenerateImage(context.Background(), ModelNanoBanana, ImageGenerateInput{Prompt: "a red circle", Category: "image"})
	if err != nil {
		t.Fatalf("GenerateImage: %v", err)
	}
	if len(res.ImageBytes) == 0 {
		t.Fatal("expected image bytes")
	}
	if res.MimeType != "image/jpeg" {
		t.Fatalf("mimeType = %q, want image/jpeg", res.MimeType)
	}
}

func TestNanoBananaGoogle_GenerateImage(t *testing.T) {
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
		UseGoogle:  true,
	})

	res, err := gen.GenerateImage(context.Background(), ModelNanoBanana, ImageGenerateInput{Prompt: "a red circle", Category: "image"})
	if err != nil {
		t.Fatalf("GenerateImage: %v", err)
	}
	if len(res.ImageBytes) == 0 {
		t.Fatal("expected image bytes")
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

func TestWaveSpeedNanoBananaEndpoint(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"google/nano-banana-pro":              "google/nano-banana-pro/text-to-image",
		"google/nano-banana-pro/text-to-image": "google/nano-banana-pro/text-to-image",
		"":                                    "google/nano-banana-pro/text-to-image",
	}
	for in, want := range cases {
		if got := waveSpeedNanoBananaEndpoint(in); got != want {
			t.Fatalf("endpoint(%q) = %q, want %q", in, got, want)
		}
	}
}
