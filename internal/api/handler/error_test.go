package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRespondError(t *testing.T) {
	tests := []struct {
		name         string
		code         int
		message      string
		err          error
		expectedBody ErrorResponse
	}{
		{
			name:    "bad request",
			code:    http.StatusBadRequest,
			message: "invalid input",
			err:     nil,
			expectedBody: ErrorResponse{
				Error:   "Bad Request",
				Message: "invalid input",
				Code:    400,
			},
		},
		{
			name:    "internal error with error",
			code:    http.StatusInternalServerError,
			message: "something went wrong",
			err:     errors.New("db connection failed"),
			expectedBody: ErrorResponse{
				Error:   "Internal Server Error",
				Message: "something went wrong",
				Code:    500,
			},
		},
		{
			name:    "not found",
			code:    http.StatusNotFound,
			message: "resource not found",
			err:     nil,
			expectedBody: ErrorResponse{
				Error:   "Not Found",
				Message: "resource not found",
				Code:    404,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			RespondError(w, tt.code, tt.message, tt.err)

			if w.Code != tt.code {
				t.Errorf("RespondError() status = %d, want %d", w.Code, tt.code)
			}

			if ct := w.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("RespondError() Content-Type = %q, want %q", ct, "application/json")
			}

			var response ErrorResponse
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if response.Error != tt.expectedBody.Error {
				t.Errorf("Response.Error = %q, want %q", response.Error, tt.expectedBody.Error)
			}
			if response.Message != tt.expectedBody.Message {
				t.Errorf("Response.Message = %q, want %q", response.Message, tt.expectedBody.Message)
			}
			if response.Code != tt.expectedBody.Code {
				t.Errorf("Response.Code = %d, want %d", response.Code, tt.expectedBody.Code)
			}
		})
	}
}

func TestRespondBadRequest(t *testing.T) {
	w := httptest.NewRecorder()
	RespondBadRequest(w, "invalid data")

	if w.Code != http.StatusBadRequest {
		t.Errorf("RespondBadRequest() status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var response ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if response.Message != "invalid data" {
		t.Errorf("Response.Message = %q, want %q", response.Message, "invalid data")
	}
}

func TestRespondNotFound(t *testing.T) {
	w := httptest.NewRecorder()
	RespondNotFound(w, "item not found")

	if w.Code != http.StatusNotFound {
		t.Errorf("RespondNotFound() status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestRespondInternalError(t *testing.T) {
	w := httptest.NewRecorder()
	RespondInternalError(w, errors.New("database error"))

	if w.Code != http.StatusInternalServerError {
		t.Errorf("RespondInternalError() status = %d, want %d", w.Code, http.StatusInternalServerError)
	}

	var response ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	// Internal error message should be generic
	if response.Message != "An internal error occurred" {
		t.Errorf("Response.Message = %q, want generic message", response.Message)
	}
}

func TestRespondUnauthorized(t *testing.T) {
	w := httptest.NewRecorder()
	RespondUnauthorized(w, "invalid token")

	if w.Code != http.StatusUnauthorized {
		t.Errorf("RespondUnauthorized() status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestRespondForbidden(t *testing.T) {
	w := httptest.NewRecorder()
	RespondForbidden(w, "access denied")

	if w.Code != http.StatusForbidden {
		t.Errorf("RespondForbidden() status = %d, want %d", w.Code, http.StatusForbidden)
	}
}
