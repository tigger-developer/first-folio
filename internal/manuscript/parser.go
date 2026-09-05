// ABOUTME: Manuscript Markdown and org-mode parsers.
// ABOUTME: Produces prose manuscript blocks rather than stage-play events.
package manuscript

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// #18 AC18.2: detect and strip a leading `Part <number>[<sep>] <name>` or
// `Chapter <number>[<sep>] <name>` prefix from H1/H2 heading text. The number
// token is captured as-is (arabic, roman, or a spelled-out word); the separator
// is `:` or `.` optionally; whitespace between number/separator/name is tolerated.
// The remainder after the separator is the semantic name.
var (
	partPrefixRE    = regexp.MustCompile(`(?i)^part\s+(\w+)\s*([:.])?\s*(.*)$`)
	chapterPrefixRE = regexp.MustCompile(`(?i)^chapter\s+(\w+)\s*([:.])?\s*(.*)$`)
)

// splitHeadingPrefix returns (name, sourceNumber, separator) for a heading, stripping
// the prefix if present. If the heading does not match the prefix pattern, the whole
// heading is returned as the semantic name and sourceNumber / separator are empty.
func splitHeadingPrefix(heading string, re *regexp.Regexp) (name, sourceNumber, separator string) {
	m := re.FindStringSubmatch(heading)
	if m == nil {
		return heading, "", ""
	}
	// m[1] = number token, m[2] = separator (: or . or ""), m[3] = remainder
	name = strings.TrimSpace(m[3])
	if name == "" {
		// The prefix matched but nothing followed (e.g. "Part 1" with no name).
		// Preserve the source-order derivation and leave the semantic name empty.
	}
	return name, m[1], m[2]
}

func Parse(format string, text string) (Document, error) {
	switch format {
	case "markdown":
		return parseMarkdown(text)
	case "org":
		return parseOrg(text)
	default:
		return Document{}, fmt.Errorf("unsupported manuscript format: %s", format)
	}
}

func parseMarkdown(text string) (Document, error) {
	var doc Document
	var err error
	text, err = parseMarkdownFrontmatter(&doc.Metadata, text)
	if err != nil {
		return Document{}, err
	}
	if err := rejectUnsupportedMarkdown(text); err != nil {
		return Document{}, err
	}
	text = removeMarkdownPrivateSections(text)
	text = promoteStandaloneMarkdownCodeSpans(text)
	typst, err := runPandoc("markdown-raw_html", "typst", text)
	if err != nil {
		return Document{}, err
	}
	blocks, err := pandocTypstBlocks(typst)
	if err != nil {
		return Document{}, err
	}
	doc.Blocks = blocks
	return doc, nil
}

func promoteStandaloneMarkdownCodeSpans(text string) string {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	out := make([]string, 0, len(lines))
	inFence := false
	fenceMarker := byte(0)
	fenceLength := 0

	for _, line := range lines {
		marker, length, closing, isFence := markdownFenceLine(line, fenceMarker, fenceLength)
		if inFence {
			out = append(out, line)
			if isFence && closing {
				inFence = false
				fenceMarker = 0
				fenceLength = 0
			}
			continue
		}
		if isFence && !closing {
			inFence = true
			fenceMarker = marker
			fenceLength = length
			out = append(out, line)
			continue
		}

		content, ok := standaloneMarkdownCodeSpan(line)
		if !ok {
			out = append(out, line)
			continue
		}
		delimiter := strings.Repeat("`", max(3, longestRun(content, '`')+1))
		out = append(out, delimiter, content, delimiter)
	}

	return strings.Join(out, "\n")
}

func markdownFenceLine(line string, openMarker byte, openLength int) (byte, int, bool, bool) {
	spaces := 0
	for spaces < len(line) && line[spaces] == ' ' {
		spaces++
	}
	if spaces > 3 || spaces == len(line) {
		return 0, 0, false, false
	}
	trimmed := line[spaces:]
	marker := trimmed[0]
	if marker != '`' && marker != '~' {
		return 0, 0, false, false
	}
	length := 0
	for length < len(trimmed) && trimmed[length] == marker {
		length++
	}
	if length < 3 {
		return 0, 0, false, false
	}
	remainder := trimmed[length:]
	if openMarker != 0 {
		closing := marker == openMarker && length >= openLength && strings.TrimSpace(remainder) == ""
		return marker, length, closing, closing
	}
	if marker == '`' && strings.Contains(remainder, "`") {
		return 0, 0, false, false
	}
	return marker, length, false, true
}

