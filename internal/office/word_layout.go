// ABOUTME: Applies native Word section, field and editable-frame layout to prose output.
// ABOUTME: Separates physical page breaks from displayed counters and running-band variants.
package office

import (
	"fmt"
	"strconv"
	"strings"
)

func wordRole(n *docxNode) string {
	pr := n.child("w:pPr")
	if pr == nil {
		return ""
	}
	style := pr.child("w:pStyle")
	if style == nil {
		return ""
	}
	return style.get("w:val")
}
func (p *Prose) wordLayout(files map[string][]byte) error {
	root, err := parseXML(files["word/document.xml"])
	if err != nil {
		return err
	}
	body := root.child("w:body")
	if body == nil {
		return fmt.Errorf("word document has no body")
	}
	p.wordHeadingAnchors(body)
	relationships, err := parseXML(files["word/_rels/document.xml.rels"])
	if err != nil {
		return err
	}
	types, err := parseXML(files["[Content_Types].xml"])
	if err != nil {
		return err
	}
	state := Boundary{NumberFormat: p.InitialNumberFormat}
	p.wordBlockSpacing(body)
	if state.NumberFormat == "" {
		state.NumberFormat = "1"
	}
	current := p.Page.wordSection()
	current.children = append(current.children, dxVal("type", "continuous"))
	sectionIndex := 0
	bandIDs := map[string]string{}
	var output []docxNode
	bodyStarted := false
	finish := func(section *docxNode, state Boundary) error {
		sectionIndex++
		format := map[string]string{"1": "decimal", "i": "lowerRoman", "I": "upperRoman"}[state.NumberFormat]
		if format == "" {
			format = "decimal"
		}
		numbers := dx("pgNumType").attr("w:fmt", format)
		if state.Reset {
			numbers = numbers.attr("w:start", "1")
		}
		section.children = append(section.children, numbers)
		switch state.Vertical {
		case "center", "middle", "horizon":
			section.children = append(section.children, dxVal("vAlign", "center"))
		case "bottom":
			section.children = append(section.children, dxVal("vAlign", "bottom"))
		}
		first := state.SkipHeader || state.SkipFooter || state.Title || state.ContinuationPadding > 0
		if first {
			section.children = append(section.children, dx("titlePg"))
		}
		for _, item := range []struct {
			kind string
			band Band
			skip bool
		}{{"header", p.Page.Header, state.SkipHeader}, {"footer", p.Page.Footer, state.SkipFooter}} {
			if !item.band.Enabled && (!state.Title || !state.TitleNumber || item.kind != "footer") {
				continue
			}
			for _, variant := range []string{"default", "even", "first"} {
				if variant == "first" && !first {
					continue
				}
				text := item.band.Odd
				if variant == "even" {
					text = item.band.Even
				}
				if !state.Body {
					if variant == "even" {
						text = item.band.FrontEven
					} else {
						text = item.band.FrontOdd
					}
				}
				if variant == "first" && item.skip {
					text = ""
				}
				if state.Blank {
					text = ""
				}
				if state.Title && item.kind == "footer" && state.TitleNumber {
					text = "[page]"
				}
				id := fmt.Sprintf("folio%s%d%s", item.kind, sectionIndex, variant)
				name := fmt.Sprintf("%sFolio%d%s.xml", item.kind, sectionIndex, variant)
				band := item.band
				if item.kind == "header" && variant != "first" {
					band.Font.After += state.ContinuationPadding
				}
				bandRoot := dxRoot(map[string]string{"header": "hdr", "footer": "ftr"}[item.kind], p.wordBand(text, band, variant, state, "FolioBand"+item.kind))
				raw, e := docxEncode(bandRoot, DataLimit)
				if e != nil {
					return e
				}
				key := item.kind + string(raw)
				if existing := bandIDs[key]; existing != "" {
					id = existing
				} else {
					bandIDs[key] = id
					files["word/"+name] = raw
					relationships.children = append(relationships.children, docxRelationship(id, item.kind, name))
					types.children = append(types.children, ox("Override").attr("PartName", "/word/"+name).attr("ContentType", "application/vnd.openxmlformats-officedocument.wordprocessingml."+item.kind+"+xml"))
				}
				section.children = append(section.children, dx(item.kind+"Reference").attr("w:type", variant).attr("r:id", id))
			}
		}
		return nil
	}
	for _, node := range body.children {
		if node.name == "w:sectPr" {
			continue
		}
		role := wordRole(&node)
		if boundary, exists := p.Boundaries[role]; exists {
			if boundary.Body {
				if bodyStarted {
					boundary.Reset = false
				}
				bodyStarted = true
			} else {
				boundary.Body = state.Body
			}
			if boundary.NumberFormat == "" {
				boundary.NumberFormat = state.NumberFormat
			}
			if boundary.Fields == nil {
				boundary.Fields = state.Fields
			}
			if len(output) > 0 {
				if err = finish(&current, state); err != nil {
					return err
				}
				last := &output[len(output)-1]
				if last.name != "w:p" {
					output = append(output, dx("p"))
					last = &output[len(output)-1]
				}
				last.ensure("w:pPr").children = append(last.ensure("w:pPr").children, current)
			}
			kind := "continuous"
			if boundary.Break {
				kind = "nextPage"
			}
			if boundary.Before == "enforce-right" {
				kind = "oddPage"
			}
			if boundary.Before == "enforce-left" {
				kind = "evenPage"
			}
			if boundary.Before == "true" {
				output = append(output, wordBlankPage())
				kind = "nextPage"
			}
			current = p.Page.wordSection()
			current.children = append(current.children, dxVal("type", kind))
			state = boundary
		}
		for _, entry := range p.Contents {
			if entry.Style == role {
				p.wordContents(&node, entry)
			}
		}
		if placement, ok := p.Placements[role]; ok {
			p.wordPlacement(&node, role, placement)
		}
		output = append(output, node)
	}
	if err = finish(&current, state); err != nil {
		return err
	}
	output = append(output, current)
	body.children = p.wordTitleFloats(output)
	p.wordRunSizes(body)
	files["word/document.xml"], err = docxEncode(*root, DataLimit)
	if err != nil {
		return err
	}
	files["word/_rels/document.xml.rels"], err = docxEncode(*relationships, DataLimit)
	if err != nil {
		return err
	}
	files["[Content_Types].xml"], err = docxEncode(*types, DataLimit)
	if err != nil {
		return err
	}
	settings, err := parseXML(files["word/settings.xml"])
	if err != nil {
		return err
	}
	if p.Page.Header.Enabled || p.Page.Footer.Enabled {
		settings.ensure("w:evenAndOddHeaders")
	}
	files["word/settings.xml"], err = docxEncode(*settings, DataLimit)
	return err
}

