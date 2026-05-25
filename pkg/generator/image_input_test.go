package generator

import (
	"strings"
	"testing"
)

func TestBuildImagePrompt_withHistory(t *testing.T) {
	got := BuildImagePrompt(ImageGenerateInput{
		Prompt: "сделай фон синим",
		Messages: []ChatMessage{
			{Role: "user", Content: "нарисуй кота"},
			{Role: "assistant", Content: "Изображение создано."},
		},
	})
	if got == "" {
		t.Fatal("empty prompt")
	}
	for _, part := range []string{"нарисуй кота", "сделай фон синим", "Сохрани общую сцену"} {
		if !strings.Contains(got, part) {
			t.Fatalf("missing %q in %q", part, got)
		}
	}
	if strings.Contains(got, "Conversation context") {
		t.Fatalf("should not contain meta header: %q", got)
	}
}

func TestBuildImagePrompt_single(t *testing.T) {
	got := BuildImagePrompt(ImageGenerateInput{Prompt: "закат над морем"})
	if !strings.Contains(got, "закат над морем") {
		t.Fatalf("got %q", got)
	}
	if !strings.Contains(got, "Создай изображение") {
		t.Fatalf("expected quality prefix: %q", got)
	}
}

func TestIsImageFollowUpPrompt(t *testing.T) {
	if !isImageFollowUpPrompt("переделай с синим небом") {
		t.Fatal("expected follow-up")
	}
	if isImageFollowUpPrompt("портрет молодой женщины в стиле масляной живописи с мягким светом") {
		t.Fatal("expected full prompt")
	}
}