func standaloneMarkdownCodeSpan(line string) (string, bool) {
	trimmed := strings.TrimSpace(line)
	if len(trimmed) < 3 || trimmed[0] != '`' {
		return "", false
	}
	delimiterLength := 0
	for delimiterLength < len(trimmed) && trimmed[delimiterLength] == '`' {
		delimiterLength++
	}
	trailingLength := 0
	for index := len(trimmed) - 1; index >= 0 && trimmed[index] == '`'; index-- {
		trailingLength++
	}
	if trailingLength != delimiterLength || len(trimmed) <= delimiterLength*2 {
		return "", false
	}
	content := trimmed[delimiterLength : len(trimmed)-delimiterLength]
	if strings.Contains(content, strings.Repeat("`", delimiterLength)) {
		return "", false
	}
	if strings.HasPrefix(content, " ") && strings.HasSuffix(content, " ") && strings.Trim(content, " ") != "" {
		content = content[1 : len(content)-1]
	}
	return content, true
}

func longestRun(text string, marker byte) int {
	longest := 0
	current := 0
	for index := 0; index < len(text); index++ {
		if text[index] == marker {
			current++
			longest = max(longest, current)
			continue
		}
		current = 0
	}
	return longest
}

func parseOrg(text string) (Document, error) {
	meta := parseOrgMetadataBlock(text)
	text = removeOrgPrivateSections(text)
	text = replaceOrgSectionBreaksWithSentinel(text)
	markdown, err := runPandoc("org", "gfm", text)
	if err != nil {
		return Document{}, err
	}
	markdown = stripRawOrgBlocks(markdown)
	markdown = restoreSectionBreakSentinel(markdown)
	markdown = normalizePandocMarkdownEscapes(markdown)
	doc, err := parseMarkdown(markdown)
	if err != nil {
		return Document{}, err
	}
	mergeMetadata(&doc.Metadata, meta)
	return doc, nil
}

func runPandoc(from string, to string, text string) (string, error) {
	cmd := exec.Command("pandoc", "-f", from, "-t", to)
	cmd.Stdin = strings.NewReader(text)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("pandoc %s to %s failed: %w: %s", from, to, err, strings.TrimSpace(stderr.String()))
	}
	return stdout.String(), nil
}

func rejectUnsupportedMarkdown(text string) error {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	for index, raw := range lines {
		lineNumber := index + 1
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		if isMarkdownSectionBreak(line) {
			if !isBlankMarkdownNeighbour(lines, index-1) || !isBlankMarkdownNeighbour(lines, index+1) {
				return fmt.Errorf("markdown section break at line %d must be surrounded by blank lines", lineNumber)
			}
			continue
		}
		if isSetextUnderline(line) && !isBlankMarkdownNeighbour(lines, index-1) {
			return fmt.Errorf("setext headings are not supported at line %d; use ATX headings", lineNumber)
		}
		if isHTMLBlockStart(line) {
			return fmt.Errorf("HTML blocks are not supported at line %d", lineNumber)
		}
	}
	return nil
}

func isMarkdownSectionBreak(line string) bool {
	return line == "***" || line == "---"
}

func isBlankMarkdownNeighbour(lines []string, index int) bool {
	return index < 0 || index >= len(lines) || strings.TrimSpace(lines[index]) == ""
}

func isSetextUnderline(line string) bool {
	if len(line) < 3 {
		return false
	}
	trimmed := strings.Trim(line, "=-")
	return trimmed == "" && (strings.HasPrefix(line, "===") || strings.HasPrefix(line, "---"))
}

func isHTMLBlockStart(line string) bool {
	if line == "" || strings.HasPrefix(line, "<!--") {
		return false
	}
	if strings.HasPrefix(line, "</") {
		return true
	}
	if !strings.HasPrefix(line, "<") || !strings.Contains(line, ">") {
		return false
	}
	if len(line) < 2 {
		return false
	}
	next := line[1]
	return next >= 'A' && next <= 'Z' || next >= 'a' && next <= 'z'
}

