package generator

import "testing"

func TestIsContentPolicy(t *testing.T) {
	err := newProviderError("alice-ai-art", "Я не могу сгенерировать это изображение. Давайте попробуем другую тему.")
	if !IsContentPolicy(err) {
		t.Fatal("expected content policy")
	}
}
