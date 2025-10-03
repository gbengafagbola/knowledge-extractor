package models

import "time"

// Analysis represents the structured output from text analysis
// JSON tags enable automatic serialization for API responses
// Fields are designed to capture key insights from unstructured text
type Analysis struct {
	ID         string    `json:"id"`         // UUID for unique identification and tracing
	RawText    string    `json:"raw_text"`   // Original input text for reference
	Summary    string    `json:"summary"`    // 1-2 sentence summary from LLM
	Title      string    `json:"title"`      // Extracted or generated title
	Topics     []string  `json:"topics"`     // 3 key topics identified by LLM
	Sentiment  string    `json:"sentiment"`  // positive/neutral/negative classification
	Keywords   []string  `json:"keywords"`   // 3 most frequent nouns (local extraction)
	Confidence float64   `json:"confidence"` // Analysis confidence score (0-1)
	CreatedAt  time.Time `json:"created_at"` // Timestamp for audit and sorting
}
