// ABOUTME: Builds a local ODF reference with resolved Folio typography and page masters.
// ABOUTME: Supplies Pandoc custom styles without accepting arbitrary templates or markup.
package app

import (
	"archive/zip"
	"strconv"
	"strings"
)

func ox(name string, children ...docxNode) docxNode { return docxNode{name: name, children: children} }
func odtRoot(name string, children ...docxNode) docxNode {
	n := ox(name, children...).attr("office:version", "1.3")
	for _, ns := range []struct{ prefix, uri string }{
		{"office", "urn:oasis:names:tc:opendocument:xmlns:office:1.0"},
		{"style", "urn:oasis:names:tc:opendocument:xmlns:style:1.0"},
		{"text", "urn:oasis:names:tc:opendocument:xmlns:text:1.0"},
		{"table", "urn:oasis:names:tc:opendocument:xmlns:table:1.0"},
		{"fo", "urn:oasis:names:tc:opendocument:xmlns:xsl-fo-compatible:1.0"},
		{"svg", "urn:oasis:names:tc:opendocument:xmlns:svg-compatible:1.0"},
	} {
		n = n.attr("xmlns:"+ns.prefix, ns.uri)
	}
	return n
}
func odtPoints(twips int) string { return strconv.FormatFloat(float64(twips)/20, 'f', -1, 64) + "pt" }
func odtParagraphProperties(props docxNode) docxNode {
	n := ox("style:paragraph-properties")
	for _, c := range props.children {
		switch c.name {
		case "w:spacing":
			for _, pair := range []struct{ w, o string }{{"before", "fo:margin-top"}, {"after", "fo:margin-bottom"}} {
				if odtValue(c, pair.w) != "" {
					n = n.attr(pair.o, odtPoints(odtNumber(c, pair.w)))
				}
			}
			if odtValue(c, "line") != "" {
				value := strconv.FormatFloat(float64(odtNumber(c, "line"))/240*100, 'f', -1, 64) + "%"
				if odtValue(c, "lineRule") == "exact" {
					value = odtPoints(odtNumber(c, "line"))
				}
				n = n.attr("fo:line-height", value)
			}
		case "w:ind":
			n = n.attr("fo:margin-left", odtPoints(odtNumber(c, "left"))).attr("fo:margin-right", odtPoints(odtNumber(c, "right")))
		case "w:jc":
			n = n.attr("fo:text-align", odtValue(c, "val"))
		case "w:keepNext":
			n = n.attr("fo:keep-with-next", "always")
		case "w:pageBreakBefore":
			n = n.attr("fo:break-before", "page")
		}
	}
	return n
}
func odtTextProperties(props docxNode) docxNode {
	n := ox("style:text-properties")
	for _, c := range props.children {
		v := odtValue(c, "val")
		switch c.name {
		case "w:rFonts":
			n = n.attr("fo:font-family", odtValue(c, "ascii")).attr("style:font-name", odtValue(c, "ascii"))
		case "w:sz":
			n = n.attr("fo:font-size", odtPoints(odtNumber(c, "val")*10))
		case "w:b":
			value := "bold"
			if v == "false" {
				value = "normal"
			}
			n = n.attr("fo:font-weight", value)
		case "w:i":
			value := "italic"
			if v == "false" {
				value = "normal"
			}
			n = n.attr("fo:font-style", value)
		case "w:spacing":
			n = n.attr("fo:letter-spacing", odtPoints(odtNumber(c, "val")))
		case "w:w":
			n = n.attr("style:text-scale", v+"%")
		case "w:smallCaps":
			n = n.attr("fo:font-variant", "small-caps")
		}
	}
	return n
}

