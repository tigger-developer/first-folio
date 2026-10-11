// ABOUTME: Builds native DOCX title paragraphs, running matter and page properties.
// ABOUTME: Uses a single section with documented title and frontmatter compromises.
package app

import "strings"

func (w *docxWriter) title() ([]docxNode, error) {
	if w.doc.Metadata["title"] == "" && w.cfg.Bool("folio.title-page.enabled", true) {
		return nil, nil
	}
	var out []docxNode
	for _, role := range []string{"Title", "Subtitle", "Author", "Date", "Version"} {
		text := w.doc.Metadata[strings.ToLower(role)]
		if text == "" {
			continue
		}
		if role == "Author" {
			text = w.cfg.String("folio.title-page.author.prefix", "") + text
		}
		p, err := w.literalParagraph(role, text)
		if err != nil {
			return nil, err
		}
		if role == "Title" && w.cfg.Bool("folio.title-page.enabled", true) {
			ratio := 0.3
			if w.cfg.String("folio.title-page.title.position", "third") != "third" {
				ratio = 0.4
			}
			p.children[0].children = append(p.children[0].children, dx("spacing").attr("w:before", dxInt(int(float64(w.height-w.top-w.bottom)*ratio))))
		}
		out = append(out, p)
	}
	return out, nil
}

type docxPart struct {
	name, contentType, relationship string
	root                            docxNode
}

func (w *docxWriter) runningParts() ([]docxPart, []docxNode, error) {
	var parts []docxPart
	var references []docxNode
	suppressFirst := !w.cfg.Bool("folio.title-page.enabled", true) || w.doc.Metadata["title"] != ""
	for _, role := range []string{"Header", "Footer"} {
		lower := strings.ToLower(role)
		path := "folio.page-" + lower
		if !w.cfg.Bool(path+".enabled", role == "Footer") {
			continue
		}
		base := w.cfg.String(path+".format", "[page]")
		alternate := w.cfg.String(path+".alt-format", "")
		if alternate == "" {
			alternate = base
		}
		align := strings.Split(w.cfg.String(path+".align", "center"), "-")
		for i, variant := range []string{"default", "even", "first"} {
			text := alternate
			alignment := align[len(align)-1]
			if variant == "even" {
				text = base
				alignment = align[0]
			}
			if variant == "first" && suppressFirst && w.cfg.Bool("folio.title-page.skip-"+lower, true) {
				text = ""
			}
			style, err := w.style(role)
			if err != nil {
				return nil, nil, err
			}
			p := dx("p", dx("pPr", dxVal("pStyle", style), dxVal("jc", alignment)))
			p.children = append(p.children, w.runningText(text)...)
			rootName := "hdr"
			if role == "Footer" {
				rootName = "ftr"
			}
			name := lower + dxInt(i+1) + ".xml"
			parts = append(parts, docxPart{name: "word/" + name, contentType: lower, relationship: lower, root: dxRoot(rootName, p)})
			references = append(references, dx(lower+"Reference").attr("w:type", variant).attr("r:id", lower+dxInt(i+1)))
		}
		if _, exists := w.cfg.Get(path + ".frontmatter-format"); exists {
			w.warnings = append(w.warnings, path+".frontmatter-format and alt-frontmatter-format use body running matter in this DOCX spike")
		}
	}
	return parts, references, nil
}

func (w *docxWriter) runningText(format string) []docxNode {
	var out []docxNode
	previous := 0
	for _, m := range scriptRunningTokenRE.FindAllStringSubmatchIndex(format, -1) {
		if m[0] > previous {
			out = append(out, docxRun(format[previous:m[0]], dx("rPr")))
		}
		token := format[m[2]:m[3]]
		switch token {
		case "title", "author":
			out = append(out, docxRun(w.doc.Metadata[token], dx("rPr")))
		case "page", "total-pages":
			instruction := "PAGE"
			if token == "total-pages" {
				instruction = "NUMPAGES"
			}
			field := dx("instrText").attr("xml:space", "preserve")
			field.text = " " + instruction + " "
			out = append(out, dx("r", dx("fldChar").attr("w:fldCharType", "begin")), dx("r", field), dx("r", dx("fldChar").attr("w:fldCharType", "separate")), docxRun("1", dx("rPr")), dx("r", dx("fldChar").attr("w:fldCharType", "end")))
		case "part", "part-number", "part-prefix", "part-full", "chapter", "chapter-number", "chapter-prefix", "chapter-full":
		default:
			out = append(out, docxRun(format[m[0]:m[1]], dx("rPr")))
		}
		previous = m[1]
	}
	if previous < len(format) {
		out = append(out, docxRun(format[previous:], dx("rPr")))
	}
	return out
}

func (w *docxWriter) section(references []docxNode) docxNode {
	section := dx("sectPr", references...)
	section.children = append(section.children,
		dx("pgSz").attr("w:w", dxInt(w.width)).attr("w:h", dxInt(w.height)),
		dx("pgMar").attr("w:top", dxInt(w.top)).attr("w:right", dxInt(w.margin)).attr("w:bottom", dxInt(w.bottom)).attr("w:left", dxInt(w.margin)).attr("w:header", dxInt(w.headerDistance)).attr("w:footer", dxInt(w.footerDistance)).attr("w:gutter", "0"),
		dx("titlePg"))
	return section
}

func (w *docxWriter) parts() ([]docxPart, error) {
	body, err := w.body()
	if err != nil {
		return nil, err
	}
	running, refs, err := w.runningParts()
	if err != nil {
		return nil, err
	}
	body = append(body, w.section(refs))
	parts := []docxPart{
		{name: "word/document.xml", contentType: "document.main", root: dxRoot("document", dx("body", body...))},
		{name: "word/styles.xml", contentType: "styles", relationship: "styles", root: dxRoot("styles", w.styles...)},
		{name: "word/settings.xml", contentType: "settings", relationship: "settings", root: dxRoot("settings", dx("evenAndOddHeaders"), dxVal("updateFields", "true"))},
	}
	if len(w.footnotes) > 0 {
		notes := []docxNode{
			dx("footnote", dx("p", dx("r", dx("separator")))).attr("w:type", "separator").attr("w:id", "-1"),
			dx("footnote", dx("p", dx("r", dx("continuationSeparator")))).attr("w:type", "continuationSeparator").attr("w:id", "0"),
		}
		notes = append(notes, w.footnotes...)
		parts = append(parts, docxPart{name: "word/footnotes.xml", contentType: "footnotes", relationship: "footnotes", root: dxRoot("footnotes", notes...)})
	}
	return append(parts, running...), nil
}
