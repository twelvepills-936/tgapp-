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

// DeepSeek via Yandex AI Studio Responses API (OpenAI-compatible).
// https://yandex.cloud/en/docs/ai-studio/concepts/responses

func newDeepSeekClient(cfg config.ConfigYandex) *deepSeekClient {
	model := strings.TrimSpace(cfg.DeepSeekModel)
	if model == "" {
		model = "deepseek-v32/latest"
	}
	modelURI := fmt.Sprintf("gpt://%s/%s", cfg.FolderID, model)

	baseURL := strings.TrimRight(cfg.ResponsesBaseURL, "/")
	if baseURL == "" {
		baseURL = "https://ai.api.cloud.yandex.net/v1"
	}

	maxOut := cfg.TextMaxOutputTokens
	if maxOut <= 0 {
		maxOut = DefaultMaxOutputTokens
	}

	return &deepSeekClient{
		apiKey:          cfg.APIKey,
		folderID:        cfg.FolderID,
		modelURI:        modelURI,
		baseURL:         baseURL,
		maxOutputTokens: maxOut,
		client:          &http.Client{Timeout: 90 * time.Second},
	}
}

type deepSeekClient struct {
	apiKey          string
	folderID        string
	modelURI        string
	baseURL         string
	maxOutputTokens int
	client          *http.Client
}

type yandexResponsesRequest struct {
	Model           string  `json:"model"`
	Temperature     float64 `json:"temperature"`
	Instructions    string  `json:"instructions,omitempty"`
	Input           string  `json:"input"`
	MaxOutputTokens int     `json:"max_output_tokens"`
}

type yandexResponsesResponse struct {
	Output []struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	} `json:"output"`
	Error json.RawMessage `json:"error"`
}

func (c *deepSeekClient) Generate(ctx context.Context, in TextGenerateInput) (Result, error) {
	in = PrepareTextInput(in)
	turns := MergePromptAndMessages(in.Messages, in.Prompt)
	multiTurn := len(turns) > 1
	instructions, input := deepSeekPromptFromMessages(turns, in.Category, multiTurn)
	if input == "" {
		return Result{}, newProviderError("deepseek", "empty prompt")
	}

	body, err := json.Marshal(yandexResponsesRequest{
		Model:           c.modelURI,
		Temperature:     0.2,
		Instructions:    instructions,
		Input:           input,
		MaxOutputTokens: c.maxOutputTokens,
	})
	if err != nil {
		return Result{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/responses", bytes.NewReader(body))
	if err != nil {
		return Result{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Api-Key "+c.apiKey)
	req.Header.Set("OpenAI-Project", c.folderID)

	resp, err := c.client.Do(req)
	if err != nil {
		return Result{}, newProviderError("deepseek", err.Error())
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return Result{}, err
	}

	if msg := parseYandexAPIError(raw); msg != "" {
		return Result{}, newProviderError("deepseek", msg)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Result{}, newProviderError("deepseek", fmt.Sprintf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(raw))))
	}

	var parsed yandexResponsesResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return Result{}, newProviderError("deepseek", fmt.Sprintf("invalid response: %s", strings.TrimSpace(string(raw))))
	}

	text := strings.TrimSpace(parsed.outputText())
	if text == "" {
		return Result{}, newProviderError("deepseek", "empty text in response")
	}

	return Result{
		Text:       text,
		TokensUsed: TokenCostFor(ModelDeepSeek),
	}, nil
}

func (r yandexResponsesResponse) outputText() string {
	for _, output := range r.Output {
		for _, part := range output.Content {
			if t := strings.TrimSpace(part.Text); t != "" {
				return t
			}
		}
	}
	return ""
}

func deepSeekPromptFromMessages(messages []ChatMessage, category string, multiTurn bool) (instructions, input string) {
	instructions = textSystemPrompt(category, multiTurn)

	var transcript strings.Builder
	for _, m := range messages {
		switch m.Role {
		case "system":
			if instructions == "" {
				instructions = m.Content
			} else {
				instructions += "\n" + m.Content
			}
		case "assistant":
			transcript.WriteString("Assistant: ")
			transcript.WriteString(m.Content)
			transcript.WriteString("\n\n")
		default:
			transcript.WriteString("User: ")
			transcript.WriteString(m.Content)
			transcript.WriteString("\n\n")
		}
	}
	return strings.TrimSpace(instructions), strings.TrimSpace(transcript.String())
}
