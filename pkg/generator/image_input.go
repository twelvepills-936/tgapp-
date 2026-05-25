package generator

import (
	"strings"
	"unicode/utf8"
)

// ImageGenerateInput is passed to image providers.
type ImageGenerateInput struct {
	Prompt   string
	Category string
	Messages []ChatMessage
}

const imageQualityTail = ". Профессиональная иллюстрация: продуманная композиция и ракурс, " +
	"выразительный свет и тени, богатые детали и текстуры, цельная палитра, глубина кадра, " +
	"без текста, UI, логотипов, водяных знаков, артефактов и искажённых лиц."

var imageAssistantPlaceholders = []string{
	"изображение создано",
	"image created",
	"image generated",
	"изображение сгенерировано",
}

// BuildImagePrompt merges multi-turn chat into one cohesive scene description for image APIs.
func BuildImagePrompt(in ImageGenerateInput) string {
	prepared := PrepareTextInput(TextGenerateInput{
		Prompt:   in.Prompt,
		Category: in.Category,
		Messages: in.Messages,
	})

	current := strings.TrimSpace(prepared.Prompt)
	var priorUsers []string
	for _, m := range prepared.Messages {
		if m.Role != "user" {
			continue
		}
		c := strings.TrimSpace(m.Content)
		if c == "" || isImageAssistantPlaceholder(c) {
			continue
		}
		priorUsers = append(priorUsers, c)
	}

	if len(priorUsers) == 0 {
		return enhanceStandaloneImagePrompt(current)
	}

	// Avoid duplicating the current prompt if the client already sent it as the last user turn.
	if len(priorUsers) > 0 && priorUsers[len(priorUsers)-1] == current {
		priorUsers = priorUsers[:len(priorUsers)-1]
	}

	if len(priorUsers) == 0 {
		return enhanceStandaloneImagePrompt(current)
	}

	scene := strings.Join(priorUsers, ". ")
	if current == "" {
		return enhanceStandaloneImagePrompt(scene)
	}

	if isImageFollowUpPrompt(current) {
		return enhanceStandaloneImagePrompt(scene + ". " + formatImageFollowUp(current))
	}

	return enhanceStandaloneImagePrompt(scene + ". " + current)
}

func isImageAssistantPlaceholder(s string) bool {
	lower := strings.ToLower(strings.TrimSpace(s))
	for _, p := range imageAssistantPlaceholders {
		if lower == p || strings.HasPrefix(lower, p) {
			return true
		}
	}
	return false
}

func isImageFollowUpPrompt(s string) bool {
	lower := strings.ToLower(strings.TrimSpace(s))
	if lower == "" {
		return false
	}

	keywords := []string{
		"передел", "измени", "добав", "убери", "сделай", "ещё", "еще", "заново",
		"другой", "другую", "поправ", "исправ", "вариант", "повтор", "уточн",
		"подправ", "перерис", "дополн", "усиль", "улучш", "ярче", "темнее", "светлее",
		"refine", "redo", "change", "modify", "again", "instead", "more ", "less ",
	}
	for _, k := range keywords {
		if strings.Contains(lower, k) {
			return true
		}
	}
	return utf8.RuneCountInString(lower) < 48
}

func formatImageFollowUp(current string) string {
	current = strings.TrimSpace(current)
	if current == "" {
		return ""
	}
	lower := strings.ToLower(current)
	if strings.HasPrefix(lower, "уточнение") || strings.HasPrefix(lower, "изменение") {
		return current
	}
	return "Сохрани общую сцену и стиль. Изменения: " + current + ". Пересобери кадр с учётом правок, не начинай с нуля."
}

func enhanceStandaloneImagePrompt(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return p
	}
	lower := strings.ToLower(p)
	if strings.HasPrefix(lower, "детализирован") || strings.HasPrefix(lower, "качествен") {
		if !strings.Contains(lower, "водяных") {
			return p + imageQualityTail
		}
		return p
	}
	return "Создай изображение по ТЗ. " + p + imageQualityTail
}
