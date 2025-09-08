package models

import "time"

type Analysis struct {
	ID         string    `json:"id"`
	RawText    string    `json:"raw_text"`
	Summary    string    `json:"summary"`
	Title      string    `json:"title"`
	Topics     []string  `json:"topics"`
	Sentiment  string    `json:"sentiment"`
	Keywords   []string  `json:"keywords"`
	Confidence float64   `json:"confidence"`
	CreatedAt  time.Time `json:"created_at"`
}
