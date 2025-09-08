package llm

// LLM is the interface for language models (real or mock).
type LLM interface {
	AnalyzeText(input string) (
		summary string,
		title string,
		topics []string,
		sentiment string,
		keywords []string,
		confidence float64,
		err error,
	)
}
