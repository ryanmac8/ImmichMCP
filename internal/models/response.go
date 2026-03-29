package models

import "github.com/google/uuid"

// Error codes used across all tools.
const (
	ErrAuthFailed          = "AUTH_FAILED"
	ErrNotFound            = "NOT_FOUND"
	ErrValidation          = "VALIDATION"
	ErrUpstreamError       = "UPSTREAM_ERROR"
	ErrRateLimit           = "RATE_LIMIT"
	ErrUnknown             = "UNKNOWN"
	ErrConfirmationRequired = "CONFIRMATION_REQUIRED"
)

// MCPResponse is the standard response envelope for all MCP tool results.
type MCPResponse struct {
	Ok       bool        `json:"ok"`
	Result   interface{} `json:"result,omitempty"`
	Error    *MCPError   `json:"error,omitempty"`
	Meta     MCPMeta     `json:"meta"`
	Warnings []string    `json:"warnings,omitempty"`
}

// MCPError holds error information.
type MCPError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// MCPMeta is metadata included with all responses.
type MCPMeta struct {
	RequestID    string  `json:"request_id"`
	Page         *int    `json:"page,omitempty"`
	PageSize     *int    `json:"page_size,omitempty"`
	Total        *int    `json:"total,omitempty"`
	Next         *string `json:"next,omitempty"`
	ImmichBaseURL string `json:"immich_base_url"`
}

// NewMeta creates a MCPMeta with a new request ID.
func NewMeta(baseURL string) MCPMeta {
	return MCPMeta{
		RequestID:     uuid.New().String(),
		ImmichBaseURL: baseURL,
	}
}

// Success returns a successful MCPResponse.
func Success(result interface{}, meta MCPMeta) MCPResponse {
	return MCPResponse{Ok: true, Result: result, Meta: meta}
}

// ErrorResponse returns an error MCPResponse.
func ErrorResponse(code, message string, details interface{}, meta MCPMeta) MCPResponse {
	return MCPResponse{
		Ok:    false,
		Error: &MCPError{Code: code, Message: message, Details: details},
		Meta:  meta,
	}
}

// BulkIDResponse is the response for bulk ID operations.
type BulkIDResponse struct {
	ID      string `json:"id"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// BulkOperationResult is returned for bulk operations.
type BulkOperationResult struct {
	AffectedIDs []string `json:"affected_ids"`
	Warnings    []string `json:"warnings,omitempty"`
	Executed    bool     `json:"executed"`
}
