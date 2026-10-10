// ABOUTME: Capitalizes declared cast names in stage directions, never dialogue.
// ABOUTME: Matches case-insensitive names at Unicode word boundaries.
package play

import (
	"regexp"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

// CapitalizeStageNames uses only character-table entries as its lookup source.
// It runs before visibility filtering so hidden cast tables still identify names.
func CapitalizeStageNames(doc *Document) {
	names := map[string]bool{}
	for _, event := range doc.Events {
		if event.Kind == EventCharacterTableRow {
			primary, aliases := ParseCastName(event.Name)
			names[primary] = true
			for _, alias := range aliases {
				names[strings.TrimRight(alias, ".")] = true
			}
		}
	}
	delete(names, "")
	if len(names) == 0 {
		return
	}
	patterns := make([]string, 0, len(names))
	for name := range names {
		patterns = append(patterns, name)
	}
	sort.Slice(patterns, func(i, j int) bool {
		if len(patterns[i]) != len(patterns[j]) {
			return len(patterns[i]) > len(patterns[j])
		}
		return patterns[i] < patterns[j]
	})
	for i, name := range patterns {
		patterns[i] = regexp.QuoteMeta(name)
	}
	matcher := regexp.MustCompile(`(?i)(` + strings.Join(patterns, "|") + `)(?:[^\p{L}\p{M}\p{N}]|$)`)
	for i := range doc.Events {
		if doc.Events[i].Kind == EventStageDirection {
			doc.Events[i].Text = capitalizeCastMatches(doc.Events[i].Text, matcher)
		}
	}
}

func capitalizeCastMatches(text string, matcher *regexp.Regexp) string {
	var output strings.Builder
	previous := 0
	for cursor := 0; cursor < len(text); {
		match := matcher.FindStringSubmatchIndex(text[cursor:])
		if match == nil {
			break
		}
		start, end := cursor+match[2], cursor+match[3]
		before, _ := utf8.DecodeLastRuneInString(text[:start])
		cursor = end // Do not consume the trailing boundary; the next name may follow it.
		if castWordRune(before) {
			continue
		}
		output.WriteString(text[previous:start])
		output.WriteString(strings.ToUpper(text[start:end]))
		previous = end
	}
	output.WriteString(text[previous:])
	return output.String()
}

func castWordRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsMark(r) || unicode.IsNumber(r)
}

// ParseCastName separates a primary cast name from a trailing comma-separated
// alias list. The source cell remains intact for editable-format round trips.
func ParseCastName(cell string) (string, []string) {
	cell = strings.TrimSpace(cell)
	open := strings.LastIndex(cell, " (")
	if open < 1 || !strings.HasSuffix(cell, ")") {
		return cell, nil
	}
	var aliases []string
	for _, alias := range strings.Split(cell[open+2:len(cell)-1], ",") {
		if alias = strings.TrimSpace(alias); alias != "" {
			aliases = append(aliases, alias)
		}
	}
	return strings.TrimSpace(cell[:open]), aliases
}