func (p *odtProjection) reference() ([]byte, error) {
	w := p.layout
	for _, role := range []string{"Body", "Footnote", "Header", "Footer"} {
		if (role == "Header" || role == "Footer") && !w.cfg.Bool("folio.page-"+strings.ToLower(role)+".enabled", role == "Footer") {
			continue
		}
		if _, err := w.style(role); err != nil {
			return nil, err
		}
	}
	bodyFont, _, err := w.font("folio.font")
	if err != nil {
		return nil, err
	}
	styles := ox("office:styles", ox("style:default-style", ox("style:paragraph-properties").attr("fo:margin-top", "0pt").attr("fo:margin-bottom", "0pt"), odtTextProperties(bodyFont)).attr("style:family", "paragraph"), ox("style:style").attr("style:name", "Standard").attr("style:family", "paragraph"))
	faces := ox("office:font-face-decls")
	seen := map[string]bool{}
	for _, s := range w.styles {
		font := odtChild(s, "rPr")
		family := odtValue(odtChild(font, "rFonts"), "ascii")
		if !seen[family] {
			faces.children = append(faces.children, ox("style:font-face").attr("style:name", family).attr("svg:font-family", family))
			seen[family] = true
		}
		styles.children = append(styles.children, ox("style:style", odtParagraphProperties(odtChild(s, "pPr")), odtTextProperties(font)).attr("style:name", odtValue(s, "styleId")).attr("style:display-name", odtValue(odtChild(s, "name"), "val")).attr("style:family", "paragraph").attr("style:parent-style-name", "Standard"))
	}
	styles.children = append(styles.children, p.styles...)
	styles.children = append(styles.children, ox("style:style", ox("style:text-properties").attr("fo:font-family", "Courier New")).attr("style:name", "FolioCode").attr("style:family", "text"))
	for _, pair := range []struct{ name, parent string }{{"Text_20_body", "FolioBody"}, {"Footnote", "FolioFootnote"}} {
		styles.children = append(styles.children, ox("style:style").attr("style:name", pair.name).attr("style:family", "paragraph").attr("style:parent-style-name", pair.parent))
	}
	styles.children = append(styles.children,
		ox("style:style").attr("style:name", "Footnote_20_Symbol").attr("style:family", "text"),
		ox("style:style", ox("style:text-properties").attr("style:text-position", "super 58%")).attr("style:name", "Footnote_20_anchor").attr("style:family", "text"))
	styles.children = append(styles.children, ox("text:notes-configuration").attr("text:citation-style-name", "Footnote_20_Symbol").attr("text:citation-body-style-name", "Footnote_20_anchor").attr("text:footnotes-position", "page").attr("text:start-numbering-at", "document").attr("text:note-class", "footnote").attr("style:num-format", "1").attr("text:start-value", "0"))
	auto, masters, err := p.pageMasters(&styles)
	if err != nil {
		return nil, err
	}
	root := odtRoot("office:document-styles", faces, styles, auto, masters)
	return odtReferencePackage(root)
}

func odtReferencePackage(root docxNode) ([]byte, error) {
	raw, err := docxEncode(root, docxXMLLimit)
	if err != nil {
		return nil, err
	}
	content, err := docxEncode(odtRoot("office:document-content", ox("office:body", ox("office:text", ox("text:p")))), docxXMLLimit)
	if err != nil {
		return nil, err
	}
	metadata, err := docxEncode(odtRoot("office:document-meta", ox("office:meta")), docxXMLLimit)
	if err != nil {
		return nil, err
	}
	manifest := ox("manifest:manifest", ox("manifest:file-entry").attr("manifest:full-path", "meta.xml").attr("manifest:media-type", "text/xml"), ox("manifest:file-entry").attr("manifest:full-path", "/").attr("manifest:media-type", odtMIME), ox("manifest:file-entry").attr("manifest:full-path", "styles.xml").attr("manifest:media-type", "text/xml"), ox("manifest:file-entry").attr("manifest:full-path", "content.xml").attr("manifest:media-type", "text/xml")).attr("xmlns:manifest", "urn:oasis:names:tc:opendocument:xmlns:manifest:1.0").attr("manifest:version", "1.3")
	manifestRaw, err := docxEncode(manifest, docxXMLLimit)
	if err != nil {
		return nil, err
	}
	out := odtBuffer{limit: docxXMLLimit}
	z := zip.NewWriter(&out)
	for _, part := range []struct {
		name string
		data []byte
	}{{"mimetype", []byte(odtMIME)}, {"styles.xml", raw}, {"meta.xml", metadata}, {"content.xml", content}, {"META-INF/manifest.xml", manifestRaw}} {
		f, err := z.CreateHeader(&zip.FileHeader{Name: part.name, Method: zip.Store})
		if err != nil {
			return nil, err
		}
		if _, err = f.Write(part.data); err != nil {
			return nil, err
		}
	}
	if err = z.Close(); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
