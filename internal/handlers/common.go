package handlers

import (
	"strconv"
)

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	TotalPages int64       `json:"total_pages"`
}

// parseIntParam parses a string parameter to int with a default value
func parseIntParam(param string, defaultValue int) int {
	if param == "" {
		return defaultValue
	}
	if value, err := strconv.Atoi(param); err == nil {
		return value
	}
	return defaultValue
}
