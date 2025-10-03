package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gbengafagbola/knowledge-extractor/internal/llm"
	"github.com/gbengafagbola/knowledge-extractor/internal/models"
	"github.com/google/uuid"
)

type Server struct {
	DB     *sql.DB
	LLM    llm.LLM
	Driver string
}

func New(db *sql.DB, llm llm.LLM, driver string) *Server {
	return &Server{DB: db, LLM: llm, Driver: driver}
}

// HANDLER
func (s *Server) AnalyzeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.Text == "" {
		http.Error(w, "invalid input", http.StatusBadRequest)
		return
	}

	// Call LLM (real or mock)
	summary, title, topics, sentiment, keywords, confidence, err :=
		s.LLM.AnalyzeText(input.Text)
	if err != nil {
		http.Error(w, "LLM analysis failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	analysis := models.Analysis{
		ID:         uuid.NewString(),
		RawText:    input.Text,
		Summary:    summary,
		Title:      title,
		Topics:     topics,
		Sentiment:  sentiment,
		Keywords:   keywords,
		Confidence: confidence,
	}

	query := `
		INSERT INTO analyses (id, raw_text, summary, title, topics, sentiment, keywords, confidence)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`
	_, err = s.DB.ExecContext(context.Background(), query,
		analysis.ID, analysis.RawText, analysis.Summary, analysis.Title,
		s.formatStringArray(analysis.Topics), analysis.Sentiment,
		s.formatStringArray(analysis.Keywords), analysis.Confidence,
	)
	if err != nil {
		http.Error(w, "failed to insert into db: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(analysis)
}

func (s *Server) SearchHandler(w http.ResponseWriter, r *http.Request) {
	topic := r.URL.Query().Get("topic")
	if topic == "" {
		http.Error(w, "missing topic query param", http.StatusBadRequest)
		return
	}

	searchQuery := s.buildSearchQuery()
	rows, err := s.DB.QueryContext(context.Background(), searchQuery, topic)
	if err != nil {
		http.Error(w, "db query failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var results []models.Analysis
	for rows.Next() {
		var a models.Analysis
		var topicsScanner, keywordsScanner interface{}

		if s.Driver == "postgres" {
			topicsScanner = &a.Topics
			keywordsScanner = &a.Keywords
		} else {
			topicsScanner = &sqliteStringArray{&a.Topics}
			keywordsScanner = &sqliteStringArray{&a.Keywords}
		}

		err := rows.Scan(
			&a.ID, &a.RawText, &a.Summary, &a.Title, topicsScanner,
			&a.Sentiment, keywordsScanner, &a.Confidence, &a.CreatedAt,
		)
		if err != nil {
			http.Error(w, "row scan failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		results = append(results, a)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(results)
}

// helpers
func (s *Server) formatStringArray(arr []string) interface{} {
	if s.Driver == "postgres" {
		return "{" + joinStrings(arr, ",") + "}"
	}
	// SQLite - store as comma-separated string
	return joinStrings(arr, ",")
}

// sqliteStringArray handles scanning comma-separated strings for SQLite
type sqliteStringArray struct {
	arr *[]string
}

func (s *sqliteStringArray) Scan(value interface{}) error {
	if value == nil {
		*s.arr = nil
		return nil
	}

	var str string
	switch v := value.(type) {
	case string:
		str = v
	case []byte:
		str = string(v)
	default:
		return fmt.Errorf("cannot scan %T into sqliteStringArray", value)
	}

	if str == "" {
		*s.arr = []string{}
		return nil
	}

	// Split comma-separated string and trim spaces
	parts := strings.Split(str, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	*s.arr = result
	return nil
}

func (s *Server) buildSearchQuery() string {
	if s.Driver == "postgres" {
		return `SELECT id, raw_text, summary, title, topics, sentiment, keywords, confidence, created_at
		 FROM analyses
		 WHERE $1 = ANY(topics) OR $1 = ANY(keywords)`
	}
	// SQLite - use LIKE with comma-separated strings
	return `SELECT id, raw_text, summary, title, topics, sentiment, keywords, confidence, created_at
		 FROM analyses
		 WHERE topics LIKE '%' || $1 || '%' OR keywords LIKE '%' || $1 || '%'`
}

func joinStrings(arr []string, sep string) string {
	out := ""
	for i, v := range arr {
		if i > 0 {
			out += sep
		}
		out += v
	}
	return out
}
