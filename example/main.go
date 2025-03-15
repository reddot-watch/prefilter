package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/reddot-watch/prefilter"
)

func main() {
	langs, err := prefilter.SupportedLanguages()
	if err != nil {
		log.Fatal("Failed to get supported languages:", err)
	}
	fmt.Printf("Available language filters: %v\n", langs)

	filter, err := prefilter.NewFilter("en", prefilter.DefaultOptions())

	// Match a security headline.
	text := "At least 11 dead in indonesian boat accidents"
	//text := "Are earmarks in DOGE sights? Previous ban saved $141 billion * WorldNetDaily * by Jeremy Portnoy, Real Clear Wire"
	if ok, matches := filter.MatchWithDetails(text); ok {
		fmt.Println("Found match")
		for _, m := range matches {
			fmt.Println(m)
		}
	} else {
		fmt.Println("No match")
	}

	// Try to create filters for different languages
	languages := []string{"en", "es", "xx"} // including an invalid one
	for _, lang := range languages {
		filter, err := prefilter.NewFilter(lang, prefilter.DefaultOptions())
		if err != nil {
			if errors.Is(err, prefilter.ErrLanguageNotSupported) {
				fmt.Printf("Language '%s' is not supported, skipping...\n", lang)
				continue
			}
			log.Printf("Error creating filter for '%s': %v\n", lang, err)
			continue
		}

		// Test some events
		events := []string{
			"At least 11 dead in indonesian boat accidents",
			"7 troops killed in fighting in eastern Ukraine",
			"Repairs in Belfairs Park Close are complete",
		}

		fmt.Printf("\nChecking events with %s keywords:\n", filter.Language())
		for _, event := range events {
			matched, keywords := filter.MatchWithDetails(event)
			if matched {
				fmt.Printf("🔴  Alert - matched keywords %v in: %s\n", keywords, event)
			} else {
				fmt.Printf("✓ Safe: %s\n", event)
			}
		}
	}

	// Example with custom keywords
	customKeywords := [][]string{
		{"root", "password"},
		{"suspicious", "login"},
	}
	customFilter := prefilter.NewFilterWithKeywords(customKeywords, prefilter.DefaultOptions())

	fmt.Printf("\nChecking with custom keywords:\n")
	event := "root password changed by admin"
	if matched, keywords := customFilter.MatchWithDetails(event); matched {
		fmt.Printf("🔴  Custom filter matched keywords %v in: %s\n", keywords, event)
	}
}
