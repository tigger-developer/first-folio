// ABOUTME: Generates ODT page masters, running matter and live page fields.
// ABOUTME: Maintains first and facing page variants without resetting numbering.
package office

import "strings"

func (p *odtProjection) pageMasters(styles *docxNode) (docxNode, docxNode, error) {
	w := p.layout
	auto := ox("office:automatic-styles")
	masters := ox("office:master-styles")
	for _, variant := range []string{"First", "Even", "Odd"} {
		page := ox("style:page-layout-properties").attr("fo:page-width", odtPoints(w.width)).attr("fo:page-height", odtPoints(w.height)).attr("fo:margin-left", odtPoints(w.margin)).attr("fo:margin-right", odtPoints(w.margin)).attr("fo:margin-top", odtPoints(w.top)).attr("fo:margin-bottom", odtPoints(w.bottom)).attr("style:print-orientation", "portrait")
		layout := ox("style:page-layout", page).attr("style:name", "FolioLayout"+variant)
		next := "Even"
		if variant == "Even" {
			next = "Odd"
		}
		master := ox("style:master-page").attr("style:name", "Folio"+variant).attr("style:page-layout-name", "FolioLayout"+variant).attr("style:next-style-name", "Folio"+next)
		for _, role := range []string{"header", "footer"} {
			path := "folio.page-" + role
			if !w.cfg.Bool(path+".enabled", role == "footer") {
				continue
			}
			suppress := variant == "First" && (!w.cfg.Bool("folio.title-page.enabled", true) || w.doc.Metadata["title"] != "") && w.cfg.Bool("folio.title-page.skip-"+role, true)
			distance, reserved := w.headerDistance, w.top
			edge := "top"
			if role == "footer" {
				distance, reserved = w.footerDistance, w.bottom
				edge = "bottom"
			}
			// ODF places running matter within the page's printable area. Reserve its
			// font height and clearance so the body retains the office layout boundary.
			font, err := w.cfg.Font(path + ".font")
			if err != nil {
				return docxNode{}, docxNode{}, err
			}
			size, err := docxPoints(font.Size, w.baseSize, 0)
			if err != nil {
				return docxNode{}, docxNode{}, err
			}
			height := int(size * 20)
			gap := reserved - distance - height
			if gap < 0 {
				gap = 0
				w.warnings = append(w.warnings, path+" clearance is smaller than the running font height")
			}
			if !suppress {
				for i, a := range layout.children[0].attrs {
					if a.Name.Local == "fo:margin-"+edge {
						layout.children[0].attrs[i].Value = odtPoints(distance)
					}
				}
				gapEdge := "bottom"
				if role == "footer" {
					gapEdge = "top"
				}
				layout.children = append(layout.children, ox("style:"+role+"-style", ox("style:header-footer-properties").attr("fo:min-height", odtPoints(height)).attr("fo:margin-"+gapEdge, odtPoints(gap)).attr("style:dynamic-spacing", "false")))
				format := w.cfg.String(path+".format", "[page]")
				alt := w.cfg.String(path+".alt-format", "")
				align := strings.Split(w.cfg.String(path+".align", "center"), "-")
				alignment := align[0]
				if variant != "Even" {
					if alt != "" {
						format = alt
					}
					alignment = align[len(align)-1]
				}
				style := "Folio" + strings.ToUpper(role[:1]) + role[1:] + variant
				styles.children = append(styles.children, ox("style:style", ox("style:paragraph-properties").attr("fo:text-align", alignment)).attr("style:name", style).attr("style:family", "paragraph").attr("style:parent-style-name", "Folio"+strings.ToUpper(role[:1])+role[1:]))
				master.children = append(master.children, ox("style:"+role, ox("text:p", p.runningText(format)...).attr("text:style-name", style)))
			}
		}
		auto.children = append(auto.children, layout)
		masters.children = append(masters.children, master)
	}
	for _, role := range []string{"header", "footer"} {
		path := "folio.page-" + role
		if w.cfg.Bool(path+".enabled", role == "footer") {
			if _, exists := w.cfg.Get(path + ".frontmatter-format"); exists {
				w.warnings = append(w.warnings, path+" frontmatter formats use body running matter")
			}
		}
	}

	return auto, masters, nil
}

func (p *odtProjection) runningText(format string) []docxNode {
	out := []docxNode{}
	pos := 0
	text := func(s string) { n := ox("text:span"); n.text = s; out = append(out, n) }
	for _, m := range scriptRunningTokenRE.FindAllStringSubmatchIndex(format, -1) {
		text(format[pos:m[0]])
		token := format[m[2]:m[3]]
		switch token {
		case "title", "author":
			text(p.layout.doc.Metadata[token])
		case "page":
			out = append(out, ox("text:page-number").attr("text:select-page", "current"))
		case "total-pages":
			out = append(out, ox("text:page-count"))
		case "part", "part-number", "part-prefix", "part-full", "chapter", "chapter-number", "chapter-prefix", "chapter-full":
		default:
			text(format[m[0]:m[1]])
		}
		pos = m[1]
	}
	text(format[pos:])
	return out
}
