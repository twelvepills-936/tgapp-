package config

import (
	"os"
	"testing"
)

func TestLoadAppConfig_PORTWinsOverAPPHTTPPort(t *testing.T) {
	t.Setenv("PORT", "8080")
	t.Setenv("APP_HTTP_PORT", "8090")

	cfg := LoadAppConfig()
	if cfg.HTTPPort != 8080 {
		t.Fatalf("HTTPPort = %d, want 8080 (PORT must override APP_HTTP_PORT)", cfg.HTTPPort)
	}
}

func TestLoadAppConfig_APPHTTPPortWhenNoPORT(t *testing.T) {
	os.Unsetenv("PORT")
	t.Setenv("APP_HTTP_PORT", "8090")

	cfg := LoadAppConfig()
	if cfg.HTTPPort != 8090 {
		t.Fatalf("HTTPPort = %d, want 8090", cfg.HTTPPort)
	}
}
