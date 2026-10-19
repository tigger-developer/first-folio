// ABOUTME: Recognizes native Org and Markdown source constructs shared by all document modes.
// ABOUTME: Keeps source syntax separate from play, manuscript and letter interpretation.
package source

import (
	"regexp"
	"strings"
)

// Line describes source syntax without assigning mode-specific semantics.
type Line struct {
	Raw, Kind, Text, Name string
	Level, Indent         int
	Cells                 []string
}

var orgKeyword = regexp.MustCompile(`^#\+([\w-]+):\s*(.*)$`)
var orgHeading = regexp.MustCompile(`^(\*+)\s+(.*)$`)
var mdHeading = regexp.MustCompile(`^(#{1,6})\s+(.*)$`)
var orgNote = regexp.MustCompile(`^\[fn:([^\]]+)\]\s*(.*)$`)
var mdNote = regexp.MustCompile(`^\[\^([^\]]+)\]:\s*(.*)$`)
var listItem = regexp.MustCompile(`^(\s*)([-+*]|[0-9]+[.)]|[a-zA-Z][.)])\s+(.*)$`)

func Lines(text string) []string {
	return strings.Split(strings.ReplaceAll(strings.TrimSuffix(text, "\n"), "\r\n", "\n"), "\n")
}

func Cells(line string) []string {
	trimmed := strings.TrimSpace(line)
	trimmed = strings.TrimPrefix(trimmed, "|")
	trimmed = strings.TrimSuffix(trimmed, "|")
	parts := strings.Split(trimmed, "|")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

// Recognize is used by the existing convert parser and the prose block parser.
func Recognize(format, line string) Line {
	token := Line{Raw: line, Text: line, Indent: len(line) - len(strings.TrimLeft(line, " \t"))}
	trim := strings.TrimSpace(line)
	if trim == "" {
		token.Kind = "blank"
		return token
	}
	heading, note := mdHeading, mdNote
	if format == "org" {
		heading, note = orgHeading, orgNote
		if m := orgKeyword.FindStringSubmatch(line); m != nil {
			token.Kind, token.Name, token.Text = "keyword", strings.ToLower(m[1]), m[2]
			return token
		}
	}
	if m := heading.FindStringSubmatch(line); m != nil {
		token.Kind, token.Level, token.Text = "heading", len(m[1]), m[2]
		return token
	}
	if m := note.FindStringSubmatch(line); m != nil {
		token.Kind, token.Name, token.Text = "note", m[1], m[2]
		return token
	}
	if strings.HasPrefix(trim, "|") {
		token.Kind, token.Cells = "table", Cells(trim)
		return token
	}
	if m := listItem.FindStringSubmatch(line); m != nil {
		token.Kind, token.Name, token.Text = "list", m[2], m[3]
		return token
	}
	if strings.HasPrefix(trim, ">") {
		token.Kind, token.Text = "quote", strings.TrimPrefix(strings.TrimPrefix(trim, ">"), " ")
		return token
	}
	token.Kind = "text"
	return token
}

// Node is a source structure whose wire representation is the supported Pandoc JSON shape.
// DOCX consumes it natively; only the ODT writer requires its JSON representation.
type Node struct {
	Type    string `json:"t"`
	Content any    `json:"c,omitempty"`
}

func node(kind string, value any) Node { return Node{kind, value} }
func attr() []any                      { return []any{"", []string{}, [][]string{}} }
