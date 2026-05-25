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

	"github.com/google/generative-ai-go/genai"
	"gitlab16.skiftrade.kz/templates/go/pkg/config"
	"google.golang.org/api/option"
)

const defaultGeminiAPIBase = "https://generativelanguage.googleapis.com/v1beta"

// Fallback IDs for Google AI Studio (v1beta). No "-latest" aliases — they return 404.
var geminiModelFallback = []string{
	"gemini-2.0-flash-lite-001",
	"gemini-2.0-flash-lite",
	"gemini-2.0-flash-001",
	"gemini-2.0-flash",
	"gemini-1.5-flash-002",
	"gemini-1.5-flash",
}

// GeminiConfig configures Google Gemini (Generative Language API).
type GeminiConfig struct {
	APIKey  string
	Model   string
	BaseURL string
}

func newGeminiClient(cfg config.ConfigGemini) *geminiClient {
	baseURL := strings.TrimRight(cfg.BaseURL, "/")
	model := cfg.Model
	if model == "" {
		model = "gemini-2.0-flash-lite"
	}

	c := &geminiClient{
		apiKey:  cfg.APIKey,
		baseURL: baseURL,
		model:   model,
		client:  &http.Client{Timeout: 90 * time.Second},
	}

	// Custom BaseURL => HTTP transport (tests, proxies). Otherwise official SDK + API key.
	if baseURL != "" && baseURL != defaultGeminiAPIBase {
		c.useHTTP = true
		return c
	}

	if cfg.APIKey == "" {
		return c
	}

	client, err := genai.NewClient(context.Background(), option.WithAPIKey(cfg.APIKey))
	if err != nil {
		c.useHTTP = true
		c.baseURL = defaultGeminiAPIBase
		return c
	}
	c.genai = client
	return c
}

type geminiClient struct {
	apiKey  string
	baseURL string
	model   string
	useHTTP bool
	client  *http.Client
	genai   *genai.Client
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
	Text       string            `json:"text,omitempty"`
	InlineData *geminiInlineData `json:"inlineData,omitempty"`
}

