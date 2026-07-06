// ABOUTME: Manuscript Markdown and org-mode parsers.
// ABOUTME: Produces prose manuscript blocks rather than stage-play events.
package manuscript

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"
)

func Parse(format string, text string) (Document, error) {
	switch format {
	case "markdown":
		return parseMarkdown(text), nil
	case "org":
		return parseOrg(text), nil
	default:
		return Document{}, fmt.Errorf("unsupported manuscript format: %s", format)
	}
}

func parseMarkdown(text string) Document {
	var doc Document
	var paragraph []string
	inNoExport := false
	noExportLevel := 0
	inCode := false
	var codeLines []string

	flushParagraph := func() {
		if len(paragraph) > 0 && !inNoExport {
			doc.Blocks = append(doc.Blocks, Block{Kind: "paragraph", Text: strings.Join(paragraph, " ")})
		}
		paragraph = nil
	}
	flushCode := func() {
		if len(codeLines) > 0 && !inNoExport {
			doc.Blocks = append(doc.Blocks, Block{Kind: "code", Text: strings.Join(codeLines, "\n")})
		}
		codeLines = nil
	}

	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\r")

		if strings.HasPrefix(line, "```") {
			flushParagraph()
			if inCode {
				flushCode()
				inCode = false
				continue
			}
			inCode = true
			continue
		}
		if inCode {
			codeLines = append(codeLines, line)
			continue
		}

		level, heading := markdownHeading(line)
		if level > 0 {
			flushParagraph()
			if strings.Contains(heading, "<!-- noexport -->") {
				inNoExport = true
				noExportLevel = level
				continue
			}
			if inNoExport && level <= noExportLevel {
				inNoExport = false
			}
			if inNoExport {
				continue
			}
			heading = strings.TrimSpace(strings.ReplaceAll(heading, "<!-- noexport -->", ""))
			addMarkdownHeading(&doc, level, heading)
			continue
		}

		if inNoExport {
			continue
		}
		if strings.TrimSpace(line) == "" {
			flushParagraph()
			continue
		}
		if strings.HasPrefix(strings.TrimSpace(line), "<!--") {
			continue
		}
		if strings.TrimSpace(line) == "***" {
			flushParagraph()
			doc.Blocks = append(doc.Blocks, Block{Kind: "scene-break"})
			continue
		}
		if name, value, ok := markdownFootnote(line); ok {
			flushParagraph()
			doc.Blocks = append(doc.Blocks, Block{Kind: "footnote", Text: name + "\t" + value})
			continue
		}
		if parseMarkdownMetadata(&doc.Metadata, line) {
			continue
		}
		paragraph = append(paragraph, strings.TrimSpace(line))
	}
	flushParagraph()
	flushCode()
	return doc
}

func parseOrg(text string) Document {
	var doc Document
	var paragraph []string
	inNoExport := false
	noExportLevel := 0
	inCode := false
	var codeLines []string

	flushParagraph := func() {
		if len(paragraph) > 0 && !inNoExport {
			doc.Blocks = append(doc.Blocks, Block{Kind: "paragraph", Text: strings.Join(paragraph, " ")})
		}
		paragraph = nil
	}
	flushCode := func() {
		if len(codeLines) > 0 && !inNoExport {
			doc.Blocks = append(doc.Blocks, Block{Kind: "code", Text: strings.Join(codeLines, "\n")})
		}
		codeLines = nil
	}

	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\r")
		upper := strings.ToUpper(strings.TrimSpace(line))

		if upper == "#+BEGIN_SRC" || strings.HasPrefix(upper, "#+BEGIN_SRC ") {
			flushParagraph()
			inCode = true
			continue
		}
		if upper == "#+END_SRC" {
			flushCode()
			inCode = false
			continue
		}
		if inCode {
			codeLines = append(codeLines, line)
			continue
		}
		if parseOrgMetadata(&doc.Metadata, line) {
			continue
		}

		level, heading := orgHeading(line)
		if level > 0 {
			flushParagraph()
			if strings.Contains(strings.ToLower(heading), ":noexport:") {
				inNoExport = true
				noExportLevel = level
				continue
			}
			if inNoExport && level <= noExportLevel {
				inNoExport = false
			}
			if inNoExport {
				continue
			}
			addOrgHeading(&doc, level, heading)
			continue
		}
		if inNoExport {
			continue
		}
		if strings.TrimSpace(line) == "" {
			flushParagraph()
			continue
		}
		if strings.TrimSpace(line) == "-----" {
			flushParagraph()
			doc.Blocks = append(doc.Blocks, Block{Kind: "scene-break"})
			continue
		}
		if name, value, ok := orgFootnote(line); ok {
			flushParagraph()
			doc.Blocks = append(doc.Blocks, Block{Kind: "footnote", Text: name + "\t" + value})
			continue
		}
		paragraph = append(paragraph, strings.TrimSpace(line))
	}
	flushParagraph()
	flushCode()
	return doc
}

