package llm

import "fmt"

// ResilientClient wraps an OpenAI client and falls back to MockClient if needed.
type ResilientClient struct {
	openai *OpenAIClient
	mock   *MockClient
}

// Ensure ResilientClient implements LLM
var _ LLM = (*ResilientClient)(nil)

// NewResilientClient returns a client that tries OpenAI first, then falls back to mock.
func NewResilientClient(openai *OpenAIClient, mock *MockClient) *ResilientClient {
	return &ResilientClient{openai: openai, mock: mock}
}

// AnalyzeText implements the Circuit Breaker pattern
// Primary client (OpenAI) is tried first, with automatic fallback to mock on failure
// This ensures the system remains functional even when external services are down
func (r *ResilientClient) AnalyzeText(input string) (
	string, string, []string, string, []string, float64, error,
) {
	// Try primary client first
	summary, title, topics, sentiment, keywords, confidence, err := r.openai.AnalyzeText(input)
	if err != nil {
		// Log failure for observability, then fallback transparently
		fmt.Println("OpenAI request failed, falling back to MockClient:", err)
		return r.mock.AnalyzeText(input)
	}
	return summary, title, topics, sentiment, keywords, confidence, nil
}
