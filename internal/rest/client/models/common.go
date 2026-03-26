package models

import "time"

const (
	PathPostingsToCancel       = "/v1/PostingsToCancel"
	PathPostingsCancelResponse = "/v1/PostingsCancelResponse"
)

type Config struct {
	Timeout time.Duration
	Host    string
	ID      string
	Secret  string
}

type Error struct {
	ErrorCode int    `json:"ErrorCode"`
	Entity    string `json:"Entity"`
	Message   string `json:"Message"`
	Details   string `json:"Details"`
	Value     string `json:"Value"`
}

type PostingsToCancelReq struct {
	IsTerminalCancel *bool
	ParcelType       string
}

type PostingsToCancelResp struct {
	Data   []PostingsToCancelData `json:"Data"`
	Errors []Error                `json:"Errors"`
}

type PostingsToCancelData struct {
	PostingNumber      string    `json:"PostingNumber"`
	TrackingNumber     string    `json:"TrackingNumber"`
	CreatedAt          time.Time `json:"CreatedAt"`
	TplIntegrationType string    `json:"TplIntegrationType"`
}

type PostingsCancelResponseReq struct {
	Body []PostingsCancelResponseBody
}

type PostingsCancelResponseBody struct {
	PostingNumber string `json:"PostingNumber"`
	NewState      string `json:"NewState"`
}

type PostingsCancelResponseResp struct {
	Errors []Error `json:"Errors"`
}
