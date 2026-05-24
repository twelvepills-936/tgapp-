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
	for _, part := range []string{"нарисуй кота", "сделай фон синим", "Current request:"} {
		if !strings.Contains(got, part) {
			t.Fatalf("missing %q in %q", part, got)
		}
	}
}
