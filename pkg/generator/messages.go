package generator

import "strings"

// ChatMessage is one turn in a multi-turn conversation.
type ChatMessage struct {
	Role    string
	Content string
}

// TextGenerateInput is passed to text providers.
type TextGenerateInput struct {
	Prompt   string
	Category string
	Messages []ChatMessage
}

// MergePromptAndMessages builds the final message list for providers.
// If messages is empty, prompt becomes a single user turn.
// If both are set, prompt is appended unless it duplicates the last user message.
func MergePromptAndMessages(messages []ChatMessage, prompt string) []ChatMessage {
	out := make([]ChatMessage, 0, len(messages)+1)
	for _, m := range messages {
		role := NormalizeChatRole(m.Role)
		content := strings.TrimSpace(m.Content)
		if role == "" || content == "" {
			continue
		}
		out = append(out, ChatMessage{Role: role, Content: content})
	}

	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		return out
	}
	if len(out) > 0 {
		last := out[len(out)-1]
		if last.Role == "user" && last.Content == prompt {
			return out
		}
	}
	return append(out, ChatMessage{Role: "user", Content: prompt})
}

// NormalizeChatRole maps client roles to provider roles.
func NormalizeChatRole(role string) string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "system":
		return "system"
	case "user", "human":
		return "user"
	case "assistant", "model", "ai", "bot":
		return "assistant"
	default:
		return ""
	}
}

func textSystemPrompt(category string, multiTurn bool) string {
	system := "Ты помощник Telegram mini app. Пиши готовый текст на языке запроса пользователя."
	if multiTurn {
		system += " Учитывай всю предыдущую переписку: сохраняй тему, стиль и факты из диалога; не начинай ответ с нуля, если пользователь продолжает беседу."
	}
	if category != "" {
		system += " Категория: " + category + "."
	}
	return system
}

func englishSystemPrompt(category string, multiTurn bool) string {
	system := "You are a helpful assistant for a Telegram mini app. Generate concise, ready-to-use text in the same language as the user prompt."
	if multiTurn {
		system += " Follow the full conversation: keep topic, tone, and prior facts consistent."
	}
	if category != "" {
		system += " Content category: " + category + "."
	}
	return system
}

// PrependSystem inserts a system message unless the client already sent one.
func PrependSystem(messages []ChatMessage, system string) []ChatMessage {
	for _, m := range messages {
		if m.Role == "system" {
			return messages
		}
	}
	out := make([]ChatMessage, 0, len(messages)+1)
	out = append(out, ChatMessage{Role: "system", Content: system})
	return append(out, messages...)
}
