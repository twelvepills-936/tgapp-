package generator

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab16.skiftrade.kz/templates/go/pkg/config"
)

func TestAliceAIArtClient_GenerateImage_OK(t *testing.T) {
	pngOnePixel := base64.StdEncoding.EncodeToString([]byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
	})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/images/generations" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Api-Key test-key" {
			t.Fatalf("auth = %q", r.Header.Get("Authorization"))
		}
		if r.Header.Get("OpenAI-Project") != "folder" {
			t.Fatalf("project = %q", r.Header.Get("OpenAI-Project"))
		}

		var req yandexImageGenerateRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req.Model != "art://folder/aliceai-image-art-3.0/latest" {
			t.Fatalf("model = %q", req.Model)
		}
		if req.Prompt == "" || req.Size != "1024x1024" {
			t.Fatalf("prompt/size: %q %q", req.Prompt, req.Size)
		}

		_ = json.NewEncoder(w).Encode(yandexImageGenerateResponse{
			Data: []struct {
				B64JSON string `json:"b64_json"`
				URL     string `json:"url"`
			}{{B64JSON: pngOnePixel}},
		})
	}))
	defer srv.Close()

	gen := NewAliceAIArtClient(config.ConfigYandex{
		APIKey:           "test-key",
		FolderID:         "folder",
		AliceAIArtModel:  "aliceai-image-art-3.0/latest",
		ResponsesBaseURL: srv.URL,
	})
	if gen == nil {
		t.Fatal("expected client")
	}

	res, err := gen.GenerateImage(context.Background(), ModelAliceAIArt, ImageGenerateInput{
		Prompt:   "Нарисуй кота",
		Category: "image",
	})
	if err != nil {
		t.Fatalf("GenerateImage: %v", err)
	}
	if len(res.ImageBytes) == 0 {
		t.Fatal("empty image")
	}
}

func TestNormalizeImageModel_AliceAliases(t *testing.T) {
	m, err := NormalizeImageModel("aliceai-image-art-3.0/latest")
	if err != nil || m != ModelAliceAIArt {
		t.Fatalf("got %q err=%v", m, err)
	}
}

func TestImageRouter_AliceAIArt(t *testing.T) {
	png := base64.StdEncoding.EncodeToString([]byte{0x89, 0x50, 0x4e, 0x47})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(yandexImageGenerateResponse{
			Data: []struct {
				B64JSON string `json:"b64_json"`
				URL     string `json:"url"`
			}{{B64JSON: png}},
		})
	}))
	defer srv.Close()

	router := NewImageRouter(config.ConfigYandex{
		APIKey:           "k",
		FolderID:         "f",
		AliceAIArtModel:  "aliceai-image-art-3.0/latest",
		ResponsesBaseURL: srv.URL,
	}, config.ConfigGemini{})

	_, err := router.GenerateImage(context.Background(), ModelAliceAIArt, ImageGenerateInput{Prompt: "cat", Category: "image"})
	if err != nil {
		t.Fatalf("router: %v", err)
	}
}
