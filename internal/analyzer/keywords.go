package analyzer

import (
	"regexp"
	"sort"
	"strings"
)

func ExtractTopKeywords(text string, topN int) []string {
	// Normalize
	normalized := strings.ToLower(text)
	re := regexp.MustCompile(`\w+`)
	words := re.FindAllString(normalized, -1)

	// Count frequency
	counts := make(map[string]int)
	for _, w := range words {
		counts[w]++
	}

	// Sort by frequency
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

	// Pick top N
	var top []string
	for i := 0; i < len(freq) && i < topN; i++ {
		top = append(top, freq[i].Key)
	}
	return top
}
