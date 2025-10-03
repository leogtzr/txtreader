package stats

import "txtreader/internal/text"

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
