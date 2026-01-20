package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandler_Check_NoDB(t *testing.T) {
	handler := NewHealthHandler(nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	w := httptest.NewRecorder()

	handler.Check(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Check() status = %d, want %d", w.Code, http.StatusOK)
	}

	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Check() Content-Type = %q, want %q", ct, "application/json")
	}

	var response HealthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Status != "ok" {
		t.Errorf("Response.Status = %q, want %q", response.Status, "ok")
	}

	if response.Version == "" {
		t.Error("Response.Version should not be empty")
	}
}

func TestHealthHandler_Check_ResponseStructure(t *testing.T) {
	handler := NewHealthHandler(nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	w := httptest.NewRecorder()

	handler.Check(w, req)

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify required fields exist
	if _, ok := response["status"]; !ok {
		t.Error("Response should have 'status' field")
	}

	if _, ok := response["version"]; !ok {
		t.Error("Response should have 'version' field")
	}
}
