package response

import (
	"encoding/json"
	"net/http"
)

// Standard response types

// Success represents a successful API response
type Success struct {
	Success bool        `json:"success" example:"true"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// Error represents an error API response
type Error struct {
	Success bool   `json:"success" example:"false"`
	Error   string `json:"error" example:"Invalid request"`
	Code    string `json:"code,omitempty" example:"INVALID_INPUT"`
}

// ValidationError represents validation error details
type ValidationError struct {
	Success bool              `json:"success" example:"false"`
	Error   string            `json:"error" example:"Validation failed"`
	Fields  map[string]string `json:"fields,omitempty"`
}

// Helper functions

// JSON writes a JSON response
func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// OK sends a 200 OK response with data
func OK(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusOK, Success{
		Success: true,
		Data:    data,
	})
}

// Created sends a 201 Created response
func Created(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusCreated, Success{
		Success: true,
		Data:    data,
		Message: "Resource created successfully",
	})
}

// NoContent sends a 204 No Content response
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// BadRequest sends a 400 Bad Request error
func BadRequest(w http.ResponseWriter, message string) {
	JSON(w, http.StatusBadRequest, Error{
		Success: false,
		Error:   message,
		Code:    "BAD_REQUEST",
	})
}

// Unauthorized sends a 401 Unauthorized error
func Unauthorized(w http.ResponseWriter, message string) {
	JSON(w, http.StatusUnauthorized, Error{
		Success: false,
		Error:   message,
		Code:    "UNAUTHORIZED",
	})
}

// Forbidden sends a 403 Forbidden error
func Forbidden(w http.ResponseWriter, message string) {
	JSON(w, http.StatusForbidden, Error{
		Success: false,
		Error:   message,
		Code:    "FORBIDDEN",
	})
}

// NotFound sends a 404 Not Found error
func NotFound(w http.ResponseWriter, message string) {
	JSON(w, http.StatusNotFound, Error{
		Success: false,
		Error:   message,
		Code:    "NOT_FOUND",
	})
}

// Conflict sends a 409 Conflict error
func Conflict(w http.ResponseWriter, message string) {
	JSON(w, http.StatusConflict, Error{
		Success: false,
		Error:   message,
		Code:    "CONFLICT",
	})
}

// UnprocessableEntity sends a 422 validation error
func UnprocessableEntity(w http.ResponseWriter, fields map[string]string) {
	JSON(w, http.StatusUnprocessableEntity, ValidationError{
		Success: false,
		Error:   "Validation failed",
		Fields:  fields,
	})
}

// InternalServerError sends a 500 error
func InternalServerError(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Internal server error"
	}
	JSON(w, http.StatusInternalServerError, Error{
		Success: false,
		Error:   message,
		Code:    "INTERNAL_ERROR",
	})
}
