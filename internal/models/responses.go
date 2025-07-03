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
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"database unreachable"`
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