func markdownHeading(line string) (int, string) {
	hashes := 0
	for hashes < len(line) && line[hashes] == '#' {
		hashes++
	}
	if hashes == 0 || hashes > 6 || len(line) <= hashes || line[hashes] != ' ' {
		return 0, ""
	}
	return hashes, strings.TrimSpace(line[hashes+1:])
}

func addMarkdownHeading(doc *Document, level int, heading string) {
	switch level {
	case 1:
		if doc.Metadata.Title == "" {
			doc.Metadata.Title = heading
		}
	case 2:
		doc.Blocks = append(doc.Blocks, Block{Kind: "part", Level: level, Text: heading})
	case 3:
		doc.Blocks = append(doc.Blocks, Block{Kind: "chapter", Level: level, Text: heading})
	default:
		doc.Blocks = append(doc.Blocks, Block{Kind: "section", Level: level, Text: heading})
	}
}

func addOrgHeading(doc *Document, level int, heading string) {
	switch level {
	case 1:
		doc.Blocks = append(doc.Blocks, Block{Kind: "part", Level: level, Text: heading})
	case 2:
		doc.Blocks = append(doc.Blocks, Block{Kind: "chapter", Level: level, Text: heading})
	default:
		doc.Blocks = append(doc.Blocks, Block{Kind: "section", Level: level, Text: heading})
	}
}

func orgHeading(line string) (int, string) {
	stars := 0
	for stars < len(line) && line[stars] == '*' {
		stars++
	}
	if stars == 0 || len(line) <= stars || line[stars] != ' ' {
		return 0, ""
	}
	return stars, strings.TrimSpace(line[stars+1:])
}

func parseMarkdownMetadata(meta *Metadata, line string) bool {
	trimmed := strings.TrimSpace(line)
	if meta.Subtitle == "" && strings.HasPrefix(trimmed, "**") && strings.HasSuffix(trimmed, "**") {
		meta.Subtitle = strings.TrimSuffix(strings.TrimPrefix(trimmed, "**"), "**")
		return true
	}
	if meta.Author == "" && strings.HasPrefix(trimmed, "*by ") && strings.HasSuffix(trimmed, "*") {
		meta.Author = strings.TrimSuffix(strings.TrimPrefix(trimmed, "*by "), "*")
		meta.AuthorAttribution = "by"
		return true
	}
	if strings.HasPrefix(trimmed, "--- ") && strings.HasSuffix(trimmed, " ---") {
		content := strings.TrimSuffix(strings.TrimPrefix(trimmed, "--- "), " ---")
		parts := strings.Split(content, "|")
		if len(parts) == 2 {
			meta.Version = strings.TrimSpace(parts[0])
			meta.Date = strings.TrimSpace(parts[1])
			return true
		}
		meta.Version = strings.TrimSpace(content)
		return true
	}
	return false
}

func parseOrgMetadata(meta *Metadata, line string) bool {
	if !strings.HasPrefix(line, "#+") {
		return false
	}
	parts := strings.SplitN(line[2:], ":", 2)
	if len(parts) != 2 {
		return false
	}
	key := strings.ToLower(strings.TrimSpace(parts[0]))
	value := strings.TrimSpace(parts[1])
	switch key {
	case "title":
		meta.Title = value
	case "subtitle":
		meta.Subtitle = value
	case "author":
		meta.Author = value
	case "date":
		meta.Date = value
	case "version":
		meta.Version = value
	case "wordcount":
		meta.WordCount = value
	case "address":
		meta.Address = value
	case "email":
		meta.Email = value
	case "website":
		meta.Website = value
	}
	return true
}

func markdownFootnote(line string) (string, string, bool) {
	re := regexp.MustCompile(`^\[\^([^\]]+)\]:\s+(.+)$`)
	match := re.FindStringSubmatch(line)
	if match == nil {
		return "", "", false
	}
	return match[1], match[2], true
}

func orgFootnote(line string) (string, string, bool) {
	re := regexp.MustCompile(`^\[fn:([^\]]+)\]\s+(.+)$`)
	match := re.FindStringSubmatch(line)
	if match == nil {
		return "", "", false
	}
	return match[1], match[2], true
}
