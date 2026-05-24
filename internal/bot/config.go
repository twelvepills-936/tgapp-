package bot

import "os"

// Config holds Telegram bot settings from environment.
type Config struct {
	Token         string
	WebhookURL    string
	WebhookSecret string
}

// LoadConfig reads TELEGRAM_* variables.
func LoadConfig() Config {
	return Config{
		Token:         os.Getenv("TELEGRAM_BOT_TOKEN"),
		WebhookURL:    os.Getenv("TELEGRAM_WEBHOOK_URL"),
		WebhookSecret: os.Getenv("TELEGRAM_WEBHOOK_SECRET"),
	}
}

// UseWebhook is true when a public webhook URL is configured (production).
func (c Config) UseWebhook() bool {
	return c.Token != "" && c.WebhookURL != ""
}
