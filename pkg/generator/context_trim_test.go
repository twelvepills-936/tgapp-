package generator

import (
	"strings"
	"testing"
)

func TestTrimChatContent_shortUnchanged(t *testing.T) {
	s := "hello"
	if got := TrimChatContent(s, 100); got != s {
		t.Fatalf("got %q", got)
	}
}

func TestTrimChatContent_longKeepsEnds(t *testing.T) {
	s := strings.Repeat("a", 500) + "MIDDLE" + strings.Repeat("z", 500)
	got := TrimChatContent(s, 200)
	if len(got) > 200 {
		t.Fatalf("len = %d", len(got))
	}
	if !strings.Contains(got, "MIDDLE") && !strings.Contains(got, "сокращено") {
		t.Fatalf("unexpected trim: len=%d", len(got))
	}
}

func TestPrepareTextInput_trimsLongAssistant(t *testing.T) {
	long := strings.Repeat("x", 8000)
	out := PrepareTextInput(TextGenerateInput{
		Prompt: "new question",
		Messages: []ChatMessage{
			{Role: "user", Content: "first"},
			{Role: "assistant", Content: long},
		},
	})
	if len(out.Messages[1].Content) > 6500 {
		t.Fatalf("assistant still too long: %d", len(out.Messages[1].Content))
	}
	if out.Prompt != "new question" {
		t.Fatalf("prompt changed: %q", out.Prompt)
	}
}
