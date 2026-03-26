package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"

	"gitlab16.skiftrade.kz/templates/go/internal"
	"gitlab16.skiftrade.kz/templates/go/internal/rest/client/models"
	"gitlab16.skiftrade.kz/templates/go/pkg/logger"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"
)

const (
	maxResponseSize = 10 << 20 // 10 MB
)

type client struct {
	config     models.Config
	httpClient *http.Client
}

func NewClient(config models.Config) internal.Client {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConnsPerHost = t.MaxIdleConns
	return &client{
		config: config,
		httpClient: &http.Client{
			Transport: otelhttp.NewTransport(
				t,
				otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
					return fmt.Sprintf("HTTP %s %s", r.Method, r.URL.Path)
				})),
			Timeout: config.Timeout,
		},
	}
}

func (c *client) PostingsToCancel(ctx context.Context, token string, req models.PostingsToCancelReq) (models.PostingsToCancelResp, error) {
	u := &url.URL{
		Scheme: "https",
		Host:   c.config.Host,
		Path:   models.PathPostingsToCancel,
	}

	// Формирование query-параметров
	q := u.Query()
	if req.ParcelType != "" {
		q.Set("parcelType", req.ParcelType)
	}
	if req.IsTerminalCancel != nil {
		q.Set("isTerminalCancel", strconv.FormatBool(*req.IsTerminalCancel))
	}
	u.RawQuery = q.Encode()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return models.PostingsToCancelResp{}, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+token)

	var resp models.PostingsToCancelResp
	if err := c.doRequest(ctx, httpReq, &resp); err != nil {
		return models.PostingsToCancelResp{}, fmt.Errorf("postings to cancel: %w", err)
	}
	return resp, nil
}

func (c *client) PostingsCancelResponse(ctx context.Context, token string, req models.PostingsCancelResponseReq) (models.PostingsCancelResponseResp, error) {
	u := &url.URL{
		Scheme: "https",
		Host:   c.config.Host,
		Path:   models.PathPostingsCancelResponse,
	}

	body, err := json.Marshal(req.Body)
	if err != nil {
		return models.PostingsCancelResponseResp{}, fmt.Errorf("marshal request body: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewReader(body))
	if err != nil {
		return models.PostingsCancelResponseResp{}, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+token)

	var resp models.PostingsCancelResponseResp
	if err := c.doRequest(ctx, httpReq, &resp); err != nil {
		return models.PostingsCancelResponseResp{}, fmt.Errorf("postings cancel response: %w", err)
	}
	return resp, nil
}

// doRequest обрабатывает HTTP-запрос и ответ
func (c *client) doRequest(ctx context.Context, req *http.Request, dest interface{}) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		logRequestError(ctx, req, "http client error", err)
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	// Защита от больших ответов
	limitReader := io.LimitReader(resp.Body, maxResponseSize)
	body, err := io.ReadAll(limitReader)
	if err != nil {
		logRequestError(ctx, req, "read body error", err)
		return fmt.Errorf("read response body: %w", err)
	}

	// Проверка статуса
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		logErrorResponse(ctx, req, resp, body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, truncateString(string(body), 500))
	}

	// Декодирование
	if err := json.Unmarshal(body, dest); err != nil {
		logErrorResponse(ctx, req, resp, body)
		return fmt.Errorf("unmarshal response: %w", err)
	}

	return nil
}

func logRequestError(ctx context.Context, req *http.Request, msg string, err error) {
	attrs := []any{
		logger.ErrorAttr(err),
		slog.String("method", req.Method),
		slog.String("url", req.URL.String()),
	}

	if span := trace.SpanFromContext(ctx); span.SpanContext().HasTraceID() {
		attrs = append(attrs, slog.String("trace_id", span.SpanContext().TraceID().String()))
	}

	slog.ErrorContext(ctx, msg, attrs...)
}

func logErrorResponse(ctx context.Context, req *http.Request, resp *http.Response, body []byte) {
	attrs := []any{
		slog.String("method", req.Method),
		slog.String("url", req.URL.String()),
		slog.Int("status", resp.StatusCode),
		slog.String("body", truncateString(string(body), 500)),
	}

	if span := trace.SpanFromContext(ctx); span.SpanContext().HasTraceID() {
		attrs = append(attrs, slog.String("trace_id", span.SpanContext().TraceID().String()))
	}

	slog.ErrorContext(ctx, "unexpected response", attrs...)
}

// truncateString обрезает строку до указанной длины
func truncateString(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
