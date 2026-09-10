// ABOUTME: Parses Org, Markdown, and Fountain stage plays into typed events.
// ABOUTME: Applies shared intro classification after format-specific parsing.
package play

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

var (
	orgKeywordRE  = regexp.MustCompile(`^#\+(\w+):\s*(.*)$`)
	orgHeadingRE  = regexp.MustCompile(`^(\*+)\s+(.*)$`)
	orgFootnoteRE = regexp.MustCompile(`^\[fn:(\S+)\]\s+(.+)$`)
	mdFootnoteRE  = regexp.MustCompile(`^\[\^(\S+)\]:\s+(.+)$`)
)

func Parse(format Format, source string, path string) (Document, []string, error) {
	var doc Document
	var warnings []string
	var err error
	switch format {
	case FormatOrg:
		doc, err = parseOrg(source)
	case FormatMarkdown:
		doc, err = parseMarkdown(source)
	case FormatFountain:
		doc, warnings, err = parseFountain(source, path)
	default:
		err = fmt.Errorf("no parser for format %q", format)
	}
	if err != nil {
		return Document{}, nil, err
	}
	doc.Events = classifyIntro(doc.Events)
	return doc, warnings, nil
}

func parseOrg(source string) (Document, error) {
	doc := Document{Metadata: map[string]string{}}
	lines := splitLines(source)
	inNoExport := false
	noExportLevel := 0
	inTable := false

	for _, line := range lines {
		if match := orgKeywordRE.FindStringSubmatch(line); match != nil {
			key := strings.ToLower(match[1])
			doc.Metadata[key] = match[2]
			doc.Events = append(doc.Events, Event{Kind: EventFrontMatter, Key: key, Text: match[2]})
			continue
		}
		if match := orgHeadingRE.FindStringSubmatch(line); match != nil {
			level, text := len(match[1]), match[2]
			if inTable {
				doc.Events = append(doc.Events, Event{Kind: EventCharacterTableEnd})
				inTable = false
			}
			if strings.Contains(text, ":noexport:") {
				inNoExport, noExportLevel = true, level
				continue
			}
			if inNoExport {
				if level > noExportLevel {
					continue
				}
				inNoExport = false
			}
			switch level {
			case 1:
				if strings.EqualFold(strings.TrimSpace(text), "CHARACTER") || strings.EqualFold(strings.TrimSpace(text), "CHARACTERS") {
					inTable = true
					doc.Events = append(doc.Events, Event{Kind: EventCharacterTableStart, Text: strings.TrimSpace(text)})
				} else {
					doc.Events = append(doc.Events, Event{Kind: EventActHeader, Text: text})
				}
			case 2:
				doc.Events = append(doc.Events, Event{Kind: EventSceneHeader, Text: text})
			case 3:
				trimmed := strings.TrimSpace(text)
				switch {
				case strings.HasSuffix(trimmed, ":prop:"):
					doc.Events = append(doc.Events, Event{Kind: EventPropText, Text: strings.TrimSpace(strings.TrimSuffix(trimmed, ":prop:"))})
				case strings.HasSuffix(trimmed, ":transition:"):
					doc.Events = append(doc.Events, Event{Kind: EventTransition, Text: strings.TrimSpace(strings.TrimSuffix(trimmed, ":transition:"))})
				default:
					doc.Events = append(doc.Events, Event{Kind: EventStageDirection, Text: text})
				}
			case 4:
				name, direction := parseCharacterLine(text)
				doc.Events = append(doc.Events, Event{Kind: EventCharacter, Name: name, Direction: direction})
			case 5:
				doc.Events = append(doc.Events, Event{Kind: EventTransition, Text: text})
			}
			continue
		}
		if inNoExport {
			continue
		}
		if inTable && strings.HasPrefix(strings.TrimSpace(line), "|") {
			trimmed := strings.TrimSpace(line)
			if strings.Contains(trimmed, "-+-") || strings.HasPrefix(trimmed, "|-") {
				continue
			}
			cells := tableCells(trimmed)
			if len(cells) >= 2 && cells[0] != "" {
				doc.Events = append(doc.Events, Event{Kind: EventCharacterTableRow, Name: cells[0], Text: cells[1]})
			}
			continue
		}
		if inTable && strings.TrimSpace(line) != "" {
			doc.Events = append(doc.Events, Event{Kind: EventCharacterTableEnd})
			inTable = false
		}
		if match := orgFootnoteRE.FindStringSubmatch(line); match != nil {
			doc.Events = append(doc.Events, Event{Kind: EventFootnote, Name: match[1], Text: match[2]})
			continue
		}
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, `- *"`) && strings.HasSuffix(trimmed, `"*`) {
			doc.Events = append(doc.Events, Event{Kind: EventPropText, Text: strings.TrimSuffix(strings.TrimPrefix(trimmed, `- *"`), `"*`)})
			continue
		}
		if trimmed != "" {
			doc.Events = append(doc.Events, Event{Kind: EventDialogue, Text: line})
		}
	}
	if inTable {
		doc.Events = append(doc.Events, Event{Kind: EventCharacterTableEnd})
	}
	return doc, nil
}

