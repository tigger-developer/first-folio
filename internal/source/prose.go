// ABOUTME: Extends convert's source recognition into structured native manuscript blocks.
// ABOUTME: Preserves editable lists, tables, notes and literals within explicit source budgets.
package source

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var privateComment = regexp.MustCompile(`(?s)<!--.*?-->`)

// Prose parses the documented prose subset without an external reader.
func Prose(format, text string) ([]Node, error) {
	if format != "markdown" && format != "org" {
		return nil, fmt.Errorf("unsupported native source format %q", format)
	}
	if len(text) > 32<<20 {
		return nil, fmt.Errorf("native source exceeds 32 MiB")
	}
	p := parser{format: format, notes: map[string]string{}, resolving: map[string]bool{}}
	lines := Lines(text)
	inCode := false
	for i := 0; i < len(lines); i++ {
		trim := strings.TrimSpace(lines[i])
		if strings.HasPrefix(trim, "```") || strings.HasPrefix(trim, "~~~") || strings.HasPrefix(strings.ToLower(trim), "#+begin_src") || strings.HasPrefix(strings.ToLower(trim), "#+end_src") {
			inCode = !inCode
			continue
		}
		if inCode {
			continue
		}
		token := Recognize(format, lines[i])
		if token.Kind != "note" {
			continue
		}
		value := token.Text
		lines[i] = ""
		for i+1 < len(lines) && strings.TrimSpace(lines[i+1]) != "" && len(lines[i+1]) > 0 && (lines[i+1][0] == ' ' || lines[i+1][0] == '\t') {
			i++
			value += " " + strings.TrimSpace(lines[i])
			lines[i] = ""
		}
		p.notes[token.Name] = value
	}
	return p.parseBlocks(lines, 0)
}

func (p *parser) appendBlock(out *[]Node, n Node) error {
	p.blocks++
	if p.blocks > 50000 {
		return fmt.Errorf("native source block limit exceeded (50000)")
	}
	*out = append(*out, n)
	return nil
}

