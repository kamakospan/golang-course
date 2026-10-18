package task

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Processor is the processing logic of a task: it validates input and
// turns it into a result. Manager depends on this interface, not on a
// concrete type, so another processor (or a fake one in tests) can be
// plugged in without changing the Manager.
type Processor interface {
	Validate(input string) error
	Process(input string) (TextStats, error)
}

// TextStats is the result of analyzing a text.
type TextStats struct {
	Words        int    // words separated by whitespace
	Runes        int    // characters (Unicode code points)
	Bytes        int    // bytes in UTF-8
	Letters      int    // letters only, any alphabet
	Sentences    int    // sentence-ending marks: . ! ?
	TopWord      string // most frequent word, lower-case, punctuation trimmed
	TopWordCount int
}

// TextAnalyzer computes TextStats for a piece of text.
// It has no fields, so a value receiver is enough: there is nothing to modify.
type TextAnalyzer struct{}

// Compile-time check that TextAnalyzer satisfies Processor.
var _ Processor = TextAnalyzer{}

// Validate rejects text that is empty or contains only whitespace:
// there is nothing to analyze in it.
func (TextAnalyzer) Validate(input string) error {
	if strings.TrimSpace(input) == "" {
		return fmt.Errorf("text is empty or whitespace only: %w", ErrInvalidInput)
	}
	return nil
}

// Process analyzes the text.
func (a TextAnalyzer) Process(input string) (TextStats, error) {
	if err := a.Validate(input); err != nil {
		return TextStats{}, err
	}

	words := strings.Fields(input)
	stats := TextStats{
		Words: len(words),
		Runes: utf8.RuneCountInString(input), // not len(): Cyrillic letters take 2 bytes
		Bytes: len(input),
	}

	for _, r := range input { // range over a string yields runes, not bytes
		if unicode.IsLetter(r) {
			stats.Letters++
		}
		if r == '.' || r == '!' || r == '?' {
			stats.Sentences++
		}
	}

	stats.TopWord, stats.TopWordCount = topWord(words)
	return stats, nil
}

// topWord returns the most frequent word. Ties go to the alphabetically
// first word, so the result is deterministic (map iteration order is random).
func topWord(words []string) (string, int) {
	counts := make(map[string]int, len(words))
	for _, w := range words {
		w = strings.ToLower(strings.TrimFunc(w, func(r rune) bool {
			return !unicode.IsLetter(r) && !unicode.IsDigit(r)
		}))
		if w != "" {
			counts[w]++
		}
	}

	best, bestCount := "", 0
	for w, c := range counts {
		if c > bestCount || (c == bestCount && w < best) {
			best, bestCount = w, c
		}
	}
	return best, bestCount
}