func parseMarkdown(source string) (Document, error) {
	doc := Document{Metadata: map[string]string{}}
	lines, err := parseMarkdownSchema(&doc, splitLines(source))
	if err != nil {
		return Document{}, err
	}
	seenTitle, expectMetadata, inTable, seenCharacter := false, false, false, false
	for _, line := range lines {
		if !seenTitle && strings.HasPrefix(line, "# ") {
			addMetadata(&doc, "title", strings.TrimPrefix(line, "# "))
			seenTitle, expectMetadata = true, true
			continue
		}
		if expectMetadata && strings.HasPrefix(line, "**") && strings.HasSuffix(line, "**") {
			addMetadata(&doc, "subtitle", strings.TrimSuffix(strings.TrimPrefix(line, "**"), "**"))
			continue
		}
		if expectMetadata && strings.HasPrefix(line, "*by ") && strings.HasSuffix(line, "*") {
			addMetadata(&doc, "author", strings.TrimSuffix(strings.TrimPrefix(line, "*by "), "*"))
			continue
		}
		if expectMetadata && strings.HasPrefix(line, "--- ") && strings.HasSuffix(line, " ---") {
			parts := strings.Split(strings.TrimSuffix(strings.TrimPrefix(line, "--- "), " ---"), "|")
			addMetadata(&doc, "version", strings.TrimSpace(parts[0]))
			if len(parts) > 1 {
				addMetadata(&doc, "date", strings.TrimSpace(parts[1]))
			}
			continue
		}
		if strings.TrimSpace(line) != "" {
			expectMetadata = false
		}
		if strings.HasPrefix(line, "## ") {
			doc.Events = append(doc.Events, Event{Kind: EventActHeader, Text: strings.TrimPrefix(line, "## ")})
			continue
		}
		if strings.HasPrefix(line, "### ") {
			doc.Events = append(doc.Events, Event{Kind: EventSceneHeader, Text: strings.TrimPrefix(line, "### ")})
			continue
		}
		if strings.HasPrefix(line, "|") {
			if strings.HasPrefix(line, "|-") || strings.Contains(line, "| -") {
				continue
			}
			cells := tableCells(line)
			if len(cells) >= 2 && !seenCharacter {
				if !inTable {
					inTable = true
					doc.Events = append(doc.Events, Event{Kind: EventCharacterTableStart, Text: "Characters"})
					if isTableHeader(cells[0]) {
						continue
					}
				}
				doc.Events = append(doc.Events, Event{Kind: EventCharacterTableRow, Name: cells[0], Text: cells[1]})
			}
			continue
		}
		if inTable {
			doc.Events = append(doc.Events, Event{Kind: EventCharacterTableEnd})
			inTable = false
		}
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(line, "> ") {
			doc.Events = append(doc.Events, Event{Kind: EventTransition, Text: strings.TrimPrefix(line, "> ")})
			continue
		}
		if match := mdFootnoteRE.FindStringSubmatch(line); match != nil {
			doc.Events = append(doc.Events, Event{Kind: EventFootnote, Name: match[1], Text: match[2]})
			continue
		}
		if strings.HasPrefix(line, `***"`) && strings.HasSuffix(line, `"***`) {
			doc.Events = append(doc.Events, Event{Kind: EventPropText, Text: strings.TrimSuffix(strings.TrimPrefix(line, `***"`), `"***`)})
			continue
		}
		if strings.HasPrefix(line, "**") {
			if close := strings.Index(line[2:], ":**"); close >= 0 {
				name := line[2 : close+2]
				rest := strings.TrimSpace(line[close+5:])
				direction := strings.TrimSuffix(strings.TrimPrefix(rest, "*("), ")*")
				if direction == rest {
					direction = ""
				}
				doc.Events = append(doc.Events, Event{Kind: EventCharacter, Name: name, Direction: direction})
				seenCharacter = true
				continue
			}
		}
		if strings.HasPrefix(line, "*") && strings.HasSuffix(line, "*") && !strings.HasPrefix(line, "**") {
			doc.Events = append(doc.Events, Event{Kind: EventStageDirection, Text: strings.TrimSuffix(strings.TrimPrefix(line, "*"), "*")})
			continue
		}
		doc.Events = append(doc.Events, Event{Kind: EventDialogue, Text: line})
	}
	if inTable {
		doc.Events = append(doc.Events, Event{Kind: EventCharacterTableEnd})
	}
	return doc, nil
}

