package keywordsearch

import (
	"reflect"
	"testing"
)

func TestTokenize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Basic tokenization",
			input:    "Hello, world!",
			expected: []string{"hello", "world"},
		},
		{
			name:     "Multiple punctuation marks",
			input:    "Hello, world! How are you?",
			expected: []string{"hello", "world", "how", "are", "you"},
		},
		{
			name:     "Numbers and special characters",
			input:    "Test123, 456+test 456-test @special#chars",
			expected: []string{"test123", "456", "test", "456-test", "special", "chars"},
		},
		{
			name:     "Empty string",
			input:    "",
			expected: []string{},
		},
		{
			name:     "Only punctuation",
			input:    ".,!?-",
			expected: []string{"-"},
		},
		{
			name:     "Multiple spaces",
			input:    "  multiple   spaces  ",
			expected: []string{"multiple", "spaces"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := Tokenize(tc.input, "-")
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("tokenize(%q) = %v, want %v", tc.input, result, tc.expected)
			}
		})
	}
}

func TestLevenshtein(t *testing.T) {
	tests := []struct {
		a        string
		b        string
		expected int
	}{
		{"kitten", "sitting", 3},
		{"saturday", "sunday", 3},
		{"", "sunday", 6},
		{"saturday", "", 8},
		{"", "", 0},
		{"same", "same", 0},
		{"a", "b", 1},
		{"ab", "ac", 1},
	}

	for _, tc := range tests {
		t.Run(tc.a+"_"+tc.b, func(t *testing.T) {
			result := levenshtein(tc.a, tc.b)
			if result != tc.expected {
				t.Errorf("levenshtein(%q, %q) = %d, want %d", tc.a, tc.b, result, tc.expected)
			}
		})
	}
}

func TestFuzzyMatch(t *testing.T) {
	tests := []struct {
		s         string
		t         string
		threshold int
		expected  bool
	}{
		{"hello", "hello", 0, true},    // Exact match
		{"hello", "helo", 1, true},     // One deletion
		{"hello", "hallo", 1, true},    // One substitution
		{"hello", "heloo", 1, true},    // One deletion + one insertion
		{"hello", "hell", 1, true},     // One deletion
		{"hello", "helloo", 1, true},   // One insertion
		{"hello", "helxo", 1, true},    // One substitution
		{"hello", "hezlo", 1, true},    // One substitution
		{"hello", "goodbye", 1, false}, // Multiple changes

		// Specifically testing the hello/world case
		{"hello", "world", 3, false}, // Should fail with threshold 3
		{"hello", "world", 4, true},  // Should pass with threshold 4
	}

	for _, tc := range tests {
		t.Run(tc.s+"_"+tc.t+"_"+string(rune('0'+tc.threshold)), func(t *testing.T) {
			result := fuzzyMatch(tc.s, tc.t, tc.threshold)
			if result != tc.expected {
				// For debugging, output the actual distance
				dist := levenshtein(tc.s, tc.t)
				t.Errorf("fuzzyMatch(%q, %q, %d) = %v, want %v (actual distance: %d)",
					tc.s, tc.t, tc.threshold, result, tc.expected, dist)
			}
		})
	}
}

func TestMatchPhraseInTokens(t *testing.T) {
	tests := []struct {
		name          string
		phrase        []string
		tokens        []string
		allowedGap    int
		typoThreshold int
		expected      bool
	}{
		{
			name:          "Exact match",
			phrase:        []string{"hello", "world"},
			tokens:        []string{"hello", "world"},
			allowedGap:    0,
			typoThreshold: 0,
			expected:      true,
		},
		{
			name:          "Fuzzy match",
			phrase:        []string{"hello", "world"},
			tokens:        []string{"helo", "worl"},
			allowedGap:    0,
			typoThreshold: 1,
			expected:      true,
		},
		{
			name:          "Match with gap",
			phrase:        []string{"hello", "world"},
			tokens:        []string{"hello", "beautiful", "world"},
			allowedGap:    1,
			typoThreshold: 0,
			expected:      true,
		},
		{
			name:          "Gap too large",
			phrase:        []string{"hello", "world"},
			tokens:        []string{"hello", "beautiful", "amazing", "world"},
			allowedGap:    1,
			typoThreshold: 0,
			expected:      false,
		},
		{
			name:          "Fuzzy match with gap",
			phrase:        []string{"hello", "world"},
			tokens:        []string{"helo", "beautiful", "worl"},
			allowedGap:    1,
			typoThreshold: 1,
			expected:      true,
		},
		{
			name:          "Empty phrase",
			phrase:        []string{},
			tokens:        []string{"hello", "world"},
			allowedGap:    0,
			typoThreshold: 0,
			expected:      true,
		},
		{
			name:          "Empty tokens",
			phrase:        []string{"hello"},
			tokens:        []string{},
			allowedGap:    0,
			typoThreshold: 0,
			expected:      false,
		},
		{
			name:          "Phrase in middle",
			phrase:        []string{"hello", "world"},
			tokens:        []string{"start", "hello", "world", "end"},
			allowedGap:    0,
			typoThreshold: 0,
			expected:      true,
		},
		{
			name:          "Long phrase",
			phrase:        []string{"one", "two", "three"},
			tokens:        []string{"zero", "one", "middle", "two", "between", "three", "end"},
			allowedGap:    1,
			typoThreshold: 0,
			expected:      true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := matchPhraseInTokens(tc.phrase, tc.tokens, tc.allowedGap, tc.typoThreshold)
			if result != tc.expected {
				t.Errorf("matchPhraseInTokens(%v, %v, %d, %d) = %v, want %v",
					tc.phrase, tc.tokens, tc.allowedGap, tc.typoThreshold, result, tc.expected)
			}
		})
	}
}

