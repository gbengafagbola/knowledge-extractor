package llm

// LLM interface demonstrates Dependency Inversion Principle
// High-level modules (server) depend on abstractions (interface), not concretions
// This enables:
// 1. Easy testing with mock implementations
// 2. Runtime switching between different LLM providers
// 3. Resilience patterns with fallback mechanisms
type LLM interface {
	AnalyzeText(input string) (
		summary string, // 1-2 sentence summary
		title string, // Extracted or generated title
		topics []string, // 3 key topics identified
		sentiment string, // positive/neutral/negative
		keywords []string, // 3 most frequent nouns
		confidence float64, // Analysis confidence score
		err error,
	)
}
