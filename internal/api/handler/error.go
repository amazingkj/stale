package handler

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"
)

// ErrorResponse represents a structured API error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"code"`
}

// RespondError sends a JSON error response and logs the error
func RespondError(w http.ResponseWriter, code int, userMessage string, err error) {
	// Log the actual error for debugging
	if err != nil {
		log.Error().Err(err).Int("status", code).Str("message", userMessage).Msg("API error")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	response := ErrorResponse{
		Error:   http.StatusText(code),
		Message: userMessage,
		Code:    code,
	}

	json.NewEncoder(w).Encode(response)
}

// RespondBadRequest sends a 400 Bad Request response
func RespondBadRequest(w http.ResponseWriter, message string) {
	RespondError(w, http.StatusBadRequest, message, nil)
}

// RespondNotFound sends a 404 Not Found response
func RespondNotFound(w http.ResponseWriter, message string) {
	RespondError(w, http.StatusNotFound, message, nil)
}

// RespondInternalError sends a 500 Internal Server Error response
// The actual error is logged but not exposed to the client
func RespondInternalError(w http.ResponseWriter, err error) {
	RespondError(w, http.StatusInternalServerError, "An internal error occurred", err)
}

// RespondUnauthorized sends a 401 Unauthorized response
func RespondUnauthorized(w http.ResponseWriter, message string) {
	RespondError(w, http.StatusUnauthorized, message, nil)
}

// RespondForbidden sends a 403 Forbidden response
func RespondForbidden(w http.ResponseWriter, message string) {
	RespondError(w, http.StatusForbidden, message, nil)
}
