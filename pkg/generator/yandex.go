package generator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gitlab16.skiftrade.kz/templates/go/pkg/config"
)

func newYandexClient(cfg config.ConfigYandex) *yandexClient {
	modelURI := cfg.ModelURI
	if modelURI == "" && cfg.FolderID != "" {
		model := cfg.Model
		if model == "" {
			model = "yandexgpt/latest"
		}
		modelURI = fmt.Sprintf("gpt://%s/%s", cfg.FolderID, model)
	}
	maxOut := cfg.TextMaxOutputTokens
	if maxOut <= 0 {
		maxOut = DefaultMaxOutputTokens
	}

	return &yandexClient{
		apiKey:    cfg.APIKey,
		folderID:  cfg.FolderID,
		modelURI:  modelURI,
		baseURL:   strings.TrimRight(cfg.BaseURL, "/"),
		maxTokens: maxOut,
		client:    &http.Client{Timeout: 90 * time.Second},
	}
}

type yandexClient struct {
	apiKey    string
	folderID  string
	modelURI  string
	baseURL   string
	maxTokens int
	client    *http.Client
}

type yandexCompletionRequest struct {
	ModelURI          string                 `json:"modelUri"`
	CompletionOptions yandexCompletionOpts   `json:"completionOptions"`
	Messages          []yandexMessage        `json:"messages"`
}

type yandexCompletionOpts struct {
	Stream      bool    `json:"stream"`
	Temperature float64 `json:"temperature"`
	MaxTokens   int     `json:"maxTokens"`
}

type yandexMessage struct {
	Role string `json:"role"`
	Text string `json:"text"`
}

type yandexCompletionResponse struct {
	Result struct {
		Alternatives []struct {
			Message struct {
				Text string `json:"text"`
			} `json:"message"`
		} `json:"alternatives"`
		Usage struct {
			InputTextTokens  json.RawMessage `json:"inputTextTokens"`
			CompletionTokens json.RawMessage `json:"completionTokens"`
			TotalTokens      json.RawMessage `json:"totalTokens"`
		} `json:"usage"`
	} `json:"result"`
}

func (c *yandexClient) Generate(ctx context.Context, in TextGenerateInput) (Result, error) {
	in = PrepareTextInput(in)

	baseURL := c.baseURL
	if baseURL == "" {
		baseURL = "https://llm.api.cloud.yandex.net/foundationModels/v1"
	}

	turns := MergePromptAndMessages(in.Messages, in.Prompt)
	multiTurn := len(turns) > 1
	messages := PrependSystem(turns, textSystemPrompt(in.Category, multiTurn))

	yandexMessages := make([]yandexMessage, 0, len(messages))
	for _, m := range messages {
		yandexMessages = append(yandexMessages, yandexMessage{Role: m.Role, Text: m.Content})
	}

	body, err := json.Marshal(yandexCompletionRequest{
		ModelURI: c.modelURI,
		CompletionOptions: yandexCompletionOpts{
			Stream:      false,
			Temperature: 0.3,
			MaxTokens:   c.maxTokens,
		},
		Messages: yandexMessages,
	})
	if err != nil {
		return Result{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/completion", bytes.NewReader(body))
	if err != nil {
		return Result{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Api-Key "+c.apiKey)
	req.Header.Set("x-folder-id", c.folderID)

	resp, err := c.client.Do(req)
	if err != nil {
		return Result{}, newProviderError("yandexgpt", err.Error())
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return Result{}, err
	}

	if msg := parseYandexAPIError(raw); msg != "" {
		return Result{}, newProviderError("yandexgpt", msg)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Result{}, newProviderError("yandexgpt", fmt.Sprintf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(raw))))
	}

	var parsed yandexCompletionResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return Result{}, newProviderError("yandexgpt", fmt.Sprintf("invalid response: %s", strings.TrimSpace(string(raw))))
	}
	if len(parsed.Result.Alternatives) == 0 {
		return Result{}, newProviderError("yandexgpt", "empty response")
	}

	text := strings.TrimSpace(parsed.Result.Alternatives[0].Message.Text)
	if text == "" {
		return Result{}, newProviderError("yandexgpt", "empty text")
	}

	tokens := parseYandexTokens(
		string(parsed.Result.Usage.TotalTokens),
		string(parsed.Result.Usage.InputTextTokens),
		string(parsed.Result.Usage.CompletionTokens),
	)
	if tokens == 0 {
		tokens = TokenCostFor(ModelYandexGPT)
	}

	return Result{Text: text, TokensUsed: tokens}, nil
}

func parseYandexAPIError(raw []byte) string {
	var envelope struct {
		Error json.RawMessage `json:"error"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil || len(envelope.Error) == 0 {
		return ""
	}
	if string(envelope.Error) == "null" {
		return ""
	}

	var asString string
	if err := json.Unmarshal(envelope.Error, &asString); err == nil && asString != "" {
		return asString
	}

	var asObject struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(envelope.Error, &asObject); err == nil && asObject.Message != "" {
		return asObject.Message
	}

	return strings.TrimSpace(string(envelope.Error))
}

func parseYandexTokens(values ...string) int64 {
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
			return n
		}
	}
	return 0
}
