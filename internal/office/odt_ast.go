// ABOUTME: Projects the existing office layout into Pandoc JSON for ODT only.
// ABOUTME: Preserves scene tables and inline content without parsing a DOCX package.
package office

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type odtASTNode struct {
	Type    string `json:"t"`
	Content any    `json:"c,omitempty"`
}

func oa(kind string, content any) odtASTNode { return odtASTNode{kind, content} }
func odtAttr(style string) []any {
	pairs := [][]string{}
	if style != "" {
		pairs = append(pairs, []string{"custom-style", style})
	}
	return []any{"", []string{}, pairs}
}
func odtChild(n docxNode, name string) docxNode {
	for _, c := range n.children {
		if c.name == "w:"+name {
			return c
		}
	}
	return docxNode{}
}
func odtValue(n docxNode, key string) string {
	for _, a := range n.attrs {
		if a.Name.Local == "w:"+key {
			return a.Value
		}
	}
	return ""
}
func odtNumber(n docxNode, key string) int {
	// Internal layout attributes are generated integers; absent optional values mean zero.
	v, _ := strconv.Atoi(odtValue(n, key))
	return v
}

type odtProjection struct {
	layout   *docxWriter
	styles   []docxNode
	variants map[string]string
}

func (p *odtProjection) blocks(nodes []docxNode) []odtASTNode {
	blocks := []odtASTNode{}
	for _, n := range nodes {
		switch n.name {
		case "w:p":
			props := odtChild(n, "pPr")
			base := odtValue(odtChild(props, "pStyle"), "val")
			if base == "" {
				base = "FolioBody"
			}
			style := base
			overrides := []docxNode{}
			for _, c := range props.children {
				if c.name != "w:pStyle" {
					overrides = append(overrides, c)
				}
			}
			if len(overrides) > 0 {
				key := base + fmt.Sprintf("%#v", overrides)
				if cached, ok := p.variants[key]; ok {
					style = cached
				} else {
					style = fmt.Sprintf("%sVariant%d", base, len(p.variants)+1)
					p.variants[key] = style
					p.styles = append(p.styles, ox("style:style", odtParagraphProperties(docxNode{children: overrides})).attr("style:name", style).attr("style:family", "paragraph").attr("style:parent-style-name", base))
				}
			}
			blocks = append(blocks, oa("Div", []any{odtAttr(style), []odtASTNode{oa("Para", p.inlines(n.children))}}))
		case "w:tbl":
			grid := odtChild(n, "tblGrid")
			columns := []any{}
			total := 0
			for _, c := range grid.children {
				total += odtNumber(c, "w")
			}
			for _, c := range grid.children {
				columns = append(columns, []any{oa("AlignDefault", nil), oa("ColWidth", float64(odtNumber(c, "w"))/float64(total))})
			}
			rows := []any{}
			for _, r := range n.children {
				if r.name != "w:tr" {
					continue
				}
				cells := []any{}
				for _, c := range r.children {
					span := odtNumber(odtChild(odtChild(c, "tcPr"), "gridSpan"), "val")
					if span == 0 {
						span = 1
					}
					cells = append(cells, []any{odtAttr(""), oa("AlignDefault", nil), 1, span, p.blocks(c.children)})
				}
				rows = append(rows, []any{odtAttr(""), cells})
			}
			blocks = append(blocks, oa("Table", []any{odtAttr(""), []any{nil, []any{}}, columns, []any{odtAttr(""), []any{}}, []any{[]any{odtAttr(""), 0, []any{}, rows}}, []any{odtAttr(""), []any{}}}))
		}
	}
	return blocks
}

func (p *odtProjection) inlines(nodes []docxNode) []odtASTNode {
	out := []odtASTNode{}
	for _, run := range nodes {
		if run.name != "w:r" {
			continue
		}
		props := odtChild(run, "rPr")
		parts := []odtASTNode{}
		for _, c := range run.children {
			switch c.name {
			case "w:t":
				parts = append(parts, oa("Str", c.text))
			case "w:br":
				parts = append(parts, oa("LineBreak", nil))
			case "w:tab":
				parts = append(parts, oa("Str", "\t"))
			case "w:footnoteReference":
				id := odtNumber(c, "id")
				if id > 0 && id <= len(p.layout.footnotes) {
					note := p.layout.footnotes[id-1].children[0]
					// ODF supplies the citation: omit Word's anchor but retain its text separator.
					note.children = append([]docxNode{note.children[0]}, note.children[2:]...)
					parts = append(parts, oa("Note", p.blocks([]docxNode{note})))
				}
			}
		}
		for _, property := range []struct{ word, pandoc string }{{"b", "Strong"}, {"i", "Emph"}, {"u", "Underline"}} {
			if odtChild(props, property.word).name != "" {
				parts = []odtASTNode{oa(property.pandoc, parts)}
			}
		}
		if odtChild(props, "rFonts").name != "" {
			parts = []odtASTNode{oa("Span", []any{odtAttr("FolioCode"), parts})}
		}
		out = append(out, parts...)
	}
	return out
}

func (p *odtProjection) json(body []docxNode) ([]byte, error) {
	blocks := p.blocks(body)
	startStyle := "FolioBody"
	if len(blocks) > 0 && blocks[0].Type == "Div" {
		content := blocks[0].Content.([]any)
		startStyle = content[0].([]any)[2].([][]string)[0][1]
		content[0] = odtAttr("FolioStart")
	} else {
		blocks = append([]odtASTNode{oa("Div", []any{odtAttr("FolioStart"), []odtASTNode{oa("Para", []odtASTNode{})}})}, blocks...)
	}
	p.styles = append(p.styles, ox("style:style").attr("style:name", "FolioStart").attr("style:family", "paragraph").attr("style:parent-style-name", startStyle).attr("style:master-page-name", "FolioFirst"))

	var b docxLimitedBuffer
	b.limit = docxXMLLimit
	err := json.NewEncoder(&b).Encode(map[string]any{"pandoc-api-version": []int{1, 23, 1, 1}, "meta": map[string]any{}, "blocks": blocks})
	return b.Bytes(), err
}
