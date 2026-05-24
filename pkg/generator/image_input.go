package generator

import "strings"

// ImageGenerateInput is passed to image providers.
type ImageGenerateInput struct {
	Prompt   string
	Category string
	Messages []ChatMessage
}

// BuildImagePrompt merges chat history into one prompt for image APIs (single prompt field).
func BuildImagePrompt(in ImageGenerateInput) string {
	prepared := PrepareTextInput(TextGenerateInput{
		Prompt:   in.Prompt,
		Category: in.Category,
		Messages: in.Messages,
	})

	prompt := prepared.Prompt
	history := prepared.Messages
	if len(history) == 0 {
		return prompt
	}

	var b strings.Builder
	b.WriteString("Conversation context for this image request:\n\n")
	for _, m := range history {
		switch m.Role {
		case "assistant":
			b.WriteString("Previous result: ")
		default:
			b.WriteString("User request: ")
		}
		b.WriteString(m.Content)
		b.WriteString("\n\n")
	}
	b.WriteString("Current request: ")
	b.WriteString(prompt)
	return b.String()
}
