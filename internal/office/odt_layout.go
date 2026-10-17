// ABOUTME: Applies ODT page masters, editable frames and live references to prose.
// ABOUTME: Uses master changes at requested breaks and approved frontmatter/body transitions.
package office

import (
	"encoding/xml"
	"fmt"
	"strings"
)

func odtRole(n *docxNode) string { return n.get("text:style-name") }
func (p *Prose) odtLayout(files map[string][]byte) error {
	root, err := parseXML(files["content.xml"])
	if err != nil {
		return err
	}
	body := root.child("office:body")
	if body == nil {
		return fmt.Errorf("ODT has no body")
	}
	flow := body.child("office:text")
	if flow == nil {
		return fmt.Errorf("ODT has no text flow")
	}
	styles, err := parseXML(files["styles.xml"])
	if err != nil {
		return err
	}
	for _, n := range []*docxNode{root, styles} {
		n.set("xmlns:draw", "urn:oasis:names:tc:opendocument:xmlns:drawing:1.0")
		n.set("xmlns:xlink", "http://www.w3.org/1999/xlink")
		n.set("xmlns:svg", "urn:oasis:names:tc:opendocument:xmlns:svg-compatible:1.0")
		n.set("xmlns:fo", "urn:oasis:names:tc:opendocument:xmlns:xsl-fo-compatible:1.0")
	}
	root.visit(func(n *docxNode) {
		if n.name != "text:h" {
			return
		}
		n.visit(func(child *docxNode) {
			if role := p.Anchors[child.get("text:name")]; role != "" {
				n.set("text:style-name", role)
			}
		})
	})
	auto := styles.ensure("office:automatic-styles")
	masters := styles.ensure("office:master-styles")
	masters.children = nil
	state := Boundary{NumberFormat: p.InitialNumberFormat}
	if state.NumberFormat == "" {
		state.NumberFormat = "1"
	}
	masterName := "FolioMaster0"
	index := 0
	makeMaster := func(name string, state Boundary) {
		layoutName := name + "Layout"
		layout := p.Page.odtLayout(layoutName)
		if state.Before == "enforce-right" {
			layout.set("style:page-usage", "right")
		}
		if state.Before == "enforce-left" {
			layout.set("style:page-usage", "left")
		}
		master := ox("style:master-page").attr("style:name", name).attr("style:page-layout-name", layoutName)
		for _, item := range []struct {
			kind string
			band Band
			skip bool
		}{{"header", p.Page.Header, state.SkipHeader}, {"footer", p.Page.Footer, state.SkipFooter}} {
			if !item.band.Enabled && !(state.Title && state.TitleNumber && item.kind == "footer") {
				continue
			}
			for _, variant := range []string{"", "-left", "-first"} {
				if variant == "-first" && !item.skip {
					continue
				}
				text := item.band.Odd
				if variant == "-left" {
					text = item.band.Even
				}
				if !state.Body {
					if variant == "-left" {
						text = item.band.FrontEven
					} else {
						text = item.band.FrontOdd
					}
				}
				if state.Title && item.kind == "footer" && state.TitleNumber {
					text = "[page]"
				}
				if state.Blank {
					text = ""
				}
				band := ox("style:"+item.kind+variant, p.odtBand(text, item.band, item.kind, variant, state, styles))
				if variant == "-first" && item.skip {
					band.children = nil
					band = band.attr("style:display", "false")
				}
				master.children = append(master.children, band)
			}
			height := item.band.Font.Line
			props := ox("style:header-footer-properties").attr("fo:min-height", point(height)).attr("fo:margin-left", "0pt").attr("fo:margin-right", "0pt")
			gap := item.band.Clearance - height
			if gap < 0 {
				gap = 0
			}
			if item.kind == "header" {
				props = props.attr("fo:margin-bottom", point(gap))
			} else {
				props = props.attr("fo:margin-top", point(gap))
			}
			layout.children = append(layout.children, ox("style:"+item.kind+"-style", props))
			pageProps := layout.child("style:page-layout-properties")
			if item.kind == "header" {
				pageProps.set("fo:margin-top", point(item.band.Edge))
			} else {
				pageProps.set("fo:margin-bottom", point(item.band.Edge))
			}
		}
		auto.children = append(auto.children, layout)
		if strings.HasPrefix(state.Before, "enforce-") || state.SkipHeader || state.SkipFooter || state.ContinuationPadding > 0 {
			followingName := name + "Following"
			followingLayout := layout
			followingLayout.attrs = append([]xml.Attr(nil), layout.attrs...)
			followingLayout.set("style:name", layoutName+"Following")
			usage := "all"
			if p.Page.Gutter != 0 {
				usage = "mirrored"
			}
			followingLayout.set("style:page-usage", usage)
			if state.ContinuationPadding > 0 {
				followingLayout.children = append([]docxNode(nil), layout.children...)
				props := followingLayout.child("style:page-layout-properties")
				props.attrs = append([]xml.Attr(nil), props.attrs...)
				if !p.Page.Header.Enabled {
					props.set("fo:margin-top", point(p.Page.Top+state.ContinuationPadding))
				} else {
					header := followingLayout.child("style:header-style")
					header.children = append([]docxNode(nil), header.children...)
					hp := header.child("style:header-footer-properties")
					hp.attrs = append([]xml.Attr(nil), hp.attrs...)
					gap, _ := Points(hp.get("fo:margin-bottom"), 0, 0)
					hp.set("fo:margin-bottom", point(gap+state.ContinuationPadding))
				}
			}
			following := master
			following.attrs = append([]xml.Attr(nil), master.attrs...)
			following.children = append([]docxNode(nil), master.children...)
			following.set("style:name", followingName)
			following.set("style:page-layout-name", layoutName+"Following")
			// A continuation is not another section opening. Omitting the
			// first-page variants lets the reader select the physical parity.
			following.remove("style:header-first")
			following.remove("style:footer-first")
			master.set("style:next-style-name", followingName)
			auto.children = append(auto.children, followingLayout)
			masters.children = append(masters.children, master, following)
			return
		}
		masters.children = append(masters.children, master)
	}
	makeMaster(masterName, state)
	bodyStarted := false
	assigned := false
	for i := range flow.children {
		n := &flow.children[i]
		if n.name != "text:p" && n.name != "text:h" {
			continue
		}
		role := odtRole(n)
		boundary, exists := p.Boundaries[role]
		change := !assigned
		if exists {
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
			change = change || boundary.Break || boundary.Before == "true" || strings.HasPrefix(boundary.Before, "enforce-") || boundary.NumberFormat != state.NumberFormat || p.odtBodyMasterChange(state, boundary)
			if change {
				index++
				masterName = fmt.Sprintf("FolioMaster%d", index)
				makeMaster(masterName, boundary)
			}
			state = boundary
		}
		if change {
			name := fmt.Sprintf("FolioPageParagraph%d", index)
			props := ox("style:paragraph-properties")
			if assigned {
				props = props.attr("fo:break-before", "page")
			}
			if state.Reset {
				props = props.attr("style:page-number", "1")
			}
			pageStyle := ox("style:style", props).attr("style:name", name).attr("style:family", "paragraph").attr("style:parent-style-name", role).attr("style:master-page-name", masterName)
			styles.ensure("office:styles").children = append(styles.ensure("office:styles").children, pageStyle)
			n.set("text:style-name", name)
			assigned = true
		}
		for _, entry := range p.Contents {
			if entry.Style == role {
				p.odtContents(n, entry, styles)
			}
		}
		if placement, ok := p.Placements[role]; ok {
			p.odtPlacement(n, role, placement, styles)
		}
		if exists || change {
			n.children = append(p.odtBandAssignments(state), n.children...)
		}
	}
	odtVariableDeclarations(flow)
	p.odtTitleFloats(flow)
	p.odtPlacementGroups(flow)
	p.odtBlockSpacing(flow, styles)
	files["content.xml"], err = docxEncode(*root, DataLimit)
	if err != nil {
		return err
	}
	files["styles.xml"], err = docxEncode(*styles, DataLimit)
	return err
}