type geminiInlineData struct {
	MIMEType string `json:"mimeType"`
	Data     string `json:"data"`
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

func (c *geminiClient) Generate(ctx context.Context, in TextGenerateInput) (Result, error) {
	in = EnsureVisionPrompt(in)
	turns := MergePromptAndMessages(in.Messages, in.Prompt)
	multiTurn := len(turns) > 1
	messages := PrependSystem(turns, englishSystemPrompt(in.Category, multiTurn))

	models := geminiModelsToTry(c.model)
	var primaryErr error
	var lastErr error
	for i, model := range models {
		var res Result
		var err error
		if c.useHTTP {
			res, err = c.generateWithHTTP(ctx, model, messages, in.ImageData, in.ImageMIME)
		} else {
			res, err = c.generateWithSDK(ctx, model, messages, in.ImageData, in.ImageMIME)
		}
		if err == nil {
			res.TokensUsed = max64(res.TokensUsed, TokenCostFor(ModelGeminiFlash))
			return res, nil
		}
		if i == 0 {
			primaryErr = err
		}
		lastErr = err
		if !isRetryableGeminiError(err) {
			return Result{}, err
		}
	}
	if primaryErr != nil {
		return Result{}, primaryErr
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
	msg := geminiErrorMessage(err)
	if msg == "" {
		return false
	}
	lower := strings.ToLower(msg)
	if isGeminiModelNotFoundMessage(lower) {
		return true
	}
	return strings.Contains(lower, "quota") ||
		strings.Contains(lower, "rate") ||
		strings.Contains(lower, "429") ||
		strings.Contains(lower, "resource_exhausted")
}

func geminiErrorMessage(err error) string {
	var pe *ProviderError
	if errors.As(err, &pe) {
		return pe.Message
	}
	if err != nil {
		return err.Error()
	}
	return ""
}

func isGeminiModelNotFoundMessage(lower string) bool {
	return strings.Contains(lower, "not found") ||
		strings.Contains(lower, "404") ||
		strings.Contains(lower, "is not supported for generatecontent")
}

func (c *geminiClient) generateWithSDK(ctx context.Context, model string, messages []ChatMessage, imageData []byte, imageMIME string) (Result, error) {
	if c.genai == nil {
		return Result{}, newProviderError("gemini", "SDK client is not initialized")
	}

	system, contents := geminiDialogFromMessages(messages)
	if len(contents) == 0 {
		return Result{}, newProviderError("gemini", "empty dialog")
	}
	attachGeminiImage(&contents[len(contents)-1], imageData, imageMIME)

	gm := c.genai.GenerativeModel(model)
	gm.SystemInstruction = &genai.Content{
		Parts: []genai.Part{genai.Text(system)},
	}

	var resp *genai.GenerateContentResponse
	var err error

	lastParts := geminiPartsToGenAI(contents[len(contents)-1].Parts)
	if len(lastParts) == 0 {
		return Result{}, newProviderError("gemini", "empty request (no text or image)")
	}

	if len(contents) == 1 {
		resp, err = gm.GenerateContent(ctx, lastParts...)
	} else {
		history := make([]*genai.Content, 0, len(contents)-1)
		for _, item := range contents[:len(contents)-1] {
			history = append(history, geminiContentToGenAI(item))
		}
		chat := gm.StartChat()
		chat.History = history
		resp, err = chat.SendMessage(ctx, lastParts...)
	}

	if err != nil {
		return Result{}, newProviderError("gemini", enhanceGeminiQuotaMessage(err.Error()))
	}

	text, tokens := extractGenAIResponse(resp)
	if text == "" {
		return Result{}, newProviderError("gemini", "empty text from model "+model)
	}
	return Result{Text: text, TokensUsed: tokens}, nil
}

func geminiContentToGenAI(c geminiContent) *genai.Content {
	role := c.Role
	if role == "" {
		role = "user"
	}
	return &genai.Content{Role: role, Parts: geminiPartsToGenAI(c.Parts)}
}

func geminiPartsToGenAI(parts []geminiPart) []genai.Part {
	out := make([]genai.Part, 0, len(parts))
	for _, p := range parts {
		if p.InlineData != nil && p.InlineData.Data != "" {
			raw, err := base64.StdEncoding.DecodeString(p.InlineData.Data)
			if err == nil {
				mime := p.InlineData.MIMEType
				if mime == "" {
					mime = "image/jpeg"
				}
				out = append(out, genai.ImageData(mime, raw))
			}
		}
		if strings.TrimSpace(p.Text) != "" {
			out = append(out, genai.Text(p.Text))
		}
	}
	return out
}

func attachGeminiImage(content *geminiContent, imageData []byte, imageMIME string) {
	if content == nil || len(imageData) == 0 {
		return
	}
	if imageMIME == "" {
		imageMIME = "image/jpeg"
	}
	imagePart := geminiPart{
		InlineData: &geminiInlineData{
			MIMEType: imageMIME,
			Data:     base64.StdEncoding.EncodeToString(imageData),
		},
	}
	hasText := false
	for _, p := range content.Parts {
		if strings.TrimSpace(p.Text) != "" {
			hasText = true
			break
		}
	}
	if hasText {
		content.Parts = append([]geminiPart{imagePart}, content.Parts...)
	} else {
		content.Parts = append([]geminiPart{imagePart}, geminiPart{Text: defaultVisionUserPrompt})
	}
}

func extractGenAIResponse(resp *genai.GenerateContentResponse) (text string, tokens int64) {
	if resp == nil || len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
		return "", 0
	}
	var b strings.Builder
	for _, part := range resp.Candidates[0].Content.Parts {
		switch v := part.(type) {
		case genai.Text:
			b.WriteString(string(v))
		default:
			b.WriteString(fmt.Sprint(v))
		}
	}
	if resp.UsageMetadata != nil {
		tokens = int64(resp.UsageMetadata.TotalTokenCount)
	}
	return strings.TrimSpace(b.String()), tokens
}

func (c *geminiClient) generateWithHTTP(ctx context.Context, model string, messages []ChatMessage, imageData []byte, imageMIME string) (Result, error) {
	system, contents := geminiDialogFromMessages(messages)
	if len(contents) == 0 {
		return Result{}, newProviderError("gemini", "empty dialog")
	}
	attachGeminiImage(&contents[len(contents)-1], imageData, imageMIME)

	body, err := json.Marshal(geminiGenerateRequest{
		SystemInstruction: &geminiContent{Parts: []geminiPart{{Text: system}}},
		Contents:          contents,
	})
	if err != nil {
		return Result{}, err
	}

	baseURL := c.baseURL
	if baseURL == "" {
		baseURL = defaultGeminiAPIBase
	}

	url := fmt.Sprintf("%s/models/%s:generateContent", baseURL, model)
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

	var textParts []string
	for _, part := range parsed.Candidates[0].Content.Parts {
		if t := strings.TrimSpace(part.Text); t != "" {
			textParts = append(textParts, t)
		}
	}
	text := strings.TrimSpace(strings.Join(textParts, "\n"))
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
