// Package response provides standardized HTTP response formats for Nivo APIs.
package response

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/1mb-dev/nivomoney/shared/errors"
)

// Response represents a standardized API response envelope.
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorData  `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// ErrorData contains error information.
type ErrorData struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// Meta contains metadata about the response.
type Meta struct {
	RequestID  string      `json:"request_id,omitempty"`
	Timestamp  string      `json:"timestamp,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

// Pagination contains pagination information.
type Pagination struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int   `json:"total_pages"`
	TotalItems int64 `json:"total_items"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// JSON writes a JSON response with the given status code and data.
// Errors during encoding are logged but not returned since the HTTP connection
// may already be broken and there's no meaningful recovery action.
func JSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("[response] Failed to encode JSON response: %v", err)
	}
}

// Success writes a success response.
func Success(w http.ResponseWriter, statusCode int, data interface{}) {
	JSON(w, statusCode, Response{
		Success: true,
		Data:    data,
	})
}

// SuccessWithMeta writes a success response with metadata.
func SuccessWithMeta(w http.ResponseWriter, statusCode int, data interface{}, meta *Meta) {
	JSON(w, statusCode, Response{
		Success: true,
		Data:    data,
		Meta:    meta,
	})
}

// Error writes an error response from an errors.Error.
func Error(w http.ResponseWriter, err *errors.Error) {
	statusCode := err.HTTPStatusCode()

	JSON(w, statusCode, Response{
		Success: false,
		Error: &ErrorData{
			Code:    string(err.Code),
			Message: err.Message,
			Details: err.Details,
		},
	})
}

// ErrorWithMeta writes an error response with metadata.
func ErrorWithMeta(w http.ResponseWriter, err *errors.Error, meta *Meta) {
	statusCode := err.HTTPStatusCode()

	JSON(w, statusCode, Response{
		Success: false,
		Error: &ErrorData{
			Code:    string(err.Code),
			Message: err.Message,
			Details: err.Details,
		},
		Meta: meta,
	})
}

// BadRequest writes a 400 Bad Request response.
func BadRequest(w http.ResponseWriter, message string) {
	Error(w, errors.BadRequest(message))
}

// Unauthorized writes a 401 Unauthorized response.
func Unauthorized(w http.ResponseWriter, message string) {
	Error(w, errors.Unauthorized(message))
}

// Forbidden writes a 403 Forbidden response.
func Forbidden(w http.ResponseWriter, message string) {
	Error(w, errors.Forbidden(message))
}

// NotFound writes a 404 Not Found response.
func NotFound(w http.ResponseWriter, resource string) {
	Error(w, errors.NotFound(resource))
}

// Conflict writes a 409 Conflict response.
func Conflict(w http.ResponseWriter, message string) {
	Error(w, errors.Conflict(message))
}

// InternalError writes a 500 Internal Server Error response.
func InternalError(w http.ResponseWriter, message string) {
	Error(w, errors.Internal(message))
}

// Created writes a 201 Created response.
func Created(w http.ResponseWriter, data interface{}) {
	Success(w, http.StatusCreated, data)
}

// OK writes a 200 OK response.
func OK(w http.ResponseWriter, data interface{}) {
	Success(w, http.StatusOK, data)
}

// NoContent writes a 204 No Content response.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// Paginated writes a paginated response with metadata.
func Paginated(w http.ResponseWriter, data interface{}, page, pageSize int, totalItems int64) {
	totalPages := int((totalItems + int64(pageSize) - 1) / int64(pageSize))
	if totalPages == 0 {
		totalPages = 1
	}

	pagination := &Pagination{
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
		TotalItems: totalItems,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}

	SuccessWithMeta(w, http.StatusOK, data, &Meta{
		Pagination: pagination,
	})
}

// ValidationError writes a validation error response.
func ValidationError(w http.ResponseWriter, details map[string]interface{}) {
	err := errors.Validation("validation failed")
	err.Details = details
	Error(w, err)
}
