package text

import (
	"strings"
	"unicode"
)

func Contains(words *[]string, word string) bool {
	for _, w := range *words {
		if w == word {
			return true
		}
	}

	return false
}

func SanitizeWord(word string) string {
	var sb strings.Builder
	for _, r := range word {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}