// parseMarkdownSchema consumes document frontmatter before the play-body parser.
// Other metadata retains the existing heading-based contract; schema is descriptive.
func parseMarkdownSchema(doc *Document, lines []string) ([]string, error) {
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return lines, nil
	}
	end := 1
	for end < len(lines) && strings.TrimSpace(lines[end]) != "---" {
		end++
	}
	if end == len(lines) {
		return nil, fmt.Errorf("Markdown frontmatter has no closing --- delimiter")
	}
	var metadata struct {
		Schema string `yaml:"schema"`
	}
	if err := yaml.Unmarshal([]byte(strings.Join(lines[1:end], "\n")), &metadata); err != nil {
		return nil, fmt.Errorf("parsing Markdown frontmatter: %w", err)
	}
	if metadata.Schema != "" {
		addMetadata(doc, "schema", metadata.Schema)
	}
	return lines[end+1:], nil
}

func parseFountain(source string, path string) (Document, []string, error) {
	doc := Document{Metadata: map[string]string{}}
	lines := splitLines(source)
	var warnings []string
	i := 0
	if len(lines) > 0 && strings.Contains(lines[0], ": ") {
		for i < len(lines) && strings.TrimSpace(lines[i]) != "" {
			if key, value, ok := strings.Cut(lines[i], ":"); ok {
				addMetadata(&doc, strings.ToLower(strings.TrimSpace(key)), strings.TrimSpace(value))
			}
			i++
		}
		for i < len(lines) && strings.TrimSpace(lines[i]) == "" {
			i++
		}
	}
	previousBlank, afterCharacter := true, false
	for i < len(lines) {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		lineNumber := i + 1
		if trimmed == "" {
			previousBlank, afterCharacter = true, false
			i++
			continue
		}
		if strings.HasPrefix(line, "/*") {
			for i < len(lines) && !strings.Contains(lines[i], "*/") {
				i++
			}
			i++
			continue
		}
		if strings.HasPrefix(line, "= ") {
			warnings = append(warnings, fmt.Sprintf("warning: %s:%d: dropping synopsis (no event-stream equivalent)", path, lineNumber))
			i++
			continue
		}
		if strings.HasPrefix(line, "===") {
			warnings = append(warnings, fmt.Sprintf("warning: %s:%d: dropping page break (no event-stream equivalent)", path, lineNumber))
			previousBlank = true
			i++
			continue
		}
		if strings.HasPrefix(line, "[[") && strings.HasSuffix(line, "]]") {
			doc.Events = append(doc.Events, Event{Kind: EventFootnote, Name: fmt.Sprintf("note_%d", lineNumber), Text: strings.TrimSuffix(strings.TrimPrefix(line, "[["), "]]")})
			i++
			continue
		}
		if strings.HasPrefix(line, "#") {
			level := len(line) - len(strings.TrimLeft(line, "#"))
			text := strings.TrimSpace(line[level:])
			kind := EventSceneHeader
			if level == 1 {
				kind = EventActHeader
			}
			doc.Events = append(doc.Events, Event{Kind: kind, Text: text})
			i++
			continue
		}
		if strings.HasPrefix(line, ".") {
			doc.Events = append(doc.Events, Event{Kind: EventSceneHeader, Text: strings.TrimPrefix(line, ".")})
			i++
			continue
		}
		if strings.HasPrefix(line, "> **") && strings.HasSuffix(line, "** <") {
			doc.Events = append(doc.Events, Event{Kind: EventActHeader, Text: strings.TrimSuffix(strings.TrimPrefix(line, "> **"), "** <")})
			i++
			continue
		}
		if strings.HasPrefix(line, ">") && strings.HasSuffix(line, "<") {
			doc.Events = append(doc.Events, Event{Kind: EventPropText, Text: strings.TrimSuffix(strings.TrimPrefix(line, ">"), "<")})
			i++
			continue
		}
		if previousBlank && isSceneHeading(line) {
			doc.Events = append(doc.Events, Event{Kind: EventSceneHeader, Text: line})
			i++
			previousBlank = false
			continue
		}
		if previousBlank && isTransition(line) {
			doc.Events = append(doc.Events, Event{Kind: EventTransition, Text: line})
			i++
			previousBlank = false
			continue
		}
		if strings.HasPrefix(line, ">") {
			doc.Events = append(doc.Events, Event{Kind: EventTransition, Text: strings.TrimPrefix(line, ">")})
			i++
			continue
		}
		if previousBlank && isUpperText(strings.TrimSuffix(strings.TrimPrefix(line, "@"), "^")) {
			forced := strings.HasPrefix(line, "@")
			if forced || (i+1 < len(lines) && strings.TrimSpace(lines[i+1]) != "") {
				name := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(line, "@"), "^"))
				direction := ""
				if open := strings.LastIndex(name, "("); open > 0 && strings.HasSuffix(name, ")") {
					direction = strings.TrimSuffix(name[open+1:], ")")
					name = strings.TrimSpace(name[:open])
				}
				if i+1 < len(lines) && strings.HasPrefix(lines[i+1], "(") && strings.HasSuffix(lines[i+1], ")") {
					direction = strings.TrimSuffix(strings.TrimPrefix(lines[i+1], "("), ")")
					i++
				}
				doc.Events = append(doc.Events, Event{Kind: EventCharacter, Name: name, Direction: direction})
				afterCharacter, previousBlank = true, false
				i++
				continue
			}
		}
		if afterCharacter {
			doc.Events = append(doc.Events, Event{Kind: EventDialogue, Text: line})
		} else {
			doc.Events = append(doc.Events, Event{Kind: EventStageDirection, Text: strings.TrimPrefix(line, "!")})
		}
		previousBlank = false
		i++
	}
	return doc, warnings, nil
}

