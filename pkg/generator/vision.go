package generator

import (
	"context"
	"strings"
)

const defaultVisionUserPrompt = "Проанализируй прикреплённое изображение и ответь пользователю."

const imageCaptionPrompt = "Кратко опиши, что изображено на картинке (1–3 предложения). " +
	"Пиши на том же языке, что и типичный запрос пользователя в чате. Только описание, без вступлений."

const imageCaptionUnavailable = "пользователь прикрепил изображение (автоописание недоступно — ответь по тексту запроса)"

// EnsureVisionPrompt sets a default text prompt when the user sends only an image.
func EnsureVisionPrompt(in TextGenerateInput) TextGenerateInput {
	if !in.HasImage() {
		return in
	}
	if strings.TrimSpace(in.Prompt) == "" {
		in.Prompt = defaultVisionUserPrompt
	}
	return in
}

func isGeminiQuotaError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "quota") ||
		strings.Contains(msg, "429") ||
		strings.Contains(msg, "limit: 0") ||
		strings.Contains(msg, "resource_exhausted")
}

// describeImage returns a text caption for an image (OpenAI vision preferred, then Gemini).
func (r *ModelRouter) describeImage(ctx context.Context, in TextGenerateInput) (string, error) {
	if r.openai != nil && r.openai.hasValidAPIKey() {
		caption, err := r.openai.DescribeImage(ctx, in.ImageData, in.ImageMIME)
		if err == nil && strings.TrimSpace(caption) != "" {
			return strings.TrimSpace(caption), nil
		}
		if err != nil && !isGeminiQuotaError(err) {
			// non-quota OpenAI errors fall through to Gemini
		}
	}

	if r.gemini != nil {
		captionIn := TextGenerateInput{
			Prompt:    imageCaptionPrompt,
			Category:  in.Category,
			ImageData: in.ImageData,
			ImageMIME: in.ImageMIME,
		}
		res, err := r.gemini.Generate(ctx, EnsureVisionPrompt(captionIn))
		if err == nil && strings.TrimSpace(res.Text) != "" {
			return strings.TrimSpace(res.Text), nil
		}
		if err != nil && isGeminiQuotaError(err) {
			return imageCaptionUnavailable, nil
		}
		if err != nil {
			return "", err
		}
	}

	return imageCaptionUnavailable, nil
}

// enrichWithImageCaption describes the image and injects text for models without native vision.
func (r *ModelRouter) enrichWithImageCaption(ctx context.Context, in TextGenerateInput) (TextGenerateInput, error) {
	caption, err := r.describeImage(ctx, in)
	if err != nil {
		return TextGenerateInput{}, err
	}
	if strings.TrimSpace(caption) == "" {
		caption = imageCaptionUnavailable
	}

	out := in
	out.ImageData = nil
	out.ImageMIME = ""
	prefix := "[Прикреплённое изображение: " + caption + "]\n\n"
	if strings.TrimSpace(out.Prompt) == "" || out.Prompt == defaultVisionUserPrompt {
		out.Prompt = prefix + "Ответь пользователю с учётом изображения."
	} else {
		out.Prompt = prefix + out.Prompt
	}
	return out, nil
}
