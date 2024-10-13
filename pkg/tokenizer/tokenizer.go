package tokenizer

import (
	"strings"
	"unicode"
)

func CountTokens(text string) int {
	return len(strings.Fields(text))
}

func SplitIntoTokens(text string) []string {
	return strings.FieldsFunc(text, func(r rune) bool {
		return unicode.IsSpace(r) || unicode.IsPunct(r)
	})
}