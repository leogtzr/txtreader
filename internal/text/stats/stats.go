// ./internal/text/stats/stats.go
package stats

import (
	"sort"
	"strings"
	"txtreader/internal/text"
)

type WordCount struct {
	Word  string
	Count int
}

func Top3FrequentWords(lines []string) []WordCount {
	wordCount := make(map[string]int)
	for _, line := range lines {
		words := strings.Fields(line)
		for _, word := range words {
			sanitized := text.SanitizeWord(strings.ToLower(word))
			if sanitized != "" {
				wordCount[sanitized] += 1
			}
		}
	}
	var pairs []WordCount
	for w, c := range wordCount {
		pairs = append(pairs, WordCount{Word: w, Count: c})
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].Count != pairs[j].Count {
			return pairs[i].Count > pairs[j].Count
		}
		return pairs[i].Word < pairs[j].Word
	})
	if len(pairs) > 3 {
		return pairs[:3]
	}
	return pairs
}

func LongestWord(words *[]string) string {
	longest := ""
	for _, word := range *words {
		sanitizedWord := text.SanitizeWord(word)
		if len(sanitizedWord) > len(longest) {
			longest = sanitizedWord
		}
	}

	return longest
}
