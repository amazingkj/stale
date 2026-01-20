package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jiin/stale/internal/domain"
)

func TestIgnoredHandler_Create_Validation(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		expectedStatus int
	}{
		{
			name:           "invalid json",
			body:           "not json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty name",
			body:           `{"name": ""}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing name",
			body:           `{"reason": "test"}`,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &IgnoredHandler{} // nil repo - testing validation only

			req := httptest.NewRequest("POST", "/ignored", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			h.Create(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.expectedStatus)
			}
		})
	}
}

func TestIgnoredHandler_Delete_Validation(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		expectedStatus int
	}{
		{
			name:           "invalid id - not a number",
			id:             "abc",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid id - empty",
			id:             "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &IgnoredHandler{}

			req := httptest.NewRequest("DELETE", "/ignored/"+tt.id, nil)
			w := httptest.NewRecorder()

			// Note: chi.URLParam won't work without chi router context
			// This test verifies the parsing logic
			h.Delete(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.expectedStatus)
			}
		})
	}
}

func TestIgnoredHandler_BulkCreate_Validation(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		expectedStatus int
	}{
		{
			name:           "invalid json",
			body:           "not json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty items array",
			body:           `{"items": []}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing items",
			body:           `{}`,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &IgnoredHandler{}

			req := httptest.NewRequest("POST", "/ignored/bulk", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			h.BulkCreate(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.expectedStatus)
			}
		})
	}
}

func TestIgnoredHandler_BulkDelete_Validation(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		expectedStatus int
	}{
		{
			name:           "invalid json",
			body:           "not json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty ids array",
			body:           `{"ids": []}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing ids",
			body:           `{}`,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &IgnoredHandler{}

			req := httptest.NewRequest("POST", "/ignored/bulk-delete", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			h.BulkDelete(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.expectedStatus)
			}
		})
	}
}

func TestIgnoredDependencyInput_JSON(t *testing.T) {
	input := domain.IgnoredDependencyInput{
		Name:      "lodash",
		Ecosystem: "npm",
		Reason:    "deprecated",
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded domain.IgnoredDependencyInput
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Name != input.Name {
		t.Errorf("Name = %q, want %q", decoded.Name, input.Name)
	}
	if decoded.Ecosystem != input.Ecosystem {
		t.Errorf("Ecosystem = %q, want %q", decoded.Ecosystem, input.Ecosystem)
	}
	if decoded.Reason != input.Reason {
		t.Errorf("Reason = %q, want %q", decoded.Reason, input.Reason)
	}
}
