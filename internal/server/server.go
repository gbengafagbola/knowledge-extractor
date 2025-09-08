package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/gbengafagbola/knowledge-extractor/internal/llm"
	"github.com/gbengafagbola/knowledge-extractor/internal/models"
	"github.com/google/uuid"
)

type Server struct {
	DB  *sql.DB
	LLM llm.LLM
}

func New(db *sql.DB, llm llm.LLM) *Server {
	return &Server{DB: db, LLM: llm}
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
		pqStringArray(analysis.Topics), analysis.Sentiment,
		pqStringArray(analysis.Keywords), analysis.Confidence,
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

	rows, err := s.DB.QueryContext(context.Background(),
		`SELECT id, raw_text, summary, title, topics, sentiment, keywords, confidence, created_at
		 FROM analyses
		 WHERE $1 = ANY(topics) OR $1 = ANY(keywords)`,
		topic,
	)
	if err != nil {
		http.Error(w, "db query failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var results []models.Analysis
	for rows.Next() {
		var a models.Analysis
		err := rows.Scan(
			&a.ID, &a.RawText, &a.Summary, &a.Title, pqArray(&a.Topics),
			&a.Sentiment, pqArray(&a.Keywords), &a.Confidence, &a.CreatedAt,
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
func pqArray(arr *[]string) interface{} {
	return (*arr)
}

func pqStringArray(arr []string) interface{} {
	return "{" + joinStrings(arr, ",") + "}"
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
