package httphandler

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"gitlab16.skiftrade.kz/templates/go/errorcodes"
	"gitlab16.skiftrade.kz/templates/go/internal"
	ucModels "gitlab16.skiftrade.kz/templates/go/internal/usecase/models"
	"gitlab16.skiftrade.kz/templates/go/pkg/generator"
)

// GenerateImageHandler serves POST /v1/generate/image.
type GenerateImageHandler struct {
	uc internal.UseCase
}

func NewGenerateImageHandler(uc internal.UseCase) *GenerateImageHandler {
	return &GenerateImageHandler{uc: uc}
}

type generateImageRequest struct {
	TelegramID      string               `json:"telegramId"`
	TelegramIDSnake string               `json:"telegram_id"`
	Prompt          string               `json:"prompt"`
	Category        string               `json:"category"`
	Model           string               `json:"model"`
	Messages        []chatMessageRequest `json:"messages"`
	SessionID       string               `json:"sessionId"`
	SessionIDSnake  string               `json:"session_id"`
}

type generateImageData struct {
	ImageURL   string `json:"imageUrl"`
	Model      string `json:"model"`
	TokensUsed int64  `json:"tokensUsed"`
}

type generateImageResponse struct {
	Data generateImageData `json:"data"`
}

func (h *GenerateImageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{
			Code:    errorcodes.InvalidArgument,
			Message: "use POST /v1/generate/image with JSON body (got " + r.Method + ")",
		})
		return
	}

	var req generateImageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Code:    errorcodes.InvalidArgument,
			Message: "invalid JSON body",
		})
		return
	}

	telegramID := strings.TrimSpace(req.TelegramID)
	if telegramID == "" {
		telegramID = strings.TrimSpace(req.TelegramIDSnake)
	}

	messages := make([]ucModels.ChatMessageInput, 0, len(req.Messages))
	for _, m := range req.Messages {
		messages = append(messages, ucModels.ChatMessageInput{
			Role:    strings.TrimSpace(m.Role),
			Content: strings.TrimSpace(m.Content),
		})
	}

	sessionID := strings.TrimSpace(req.SessionID)
	if sessionID == "" {
		sessionID = strings.TrimSpace(req.SessionIDSnake)
	}

	out, err := h.uc.GenerateImage(r.Context(), ucModels.GenerateImageInput{
		TelegramID: telegramID,
		Prompt:     strings.TrimSpace(req.Prompt),
		Category:   strings.TrimSpace(req.Category),
		Model:      strings.TrimSpace(req.Model),
		Messages:   messages,
		SessionID:  sessionID,
	})
	if err != nil {
		slog.ErrorContext(r.Context(), "generate image failed", slog.String("error", err.Error()))
		writeGenerateImageError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, generateImageResponse{Data: generateImageData{
		ImageURL:   out.ImageURL,
		Model:      out.Model,
		TokensUsed: out.TokensUsed,
	}})
}

func writeGenerateImageError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, generator.ErrImageGeneratorUnavailable):
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{
			Code:    errorcodes.AIProviderError,
			Message: "image generation is not configured (set WAVESPEED_API_KEY or GEMINI_API_KEY for nano-banana, or YANDEX_GPT_* for alice-ai-art)",
		})
	case errors.Is(err, ucModels.ErrProfileNotFound):
		writeJSON(w, http.StatusNotFound, errorResponse{
			Code:    errorcodes.ProfileNotFound,
			Message: ucModels.ErrProfileNotFound.Error(),
		})
	case errors.Is(err, ucModels.ErrInsufficientBalance):
		writeJSON(w, http.StatusPaymentRequired, errorResponse{
			Code:    errorcodes.InsufficientBalance,
			Message: ucModels.ErrInsufficientBalance.Error(),
		})
	case errors.Is(err, ucModels.ErrInvalidInput), errors.Is(err, generator.ErrUnsupportedModel):
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Code:    errorcodes.InvalidArgument,
			Message: err.Error(),
		})
	case errors.Is(err, generator.ErrContentPolicy):
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Code:    errorcodes.ContentPolicy,
			Message: aiProviderErrorMessage(err),
		})
	case errors.Is(err, generator.ErrProvider):
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{
			Code:    errorcodes.AIProviderError,
			Message: aiProviderErrorMessage(err),
		})
	default:
		msg := ucModels.ErrInternalServerError.Error()
		if os.Getenv("ENVIRONMENT") == "development" {
			msg = err.Error()
		}
		writeJSON(w, http.StatusInternalServerError, errorResponse{
			Code:    errorcodes.Internal,
			Message: msg,
		})
	}
}
