package generator

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

func isPlaceholderAPIKey(key string) bool {
	key = strings.TrimSpace(key)
	return key == "" || key == "..." || key == "your-key" || key == "changeme"
}

func (c *openAIClient) hasValidAPIKey() bool {
	return !isPlaceholderAPIKey(c.apiKey)
}

func imageDataURL(mime string, data []byte) string {
	if mime == "" {
		mime = "image/jpeg"
	}
	return fmt.Sprintf("data:%s;base64,%s", mime, base64.StdEncoding.EncodeToString(data))
}

// DescribeImage returns a short caption for use by text-only models.
func (c *openAIClient) DescribeImage(ctx context.Context, imageData []byte, imageMIME string) (string, error) {
	if !c.hasValidAPIKey() {
		return "", newProviderError("openai", "OPENAI_API_KEY is not configured")
	}
	body, err := c.buildVisionChatBody(imageCaptionPrompt, imageData, imageMIME, 256)
	if err != nil {
		return "", err
	}
	text, _, err := c.postChatCompletions(ctx, body)
	return text, err
}

func (c *openAIClient) buildVisionChatBody(userText string, imageData []byte, imageMIME string, maxTokens int) ([]byte, error) {
	userContent := []map[string]any{
		{"type": "text", "text": userText},
		{"type": "image_url", "image_url": map[string]string{"url": imageDataURL(imageMIME, imageData)}},
	}
	req := map[string]any{
		"model": c.model,
		"messages": []map[string]any{
			{"role": "user", "content": userContent},
		},
	}
	if maxTokens > 0 {
		req["max_tokens"] = maxTokens
	}
	return json.Marshal(req)
}

func (c *openAIClient) buildOpenAIChatMessages(messages []ChatMessage, imageData []byte, imageMIME string) ([]map[string]any, error) {
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
			out = append(out, map[string]any{
				"role": "user",
				"content": []map[string]any{
					{"type": "text", "text": m.Content},
					{"type": "image_url", "image_url": map[string]string{"url": imageDataURL(imageMIME, imageData)}},
				},
			})
			continue
		}
		out = append(out, map[string]any{"role": role, "content": m.Content})
	}
	return out, nil
}
