package generator

import "strings"

func enhanceGeminiQuotaMessage(msg string) string {
	lower := strings.ToLower(msg)
	if strings.Contains(lower, "limit: 0") || strings.Contains(lower, "quota exceeded") {
		return "Бесплатная квота Gemini не активна (limit: 0). " +
			"Привяжите биллинг в Google AI Studio: https://aistudio.google.com/apikey — " +
			"или используйте model=yandexgpt. Детали: " + msg
	}
	if strings.Contains(lower, "location is not supported") {
		return "Gemini недоступен в вашем регионе. Используйте model=yandexgpt или VPN. Детали: " + msg
	}
	return msg
}
