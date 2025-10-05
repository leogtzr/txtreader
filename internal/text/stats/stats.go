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

func isCommonWord(word string) bool {
	commonWords := map[string]bool{
		"the": true, "and": true, "is": true, "in": true, "to": true,
		"a": true, "of": true, "that": true, "it": true, "on": true,
		"for": true, "with": true, "as": true, "was": true, "at": true,
		"by": true, "an": true, "be": true, "this": true, "from": true,
		"de": true, "que": true, "la": true, "el": true, "y": true, "en": true,
		"se": true, "no": true, "un": true, "lo": true, "una": true, "los": true,
		"con": true, "por": true, "su": true, "las": true, "es": true, "me": true, "del": true,
		"le": true, "al": true, "como": true, "más": true, "para": true, "pero": true, "si": true,
		"yo": true, "porque": true, "nos": true, "ha": true, "o": true, "cuando": true, "está": true,
	}

	return commonWords[word]
}

func TopNFrequentWords(lines []string) []WordCount {
	wordCount := make(map[string]int)
	for _, line := range lines {
		words := strings.Fields(line)
		for _, word := range words {
			sanitized := text.SanitizeWord(word)
			if sanitized != "" && !isCommonWord(sanitized) {
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
	if len(pairs) > 10 {
		return pairs[:10]
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
