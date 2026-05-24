package generator

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"gitlab16.skiftrade.kz/templates/go/pkg/config"
)

// GeminiConfig configures Google Gemini (Generative Language API).
type GeminiConfig struct {
	APIKey  string
	Model   string
	BaseURL string
}

func newGeminiClient(cfg config.ConfigGemini) *geminiClient {
	baseURL := strings.TrimRight(cfg.BaseURL, "/")
	if baseURL == "" {
		baseURL = "https://generativelanguage.googleapis.com/v1beta"
	}
	model := cfg.Model
	if model == "" {
		model = "gemini-2.0-flash"
	}
	return &geminiClient{
		apiKey:  cfg.APIKey,
		baseURL: baseURL,
		model:   model,
		client:  &http.Client{Timeout: 90 * time.Second},
	}
}

type geminiClient struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

type geminiGenerateRequest struct {
	SystemInstruction *geminiContent  `json:"systemInstruction,omitempty"`
	Contents          []geminiContent `json:"contents"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiGenerateResponse struct {
	Candidates []struct {
		Content struct {
			Parts []geminiPart `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	UsageMetadata *struct {
		TotalTokenCount int64 `json:"totalTokenCount"`
	} `json:"usageMetadata"`
}

var geminiModelFallback = []string{
	"gemini-2.0-flash",
	"gemini-2.0-flash-lite",
	"gemini-2.0-flash-001",
	"gemini-2.0-flash-lite-001",
}

func (c *geminiClient) Generate(ctx context.Context, in TextGenerateInput) (Result, error) {
	turns := MergePromptAndMessages(in.Messages, in.Prompt)
	multiTurn := len(turns) > 1
	messages := PrependSystem(turns, englishSystemPrompt(in.Category, multiTurn))

	models := geminiModelsToTry(c.model)
	var lastErr error
	for _, model := range models {
		res, err := c.generateWithModel(ctx, model, messages)
		if err == nil {
			res.TokensUsed = max64(res.TokensUsed, TokenCostFor(ModelGeminiFlash))
			return res, nil
		}
		lastErr = err
		if !isRetryableGeminiError(err) {
			return Result{}, err
		}
	}
	if lastErr != nil {
		return Result{}, lastErr
	}
	return Result{}, newProviderError("gemini", "generation failed")
}

func geminiModelsToTry(primary string) []string {
	seen := make(map[string]struct{})
	out := make([]string, 0, len(geminiModelFallback)+1)
	add := func(m string) {
		m = strings.TrimSpace(m)
		if m == "" {
			return
		}
		if _, ok := seen[m]; ok {
			return
		}
		seen[m] = struct{}{}
		out = append(out, m)
	}
	add(primary)
	for _, m := range geminiModelFallback {
		add(m)
	}
	return out
}

func isRetryableGeminiError(err error) bool {
	var pe *ProviderError
	if !errors.As(err, &pe) {
		return false
	}
	msg := strings.ToLower(pe.Message)
	return strings.Contains(msg, "quota") ||
		strings.Contains(msg, "rate") ||
		strings.Contains(msg, "429") ||
		strings.Contains(msg, "resource_exhausted")
}

func (c *geminiClient) generateWithModel(ctx context.Context, model string, messages []ChatMessage) (Result, error) {
	system, contents := geminiDialogFromMessages(messages)
	if len(contents) == 0 {
		return Result{}, newProviderError("gemini", "empty dialog")
	}

	body, err := json.Marshal(geminiGenerateRequest{
		SystemInstruction: &geminiContent{Parts: []geminiPart{{Text: system}}},
		Contents:          contents,
	})
	if err != nil {
		return Result{}, err
	}

	url := fmt.Sprintf("%s/models/%s:generateContent", c.baseURL, model)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return Result{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return Result{}, newProviderError("gemini", err.Error())
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return Result{}, err
	}

	msg := parseGeminiErrorMessage(raw)
	if msg != "" {
		return Result{}, newProviderError("gemini", enhanceGeminiQuotaMessage(msg))
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Result{}, newProviderError("gemini", fmt.Sprintf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(raw))))
	}

	var parsed geminiGenerateResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return Result{}, err
	}
	if len(parsed.Candidates) == 0 || len(parsed.Candidates[0].Content.Parts) == 0 {
		return Result{}, newProviderError("gemini", "empty response from model "+model)
	}

	text := strings.TrimSpace(parsed.Candidates[0].Content.Parts[0].Text)
	if text == "" {
		return Result{}, newProviderError("gemini", "empty text from model "+model)
	}

	var tokens int64
	if parsed.UsageMetadata != nil && parsed.UsageMetadata.TotalTokenCount > 0 {
		tokens = parsed.UsageMetadata.TotalTokenCount
	}
	return Result{Text: text, TokensUsed: tokens}, nil
}

func parseGeminiErrorMessage(raw []byte) string {
	var envelope struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(raw, &envelope); err == nil && envelope.Error.Message != "" {
		return envelope.Error.Message
	}
	return ""
}

func max64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func geminiDialogFromMessages(messages []ChatMessage) (system string, contents []geminiContent) {
	for _, m := range messages {
		switch m.Role {
		case "system":
			if system == "" {
				system = m.Content
			}
		case "assistant":
			contents = append(contents, geminiContent{
				Role:  "model",
				Parts: []geminiPart{{Text: m.Content}},
			})
		default:
			contents = append(contents, geminiContent{
				Role:  "user",
				Parts: []geminiPart{{Text: m.Content}},
			})
		}
	}
	if system == "" {
		system = englishSystemPrompt("", false)
	}
	return system, contents
}