func removeMarkdownPrivateSections(text string) string {
	var out []string
	inNoExport := false
	noExportLevel := 0
	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		line := scanner.Text()
		level, heading := markdownHeading(line)
		if level > 0 {
			if inNoExport && level <= noExportLevel {
				inNoExport = false
			}
			if strings.Contains(heading, "<!-- noexport -->") {
				inNoExport = true
				noExportLevel = level
				continue
			}
		}
		if inNoExport || strings.HasPrefix(strings.TrimSpace(line), "<!--") {
			continue
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

func removeOrgPrivateSections(text string) string {
	var out []string
	inNoExport := false
	noExportLevel := 0
	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		line := scanner.Text()
		level, heading := orgHeading(line)
		if level > 0 {
			if inNoExport && level <= noExportLevel {
				inNoExport = false
			}
			if strings.Contains(strings.ToLower(heading), ":noexport:") {
				inNoExport = true
				noExportLevel = level
				continue
			}
		}
		if inNoExport {
			continue
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

func stripRawOrgBlocks(text string) string {
	var out []string
	inRawOrg := false
	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "```{=org}" {
			inRawOrg = true
			continue
		}
		if inRawOrg {
			if strings.TrimSpace(line) == "```" {
				inRawOrg = false
			}
			continue
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

func replaceOrgSectionBreaksWithSentinel(text string) string {
	var out []string
	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "-----" || trimmed == "_____" {
			out = append(out, "FOLIOSECTIONBREAKTOKEN")
			continue
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

func restoreSectionBreakSentinel(text string) string {
	var out []string
	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "FOLIOSECTIONBREAKTOKEN" {
			out = append(out, "***")
			continue
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

func normalizePandocMarkdownEscapes(text string) string {
	replacer := strings.NewReplacer(
		`\"`, `"`,
		`\_`, `_`,
		`\*`, `*`,
		`\[`, `[`,
		`\]`, `]`,
		`\(`, `(`,
		`\)`, `)`,
	)
	return replacer.Replace(text)
}

func parseOrgMetadataBlock(text string) Metadata {
	var meta Metadata
	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		parseOrgMetadata(&meta, scanner.Text())
	}
	return meta
}

func mergeMetadata(target *Metadata, source Metadata) {
	if source.Title != "" {
		target.Title = source.Title
	}
	if source.Subtitle != "" {
		target.Subtitle = source.Subtitle
	}
	if source.Author != "" {
		target.Author = source.Author
	}
	if source.AuthorAttribution != "" {
		target.AuthorAttribution = source.AuthorAttribution
	}
	if source.Date != "" {
		target.Date = source.Date
	}
	if source.Version != "" {
		target.Version = source.Version
	}
	if source.WordCount != "" {
		target.WordCount = source.WordCount
	}
	if source.ContactName != "" {
		target.ContactName = source.ContactName
	}
	if source.Address != "" {
		target.Address = source.Address
	}
	if source.Phone != "" {
		target.Phone = source.Phone
	}
	if source.Email != "" {
		target.Email = source.Email
	}
	if source.Website != "" {
		target.Website = source.Website
	}
}

func pandocTypstBlocks(typst string) ([]Block, error) {
	var blocks []Block
	var raw []string
	// AC18.1: derive part and chapter numbers from source order.
	// The chapter counter resets at each new part.
	partNumber := 0
	chapterNumber := 0
	flushRaw := func() {
		var chunk []string
		flushChunk := func() {
			text := strings.TrimSpace(strings.Join(chunk, "\n"))
			if text != "" {
				blocks = append(blocks, Block{Kind: "raw-typst", Text: text})
			}
			chunk = nil
		}
		for _, line := range raw {
			if strings.TrimSpace(line) == "#horizontalrule" {
				flushChunk()
				blocks = append(blocks, Block{Kind: "scene-break"})
				continue
			}
			chunk = append(chunk, line)
		}
		flushChunk()
		raw = nil
	}
	scanner := bufio.NewScanner(strings.NewReader(typst))
	for scanner.Scan() {
		line := scanner.Text()
		if isPandocTypstLabel(line) {
			continue
		}
		if level, heading := pandocTypstHeading(line); level > 0 {
			flushRaw()
			switch level {
			case 1:
				partNumber++
				chapterNumber = 0
				name, sourceNum, sep := splitHeadingPrefix(heading, partPrefixRE)
				blocks = append(blocks, Block{
					Kind:            "part",
					Level:           level,
					Text:            heading,
					Name:            name,
					Number:          partNumber,
					SourceNumber:    sourceNum,
					SourceSeparator: sep,
				})
			case 2:
				chapterNumber++
				name, sourceNum, sep := splitHeadingPrefix(heading, chapterPrefixRE)
				blocks = append(blocks, Block{
					Kind:            "chapter",
					Level:           level,
					Text:            heading,
					Name:            name,
					Number:          chapterNumber,
					SourceNumber:    sourceNum,
					SourceSeparator: sep,
				})
			default:
				blocks = append(blocks, Block{Kind: "section", Level: level, Text: heading})
			}
			continue
		}
		raw = append(raw, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	flushRaw()
	return blocks, nil
}

func pandocTypstHeading(line string) (int, string) {
	level := 0
	for level < len(line) && line[level] == '=' {
		level++
	}
	if level == 0 || len(line) <= level || line[level] != ' ' {
		return 0, ""
	}
	return level, strings.TrimSpace(line[level+1:])
}

func isPandocTypstLabel(line string) bool {
	trimmed := strings.TrimSpace(line)
	return strings.HasPrefix(trimmed, "<") && strings.HasSuffix(trimmed, ">") && !strings.Contains(trimmed, " ")
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

func parseMarkdownFrontmatter(meta *Metadata, text string) (string, error) {
	normalized := strings.ReplaceAll(text, "\r\n", "\n")
	if !strings.HasPrefix(normalized, "---\n") {
		return text, nil
	}
	end := strings.Index(normalized[len("---\n"):], "\n---\n")
	closingLen := len("\n---\n")
	if end < 0 && strings.HasSuffix(normalized, "\n---") {
		end = len(normalized) - len("---\n") - len("\n---")
		closingLen = len("\n---")
	}
	if end < 0 {
		return "", fmt.Errorf("markdown frontmatter starts with --- but has no closing ---")
	}
	content := normalized[len("---\n") : len("---\n")+end]
	remaining := normalized[len("---\n")+end+closingLen:]
	values := map[string]any{}
	if err := yaml.Unmarshal([]byte(content), &values); err != nil {
		return "", fmt.Errorf("parsing markdown frontmatter: %w", err)
	}
	applyMarkdownFrontmatter(meta, values)
	return remaining, nil
}

func applyMarkdownFrontmatter(meta *Metadata, values map[string]any) {
	for key, value := range values {
		switch strings.ToLower(strings.TrimSpace(key)) {
		case "title":
			meta.Title = frontmatterString(value)
		case "subtitle":
			meta.Subtitle = frontmatterString(value)
		case "author":
			meta.Author = frontmatterString(value)
		case "attribution", "author-attribution", "author_attribution":
			meta.AuthorAttribution = frontmatterString(value)
		case "date":
			meta.Date = frontmatterString(value)
		case "version", "draft":
			meta.Version = frontmatterString(value)
		case "wordcount", "word count", "word-count":
			meta.WordCount = frontmatterString(value)
		case "contact-name", "contact_name", "contact":
			meta.ContactName = frontmatterString(value)
		case "address":
			meta.Address = frontmatterString(value)
		case "phone", "telephone":
			meta.Phone = frontmatterString(value)
		case "email":
			meta.Email = frontmatterString(value)
		case "website":
			meta.Website = frontmatterString(value)
		}
	}
}

func frontmatterString(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return typed
	case int:
		return fmt.Sprintf("%d", typed)
	case int64:
		return fmt.Sprintf("%d", typed)
	case float64:
		return fmt.Sprintf("%.0f", typed)
	case time.Time:
		return typed.Format("2006-01-02")
	case []any:
		parts := make([]string, 0, len(typed))
		for _, item := range typed {
			if rendered := frontmatterString(item); rendered != "" {
				parts = append(parts, rendered)
			}
		}
		return strings.Join(parts, " / ")
	default:
		return fmt.Sprint(typed)
	}
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
	case "contact-name", "contact_name", "contact":
		meta.ContactName = value
	case "address":
		meta.Address = value
	case "phone":
		meta.Phone = value
	case "email":
		meta.Email = value
	case "website":
		meta.Website = value
	}
	return true
}
