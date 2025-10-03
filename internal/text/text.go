package text

import (
	"strings"
	"unicode"
)

func Contains(words *[]string, word string) bool {
	found := false
	for _, v := range *words {
		if v == word {
			found = true
			break
		}
	}

	return found
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
