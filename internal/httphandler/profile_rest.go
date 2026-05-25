package httphandler

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"gitlab16.skiftrade.kz/templates/go/errorcodes"
	"gitlab16.skiftrade.kz/templates/go/internal"
	ucModels "gitlab16.skiftrade.kz/templates/go/internal/usecase/models"
)

const restQueryTimeout = 25 * time.Second

// ProfileRESTHandler serves high-traffic GET routes directly (no gRPC hop).
type ProfileRESTHandler struct {
	uc internal.UseCase
}

func NewProfileRESTHandler(uc internal.UseCase) *ProfileRESTHandler {
	return &ProfileRESTHandler{uc: uc}
}

func (h *ProfileRESTHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/users/telegram/{telegramId}", h.getUser)
	mux.HandleFunc("GET /v1/wallet/telegram/{telegramId}", h.getWallet)
	mux.HandleFunc("GET /v1/referrals/telegram/{telegramId}", h.getReferrals)
	mux.HandleFunc("GET /v1/prompts/history/telegram/{telegramId}", h.getPromptHistory)
	mux.HandleFunc("GET /health", h.health)
}

func (h *ProfileRESTHandler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *ProfileRESTHandler) getUser(w http.ResponseWriter, r *http.Request) {
	telegramID := strings.TrimSpace(r.PathValue("telegramId"))
	ctx, cancel := context.WithTimeout(r.Context(), restQueryTimeout)
	defer cancel()

	out, err := h.uc.GetUserByTelegramID(ctx, telegramID)
	if err != nil {
		writeProfileRESTError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"id":               out.Data.ID,
			"name":             out.Data.Name,
			"surname":          out.Data.Username,
			"username":         out.Data.Username,
			"subscriptionPlan": normalizeSubscriptionPlan(out.Data.Role),
			"verified":         out.Data.Verified,
		},
	})
}

func normalizeSubscriptionPlan(role string) string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "pro", "ultra", "free":
		return strings.ToLower(strings.TrimSpace(role))
	default:
		return "free"
	}
}

func (h *ProfileRESTHandler) getWallet(w http.ResponseWriter, r *http.Request) {
	telegramID := strings.TrimSpace(r.PathValue("telegramId"))
	ctx, cancel := context.WithTimeout(r.Context(), restQueryTimeout)
	defer cancel()

	out, err := h.uc.GetWalletByTelegramID(ctx, telegramID)
	if err != nil {
		writeProfileRESTError(w, err)
		return
	}

	transactions := make([]map[string]any, 0, len(out.Transactions))
	for _, item := range out.Transactions {
		transactions = append(transactions, map[string]any{
			"id":          item.ID,
			"date":        item.Date,
			"type":        item.Type,
			"amount":      item.Amount,
			"status":      item.Status,
			"description": item.Description,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"wallet": map[string]any{
			"id":               out.Wallet.ID,
			"profileId":        out.Wallet.ProfileID,
			"balance":          out.Wallet.Balance,
			"totalEarned":      out.Wallet.TotalEarned,
			"balanceAvailable": out.Wallet.BalanceAvailable,
		},
		"transactions": transactions,
	})
}

func (h *ProfileRESTHandler) getReferrals(w http.ResponseWriter, r *http.Request) {
	telegramID := strings.TrimSpace(r.PathValue("telegramId"))
	ctx, cancel := context.WithTimeout(r.Context(), restQueryTimeout)
	defer cancel()

	out, err := h.uc.GetReferralsByTelegramID(ctx, telegramID)
	if err != nil {
		writeProfileRESTError(w, err)
		return
	}

	items := make([]map[string]any, 0, len(out.Items))
	for _, item := range out.Items {
		items = append(items, map[string]any{
			"id":                  item.ID,
			"telegramId":          item.TelegramID,
			"name":                item.Name,
			"username":            item.Username,
			"completedTasksCount": item.CompletedTasksCount,
			"earnings":            item.Earnings,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *ProfileRESTHandler) getPromptHistory(w http.ResponseWriter, r *http.Request) {
	telegramID := strings.TrimSpace(r.PathValue("telegramId"))
	ctx, cancel := context.WithTimeout(r.Context(), restQueryTimeout)
	defer cancel()

	out, err := h.uc.GetPromptHistoryByTelegramID(ctx, telegramID)
	if err != nil {
		writeProfileRESTError(w, err)
		return
	}

	items := make([]map[string]any, 0, len(out.Items))
	for _, item := range out.Items {
		items = append(items, map[string]any{
			"id":        item.ID,
			"prompt":    item.Prompt,
			"category":  item.Category,
			"model":     item.Model,
			"createdAt": item.CreatedAt,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func writeProfileRESTError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ucModels.ErrProfileNotFound):
		writeJSON(w, http.StatusNotFound, errorResponse{
			Code:    errorcodes.ProfileNotFound,
			Message: ucModels.ErrProfileNotFound.Error(),
		})
	case errors.Is(err, ucModels.ErrInvalidInput):
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Code:    errorcodes.InvalidArgument,
			Message: err.Error(),
		})
	case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
		writeJSON(w, http.StatusGatewayTimeout, errorResponse{
			Code:    errorcodes.Internal,
			Message: "request timed out",
		})
	default:
		writeJSON(w, http.StatusInternalServerError, errorResponse{
			Code:    errorcodes.Internal,
			Message: ucModels.ErrInternalServerError.Error(),
		})
	}
}
