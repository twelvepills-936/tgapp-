package generator

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"gitlab16.skiftrade.kz/templates/go/pkg/config"
)

const defaultWaveSpeedImageBase = "https://api.wavespeed.ai/api/v3"

type nanoBananaWaveSpeedClient struct {
	apiKey   string
	baseURL  string
	endpoint string
	client   *http.Client
}

func newNanoBananaWaveSpeedClient(cfg config.ConfigGemini) ImageGenerator {
	baseURL := strings.TrimRight(cfg.WaveSpeedImageBase, "/")
	if baseURL == "" {
		baseURL = defaultWaveSpeedImageBase
	}
	return &nanoBananaWaveSpeedClient{
		apiKey:   cfg.APIKey,
		baseURL:  baseURL,
		endpoint: waveSpeedNanoBananaEndpoint(cfg.ImageModel),
		client:   &http.Client{Timeout: 180 * time.Second},
	}
}

func waveSpeedNanoBananaEndpoint(model string) string {
	model = strings.TrimSpace(model)
	if model == "" {
		model = "google/nano-banana-pro"
	}
	if strings.Contains(model, "/text-to-image") || strings.Contains(model, "/edit") {
		return model
	}
	if strings.HasPrefix(model, "google/") {
		return model + "/text-to-image"
	}
	return "google/" + model + "/text-to-image"
}

type waveSpeedImageRequest struct {
	Prompt             string `json:"prompt"`
	Resolution         string `json:"resolution,omitempty"`
	OutputFormat       string `json:"output_format,omitempty"`
	AspectRatio        string `json:"aspect_ratio,omitempty"`
	EnableSyncMode     bool   `json:"enable_sync_mode"`
	EnableBase64Output bool   `json:"enable_base64_output"`
}

type waveSpeedImageResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		ID      string   `json:"id"`
		Status  string   `json:"status"`
		Outputs []string `json:"outputs"`
		Error   string   `json:"error"`
		URLs    struct {
			Get string `json:"get"`
		} `json:"urls"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (c *nanoBananaWaveSpeedClient) GenerateImage(ctx context.Context, _ string, in ImageGenerateInput) (ImageResult, error) {
	text := strings.TrimSpace(in.Prompt)
	if text == "" {
		text = strings.TrimSpace(BuildImagePromptCompact(in))
	}
	if text == "" {
		return ImageResult{}, newProviderError("nano-banana", "empty prompt")
	}
	if in.Category != "" && !strings.HasPrefix(text, "Category:") {
		text = fmt.Sprintf("Category: %s. %s", in.Category, text)
	}

	resolution := strings.TrimSpace(envOr("NANO_BANANA_RESOLUTION", "1k"))
	outputFormat := strings.TrimSpace(envOr("NANO_BANANA_OUTPUT_FORMAT", "jpeg"))
	if outputFormat == "" {
		outputFormat = "jpeg"
	}
	aspectRatio := strings.TrimSpace(os.Getenv("NANO_BANANA_ASPECT_RATIO"))
	syncMode := EnvBool("NANO_BANANA_SYNC_MODE")
	base64Out := EnvBool("NANO_BANANA_BASE64_OUTPUT")

	body, err := json.Marshal(waveSpeedImageRequest{
		Prompt:             text,
		Resolution:         resolution,
		OutputFormat:       outputFormat,
		AspectRatio:        aspectRatio,
		EnableSyncMode:     syncMode,
		EnableBase64Output: base64Out,
	})
	if err != nil {
		return ImageResult{}, err
	}

	url := c.baseURL + "/" + c.endpoint
	raw, err := c.postJSON(ctx, http.MethodPost, url, body)
	if err != nil {
		return ImageResult{}, err
	}

	parsed, err := parseWaveSpeedImageResponse(raw)
	if err != nil {
		return ImageResult{}, err
	}

	if strings.EqualFold(parsed.Data.Status, "completed") && len(parsed.Data.Outputs) > 0 {
		return c.outputsToResult(ctx, parsed.Data.Outputs, outputFormat)
	}

	if parsed.Data.ID == "" {
		return ImageResult{}, newProviderError("nano-banana", waveSpeedImageErrorMessage(parsed, raw))
	}

	return c.pollResult(ctx, parsed, outputFormat)
}

func (c *nanoBananaWaveSpeedClient) postJSON(ctx context.Context, method, url string, body []byte) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, newProviderError("nano-banana", err.Error())
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg := strings.TrimSpace(string(raw))
		var envelope waveSpeedImageResponse
		if json.Unmarshal(raw, &envelope) == nil {
			if m := waveSpeedImageErrorMessage(&envelope, raw); m != "" {
				msg = m
			}
		}
		return nil, newProviderError("nano-banana", fmt.Sprintf("HTTP %d: %s", resp.StatusCode, msg))
	}
	return raw, nil
}

func parseWaveSpeedImageResponse(raw []byte) (*waveSpeedImageResponse, error) {
	var parsed waveSpeedImageResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, err
	}
	if parsed.Code != 0 && parsed.Code != 200 {
		return &parsed, newProviderError("nano-banana", waveSpeedImageErrorMessage(&parsed, raw))
	}
	if parsed.Error != nil && parsed.Error.Message != "" {
		return &parsed, newProviderError("nano-banana", parsed.Error.Message)
	}
	if msg := strings.TrimSpace(parsed.Data.Error); msg != "" && !strings.EqualFold(parsed.Data.Status, "completed") {
		return &parsed, newProviderError("nano-banana", msg)
	}
	return &parsed, nil
}

func waveSpeedImageErrorMessage(parsed *waveSpeedImageResponse, raw []byte) string {
	if parsed == nil {
		return strings.TrimSpace(string(raw))
	}
	if parsed.Error != nil && parsed.Error.Message != "" {
		return parsed.Error.Message
	}
	if msg := strings.TrimSpace(parsed.Data.Error); msg != "" {
		return msg
	}
	if msg := strings.TrimSpace(parsed.Message); msg != "" && !strings.EqualFold(msg, "success") {
		return msg
	}
	return strings.TrimSpace(string(raw))
}

func (c *nanoBananaWaveSpeedClient) pollResult(ctx context.Context, initial *waveSpeedImageResponse, outputFormat string) (ImageResult, error) {
	pollURL := strings.TrimSpace(initial.Data.URLs.Get)
	if pollURL == "" && initial.Data.ID != "" {
		pollURL = c.baseURL + "/predictions/" + initial.Data.ID + "/result"
	}
	if pollURL == "" {
		return ImageResult{}, newProviderError("nano-banana", "empty task id in WaveSpeed response")
	}

	interval := waveSpeedPollInterval()
	const attempts = 120
	for i := 0; i < attempts; i++ {
		if err := ctx.Err(); err != nil {
			return ImageResult{}, err
		}
		if i > 0 {
			select {
			case <-ctx.Done():
				return ImageResult{}, ctx.Err()
			case <-time.After(interval):
			}
		}

		raw, err := c.postJSON(ctx, http.MethodGet, pollURL, nil)
		if err != nil {
			return ImageResult{}, err
		}
		parsed, err := parseWaveSpeedImageResponse(raw)
		if err != nil {
			return ImageResult{}, err
		}
		switch strings.ToLower(strings.TrimSpace(parsed.Data.Status)) {
		case "completed":
			if len(parsed.Data.Outputs) == 0 {
				return ImageResult{}, newProviderError("nano-banana", "no image in completed task")
			}
			return c.outputsToResult(ctx, parsed.Data.Outputs, outputFormat)
		case "failed":
			return ImageResult{}, newProviderError("nano-banana", waveSpeedImageErrorMessage(parsed, raw))
		}
	}
	return ImageResult{}, newProviderError("nano-banana", "image generation timed out")
}

func envOr(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

func waveSpeedPollInterval() time.Duration {
	if ms := strings.TrimSpace(os.Getenv("NANO_BANANA_POLL_MS")); ms != "" {
		if d, err := time.ParseDuration(ms + "ms"); err == nil && d >= 200*time.Millisecond {
			return d
		}
	}
	return 800 * time.Millisecond
}

func (c *nanoBananaWaveSpeedClient) outputsToResult(ctx context.Context, outputs []string, outputFormat string) (ImageResult, error) {
	mimeDefault := "image/png"
	if outputFormat == "jpeg" || outputFormat == "jpg" {
		mimeDefault = "image/jpeg"
	}

	for _, out := range outputs {
		out = strings.TrimSpace(out)
		if out == "" {
			continue
		}
		imageBytes, mime, err := decodeWaveSpeedImageOutput(ctx, c.client, c.apiKey, out, mimeDefault)
		if err != nil {
			return ImageResult{}, err
		}
		if len(imageBytes) > 0 {
			return ImageResult{
				ImageBytes: imageBytes,
				MimeType:   mime,
				TokensUsed: ImageTokenCostFor(ModelNanoBanana),
			}, nil
		}
	}
	return ImageResult{}, newProviderError("nano-banana", "no image in model response")
}

func decodeWaveSpeedImageOutput(ctx context.Context, client *http.Client, apiKey, out, mimeDefault string) ([]byte, string, error) {
	if strings.HasPrefix(out, "http://") || strings.HasPrefix(out, "https://") {
		return downloadWaveSpeedImage(ctx, client, apiKey, out, mimeDefault)
	}
	if strings.HasPrefix(out, "data:") {
		return decodeDataURLImage(out, mimeDefault)
	}
	raw, err := base64.StdEncoding.DecodeString(out)
	if err != nil {
		return nil, "", newProviderError("nano-banana", "invalid image data in response")
	}
	return raw, mimeDefault, nil
}

func downloadWaveSpeedImage(ctx context.Context, client *http.Client, apiKey, url, mimeDefault string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, "", err
	}
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", newProviderError("nano-banana", err.Error())
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", newProviderError("nano-banana", fmt.Sprintf("failed to download image: HTTP %d", resp.StatusCode))
	}
	mime := resp.Header.Get("Content-Type")
	if mime == "" {
		mime = mimeDefault
	}
	return raw, mime, nil
}

func decodeDataURLImage(dataURL, mimeDefault string) ([]byte, string, error) {
	comma := strings.Index(dataURL, ",")
	if comma < 0 {
		return nil, "", newProviderError("nano-banana", "invalid data URL")
	}
	meta := dataURL[:comma]
	payload := dataURL[comma+1:]
	mime := mimeDefault
	if semi := strings.Index(meta, ";"); semi > 0 && strings.HasPrefix(meta, "data:") {
		mime = strings.TrimPrefix(meta[:semi], "data:")
	}
	raw, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return nil, "", newProviderError("nano-banana", "invalid base64 image data")
	}
	return raw, mime, nil
}
