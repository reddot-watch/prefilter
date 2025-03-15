package prefilter

import (
	"bufio"
	"embed"
	"fmt"
	"regexp"
	"strings"

	"github.com/reddot-watch/prefilter/internal/keywordsearch"
)

//go:embed keywords/*.txt
var defaultKeywords embed.FS

// ErrLanguageNotSupported is returned when keywords for a specific language are not available
var ErrLanguageNotSupported = fmt.Errorf("language keywords not supported")

type Options struct {
	// AllowedGap defines how many extra tokens can appear between consecutive
	// phrase tokens when matching.
	AllowedGap int

	// TypoThreshold defines the maximum Levenshtein distance allowed for
	// tokens to be considered a match.
	TypoThreshold int

	// PreserveHyphens determines whether hyphens should be preserved
	// in the text during matching
	PreserveHyphens bool
}

// Filter represents a security event keyword filter
type Filter struct {
	keywords    [][]string
	searcher    *keywordsearch.KeywordSearcher
	langCode    string // ISO 639-1 language code
	customWords bool   // indicates if using custom keywords instead of embedded
}

// DefaultOptions returns the default configuration for keyword search.
func DefaultOptions() Options {
	return Options{
		AllowedGap:      2,
		TypoThreshold:   1,
		PreserveHyphens: true,
	}
}

// NewFilter creates a new security event filter for the specified language
func NewFilter(langCode string, opts Options) (*Filter, error) {
	keywords, err := loadLanguageKeywords(langCode)
	if err != nil {
		return nil, fmt.Errorf("loading keywords for language %s: %w", langCode, err)
	}

	preserveChars := ""
	if opts.PreserveHyphens {
		preserveChars = "-"
	}

	return &Filter{
		keywords: keywords,
		searcher: keywordsearch.NewKeywordSearcher(keywordsearch.Options{
			AllowedGap:           opts.AllowedGap,
			TypoThreshold:        opts.TypoThreshold,
			PreserveCertainChars: preserveChars,
		}),
		langCode: langCode,
	}, nil
}

// NewFilterWithKeywords creates a filter with custom keywords
func NewFilterWithKeywords(keywords [][]string, opts Options) *Filter {
	preserveChars := ""
	if opts.PreserveHyphens {
		preserveChars = "-"
	}

	return &Filter{
		keywords: keywords,
		searcher: keywordsearch.NewKeywordSearcher(keywordsearch.Options{
			AllowedGap:           opts.AllowedGap,
			TypoThreshold:        opts.TypoThreshold,
			PreserveCertainChars: preserveChars,
		}),
		customWords: true,
	}
}

// Language returns the ISO 639-1 language code of the loaded keywords
// Returns empty string if using custom keywords
func (f *Filter) Language() string {
	return f.langCode
}

// IsCustomKeywords returns true if the filter uses custom keywords
func (f *Filter) IsCustomKeywords() bool {
	return f.customWords
}

// Match checks if the given text matches any of the security keywords
func (f *Filter) Match(text string) bool {
	matches := f.searcher.Search(text, f.keywords)
	return len(matches) > 0
}

// MatchWithDetails returns both whether there's a match and the matching keywords
func (f *Filter) MatchWithDetails(text string) (bool, []string) {
	matches := f.searcher.Search(text, f.keywords)
	return len(matches) > 0, matches
}

// SupportedLanguages returns a list of available language codes
func SupportedLanguages() ([]string, error) {
	files, err := defaultKeywords.ReadDir("keywords")
	if err != nil {
		return nil, fmt.Errorf("reading keywords directory: %w", err)
	}

	languages := make([]string, 0, len(files))
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".txt") {
			continue
		}
		// Extract language code from filename (e.g., "en.txt" -> "en")
		langCode := strings.TrimSuffix(file.Name(), ".txt")
		if len(langCode) == 2 { // Only include valid ISO 639-1 codes
			languages = append(languages, langCode)
		}
	}

	return languages, nil
}

// loadLanguageKeywords reads keywords for a specific language
func loadLanguageKeywords(langCode string) ([][]string, error) {
	if len(langCode) != 2 {
		return nil, fmt.Errorf("invalid language code: must be ISO 639-1 (2 letters)")
	}

	filename := fmt.Sprintf("keywords/%s.txt", strings.ToLower(langCode))
	content, err := defaultKeywords.ReadFile(filename)
	if err != nil {
		// Check if the language file exists
		if strings.Contains(err.Error(), "file does not exist") {
			return nil, ErrLanguageNotSupported
		}
		return nil, fmt.Errorf("reading keyword file: %w", err)
	}

	var keywords [][]string
	seenPhrases := make(map[string]bool)
	spaceRegex := regexp.MustCompile(`\s+`)

	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	for scanner.Scan() {
		phrase := strings.TrimSpace(scanner.Text())
		if phrase == "" || phrase[0] == '#' { // Skip empty lines and comments
			continue
		}

		// Normalize spaces (trim and replace consecutive spaces with a single space)
		normalized := strings.TrimSpace(spaceRegex.ReplaceAllString(phrase, " "))

		if !seenPhrases[normalized] {
			tokens := keywordsearch.Tokenize(normalized, "-")
			if len(tokens) > 0 {
				keywords = append(keywords, tokens)
				seenPhrases[normalized] = true
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanning keyword file: %w", err)
	}

	return keywords, nil
}
