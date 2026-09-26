// ABOUTME: Supplies English discretionary breaks for geometrically stretched prose.
// ABOUTME: Embeds patterns and document-specific break positions for offline Typst output.
package manuscript

import (
	"embed"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/speedata/hyphenation"
)

//go:embed hyphenation/*.txt
var hyphenationAssets embed.FS

var englishHyphenator = sync.OnceValues(func() (*hyphenation.Lang, error) {
	patterns, err := hyphenationAssets.ReadFile("hyphenation/en-us.pat.txt")
	if err != nil {
		return nil, err
	}
	language, err := hyphenation.New(strings.NewReader(string(patterns)))
	if err != nil {
		return nil, err
	}
	return language, nil
})

var hyphenationWordRE = regexp.MustCompile(`[A-Za-z]+`)

// Typst's default language is English. Keep the same language when boxes require
// explicit break opportunities; the default-width path still uses native Typst.
func manuscriptHyphenationDictionary(body string) (string, error) {
	language, err := englishHyphenator()
	if err != nil {
		return "", fmt.Errorf("loading manuscript hyphenation patterns: %w", err)
	}
	exceptionData, err := hyphenationAssets.ReadFile("hyphenation/en-us.hyp.txt")
	if err != nil {
		return "", err
	}
	exceptions := map[string][]string{}
	for _, exception := range strings.Fields(string(exceptionData)) {
		exceptions[strings.ReplaceAll(exception, "-", "")] = strings.Split(exception, "-")
	}
	seen := map[string]bool{}
	for _, word := range hyphenationWordRE.FindAllString(body, -1) {
		// Bound the pattern library's substring search for pathological long tokens.
		if len(word) >= 5 && len(word) <= 128 {
			seen[strings.ToLower(word)] = true
		}
	}
	words := make([]string, 0, len(seen))
	for word := range seen {
		words = append(words, word)
	}
	sort.Strings(words)
	var out strings.Builder
	out.WriteString("(")
	for _, word := range words {
		positions := language.Hyphenate(word)
		if parts, ok := exceptions[word]; ok {
			positions = nil
			position := 0
			for _, part := range parts[:len(parts)-1] {
				position += len(part)
				positions = append(positions, position)
			}
		}
		var valid []int
		for _, position := range positions {
			if position >= 2 && len(word)-position >= 3 {
				valid = append(valid, position)
			}
		}
		if len(valid) == 0 {
			continue
		}
		fmt.Fprintf(&out, "\n%q: (", word)
		for _, position := range valid {
			fmt.Fprintf(&out, "%d,", position)
		}
		out.WriteString("),")
	}
	if out.Len() == 1 {
		return "(:)", nil
	}
	out.WriteString("\n)")
	return out.String(), nil
}
