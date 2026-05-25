package generator

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"gitlab16.skiftrade.kz/templates/go/pkg/config"
)

// Alice AI ART via Yandex AI Studio (OpenAI-compatible Images API).
// https://yandex.cloud/en/docs/ai-studio/operations/generation/aliceai-image

func NewAliceAIArtClient(cfg config.ConfigYandex) ImageGenerator {
	if strings.TrimSpace(cfg.APIKey) == "" || strings.TrimSpace(cfg.FolderID) == "" {
		return nil
	}

	model := strings.TrimSpace(cfg.AliceAIArtModel)
	if model == "" {
		model = "aliceai-image-art-3.0/latest"
	}
	modelURI := fmt.Sprintf("art://%s/%s", cfg.FolderID, model)

	baseURL := strings.TrimRight(cfg.ResponsesBaseURL, "/")
	if baseURL == "" {
		baseURL = "https://ai.api.cloud.yandex.net/v1"
	}

	size := strings.TrimSpace(cfg.ImageSize)
	if size == "" {
		size = "1024x1024"
	}

	return &aliceAIArtClient{
		apiKey:   cfg.APIKey,
		folderID: cfg.FolderID,
		modelURI: modelURI,
		baseURL:  baseURL,
		size:     size,
		client:   &http.Client{Timeout: 120 * time.Second},
	}
}

type aliceAIArtClient struct {
	apiKey   string
	folderID string
	modelURI string
	baseURL  string
	size     string
	client   *http.Client
}

type yandexImageGenerateRequest struct {
	Model          string `json:"model"`
	Prompt         string `json:"prompt"`
	Size           string `json:"size"`
	ResponseFormat string `json:"response_format,omitempty"`
}

type yandexImageGenerateResponse struct {
	Data []struct {
		B64JSON string `json:"b64_json"`
		URL     string `json:"url"`
	} `json:"data"`
	Error json.RawMessage `json:"error"`
}

func (c *aliceAIArtClient) GenerateImage(ctx context.Context, _ string, in ImageGenerateInput) (ImageResult, error) {
	text := strings.TrimSpace(BuildImagePrompt(in))
	if text == "" {
		return ImageResult{}, newProviderError("alice-ai-art", "empty prompt")
	}

	res, err := c.postGenerate(ctx, text)
	if err != nil && IsContentPolicy(err) && len(in.Messages) > 0 {
		fallback := strings.TrimSpace(enhanceStandaloneImagePrompt(in.Prompt))
		if fallback != "" && fallback != text {
			if res2, err2 := c.postGenerate(ctx, fallback); err2 == nil {
				return res2, nil
			}
		}
	}
	return res, err
}

func (c *aliceAIArtClient) postGenerate(ctx context.Context, prompt string) (ImageResult, error) {
	body, err := json.Marshal(yandexImageGenerateRequest{
		Model:          c.modelURI,
		Prompt:         prompt,
		Size:           c.size,
		ResponseFormat: "b64_json",
	})
	if err != nil {
		return ImageResult{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/images/generations", bytes.NewReader(body))
	if err != nil {
		return ImageResult{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Api-Key "+c.apiKey)
	req.Header.Set("OpenAI-Project", c.folderID)

	resp, err := c.client.Do(req)
	if err != nil {
		return ImageResult{}, newProviderError("alice-ai-art", err.Error())
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return ImageResult{}, err
	}

	if msg := parseYandexAPIError(raw); msg != "" {
		err := newProviderError("alice-ai-art", msg)
		if IsContentPolicy(err) {
			return ImageResult{}, errors.Join(ErrContentPolicy, err)
		}
		return ImageResult{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ImageResult{}, newProviderError("alice-ai-art", fmt.Sprintf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(raw))))
	}

	var parsed yandexImageGenerateResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return ImageResult{}, newProviderError("alice-ai-art", fmt.Sprintf("invalid response: %s", strings.TrimSpace(string(raw))))
	}

	var imageBytes []byte
	for _, item := range parsed.Data {
		if item.B64JSON == "" {
			continue
		}
		decoded, err := base64.StdEncoding.DecodeString(item.B64JSON)
		if err != nil {
			return ImageResult{}, newProviderError("alice-ai-art", "invalid base64 image data")
		}
		imageBytes = decoded
		break
	}
	if len(imageBytes) == 0 {
		return ImageResult{}, newProviderError("alice-ai-art", "no image in model response")
	}

	return ImageResult{
		ImageBytes: imageBytes,
		MimeType:   "image/png",
		TokensUsed: ImageTokenCostFor(ModelAliceAIArt),
	}, nil
}
