// Package lang detects the natural language of a text so the UI can pick the
// right readability strategy (English-tuned formulas vs. Spanish-tuned ones).
package lang

import "github.com/abadojack/whatlanggo"

// IsEnglish reports whether text is confidently detected as English.
// Short or ambiguous text (or detection failure) defaults to true, since the
// existing English-based readability formulas were the prior behavior.
func IsEnglish(text string) bool {
	info := whatlanggo.Detect(text)
	if info.Lang == -1 {
		return true
	}
	return info.Lang == whatlanggo.Eng
}