// Top and bottom title floats occupy successive vertical space, even when
// their text aligns to opposite edges. Independent full-width frames would
// obscure one another. Keep paragraph alignment inside one expanding frame.
func (p *Prose) odtTitleFloats(flow *docxNode) {
	var output []docxNode
	groups := map[string]int{}
	for _, node := range flow.children {
		if strings.HasPrefix(odtRole(&node), "FolioPageParagraph") {
			groups = map[string]int{}
		}
		frame := node.child("draw:frame")
		if frame != nil {
			role := strings.TrimPrefix(frame.get("draw:name"), "FolioFrame")
			placement := p.Placements[role]
			if (placement.Group == "title" || placement.Group == "footer") && (placement.Vertical == "top" || placement.Vertical == "bottom") {
				if index, exists := groups[placement.Vertical]; exists {
					box := output[index].child("draw:frame").child("draw:text-box")
					box.children = append(box.children, frame.child("draw:text-box").children...)
					continue
				}
				groups[placement.Vertical] = len(output)
			}
		}
		output = append(output, node)
	}
	flow.children = output
}
func (p *Prose) odtBand(text string, band Band, kind, variant string, state Boundary, styles *docxNode) docxNode {
	style := band.Font
	style.Align = band.Align
	if pair := strings.Split(band.Align, "-"); len(pair) == 2 {
		style.Align = pair[1]
		if variant == "-left" {
			style.Align = pair[0]
		}
	}
	name := fmt.Sprintf("FolioBand%d", len(styles.ensure("office:styles").children))
	styles.ensure("office:styles").children = append(styles.ensure("office:styles").children, ox("style:style", style.odtParagraph(), style.odtFont()).attr("style:name", name).attr("style:family", "paragraph"))
	n := ox("text:p").attr("text:style-name", name)
	for index, token := range p.odtBandTokens(text, state) {
		if token.field == "page" || token.field == "total-pages" {
			kind := "text:page-number"
			if token.field == "total-pages" {
				kind = "text:page-count"
			}
			field := ox(kind).attr("style:num-format", state.NumberFormat)
			field.text = "1"
			n.children = append(n.children, field)
			continue
		}
		if state.Blank || state.Title && state.TitleNumber && kind == "footer" {
			if token.text != "" {
				n.children = append(n.children, docxNode{text: token.text})
			}
			continue
		}
		field := ox("text:variable-get").attr("text:name", odtBandVariable(kind, variant, index)).attr("office:value-type", "string")
		field.text = token.text
		n.children = append(n.children, field)
	}
	return n
}
func (p *Prose) odtContents(n *docxNode, entry ContentsEntry, styles *docxNode) {
	if entry.Leaders {
		style := styles.ensure("office:styles")
		for i := range style.children {
			if style.children[i].get("style:name") == entry.Style {
				style.children[i].ensure("style:paragraph-properties").children = append(style.children[i].ensure("style:paragraph-properties").children, ox("style:tab-stops", ox("style:tab-stop").attr("style:type", "right").attr("style:position", point(p.Page.Width-p.Page.Left-p.Page.Right-p.Page.Gutter)).attr("style:leader-style", "dotted").attr("style:leader-text", ".")))
			}
		}
	}
	if entry.Links {
		n.children = []docxNode{ox("text:a", n.children...).attr("xlink:href", "#"+entry.Anchor).attr("xlink:type", "simple")}
	}
	if entry.Pages {
		field := ox("text:bookmark-ref").attr("text:reference-format", "page").attr("text:ref-name", entry.Anchor)
		field.text = "1"
		n.children = append(n.children, ox("text:tab"), field)
	}
}
func (p *Prose) odtPlacement(n *docxNode, role string, placement Placement, styles *docxNode) {
	vertical := placement.Vertical
	if vertical == "center" {
		vertical = "middle"
	}
	name := "FolioFrame" + role
	graphic := ox("style:style", ox("style:graphic-properties").attr("style:vertical-pos", vertical).attr("style:vertical-rel", "page-content").attr("style:horizontal-pos", placement.Horizontal).attr("style:horizontal-rel", "page-content").attr("style:wrap", "none").attr("draw:auto-grow-height", "true")).attr("style:name", name).attr("style:family", "graphic")
	styles.ensure("office:styles").children = append(styles.ensure("office:styles").children, graphic)
	inner := *n
	inner.attrs = append([]xml.Attr(nil), n.attrs...)
	inner.set("text:style-name", role)
	box := ox("draw:text-box", inner).attr("fo:min-height", point(p.Styles[role].Line))
	frame := ox("draw:frame", box).attr("draw:style-name", name).attr("draw:name", name).attr("text:anchor-type", "page").attr("svg:width", point(p.Page.Width-p.Page.Left-p.Page.Right-p.Page.Gutter))
	n.name = "text:p"
	n.text = ""
	n.children = []docxNode{frame}
}

// Consecutive paragraphs in one positioned group share an expanding text box.
// Separate frames at the same anchor would overlap instead of stacking.
func (p *Prose) odtPlacementGroups(flow *docxNode) {
	var output []docxNode
	var previous Placement
	groupIndex := -1
	for _, node := range flow.children {
		if node.name == "" && strings.TrimSpace(node.text) == "" {
			continue
		}
		frame := node.child("draw:frame")
		var placement Placement
		if frame != nil {
			role := strings.TrimPrefix(frame.get("draw:name"), "FolioFrame")
			placement = p.Placements[role]
		}
		if frame != nil && placement.Group != "" && placement == previous && groupIndex >= 0 {
			box := output[groupIndex].child("draw:frame").child("draw:text-box")
			box.children = append(box.children, frame.child("draw:text-box").children...)
			continue
		}
		output = append(output, node)
		groupIndex = len(output) - 1
		previous = placement
	}
	flow.children = output
}
