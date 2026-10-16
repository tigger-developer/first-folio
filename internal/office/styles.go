// ABOUTME: Serializes resolved prose typography into named DOCX and ODT styles.
// ABOUTME: Keeps point rounding at the native format boundary and role values independent.
package office

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

func point(value float64) string {
	return strconv.FormatFloat(math.Round(value*1e6)/1e6, 'f', -1, 64) + "pt"
}
func twip(value float64) string { return strconv.Itoa(int(math.Round(value * 20))) }
func boolean(value bool) string { return strconv.FormatBool(value) }
func (s Style) wordFont() docxNode {
	stretch := strings.TrimSuffix(s.Font.Stretch, "%")
	if stretch == "" {
		stretch = "100"
	}
	n := dx("rPr", dx("rFonts").attr("w:ascii", s.Font.Family).attr("w:hAnsi", s.Font.Family).attr("w:eastAsia", s.Font.Family).attr("w:cs", s.Font.Family), dxVal("b", boolean(s.Bold)), dxVal("i", boolean(s.Italic)), dxVal("spacing", twip(s.Tracking)), dxVal("w", stretch))
	if s.BaselineShift != 0 {
		n.children = append(n.children, dxVal("position", s.wordPosition()))
	}
	n.children = append(n.children, dxVal("sz", strconv.Itoa(int(math.Round(s.Size*2)))), dxVal("szCs", strconv.Itoa(int(math.Round(s.Size*2)))))
	return n
}

func (s Style) wordPosition() string {
	shift := s.BaselineShift
	if s.FontBox > 0 {
		shift *= s.Size / s.FontBox
	}
	return strconv.Itoa(int(math.Round(shift * 2)))
}
func (s Style) wordParagraph() docxNode {
	align := s.Align
	if align == "justify" {
		align = "both"
	}
	if align == "" {
		align = "left"
	}
	spacing := dx("spacing").attr("w:before", twip(s.Before)).attr("w:after", twip(s.After))
	if s.Line > 0 {
		spacing = spacing.attr("w:line", twip(s.Line)).attr("w:lineRule", "exact")
	}
	p := dx("pPr", spacing, dx("ind").attr("w:left", twip(s.Left)).attr("w:right", twip(s.Right)).attr("w:firstLine", twip(s.First)), dxVal("jc", align), dxVal("suppressAutoHyphens", boolean(!s.Hyphenate)), dxVal("widowControl", boolean(s.Widow)))
	if s.KeepNext {
		p.children = append(p.children, dx("keepNext"))
	}
	return p
}
func (s Style) odtFont() docxNode {
	weight := s.Font.Weight
	if weight == "regular" {
		weight = "normal"
	}
	if s.Bold {
		weight = "bold"
	}
	italic := "normal"
	if s.Italic {
		italic = "italic"
	}
	stretch := s.Font.Stretch
	if stretch == "" {
		stretch = "100%"
	}
	if !strings.HasSuffix(stretch, "%") {
		stretch += "%"
	}
	n := ox("style:text-properties").attr("style:font-name", s.Font.Family).attr("fo:font-family", s.Font.Family).attr("fo:font-size", point(s.Size)).attr("fo:font-weight", weight).attr("fo:font-style", italic).attr("style:text-scale", stretch).attr("fo:letter-spacing", point(s.Tracking)).attr("fo:hyphenate", boolean(s.Hyphenate))
	if s.BaselineShift != 0 && s.FontBox > 0 {
		n = n.attr("style:text-position", strconv.FormatFloat(s.BaselineShift/s.FontBox*100, 'f', 5, 64)+"% 100%")
	}
	return n
}
func (s Style) odtParagraph() docxNode {
	align := s.Align
	if align == "" || align == "left" {
		align = "start"
	}
	widows := "0"
	if s.Widow {
		widows = "2"
	}
	p := ox("style:paragraph-properties").attr("fo:margin-top", point(s.Before)).attr("fo:margin-bottom", point(s.After)).attr("fo:margin-left", point(s.Left)).attr("fo:margin-right", point(s.Right)).attr("fo:text-indent", point(s.First)).attr("fo:text-align", align).attr("fo:widows", widows).attr("fo:orphans", widows)
	if s.Line > 0 {
		p = p.attr("fo:line-height", point(s.Line))
	}
	if s.KeepNext {
		p = p.attr("fo:keep-with-next", "always")
	}
	return p
}