func classifyIntro(events []Event) []Event {
	firstCharacter := -1
	for i, event := range events {
		if event.Kind == EventCharacter {
			firstCharacter = i
			break
		}
	}
	if firstCharacter < 0 {
		return events
	}
	boundary := -1
	for i := firstCharacter - 1; i >= 0; i-- {
		if events[i].Kind == EventActHeader {
			boundary = i
			break
		}
	}
	if boundary < 0 {
		return events
	}
	result := make([]Event, 0, len(events))
	for i, event := range events {
		if i < boundary {
			switch event.Kind {
			case EventActHeader:
				event.Kind = EventIntroHeader
			case EventDialogue, EventStageDirection, EventSceneHeader:
				event.Kind = EventIntroText
			}
		}
		result = append(result, event)
	}
	return result
}

func addMetadata(doc *Document, key string, value string) {
	doc.Metadata[key] = value
	doc.Events = append(doc.Events, Event{Kind: EventFrontMatter, Key: key, Text: value})
}

func splitLines(source string) []string {
	return strings.Split(strings.ReplaceAll(strings.TrimSuffix(source, "\n"), "\r\n", "\n"), "\n")
}

func tableCells(line string) []string {
	trimmed := strings.Trim(strings.TrimSpace(line), "|")
	parts := strings.Split(trimmed, "|")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

func isTableHeader(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "character", "characters", "name", "names", "cast", "casts", "role", "roles":
		return true
	default:
		return false
	}
}

func parseCharacterLine(text string) (string, string) {
	fields := strings.Fields(text)
	nameEnd := len(fields)
	for i, field := range fields {
		r, _ := utf8.DecodeRuneInString(strings.TrimLeft(field, "(:,"))
		if i > 0 && !unicode.IsUpper(r) {
			nameEnd = i
			break
		}
	}
	name := strings.Join(fields[:nameEnd], " ")
	direction := strings.Trim(strings.Join(fields[nameEnd:], " "), "(),: ")
	return name, direction
}

func isUpperText(text string) bool {
	hasLetter := false
	for _, r := range text {
		if unicode.IsLetter(r) {
			hasLetter = true
			if unicode.IsLower(r) {
				return false
			}
		}
	}
	return hasLetter
}

func isSceneHeading(line string) bool {
	upper := strings.ToUpper(line)
	for _, prefix := range []string{"INT.", "INT ", "EXT.", "EXT ", "EST.", "EST ", "I/E", "INT./EXT", "INT/EXT"} {
		if strings.HasPrefix(upper, prefix) {
			return true
		}
	}
	return false
}

func isTransition(line string) bool {
	return isUpperText(line) && strings.HasSuffix(strings.TrimSpace(line), "TO:")
}
