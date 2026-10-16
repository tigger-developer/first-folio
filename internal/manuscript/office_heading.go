// ABOUTME: Composes office headings without flattening their source inline markup.
// ABOUTME: Applies configured name casing to visible text while preserving links and spans.
package manuscript

import (
	"strings"
	"unicode/utf8"

	"github.com/tigger-developer/first-folio/internal/office"
)

func officeHeadingInlines(value any, block Block, parts HeadingParts, hc HeadingConfig) any {
	transform := func(s string) string {
		if hc.CaseTransform == "upper" {
			return strings.ToUpper(s)
		}
		return s
	}
	if hc.ShowName != nil && !*hc.ShowName {
		return office.Text(transform(parts.Full))
	}
	name := block.Name
	if name == "" {
		name = block.Text
	}
	start := strings.LastIndex(block.Text, name)
	if start < 0 {
		start = 0
	}
	start = utf8.RuneCountInString(block.Text[:start])
	nameRunes := []rune(transform(parts.Name))
	position := 0
	sliceText := func(s string) string {
		var out strings.Builder
		for range []rune(s) {
			i := position - start
			if i >= 0 && i < len(nameRunes) {
				out.WriteRune(nameRunes[i])
			}
			position++
		}
		return out.String()
	}
	nodes, ok := value.([]any)
	if !ok {
		return office.Text(transform(parts.Full))
	}
	var rewrite func([]any) []any
	rewrite = func(items []any) []any {
		var result []any
		for _, item := range items {
			node, ok := item.(map[string]any)
			if !ok {
				continue
			}
			kind, _ := node["t"].(string)
			keep := true
			switch kind {
			case "Str":
				text := sliceText(node["c"].(string))
				node["c"], keep = text, text != ""
			case "Space", "SoftBreak", "LineBreak":
				keep = sliceText(" ") != ""
			case "Code":
				c := node["c"].([]any)
				text := sliceText(c[1].(string))
				c[1], keep = text, text != ""
			case "Link", "Span", "Quoted":
				c := node["c"].([]any)
				children := rewrite(c[1].([]any))
				c[1], keep = children, len(children) > 0
			default:
				if c, ok := node["c"].([]any); ok {
					children := rewrite(c)
					node["c"], keep = children, len(children) > 0
				}
			}
			if keep {
				result = append(result, node)
			}
		}
		return result
	}
	// composeHeadingParts places the semantic name immediately before suffix.
	prefixLength := len(parts.Full) - len(hc.Suffix) - len(parts.Name)
	if prefixLength < 0 {
		prefixLength = 0
	}
	var result []any
	for _, n := range office.Text(transform(parts.Full[:prefixLength])) {
		result = append(result, n)
	}
	result = append(result, rewrite(nodes)...)
	for _, n := range office.Text(transform(hc.Suffix)) {
		result = append(result, n)
	}
	return result
}
