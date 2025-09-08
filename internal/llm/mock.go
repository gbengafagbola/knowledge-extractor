package llm

// MockClient is a fake implementation of LLM for testing.
type MockClient struct{}

// Ensure MockClient implements the LLM interface
var _ LLM = (*MockClient)(nil)

// NewMockClient returns a new MockClient
func NewMockClient() *MockClient {
	return &MockClient{}
}

// AnalyzeText implements the LLM interface with static values
func (m *MockClient) AnalyzeText(text string) (
	summary string,
	title string,
	topics []string,
	sentiment string,
	keywords []string,
	confidence float64,
	err error,
) {
	return "mock summary",
		"mock title",
		[]string{"mock", "topic"},
		"neutral",
		[]string{"keyword"},
		0.99,
		nil
}
