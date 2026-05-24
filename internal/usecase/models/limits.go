package models

// Limits for text generation API (bytes, UTF-8).
const (
	MaxGeneratePromptBytes   = 16000
	MaxGenerateMessageBytes  = 16000
	MaxGenerateCategoryBytes = 100
	MaxGenerateModelBytes    = 50
	MaxGenerateMessagesCount = 40
)
