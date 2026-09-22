// Package esreadability computes readability metrics for Spanish text.
//
// The classic Flesch Reading Ease formula (and the syllable-counting rules
// behind it) are tuned for English phonics and systematically misjudge
// Spanish text, which averages more syllables per word. This package uses
// the Szigriszt-Pazos formula, the standard Spanish adaptation of Flesch,
// together with a syllable counter built on Spanish diphthong/hiatus rules
// instead of English silent-e heuristics.
package esreadability

import (
	_ "embed"
	"strings"

	"github.com/neurosnap/sentences"
	txttools "txtreader/internal/text"
)

//go:embed data/spanish.json
var spanishTrainingData []byte

var strongVowels = map[rune]bool{
	'a': true, 'e': true, 'o': true,
	'á': true, 'é': true, 'í': true, 'ó': true, 'ú': true,
}

var weakVowels = map[rune]bool{
	'i': true, 'u': true, 'ü': true,
}

func isVowel(r rune) bool {
	return strongVowels[r] || weakVowels[r]
}

// CountSyllablesWord counts the syllables in a single Spanish word.
//
// Rule: two "strong" vowels (a, e, o, or any accented vowel) next to each
// other always break into separate syllables (hiatus). Any other vowel
// pairing (weak-weak, weak-strong, strong-weak) merges into one syllable
// nucleus (diphthong/triphthong). "y" behaves as a weak vowel when it closes
// a diphthong (preceded by a vowel, not followed by one), and as a consonant
// otherwise. Intervocalic "h" is silent and ignored, since it doesn't block
// diphthong formation in Spanish (e.g. "búho", "ahumado").
func CountSyllablesWord(word string) int {
	word = strings.ToLower(word)
	if word == "" {
		return 0
	}

	runes := make([]rune, 0, len(word))
	for _, r := range word {
		if r != 'h' {
			runes = append(runes, r)
		}
	}

	isVowelAt := func(i int) bool {
		r := runes[i]
		if isVowel(r) {
			return true
		}
		if r == 'y' {
			prevIsVowel := i > 0 && isVowel(runes[i-1])
			nextIsVowel := i+1 < len(runes) && isVowel(runes[i+1])
			return prevIsVowel && !nextIsVowel
		}
		return false
	}

	isStrongAt := func(i int) bool {
		return strongVowels[runes[i]]
	}

	count := 0
	prevVowelIdx := -1
	for i := range runes {
		if !isVowelAt(i) {
			prevVowelIdx = -1
			continue
		}
		if prevVowelIdx == -1 || (isStrongAt(i) && isStrongAt(prevVowelIdx)) {
			count++
		}
		prevVowelIdx = i
	}

	if count == 0 {
		count = 1
	}
	return count
}

// Analysis holds the counts behind the Szigriszt-Pazos formula.
type Analysis struct {
	Ease          float64
	WordCount     int
	SentenceCount int
	SyllableCount int
}

// Analyze computes Spanish readability stats for text.
func Analyze(text string) Analysis {
	var a Analysis

	for _, w := range strings.Fields(text) {
		sanitized := txttools.SanitizeWord(w)
		if sanitized == "" {
			continue
		}
		a.WordCount++
		a.SyllableCount += CountSyllablesWord(sanitized)
	}

	a.SentenceCount = countSentences(text)

	if a.WordCount > 0 && a.SentenceCount > 0 {
		syllablesPerWord := float64(a.SyllableCount) / float64(a.WordCount)
		wordsPerSentence := float64(a.WordCount) / float64(a.SentenceCount)
		a.Ease = 206.84 - 62.3*syllablesPerWord - wordsPerSentence
	}

	return a
}

func countSentences(text string) int {
	storage, err := sentences.LoadTraining(spanishTrainingData)
	if err != nil {
		return 0
	}
	tokenizer := sentences.NewSentenceTokenizer(storage)
	return len(tokenizer.Tokenize(text))
}

// InflesZLabel returns the INFLESZ scale label for a Szigriszt-Pazos score.
func InflesZLabel(score float64) string {
	switch {
	case score > 80:
		return "(muy fácil)"
	case score > 65:
		return "(bastante fácil)"
	case score > 55:
		return "(normal)"
	case score > 40:
		return "(algo difícil)"
	default:
		return "(muy difícil)"
	}
}
