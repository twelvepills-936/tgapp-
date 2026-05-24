package generator

import "strings"

const (
	// DefaultMaxOutputTokens is the completion cap for Yandex / DeepSeek text calls.
	DefaultMaxOutputTokens = 4096
	// MaxContextMessages limits turns sent to providers (user+assistant pairs).
	MaxContextMessages = 24
)

// TrimChatContent shortens text for API context while keeping start and end.
func TrimChatContent(s string, maxLen int) string {
	if maxLen <= 0 || len(s) <= maxLen {
		return s
	}
	const marker = "\n\n…[сокращено]…\n\n"
	if maxLen <= len(marker)+100 {
		return s[:maxLen]
	}
	head := 600
	tail := maxLen - head - len(marker)
	if tail < 400 {
		head = 0
		tail = maxLen - len(marker)
	}
	if tail > len(s) {
		tail = len(s)
	}
	return s[:head] + marker + s[len(s)-tail:]
}

// PrepareTextInput trims prompt and chat history before calling a provider.
func PrepareTextInput(in TextGenerateInput) TextGenerateInput {
	const (
		maxPromptBytes        = 16000
		maxHistoryUserBytes   = 12000
		maxHistoryAssistant   = 6000
	)

	in.Prompt = strings.TrimSpace(TrimChatContent(in.Prompt, maxPromptBytes))
	if len(in.Messages) > MaxContextMessages {
		in.Messages = in.Messages[len(in.Messages)-MaxContextMessages:]
	}
	for i := range in.Messages {
		limit := maxHistoryUserBytes
		if in.Messages[i].Role == "assistant" {
			limit = maxHistoryAssistant
		}
		in.Messages[i].Content = strings.TrimSpace(TrimChatContent(in.Messages[i].Content, limit))
	}
	return in
}
