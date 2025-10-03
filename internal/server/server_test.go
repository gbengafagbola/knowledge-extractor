package server_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/gbengafagbola/knowledge-extractor/internal/llm"
	"github.com/gbengafagbola/knowledge-extractor/internal/models"
	"github.com/gbengafagbola/knowledge-extractor/internal/server"
)

// helpers
func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	schema := `
	CREATE TABLE analyses (
		id TEXT PRIMARY KEY,
		raw_text TEXT,
		summary TEXT,
		title TEXT,
		topics TEXT,
		sentiment TEXT,
		keywords TEXT,
		confidence REAL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return db
}

// tests
func TestAnalyzeHandler(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	s := server.New(db, llm.NewMockClient(), "sqlite3")

	body := []byte(`{"text": "This is a test document about AI and Go"}`)
	req := httptest.NewRequest(http.MethodPost, "/analyze", bytes.NewReader(body))
	w := httptest.NewRecorder()

	s.AnalyzeHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var result models.Analysis
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if result.Summary == "" || result.Title == "" {
		t.Errorf("expected summary and title, got empty")
	}

	// Verify row inserted
	row := db.QueryRowContext(context.Background(), `SELECT id FROM analyses WHERE id = ?`, result.ID)
	var id string
	if err := row.Scan(&id); err != nil {
		t.Errorf("expected row in db, got error: %v", err)
	}
}

func TestSearchHandler(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert dummy analysis
	_, err := db.Exec(`
		INSERT INTO analyses (id, raw_text, summary, title, topics, sentiment, keywords, confidence, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"test-id", "AI text", "summary", "title", `{"AI"}`, "positive", `{"Go"}`, 0.95, time.Now(),
	)
	if err != nil {
		t.Fatalf("failed to insert test row: %v", err)
	}

	s := server.New(db, llm.NewMockClient(), "sqlite3")

	req := httptest.NewRequest(http.MethodGet, "/search?topic=AI", nil)
	w := httptest.NewRecorder()

	s.SearchHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var results []models.Analysis
	if err := json.NewDecoder(w.Body).Decode(&results); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(results) == 0 {
		t.Errorf("expected at least 1 result, got 0")
	}
}