func TestKeywordSearcher_Search(t *testing.T) {
	tests := []struct {
		name           string
		text           string
		keywordPhrases [][]string
		options        Options
		expected       []string
	}{
		{
			name: "Basic search",
			text: "This is a sample text with a smaple and some keyword phrases.",
			keywordPhrases: [][]string{
				{"sample", "text"},
				{"keyword", "phrases"},
				{"smaple", "and"},
				{"not", "present"},
			},
			options:  DefaultOptions(),
			expected: []string{"sample text", "keyword phrases", "smaple and"},
		},
		{
			name: "No matches",
			text: "This is a simple test",
			keywordPhrases: [][]string{
				{"complex", "algorithm"},
				{"not", "present"},
			},
			options:  DefaultOptions(),
			expected: []string{},
		},
		{
			name: "More permissive options",
			text: "This is some text with keywords and phrases spread apart.",
			keywordPhrases: [][]string{
				{"text", "keywords"},
				{"keywords", "phrases"},
			},
			options: Options{
				AllowedGap:    3,
				TypoThreshold: 1,
			},
			expected: []string{"text keywords", "keywords phrases"},
		},
		{
			name: "Less permissive options",
			text: "This is a sample text with a sample and some keyword phrases.",
			keywordPhrases: [][]string{
				{"sample", "text"},
				{"keyword", "phrases"},
				{"smaple", "and"},
			},
			options: Options{
				AllowedGap:    0,
				TypoThreshold: 0,
			},
			expected: []string{"sample text", "keyword phrases"},
		},
		{
			name:           "Empty text",
			text:           "",
			keywordPhrases: [][]string{{"sample", "text"}},
			options:        DefaultOptions(),
			expected:       []string{},
		},
		{
			name:           "Empty phrases",
			text:           "This is a sample text",
			keywordPhrases: [][]string{},
			options:        DefaultOptions(),
			expected:       []string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			searcher := NewKeywordSearcher(tc.options)
			result := searcher.Search(tc.text, tc.keywordPhrases)

			if len(result) == 0 && len(tc.expected) == 0 {
				return // Test passes if both are empty
			}

			// Sort both slices for comparison
			result = sortStrings(result)
			expected := sortStrings(tc.expected)

			if !reflect.DeepEqual(result, expected) {
				t.Errorf("Search() = %v, want %v", result, expected)
			}
		})
	}
}

// Helper function to sort string slices for consistent comparison
func sortStrings(strs []string) []string {
	if len(strs) <= 1 {
		return strs
	}

	// Simple bubble sort for small slices
	result := make([]string, len(strs))
	copy(result, strs)

	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i] > result[j] {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result
}

func BenchmarkSearch(b *testing.B) {
	text := "This is a relatively long text that contains multiple sentences. " +
		"It has some sample phrases that might match our keywords. " +
		"We want to test the performance of our keyword search algorithm. " +
		"This benchmark will help us identify potential bottlenecks. " +
		"Sample and text appear together, and keyword phrases also appear."

	keywordPhrases := [][]string{
		{"sample", "text"},
		{"keyword", "phrases"},
		{"performance", "algorithm"},
		{"benchmark", "bottlenecks"},
		{"not", "present"},
	}

	searcher := NewKeywordSearcher(DefaultOptions())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		searcher.Search(text, keywordPhrases)
	}
}

func BenchmarkLevenshtein(b *testing.B) {
	s1 := "performance"
	s2 := "performence"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		levenshtein(s1, s2)
	}
}

func BenchmarkTokenize(b *testing.B) {
	text := "This is a relatively long text that contains multiple sentences. " +
		"It has some sample phrases that might match our keywords. " +
		"We want to test the performance of our keyword-search algorithm. " +
		"This benchmark will help us identify potential bottlenecks. " +
		"Sample and text appear together, and keyword phrases also appear."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Tokenize(text, "-")
	}
}