func (p *parser) parseBlocks(lines []string, depth int) ([]Node, error) {
	if depth > 128 {
		return nil, fmt.Errorf("native block nesting exceeds 128")
	}
	out := []Node{}
	for i := 0; i < len(lines); {
		token := Recognize(p.format, lines[i])
		trim := strings.TrimSpace(lines[i])
		if token.Kind == "blank" {
			i++
			continue
		}
		if p.format == "org" && (strings.HasPrefix(trim, ":PROPERTIES:") || strings.HasPrefix(trim, ":LOGBOOK:")) {
			for i < len(lines) && strings.TrimSpace(lines[i]) != ":END:" {
				i++
			}
			if i < len(lines) {
				i++
			}
			continue
		}
		if p.format == "org" && (strings.HasPrefix(trim, "# ") || trim == "#") {
			i++
			continue
		}
		if p.format == "org" && strings.HasPrefix(strings.ToLower(trim), "#+begin_") {
			parts := strings.Fields(trim)
			kind := strings.ToLower(strings.TrimPrefix(parts[0], "#+begin_"))
			end := "#+end_" + kind
			i++
			start := i
			for i < len(lines) && !strings.EqualFold(strings.TrimSpace(lines[i]), end) {
				i++
			}
			if i == len(lines) {
				return nil, fmt.Errorf("unterminated Org %s block", kind)
			}
			content := lines[start:i]
			i++
			var n Node
			switch kind {
			case "src", "example":
				n = node("CodeBlock", []any{attr(), strings.Join(content, "\n")})
			case "quote":
				children, err := p.parseBlocks(content, depth+1)
				if err != nil {
					return nil, err
				}
				n = node("BlockQuote", children)
			case "comment":
				continue
			default:
				return nil, fmt.Errorf("unsupported Org block %q", kind)
			}
			if err := p.appendBlock(&out, n); err != nil {
				return nil, err
			}
			continue
		}
		if token.Kind == "keyword" {
			i++
			continue
		}
		if mark, ok := fence(trim); ok && p.format != "org" {
			info := strings.TrimSpace(trim[len(mark):])
			i++
			start := i
			for i < len(lines) && !strings.HasPrefix(strings.TrimSpace(lines[i]), mark) {
				i++
			}
			if i == len(lines) {
				return nil, fmt.Errorf("unterminated fenced code block")
			}
			a := attr()
			if info != "" {
				a[1] = []string{strings.Fields(info)[0]}
			}
			n := node("CodeBlock", []any{a, strings.Join(lines[start:i], "\n")})
			i++
			if err := p.appendBlock(&out, n); err != nil {
				return nil, err
			}
			continue
		}
		if token.Kind == "heading" {
			text := strings.TrimSpace(privateComment.ReplaceAllString(token.Text, ""))
			spans, err := p.inline(text, depth+1)
			if err != nil {
				return nil, err
			}
			if err = p.appendBlock(&out, node("Header", []any{float64(token.Level), attr(), spans})); err != nil {
				return nil, err
			}
			i++
			continue
		}
		if rule(trim, p.format) {
			if err := p.appendBlock(&out, node("HorizontalRule", nil)); err != nil {
				return nil, err
			}
			i++
			continue
		}
		if token.Kind == "table" {
			start := i
			for i < len(lines) && Recognize(p.format, lines[i]).Kind == "table" {
				i++
			}
			n, err := p.table(lines[start:i], depth+1)
			if err != nil {
				return nil, err
			}
			if err = p.appendBlock(&out, n); err != nil {
				return nil, err
			}
			continue
		}
		if token.Kind == "quote" {
			var quoted []string
			for i < len(lines) {
				q := Recognize(p.format, lines[i])
				if q.Kind != "quote" {
					break
				}
				quoted = append(quoted, q.Text)
				i++
			}
			children, err := p.parseBlocks(quoted, depth+1)
			if err != nil {
				return nil, err
			}
			if err = p.appendBlock(&out, node("BlockQuote", children)); err != nil {
				return nil, err
			}
			continue
		}
		if token.Kind == "list" {
			n, end, err := p.list(lines, i, depth+1)
			if err != nil {
				return nil, err
			}
			i = end
			if err = p.appendBlock(&out, n); err != nil {
				return nil, err
			}
			continue
		}
		// A standalone literal is a code block; mixed prose retains an inline Code span.
		if literal, ok := standalone(trim, p.format); ok {
			if err := p.appendBlock(&out, node("CodeBlock", []any{attr(), literal})); err != nil {
				return nil, err
			}
			i++
			continue
		}
		var paragraph []string
		for i < len(lines) {
			line := lines[i]
			next := Recognize(p.format, line)
			value := strings.TrimSpace(line)
			if value == "" {
				break
			}
			if len(paragraph) > 0 && (next.Kind != "text" || rule(value, p.format) || strings.HasPrefix(strings.ToLower(value), "#+begin_") || strings.HasPrefix(value, "```")) {
				break
			}
			paragraph = append(paragraph, value)
			i++
		}
		text := strings.TrimSpace(privateComment.ReplaceAllString(strings.Join(paragraph, " "), ""))
		if text == "" {
			if len(paragraph) == 0 {
				i++
			}
			continue
		}
		spans, err := p.inline(text, depth+1)
		if err != nil {
			return nil, err
		}
		if err = p.appendBlock(&out, node("Para", spans)); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func fence(text string) (string, bool) {
	if len(text) < 3 || text[0] != '`' && text[0] != '~' {
		return "", false
	}
	i := 0
	for i < len(text) && text[i] == text[0] {
		i++
	}
	return text[:i], i >= 3
}

func standalone(text, format string) (string, bool) {
	if len(text) < 2 {
		return "", false
	}
	mark := text[0]
	if mark != '`' && (format != "org" || mark != '=' && mark != '~') {
		return "", false
	}
	if text[len(text)-1] != mark || strings.ContainsRune(text[1:len(text)-1], rune(mark)) {
		return "", false
	}
	return text[1 : len(text)-1], true
}

func rule(text, format string) bool {
	if format == "org" {
		return len(text) >= 5 && (strings.Trim(text, "-") == "" || strings.Trim(text, "_") == "")
	}
	trim := strings.ReplaceAll(text, " ", "")
	return len(trim) >= 3 && (strings.Trim(trim, "*") == "" || strings.Trim(trim, "-") == "" || strings.Trim(trim, "_") == "")
}

func ordered(mark string) bool { return len(mark) > 1 }

func (p *parser) list(lines []string, start, depth int) (Node, int, error) {
	first := Recognize(p.format, lines[start])
	i := start
	var items [][]Node
	for i < len(lines) {
		token := Recognize(p.format, lines[i])
		if token.Kind != "list" || token.Indent != first.Indent || ordered(token.Name) != ordered(first.Name) {
			break
		}
		content := []string{token.Text}
		i++
		for i < len(lines) {
			next := Recognize(p.format, lines[i])
			if next.Kind == "blank" {
				if i+1 < len(lines) && Recognize(p.format, lines[i+1]).Kind == "list" {
					i++
					break
				}
				break
			}
			if next.Indent <= first.Indent {
				break
			}
			cut := min(first.Indent+2, len(lines[i]))
			content = append(content, lines[i][cut:])
			i++
		}
		children, err := p.parseBlocks(content, depth+1)
		if err != nil {
			return Node{}, i, err
		}
		if len(children) > 0 && children[0].Type == "Para" {
			children[0].Type = "Plain"
		}
		items = append(items, children)
	}
	if !ordered(first.Name) {
		return node("BulletList", items), i, nil
	}
	style, delimiter := "Decimal", "Period"
	if strings.HasSuffix(first.Name, ")") {
		delimiter = "OneParen"
	}
	label := strings.TrimRight(first.Name, ".)")
	number, err := strconv.Atoi(label)
	if err != nil {
		number = int(strings.ToLower(label)[0]-'a') + 1
		style = "LowerAlpha"
		if label == strings.ToUpper(label) {
			style = "UpperAlpha"
		}
	}
	return node("OrderedList", []any{[]any{number, node(style, nil), node(delimiter, nil)}, items}), i, nil
}

func separator(cells []string) bool {
	for _, cell := range cells {
		if strings.Trim(cell, "-+: ") == "" {
			continue
		}
		return false
	}
	return true
}

func (p *parser) table(lines []string, depth int) (Node, error) {
	count := len(Cells(lines[0]))
	aligns := make([]Node, count)
	for i := range aligns {
		aligns[i] = node("AlignDefault", nil)
	}
	var head, body []any
	for index, line := range lines {
		cells := Cells(line)
		if separator(cells) {
			for i, cell := range cells {
				if i >= count {
					break
				}
				left, right := strings.HasPrefix(cell, ":"), strings.HasSuffix(cell, ":")
				kind := "AlignDefault"
				if left && right {
					kind = "AlignCenter"
				} else if left {
					kind = "AlignLeft"
				} else if right {
					kind = "AlignRight"
				}
				aligns[i] = node(kind, nil)
			}
			continue
		}
		if len(cells) != count {
			return Node{}, fmt.Errorf("inconsistent table column count")
		}
		var entries []any
		for _, text := range cells {
			spans, err := p.inline(text, depth+1)
			if err != nil {
				return Node{}, err
			}
			cellBlocks := []Node{}
			if err := p.appendBlock(&cellBlocks, node("Plain", spans)); err != nil {
				return Node{}, err
			}
			entries = append(entries, []any{attr(), node("AlignDefault", nil), 1, 1, cellBlocks})
		}
		row := []any{attr(), entries}
		if index == 0 && len(lines) > 1 && separator(Cells(lines[1])) {
			head = append(head, row)
		} else {
			body = append(body, row)
		}
	}
	if head == nil {
		head = []any{}
	}
	if body == nil {
		body = []any{}
	}
	columns := []any{}
	for _, alignment := range aligns {
		columns = append(columns, []any{alignment, node("ColWidthDefault", nil)})
	}
	return node("Table", []any{attr(), []any{nil, []Node{}}, columns, []any{attr(), head}, []any{[]any{attr(), 0, []any{}, body}}, []any{attr(), []any{}}}), nil
}
