package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gbengafagbola/knowledge-extractor/internal/llm"
	"github.com/gbengafagbola/knowledge-extractor/internal/server"
)

// newMockServer sets up a test server with sqlmock + a mock LLM
func newMockServer(t *testing.T) (*server.Server, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	// Use a mock wrapper instead of hitting OpenAI
	return server.New(db, llm.NewMockClient()), mock
}

func TestAnalyzeHandler(t *testing.T) {
	s, mock := newMockServer(t)

	// Expect insert query since AnalyzeHandler writes results to DB
	mock.ExpectExec("INSERT INTO analyses").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Simulate POST /analyze with some text
	body := `{"text":"Go is fast"}`
	req := httptest.NewRequest(http.MethodPost, "/analyze", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.AnalyzeHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Verify expectations (ensures query was actually run)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet db expectations: %v", err)
	}
}

func TestSearchHandler(t *testing.T) {
	s, mock := newMockServer(t)

	// Mock a row that matches search
	rows := sqlmock.NewRows([]string{
		"id", "raw_text", "summary", "title", "topics", "sentiment", "keywords", "confidence", "created_at",
	}).AddRow(
		"1", "raw", "sum", "title",
		[]string{"go"}, "neutral", []string{"fast"},
		0.9, "2024-01-01T00:00:00Z",
	)

	mock.ExpectQuery("SELECT id, raw_text").
		WithArgs("go").
		WillReturnRows(rows)

	req := httptest.NewRequest(http.MethodGet, "/search?topic=go", nil)
	w := httptest.NewRecorder()

	s.SearchHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet db expectations: %v", err)
	}
}
