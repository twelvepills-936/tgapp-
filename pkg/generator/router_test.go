package generator

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab16.skiftrade.kz/templates/go/pkg/config"
)

func TestModelRouter_Yandex(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"result":{"alternatives":[{"message":{"text":"ok"}}],"usage":{"totalTokens":"3"}}}`))
	}))
	defer srv.Close()

	r := NewModelRouter(
		config.ConfigYandex{APIKey: "k", FolderID: "f", BaseURL: srv.URL},
		config.ConfigGemini{},
		config.ConfigAI{},
	)
	res, err := r.Generate(context.Background(), ModelYandexGPT, TextGenerateInput{Prompt: "hi", Category: "text"})
	if err != nil || res.Text != "ok" {
		t.Fatalf("got %+v err=%v", res, err)
	}
}
