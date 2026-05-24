package generator

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"gitlab16.skiftrade.kz/templates/go/pkg/config"
)

// NewNanoBananaClient creates a Gemini image generator (Nano Banana).
func NewNanoBananaClient(cfg config.ConfigGemini) ImageGenerator {
	if cfg.APIKey == "" {
		return nil
	}
	baseURL := strings.TrimRight(cfg.BaseURL, "/")
	if baseURL == "" {
		baseURL = "https://generativelanguage.googleapis.com/v1beta"
	}
	model := cfg.ImageModel
	if model == "" {
		model = "gemini-2.5-flash-image"
	}
	return &nanoBananaClient{
		apiKey:  cfg.APIKey,
		baseURL: baseURL,
		model:   model,
		client:  &http.Client{Timeout: 120 * time.Second},
	}
}

type nanoBananaClient struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

type nanoBananaRequest struct {
	Contents         []nanoBananaContent `json:"contents"`
	GenerationConfig nanoBananaGenConfig `json:"generationConfig"`
}

type nanoBananaContent struct {
	Role  string           `json:"role,omitempty"`
	Parts []nanoBananaPart `json:"parts"`
}

type nanoBananaPart struct {
	Text       string              `json:"text,omitempty"`
	InlineData *nanoBananaInline   `json:"inlineData,omitempty"`
}

type nanoBananaInline struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
}

type nanoBananaGenConfig struct {
	ResponseModalities []string `json:"responseModalities"`
}

type nanoBananaResponse struct {
	Candidates []struct {
		Content struct {
			Parts []nanoBananaPart `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	UsageMetadata *struct {
		TotalTokenCount int64 `json:"totalTokenCount"`
	} `json:"usageMetadata"`
}

func (c *nanoBananaClient) GenerateImage(ctx context.Context, prompt, category string) (ImageResult, error) {
	if c.apiKey == "" {
		return ImageResult{}, ErrImageGeneratorUnavailable
	}

	text := strings.TrimSpace(prompt)
	if category != "" {
		text = fmt.Sprintf("Category: %s. %s", category, text)
	}

	body, err := json.Marshal(nanoBananaRequest{
		Contents: []nanoBananaContent{
			{Role: "user", Parts: []nanoBananaPart{{Text: text}}},
		},
		GenerationConfig: nanoBananaGenConfig{
			ResponseModalities: []string{"TEXT", "IMAGE"},
		},
	})
	if err != nil {
		return ImageResult{}, err
	}

	url := fmt.Sprintf("%s/models/%s:generateContent", c.baseURL, c.model)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return ImageResult{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return ImageResult{}, newProviderError("gemini-image", err.Error())
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return ImageResult{}, err
	}

	if msg := parseGeminiErrorMessage(raw); msg != "" {
		return ImageResult{}, newProviderError("gemini-image", enhanceGeminiQuotaMessage(msg))
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ImageResult{}, newProviderError("gemini-image", fmt.Sprintf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(raw))))
	}

	var parsed nanoBananaResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return ImageResult{}, err
	}

	var imageBytes []byte
	var mimeType string
	for _, cand := range parsed.Candidates {
		for _, part := range cand.Content.Parts {
			if part.InlineData == nil || part.InlineData.Data == "" {
				continue
			}
			decoded, err := base64.StdEncoding.DecodeString(part.InlineData.Data)
			if err != nil {
				return ImageResult{}, newProviderError("gemini-image", "invalid image data in response")
			}
			imageBytes = decoded
			mimeType = part.InlineData.MimeType
			if mimeType == "" {
				mimeType = "image/png"
			}
			break
		}
		if len(imageBytes) > 0 {
			break
		}
	}
	if len(imageBytes) == 0 {
		return ImageResult{}, newProviderError("gemini-image", "no image in model response")
	}

	var tokens int64
	if parsed.UsageMetadata != nil && parsed.UsageMetadata.TotalTokenCount > 0 {
		tokens = parsed.UsageMetadata.TotalTokenCount
	}
	if tokens == 0 {
		tokens = ImageTokenCostFor(ModelNanoBanana)
	}

	return ImageResult{
		ImageBytes: imageBytes,
		MimeType:   mimeType,
		TokensUsed: tokens,
	}, nil
}
