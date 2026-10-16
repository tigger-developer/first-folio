// ABOUTME: Lowers generated-page endings and explicit blank pages into native boundaries.
// ABOUTME: Retains ordinary content roles while giving each inserted blank suppressed bands.
package office

import "fmt"

func blockRole(node Node) string {
	c, ok := node.Content.([]any)
	if node.Type != "Div" || !ok || len(c) != 2 {
		return ""
	}
	a, ok := c[0].([]any)
	if !ok || len(a) != 3 {
		return ""
	}
	pairs, ok := a[2].([][]string)
	if !ok {
		return ""
	}
	for _, pair := range pairs {
		if len(pair) == 2 && pair[0] == "custom-style" {
			return pair[1]
		}
	}
	return ""
}

func (p *Prose) flowBoundaries(blocks []Node) []Node {
	if len(p.Boundaries) == 0 {
		return blocks
	}
	var out []Node
	var pending Boundary
	for i, node := range blocks {
		role := blockRole(node)
		originalRole := role
		boundary, exists := p.Boundaries[role]
		if !exists && (pending.Break || pending.Before != "" && pending.Before != "false") {
			role = fmt.Sprintf("FolioFlow%d", i)
			p.Styles[role] = p.Styles[originalRole]
			c := node.Content.([]any)
			node.Content = []any{Attr("", nil, []string{"custom-style", role}), c[1]}
		}
		if pending.Break {
			boundary.Break = true
			exists = true
		}
		before := boundary.Before
		if pending.Before != "" && pending.Before != "false" {
			boundary.Before = pending.Before
			exists = true
		}
		for _, request := range []string{pending.Before, before} {
			if request != "true" {
				continue
			}
			name := fmt.Sprintf("FolioBlank%d_%d", i, len(out))
			style := p.Styles["FolioBody"]
			style.Size, style.Line, style.Before, style.After = 1, 1, 0, 0
			p.Styles[name] = style
			p.Boundaries[name] = Boundary{Break: true, Blank: true, SkipHeader: true, SkipFooter: true}
			out = append(out, Paragraph(name, nil))
			boundary.Break = true
			boundary.Before = "false"
			exists = true
		}
		if exists {
			p.Boundaries[role] = boundary
		}
		out = append(out, node)
		pending = p.Ends[originalRole]
	}
	return out
}
