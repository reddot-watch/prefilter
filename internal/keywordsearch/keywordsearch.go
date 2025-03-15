package keywordsearch

import (
	"strings"
	"unicode"
)

// Options defines the configuration parameters for keyword search.
type Options struct {
	// AllowedGap defines how many extra tokens can appear between consecutive
	// phrase tokens when matching.
	AllowedGap int

	// TypoThreshold defines the maximum Levenshtein distance allowed for
	// tokens to be considered a match.
	TypoThreshold int

	// PreserveCertainChars specifies characters that should be preserved
	// rather than replaced with spaces during tokenization
	PreserveCertainChars string
}

// DefaultOptions returns the default configuration for keyword search.
func DefaultOptions() Options {
	return Options{
		AllowedGap:           2,
		TypoThreshold:        1,
		PreserveCertainChars: "-", // Preserve hyphens by default
	}
}

// KeywordSearcher provides functionality to search for keywords in text.
type KeywordSearcher struct {
	options Options
}

// NewKeywordSearcher creates a new KeywordSearcher with the given options.
func NewKeywordSearcher(options Options) *KeywordSearcher {
	return &KeywordSearcher{
		options: options,
	}
}

// Search takes a target text and a slice of keyword phrases,
// and returns phrases that were found in the text.
func (ks *KeywordSearcher) Search(text string, keywordPhrases [][]string) []string {
	tokens := Tokenize(text, ks.options.PreserveCertainChars)
	var matchedKeywords []string

	for _, phrase := range keywordPhrases {
		if matchPhraseInTokens(phrase, tokens, ks.options.AllowedGap, ks.options.TypoThreshold) {
			matchedKeywords = append(matchedKeywords, strings.Join(phrase, " "))
		}
	}

	return matchedKeywords
}

// Tokenize converts the text to lowercase, preserves specified characters,
// and splits it into tokens.
func Tokenize(text string, preserveChars string) []string {
	if text == "" {
		return []string{}
	}

	text = strings.ToLower(text)
	var sb strings.Builder

	// Create a map of characters to preserve for O(1) lookups
	preserve := make(map[rune]bool)
	for _, char := range preserveChars {
		preserve[char] = true
	}

	for _, r := range text {
		// Keep letters, digits, spaces, and preserved characters
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) || preserve[r] {
			sb.WriteRune(r)
		} else {
			sb.WriteRune(' ')
		}
	}
	return strings.Fields(sb.String())
}

// Minimum of three integers
func min3(a, b, c int) int {
	if a <= b && a <= c {
		return a
	} else if b <= a && b <= c {
		return b
	}
	return c
}

// levenshtein calculates the Levenshtein distance between two strings.
// It uses O(min(m,n)) space complexity where m and n are string lengths.
func levenshtein(a, b string) int {
	m, n := len(a), len(b)

	// Handle edge cases
	if m == 0 {
		return n
	}
	if n == 0 {
		return m
	}

	// Optimization: if strings are equal, distance is 0
	if a == b {
		return 0
	}

	// Optimization: ensure a is the shorter string to minimize memory usage
	if m > n {
		a, b = b, a
		m, n = n, m
	}

	// Create just two rows instead of a full matrix
	previousRow := make([]int, n+1)
	currentRow := make([]int, n+1)

	// Initialize the first row
	for j := 0; j <= n; j++ {
		previousRow[j] = j
	}

	// Fill in the distance matrix one row at a time
	for i := 1; i <= m; i++ {
		currentRow[0] = i

		for j := 1; j <= n; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}

			// Calculate minimum cost from three possible operations
			deletion := previousRow[j] + 1
			insertion := currentRow[j-1] + 1
			substitution := previousRow[j-1] + cost

			currentRow[j] = min3(deletion, insertion, substitution)
		}

		// Swap rows for next iteration
		previousRow, currentRow = currentRow, previousRow
	}

	// The result is in the last cell we calculated
	return previousRow[n]
}

// fuzzyMatch returns true if the Levenshtein distance between s and t
// is within the given threshold. It performs quick checks before computing
// the full Levenshtein distance for better performance.
func fuzzyMatch(s, t string, threshold int) bool {
	// Quick check for exact match
	if s == t {
		return true
	}

	// If threshold is 0, only exact matches are allowed
	if threshold == 0 {
		return false // We already checked for exact match above
	}

	// Quick check for length difference exceeding threshold
	if abs(len(s)-len(t)) > threshold {
		return false
	}

	return levenshtein(s, t) <= threshold
}

// matchPhraseInTokens checks if the phrase appears in the tokens in order.
// It allows for fuzzy matching of individual tokens and gaps between matches.
// Returns true if a match is found, false otherwise.
func matchPhraseInTokens(phrase, tokens []string, allowedGap, typoThreshold int) bool {
	// Handle edge cases
	if len(phrase) == 0 {
		return true // Empty phrase always matches
	}
	if len(tokens) == 0 {
		return false // Can't match anything in empty tokens
	}

	n := len(tokens)
	m := len(phrase)

	// Try every starting position in tokens.
	for i := 0; i < n; i++ {
		if !fuzzyMatch(tokens[i], phrase[0], typoThreshold) {
			continue
		}

		j := i // current position in tokens for matching
		matched := true

		// For each subsequent token in the phrase...
		for k := 1; k < m; k++ {
			found := false

			// Search from the next token up to allowedGap+1 tokens ahead.
			start := j + 1
			end := min(j+allowedGap+2, n) // +2 because we want to include j+allowedGap+1

			for l := start; l < end; l++ {
				if fuzzyMatch(tokens[l], phrase[k], typoThreshold) {
					found = true
					j = l // update current matched position
					break
				}
			}

			if !found {
				matched = false
				break
			}
		}

		if matched {
			return true
		}
	}

	return false
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
