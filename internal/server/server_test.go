package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/zaynereo/admission-policy-webhook/internal/admission"
)

func TestHealthz(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	NewMux().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d", rr.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["status"] != "ok" {
		t.Fatalf("body=%v", body)
	}
}

func TestValidateBoundedBody(t *testing.T) {
	big := bytes.Repeat([]byte("a"), admission.MaxAdmissionBodyBytes+1)
	req := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewReader(big))
	rr := httptest.NewRecorder()
	NewMux().ServeHTTP(rr, req)
	if rr.Code != 400 {
		t.Fatalf("status %d", rr.Code)
	}
}