func (p *Prose) configureStyles(files map[string][]byte, format string) error {
	path := "word/styles.xml"
	if format == "odt" {
		path = "styles.xml"
	}
	root, err := parseXML(files[path])
	if err != nil {
		return err
	}
	styles := root
	if format == "odt" {
		styles = root.ensure("office:styles")
		root.set("xmlns:fo", "urn:oasis:names:tc:opendocument:xmlns:xsl-fo-compatible:1.0")
		root.set("xmlns:svg", "urn:oasis:names:tc:opendocument:xmlns:svg-compatible:1.0")
	}
	names := make([]string, 0, len(p.Styles))
	for name := range p.Styles {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		s := p.Styles[name]
		var n docxNode
		if format == "docx" {
			kind := "paragraph"
			if s.Character {
				kind = "character"
			}
			n = dx("style", dxVal("name", name)).attr("w:type", kind).attr("w:styleId", name)
			if !s.Character {
				n.children = append(n.children, s.wordParagraph())
			}
			n.children = append(n.children, s.wordFont())
			if !s.Character {
				n.child("w:rPr").remove("w:position")
			}
		} else {
			kind := "paragraph"
			if s.Character {
				kind = "text"
			}
			n = ox("style:style").attr("style:name", name).attr("style:display-name", name).attr("style:family", kind)
			if !s.Character {
				n.children = append(n.children, s.odtParagraph())
			}
			n.children = append(n.children, s.odtFont())
		}
		key := "w:styleId"
		if format == "odt" {
			key = "style:name"
		}
		found := false
		for i := range styles.children {
			if styles.children[i].get(key) == name {
				styles.children[i] = n
				found = true
				break
			}
		}
		if !found {
			styles.children = append(styles.children, n)
		}
	}
	if format == "odt" {
		faces := root.ensure("office:font-face-decls")
		seen := map[string]bool{}
		for _, face := range faces.children {
			seen[face.get("style:name")] = true
		}
		for _, s := range p.Styles {
			if seen[s.Font.Family] {
				continue
			}
			seen[s.Font.Family] = true
			faces.children = append(faces.children, ox("style:font-face").attr("style:name", s.Font.Family).attr("svg:font-family", s.Font.Family))
		}
	}
	files[path], err = docxEncode(*root, DataLimit)
	return err
}
func (p Page) wordSection() docxNode {
	return dx("sectPr", dx("pgSz").attr("w:w", twip(p.Width)).attr("w:h", twip(p.Height)), dx("pgMar").attr("w:top", twip(p.Top)).attr("w:bottom", twip(p.Bottom)).attr("w:left", twip(p.Left)).attr("w:right", twip(p.Right)).attr("w:gutter", twip(p.Gutter)).attr("w:header", twip(p.Header.Edge)).attr("w:footer", twip(p.Footer.Edge)))
}
func (p Page) odtLayout(name string) docxNode {
	usage := "all"
	if p.Gutter != 0 {
		usage = "mirrored"
	}
	props := ox("style:page-layout-properties").attr("fo:page-width", point(p.Width)).attr("fo:page-height", point(p.Height)).attr("fo:margin-top", point(p.Top)).attr("fo:margin-bottom", point(p.Bottom)).attr("fo:margin-left", point(p.Left+p.Gutter)).attr("fo:margin-right", point(p.Right))
	return ox("style:page-layout", props).attr("style:name", name).attr("style:page-usage", usage)
}
func (p *Prose) configurePage(files map[string][]byte, format string) error {
	if p.Page.Width <= 0 || p.Page.Height <= 0 {
		return fmt.Errorf("invalid office paper dimensions")
	}
	if format == "docx" {
		root, err := parseXML(files["word/document.xml"])
		if err != nil {
			return err
		}
		body := root.child("w:body")
		if body == nil {
			return fmt.Errorf("DOCX has no body")
		}
		body.remove("w:sectPr")
		body.children = append(body.children, p.Page.wordSection())
		files["word/document.xml"], err = docxEncode(*root, DataLimit)
		if err != nil {
			return err
		}
		settings, err := parseXML(files["word/settings.xml"])
		if err != nil {
			return err
		}
		if p.Page.Gutter != 0 {
			settings.ensure("w:mirrorMargins")
		}
		settings.ensure("w:updateFields").set("w:val", "true")
		files["word/settings.xml"], err = docxEncode(*settings, DataLimit)
		return err
	}
	root, err := parseXML(files["styles.xml"])
	if err != nil {
		return err
	}
	auto := root.ensure("office:automatic-styles")
	auto.children = append(auto.children, p.Page.odtLayout("FolioPage"))
	masters := root.ensure("office:master-styles")
	masters.children = []docxNode{ox("style:master-page").attr("style:name", "Standard").attr("style:page-layout-name", "FolioPage")}
	files["styles.xml"], err = docxEncode(*root, DataLimit)
	return err
}
