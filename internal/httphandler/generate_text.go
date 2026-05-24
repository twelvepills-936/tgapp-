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

const generateTextPath = "/v1/generate/text"

// GenerateTextHandler serves POST /v1/generate/text.
type GenerateTextHandler struct {
	uc internal.UseCase
}

func NewGenerateTextHandler(uc internal.UseCase) *GenerateTextHandler {
	return &GenerateTextHandler{uc: uc}
}

type chatMessageRequest struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type generateTextRequest struct {
	TelegramID      string               `json:"telegramId"`
	TelegramIDSnake string               `json:"telegram_id"`
	Prompt          string               `json:"prompt"`
	Category        string               `json:"category"`
	Model           string               `json:"model"`
	Messages        []chatMessageRequest `json:"messages"`
}

type generateTextData struct {
	Text       string `json:"text"`
	Model      string `json:"model"`
	TokensUsed int64  `json:"tokensUsed"`
}

type generateTextResponse struct {
	Data generateTextData `json:"data"`
}

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (h *GenerateTextHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{
			Code:    errorcodes.InvalidArgument,
			Message: "method not allowed",
		})
		return
	}

	var req generateTextRequest
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

	out, err := h.uc.GenerateText(r.Context(), ucModels.GenerateTextInput{
		TelegramID: telegramID,
		Prompt:     strings.TrimSpace(req.Prompt),
		Category:   strings.TrimSpace(req.Category),
		Model:      strings.TrimSpace(req.Model),
		Messages:   messages,
	})
	if err != nil {
		slog.ErrorContext(r.Context(), "generate text failed", slog.String("error", err.Error()))
		writeGenerateError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, generateTextResponse{Data: generateTextData{
		Text:       out.Text,
		Model:      out.Model,
		TokensUsed: out.TokensUsed,
	}})
}

func writeGenerateError(w http.ResponseWriter, err error) {
	switch {
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

func aiProviderErrorMessage(err error) string {
	var pe *generator.ProviderError
	if errors.As(err, &pe) {
		return pe.Error()
	}
	if os.Getenv("ENVIRONMENT") == "development" {
		return err.Error()
	}
	return "AI provider is temporarily unavailable"
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
