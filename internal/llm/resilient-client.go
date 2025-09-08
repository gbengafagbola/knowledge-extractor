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

func (r *ResilientClient) AnalyzeText(input string) (
	string, string, []string, string, []string, float64, error,
) {
	summary, title, topics, sentiment, keywords, confidence, err := r.openai.AnalyzeText(input)
	if err != nil {
		fmt.Println("OpenAI request failed, falling back to MockClient:", err)
		return r.mock.AnalyzeText(input)
	}
	return summary, title, topics, sentiment, keywords, confidence, nil
}
