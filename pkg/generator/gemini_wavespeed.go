package generator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const defaultWaveSpeedBase = "https://llm.wavespeed.ai/v1"

const defaultWaveSpeedModel = "google/gemini-2.0-flash-001"

func isWaveSpeedBaseURL(base string) bool {
	b := strings.ToLower(strings.TrimSpace(base))
	return strings.Contains(b, "wavespeed.ai")
}

func isWaveSpeedMode(baseURL, model string) bool {
	if isWaveSpeedBaseURL(baseURL) {
		return true
	}
	return strings.HasPrefix(strings.TrimSpace(model), "google/")
}

func normalizeWaveSpeedModel(model string) string {
	model = strings.TrimSpace(model)
	if model == "" || model == ModelGeminiFlash {
		return defaultWaveSpeedModel
	}
	return model
}

func (c *geminiClient) generateWithChatCompletions(ctx context.Context, model string, messages []ChatMessage, imageData []byte, imageMIME string) (Result, error) {
	var body []byte
	var err error
	if len(imageData) > 0 {
		oaiMessages, buildErr := buildChatCompletionMessages(messages, imageData, imageMIME)
		if buildErr != nil {
			return Result{}, buildErr
		}
		body, err = json.Marshal(map[string]any{
			"model":    model,
			"messages": oaiMessages,
		})
	} else {
		chatMessages := make([]chatMessage, 0, len(messages))
		for _, m := range messages {
			role := m.Role
			if role == "model" {
				role = "assistant"
			}
			chatMessages = append(chatMessages, chatMessage{Role: role, Content: m.Content})
		}
		body, err = json.Marshal(chatRequest{
			Model:    model,
			Messages: chatMessages,
		})
	}
	if err != nil {
		return Result{}, err
	}

	baseURL := c.baseURL
	if baseURL == "" {
		baseURL = defaultWaveSpeedBase
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return Result{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return Result{}, newProviderError("gemini", err.Error())
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
		return Result{}, newProviderError("gemini", enhanceGeminiQuotaMessage(parsed.Error.Message))
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg := strings.TrimSpace(string(raw))
		if parsed.Error != nil && parsed.Error.Message != "" {
			msg = parsed.Error.Message
		}
		return Result{}, newProviderError("gemini", enhanceGeminiQuotaMessage(fmt.Sprintf("HTTP %d: %s", resp.StatusCode, msg)))
	}
	if len(parsed.Choices) == 0 || strings.TrimSpace(parsed.Choices[0].Message.Content) == "" {
		return Result{}, newProviderError("gemini", "empty response from model "+model)
	}

	var tokens int64
	if parsed.Usage != nil {
		tokens = parsed.Usage.TotalTokens
	}
	return Result{
		Text:       strings.TrimSpace(parsed.Choices[0].Message.Content),
		TokensUsed: tokens,
	}, nil
}

func buildChatCompletionMessages(messages []ChatMessage, imageData []byte, imageMIME string) ([]map[string]any, error) {
	out := make([]map[string]any, 0, len(messages))
	attachImage := len(imageData) > 0
	lastIdx := len(messages) - 1
	for i, m := range messages {
		role := m.Role
		if role == "model" {
			role = "assistant"
		}
		isLastUser := attachImage && role == "user" && i == lastIdx
		if isLastUser {
			text := m.Content
			if strings.TrimSpace(text) == "" {
				text = defaultVisionUserPrompt
			}
			out = append(out, map[string]any{
				"role": "user",
				"content": []map[string]any{
					{"type": "text", "text": text},
					{"type": "image_url", "image_url": map[string]string{"url": imageDataURL(imageMIME, imageData)}},
				},
			})
			continue
		}
		out = append(out, map[string]any{"role": role, "content": m.Content})
	}
	return out, nil
}
