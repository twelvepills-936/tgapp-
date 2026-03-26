package logger

import (
	"log/slog"
)

// ErrorAttr creates a slog attribute for an error.
// Replaces gitlab16.skiftrade.kz/libs-go/logger.ErrorAttr
func ErrorAttr(err error) slog.Attr {
	if err == nil {
		return slog.Attr{}
	}
	return slog.Any("error", err)
}

// InputAttr creates a slog attribute for input values.
// Replaces gitlab16.skiftrade.kz/libs-go/logger.InputAttr
func InputAttr(key string, value any) slog.Attr {
	return slog.Any(key, value)
}

// StringAttr creates a slog attribute for string values.
func StringAttr(key, value string) slog.Attr {
	return slog.String(key, value)
}

// IntAttr creates a slog attribute for int values.
func IntAttr(key string, value int) slog.Attr {
	return slog.Int(key, value)
}

// Int64Attr creates a slog attribute for int64 values.
func Int64Attr(key string, value int64) slog.Attr {
	return slog.Int64(key, value)
}

