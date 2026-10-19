// ABOUTME: Applies additive clearance at native list and table boundaries.
// ABOUTME: Preserves ordinary text flow and scopes spacing to the outer block edges.
package office

import "fmt"

func (p *Prose) wordBlockSpacing(body *docxNode) {
	var paragraphs []*docxNode
	flush := func() {
		if len(paragraphs) == 0 {
			return
		}
		p.wordAddSpace(paragraphs[0], p.ListBefore, 0)
		p.wordAddSpace(paragraphs[len(paragraphs)-1], 0, p.ListAfter)
		paragraphs = nil
	}
	for i := range body.children {
		n := &body.children[i]
		if n.name == "" || n.name == "w:bookmarkStart" || n.name == "w:bookmarkEnd" {
			continue
		}
		pr := n.child("w:pPr")
		if n.name == "w:p" && pr != nil && pr.child("w:numPr") != nil {
			base := p.Styles["FolioBody"]
			pr.ensure("w:spacing").set("w:before", twip(base.Before))
			pr.ensure("w:spacing").set("w:after", twip(base.After))
			paragraphs = append(paragraphs, n)
			continue
		}
		flush()
		if n.name != "w:tbl" {
			continue
		}
		for j := i - 1; j >= 0; j-- {
			if body.children[j].name == "w:p" {
				p.wordAddSpace(&body.children[j], 0, p.TableBefore)
				break
			}
		}
		for j := i + 1; j < len(body.children); j++ {
			if body.children[j].name == "w:p" {
				p.wordAddSpace(&body.children[j], p.TableAfter, 0)
				break
			}
		}
	}
	flush()
}

func (p *Prose) wordAddSpace(n *docxNode, before, after float64) {
	style, ok := p.Styles[wordRole(n)]
	if !ok {
		style = p.Styles["FolioBody"]
	}
	spacing := n.ensure("w:pPr").ensure("w:spacing")
	if before != 0 {
		spacing.set("w:before", twip(style.Before+before))
	}
	if after != 0 {
		spacing.set("w:after", twip(style.After+after))
	}
}

func (p *Prose) odtBlockSpacing(flow, styles *docxNode) {
	index := 0
	add := func(n *docxNode, before, after float64) {
		role := odtRole(n)
		style, ok := p.Styles[role]
		if !ok {
			style = p.Styles["FolioBody"]
		}
		style.Before += before
		style.After += after
		name := fmt.Sprintf("FolioBlockEdge%d", index)
		index++
		// A paragraph can border two blocks; retain the first edge's resolved
		// properties when the second edge adds its independent clearance.
		p.Styles[name] = style
		styles.ensure("office:styles").children = append(styles.ensure("office:styles").children,
			ox("style:style", style.odtParagraph(), style.odtFont()).attr("style:name", name).attr("style:family", "paragraph").attr("style:parent-style-name", role))
		n.set("text:style-name", name)
	}
	for i := range flow.children {
		n := &flow.children[i]
		if n.name == "text:list" {
			var paragraphs []*docxNode
			n.visit(func(child *docxNode) {
				if child.name == "text:p" {
					paragraphs = append(paragraphs, child)
				}
			})
			if len(paragraphs) == 1 {
				add(paragraphs[0], p.ListBefore, p.ListAfter)
			} else if len(paragraphs) > 1 {
				add(paragraphs[0], p.ListBefore, 0)
				add(paragraphs[len(paragraphs)-1], 0, p.ListAfter)
			}
		}
		if n.name != "table:table" {
			continue
		}
		for j := i - 1; j >= 0; j-- {
			if flow.children[j].name == "text:p" {
				add(&flow.children[j], 0, p.TableBefore)
				break
			}
		}
		for j := i + 1; j < len(flow.children); j++ {
			if flow.children[j].name == "text:p" {
				add(&flow.children[j], p.TableAfter, 0)
				break
			}
		}
	}
}
