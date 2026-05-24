package generator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"gitlab16.skiftrade.kz/templates/go/pkg/config"
)

func newOpenAIClient(cfg config.ConfigAI) *openAIClient {
	baseURL := strings.TrimRight(cfg.BaseURL, "/")
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	model := cfg.Model
	if model == "" {
		model = "gpt-4o-mini"
	}
	return &openAIClient{
		apiKey:  cfg.APIKey,
		baseURL: baseURL,
		model:   model,
		client:  &http.Client{Timeout: 90 * time.Second},
	}
}

type openAIClient struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage *struct {
		TotalTokens int64 `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (c *openAIClient) Generate(ctx context.Context, in TextGenerateInput) (Result, error) {
	turns := MergePromptAndMessages(in.Messages, in.Prompt)
	multiTurn := len(turns) > 1
	messages := PrependSystem(turns, englishSystemPrompt(in.Category, multiTurn))

	chatMessages := make([]chatMessage, 0, len(messages))
	for _, m := range messages {
		chatMessages = append(chatMessages, chatMessage{Role: m.Role, Content: m.Content})
	}

	body, err := json.Marshal(chatRequest{
		Model:    c.model,
		Messages: chatMessages,
	})
	if err != nil {
		return Result{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return Result{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return Result{}, newProviderError("openai", err.Error())
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return Result{}, err
	}

	var parsed chatResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return Result{}, err
	}
	if parsed.Error != nil && parsed.Error.Message != "" {
		return Result{}, newProviderError("openai", parsed.Error.Message)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Result{}, newProviderError("openai", fmt.Sprintf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(raw))))
	}
	if len(parsed.Choices) == 0 || strings.TrimSpace(parsed.Choices[0].Message.Content) == "" {
		return Result{}, newProviderError("openai", "empty response")
	}

	var tokens int64
	if parsed.Usage != nil {
		tokens = parsed.Usage.TotalTokens
	}
	if tokens == 0 {
		tokens = TokenCostFor(ModelOpenAI)
	}

	return Result{
		Text:       strings.TrimSpace(parsed.Choices[0].Message.Content),
		TokensUsed: tokens,
	}, nil
}
