package analyzer

import (
	"regexp"
	"sort"
	"strings"
)

// ExtractTopKeywords implements local keyword extraction using frequency analysis
// This provides a fallback when LLM-based extraction fails or for performance reasons
// Algorithm: normalize -> tokenize -> count -> sort -> select top N
func ExtractTopKeywords(text string, topN int) []string {
	// STEP 1: Text normalization - convert to lowercase for case-insensitive matching
	normalized := strings.ToLower(text)

	// STEP 2: Tokenization using regex to extract word boundaries
	re := regexp.MustCompile(`\w+`)
	words := re.FindAllString(normalized, -1)

	// STEP 3: Frequency counting using map for O(1) lookups
	counts := make(map[string]int)
	for _, w := range words {
		counts[w]++
	}

	// STEP 4: Sort by frequency (descending) using custom comparator
	type kv struct {
		Key   string
		Value int
	}
	var freq []kv
	for k, v := range counts {
		freq = append(freq, kv{k, v})
	}
	sort.Slice(freq, func(i, j int) bool {
		return freq[i].Value > freq[j].Value
	})

	// STEP 5: Select top N most frequent words
	var top []string
	for i := 0; i < len(freq) && i < topN; i++ {
		top = append(top, freq[i].Key)
	}
	return top
}
