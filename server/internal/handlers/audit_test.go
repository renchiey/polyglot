package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/renchieyang/polyglot/server/internal/lexaudit"
)

func newTestHandler() *AuditHandler {
	return &AuditHandler{Registry: lexaudit.NewRegistry()}
}

func post(h http.HandlerFunc, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/audit", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h(rec, req)
	return rec
}

func TestAuditHandlerOK(t *testing.T) {
	rec := post(newTestHandler().Audit, `{"language":"zh","sentence":"我喜欢喝咖啡","target_level":3}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (%s)", rec.Code, rec.Body.String())
	}
	var report lexaudit.Report
	if err := json.Unmarshal(rec.Body.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if !report.Passed {
		t.Errorf("expected pass, got %+v", report)
	}
}

func TestAuditHandlerUnsupportedLanguage(t *testing.T) {
	rec := post(newTestHandler().Audit, `{"language":"xx","sentence":"hi","target_level":1}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestAuditHandlerBadBody(t *testing.T) {
	rec := post(newTestHandler().Audit, `not json`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}
