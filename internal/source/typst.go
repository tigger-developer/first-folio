// ABOUTME: Projects parsed native inline semantics into safe Typst content.
// ABOUTME: Keeps source recognition out of script and letter presentation layers.
package source

import (
	"fmt"
	"strings"

	typstutil "github.com/tigger-developer/first-folio/internal/typst"
)

// TypstInline renders already parsed inline nodes, resolving script note references
// from the same structured note map used by editable office output.
func TypstInline(nodes []Node, notes map[string][]string, format string) (string, error) {
	return typstNodes(nodes, notes, format, map[string]bool{}, 0)
}

func typstNodes(nodes []Node, notes map[string][]string, format string, resolving map[string]bool, depth int) (string, error) {
	if depth > 128 {
		return "", fmt.Errorf("Typst inline nesting exceeds 128")
	}
	var out strings.Builder
	for _, n := range nodes {
		wrap := func(open, close string, children []Node) error {
			value, err := typstNodes(children, notes, format, resolving, depth+1)
			if err != nil {
				return err
			}
			out.WriteString(open)
			out.WriteString(value)
			out.WriteString(close)
			return nil
		}
		switch n.Type {
		case "Str":
			out.WriteString(typstutil.EscapeContent(n.Content.(string)))
		case "Code":
			content := n.Content.([]any)
			out.WriteString(`#mono-text[#raw("`)
			out.WriteString(typstutil.EscapeString(content[1].(string)))
			out.WriteString(`")]`)
		case "Strong":
			if err := wrap("*", "*", n.Content.([]Node)); err != nil {
				return "", err
			}
		case "Emph":
			if err := wrap("_", "_", n.Content.([]Node)); err != nil {
				return "", err
			}
		case "Underline":
			if err := wrap("#underline[", "]", n.Content.([]Node)); err != nil {
				return "", err
			}
		case "Strikeout":
			if err := wrap("#strike[", "]", n.Content.([]Node)); err != nil {
				return "", err
			}
		case "LineBreak":
			out.WriteString(" \\\n")
		case "NoteReference":
			id := n.Content.(string)
			paragraphs, ok := notes[id]
			if !ok {
				return "", fmt.Errorf("unknown footnote %q", id)
			}
			if resolving[id] {
				return "", fmt.Errorf("recursive footnote %q", id)
			}
			resolving[id] = true
			out.WriteString("#footnote[")
			for i, paragraph := range paragraphs {
				parsed, err := ScriptInline(format, paragraph, notes)
				if err != nil {
					delete(resolving, id)
					return "", err
				}
				value, err := typstNodes(parsed, notes, format, resolving, depth+1)
				if err != nil {
					delete(resolving, id)
					return "", err
				}
				if i > 0 {
					out.WriteString("#parbreak()")
				}
				out.WriteString(value)
			}
			out.WriteByte(']')
			delete(resolving, id)
		default:
			return "", fmt.Errorf("unsupported Typst inline %q", n.Type)
		}
	}
	return out.String(), nil
}
