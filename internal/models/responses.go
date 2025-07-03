package models

import "time"

// PingResponse represents the response from ping endpoints
// @Description Response from ping endpoints
type PingResponse struct {
	Status  string `json:"status" example:"ok"`
	Message string `json:"message" example:"pong"`
}

// ErrorResponse represents an error response
// @Description Error response structure
type ErrorResponse struct {
	Status   string   `json:"status" example:"error"`
	Messages []string `json:"messages" example:"[\"Invalid email format\", \"Password too short\"]"`
}

// HealthResponse represents a comprehensive health check response
// @Description Comprehensive health check response
type HealthResponse struct {
	Status    string            `json:"status" example:"ok"`
	Message   string            `json:"message" example:"all services healthy"`
	Timestamp time.Time         `json:"timestamp" example:"2024-01-15T10:30:00Z"`
	Services  map[string]string `json:"services" example:"{\"server\":\"healthy\",\"database\":\"healthy\"}"`
}

// SuccessResponse represents a successful response
// @Description Success response structure
type SuccessResponse struct {
	Status  string `json:"status" example:"success"`
	Message string `json:"message" example:"Operation completed successfully"`
}

// NewErrorResponse creates a new error response with a single message
func NewErrorResponse(message string) ErrorResponse {
	return ErrorResponse{
		Status:   "error",
		Messages: []string{message},
	}
}

// NewErrorResponseWithMessages creates a new error response with multiple messages
func NewErrorResponseWithMessages(messages []string) ErrorResponse {
	return ErrorResponse{
		Status:   "error",
		Messages: messages,
	}
}

// NewErrorResponseFromError creates a new error response from an error
func NewErrorResponseFromError(err error) ErrorResponse {
	return NewErrorResponse(err.Error())
}
