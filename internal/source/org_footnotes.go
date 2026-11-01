// ABOUTME: Collects named Org footnotes without assigning document-mode semantics.
// ABOUTME: Preserves blank-separated paragraphs and explicit structural boundaries.
package source

import "strings"

// OrgFootnote is one named definition and the first line after its consumed body.
type OrgFootnote struct {
	Name       string
	Paragraphs []string
	End        int
}

// CollectOrgFootnote collects a definition at start through the next definition,
// headline, two consecutive blank lines, or end of input.
func CollectOrgFootnote(lines []string, start int) (OrgFootnote, bool) {
	if start < 0 || start >= len(lines) {
		return OrgFootnote{}, false
	}
	token := Recognize("org", lines[start])
	if token.Kind != "note" {
		return OrgFootnote{}, false
	}
	note := OrgFootnote{Name: token.Name, End: start + 1}
	current := []string{}
	if text := strings.TrimSpace(token.Text); text != "" {
		current = append(current, text)
	}
	flush := func() {
		if len(current) == 0 {
			return
		}
		note.Paragraphs = append(note.Paragraphs, strings.Join(current, " "))
		current = nil
	}
	for i := start + 1; i < len(lines); {
		next := Recognize("org", lines[i])
		if next.Kind == "note" || next.Kind == "heading" {
			flush()
			note.End = i
			return note, true
		}
		if next.Kind == "blank" {
			flush()
			if i+1 >= len(lines) {
				note.End = len(lines)
				return note, true
			}
			if Recognize("org", lines[i+1]).Kind == "blank" {
				note.End = i + 2
				return note, true
			}
			i++
			note.End = i
			continue
		}
		current = append(current, strings.TrimSpace(lines[i]))
		i++
		note.End = i
	}
	flush()
	note.End = len(lines)
	return note, true
}

// NormalizeOrgFootnotes emits continuation paragraphs in the indented form
// understood by standards-based Org readers while retaining named references.
func NormalizeOrgFootnotes(text string) string {
	lines := Lines(text)
	out := make([]string, 0, len(lines))
	for i := 0; i < len(lines); {
		note, ok := CollectOrgFootnote(lines, i)
		if !ok {
			out = append(out, lines[i])
			i++
			continue
		}
		if len(note.Paragraphs) == 0 {
			out = append(out, lines[i])
		} else {
			out = append(out, "[fn:"+note.Name+"] "+note.Paragraphs[0])
			for _, paragraph := range note.Paragraphs[1:] {
				out = append(out, "", "  "+paragraph)
			}
		}
		if note.End > i+1 && note.End < len(lines) {
			out = append(out, "")
		}
		i = note.End
	}
	return strings.Join(out, "\n")
}

// ExtractOrgFootnotes removes definitions from body text while retaining their
// named, structured paragraphs for native serializers.
func ExtractOrgFootnotes(text string) (string, map[string][]string) {
	lines := Lines(text)
	body := make([]string, 0, len(lines))
	notes := map[string][]string{}
	for i := 0; i < len(lines); {
		note, ok := CollectOrgFootnote(lines, i)
		if !ok {
			body = append(body, lines[i])
			i++
			continue
		}
		notes[note.Name] = append([]string(nil), note.Paragraphs...)
		i = note.End
	}
	return strings.TrimSpace(strings.Join(body, "\n")), notes
}