// Explicit run sizes prevent readers from applying their automatic superscript
// size reduction when a paragraph's baseline position is adjusted.
func (p *Prose) wordRunSizes(root *docxNode) {
	root.visit(func(paragraph *docxNode) {
		if paragraph.name != "w:p" {
			return
		}
		style, ok := p.Styles[wordRole(paragraph)]
		if !ok {
			return
		}
		// List labels share the paragraph's native baseline. Readers do not
		// consistently apply run-position offsets to generated numbering.
		if pr := paragraph.child("w:pPr"); pr != nil && pr.child("w:numPr") != nil {
			style.BaselineShift = 0
		}
		// Native footnote anchors use the reader's paragraph baseline, as list
		// labels do. Keep their surrounding text on that same baseline.
		paragraph.visit(func(node *docxNode) {
			if node.name == "w:footnoteReference" {
				style.BaselineShift = 0
			}
		})
		paragraph.visit(func(run *docxNode) {
			if run.name != "w:r" || run.child("w:t") == nil && run.child("w:footnoteReference") == nil {
				return
			}
			pr := run.child("w:rPr")
			if pr == nil {
				run.children = append([]docxNode{dx("rPr")}, run.children...)
				pr = &run.children[0]
			}
			font := style
			if ref := pr.child("w:rStyle"); ref != nil {
				if s, exists := p.Styles[ref.get("w:val")]; exists {
					font = s
				}
			}
			pr.ensure("w:position").set("w:val", style.wordPosition())
			if pr.child("w:sz") == nil {
				pr.children = append(pr.children, dxVal("sz", strconv.Itoa(int(font.Size*2+0.5))))
			}
		})
	})
}

func (p *Prose) wordHeadingAnchors(body *docxNode) {
	var out, bookmarks []docxNode
	for _, node := range body.children {
		if node.name == "" && strings.TrimSpace(node.text) == "" {
			continue
		}
		if node.name == "w:bookmarkStart" && p.Anchors[node.get("w:name")] != "" {
			bookmarks = append(bookmarks, node)
			continue
		}
		if node.name == "w:p" {
			for _, bookmark := range bookmarks {
				node.ensure("w:pPr").ensure("w:pStyle").set("w:val", p.Anchors[bookmark.get("w:name")])
				node.children = append(node.children[:1], append([]docxNode{bookmark}, node.children[1:]...)...)
			}
			bookmarks = nil
		}
		out = append(out, node)
	}
	body.children = append(out, bookmarks...)
}
func wordBlankPage() docxNode {
	return dx("p", dx("pPr", dx("spacing").attr("w:line", "1").attr("w:lineRule", "exact")), dx("r", dx("br").attr("w:type", "page")))
}
func (p *Prose) wordBand(text string, band Band, variant string, state Boundary, role string) docxNode {
	style := band.Font
	if role == "FolioBandheader" {
		// Position the whole band at its baseline; refreshed fields discard run shifts.
		style.BaselineShift = 0
	}
	style.Align = band.Align
	if pair := strings.Split(band.Align, "-"); len(pair) == 2 {
		style.Align = pair[1]
		if variant == "even" {
			style.Align = pair[0]
		}
	}
	n := dx("p", style.wordParagraph())
	// Refreshed field results inherit paragraph typography in office readers.
	pr := n.child("w:pPr")
	pr.children = append([]docxNode{dxVal("pStyle", role)}, pr.children...)
	for _, token := range fieldTokens(text) {
		if token.field == "page" || token.field == "total-pages" {
			instruction := "PAGE"
			if token.field == "total-pages" {
				instruction = "NUMPAGES"
			}
			numberFormat := map[string]string{"i": "roman", "I": "ROMAN"}[state.NumberFormat]
			if numberFormat == "" {
				numberFormat = "ARABIC"
			}
			instruction += ` \* ` + numberFormat
			n.children = append(n.children, dx("fldSimple", dx("r", style.wordFont(), dxText("1"))).attr("w:instr", instruction))
			continue
		}
		value := token.text
		if token.field != "" {
			value = p.runningValue(token.field, state)
		}
		n.children = append(n.children, dx("r", style.wordFont(), dxText(value)))
	}
	return n
}

func (p *Prose) runningValue(field string, state Boundary) string {
	if value, exists := state.Fields[field]; exists {
		return value
	}
	if value, exists := p.Metadata[field]; exists {
		return value
	}
	switch field {
	case "part", "part-number", "part-prefix", "part-full", "chapter", "chapter-number", "chapter-prefix", "chapter-full":
		return ""
	}
	return "[" + field + "]"
}

type fieldToken struct{ text, field string }

func fieldTokens(value string) []fieldToken {
	var tokens []fieldToken
	for value != "" {
		start := strings.Index(value, "[")
		if start < 0 {
			tokens = append(tokens, fieldToken{text: value})
			break
		}
		if start > 0 {
			tokens = append(tokens, fieldToken{text: value[:start]})
		}
		end := strings.Index(value[start:], "]")
		if end < 0 {
			tokens = append(tokens, fieldToken{text: value[start:]})
			break
		}
		end += start
		tokens = append(tokens, fieldToken{field: value[start+1 : end]})
		value = value[end+1:]
	}
	return tokens
}
func (p *Prose) wordContents(n *docxNode, entry ContentsEntry) {
	if entry.Leaders {
		n.ensure("w:pPr").children = append(n.ensure("w:pPr").children, dx("tabs", dx("tab").attr("w:val", "right").attr("w:leader", "dot").attr("w:pos", twip(p.Page.Width-p.Page.Left-p.Page.Right-p.Page.Gutter))))
	}
	if entry.Links {
		var children []docxNode
		var runs []docxNode
		for _, c := range n.children {
			if c.name == "w:pPr" {
				children = append(children, c)
			} else {
				runs = append(runs, c)
			}
		}
		children = append(children, dx("hyperlink", runs...).attr("w:anchor", entry.Anchor))
		n.children = children
	}
	if entry.Pages {
		n.children = append(n.children, dx("r", dx("tab")), dx("fldSimple", dx("r", dxText("1"))).attr("w:instr", "PAGEREF "+entry.Anchor))
	}
}
func (p *Prose) wordPlacement(n *docxNode, role string, placement Placement) {
	// Native paragraph frames remain editable and keep text in document order.
	pr := n.ensure("w:pPr")
	vertical := placement.Vertical
	if vertical == "" {
		vertical = "center"
	}
	horizontal := placement.Horizontal
	if placement.Group == "title" || placement.Group == "footer" {
		// These frames span the full content width; paragraph alignment owns
		// the horizontal position. Identical frame properties join neighbours.
		horizontal = "left"
	}
	frame := dx("framePr").attr("w:hAnchor", "margin").attr("w:vAnchor", "margin").attr("w:yAlign", vertical).attr("w:xAlign", horizontal).attr("w:wrap", "around").attr("w:w", twip(p.Page.Width-p.Page.Left-p.Page.Right-p.Page.Gutter))
	pr.children = append(pr.children, frame)
}

// Word joins adjacent paragraphs with identical frame properties. Collect each
// title band and copyright bottom group so their section markers stay outside frames.
func (p *Prose) wordTitleFloats(nodes []docxNode) []docxNode {
	var output, title []docxNode
	flush := func() {
		var section *docxNode
		for i := range title {
			pr := title[i].child("w:pPr")
			if found := pr.child("w:sectPr"); found != nil {
				copy := *found
				section = &copy
				pr.remove("w:sectPr")
			}
		}
		for _, vertical := range []string{"top", "center", "bottom"} {
			for _, node := range title {
				if p.Placements[wordRole(&node)].Vertical == vertical {
					output = append(output, node)
				}
			}
		}
		if section != nil {
			// A section-ending paragraph is ordinary flow in some readers even
			// with framePr. Keep the section marker outside the positioned text.
			output = append(output, dx("p", dx("pPr", dx("spacing").attr("w:line", "1").attr("w:lineRule", "exact"), *section)))
		}
		title = nil
	}
	for _, node := range nodes {
		placement := p.Placements[wordRole(&node)]
		if placement.Group == "title" || placement.Group == "footer" || placement.Group == "copyright-bottom" {
			title = append(title, node)
			continue
		}
		flush()
		output = append(output, node)
	}
	flush()
	return output
}
