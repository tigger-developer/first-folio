// ABOUTME: Keeps native list labels aligned with the resolved prose baseline.
// ABOUTME: Preserves editable numbering and bullet glyph families.
package office

import "strconv"

func (p *Prose) alignListMarkers(files map[string][]byte, format string) error {
	style, ok := p.Styles["FolioBody"]
	if !ok || style.BaselineShift == 0 {
		return nil
	}
	paths := []string{"word/numbering.xml"}
	if format == "odt" {
		paths = []string{"styles.xml", "content.xml"}
	}
	for _, path := range paths {
		if len(files[path]) == 0 {
			continue
		}
		root, err := parseXML(files[path])
		if err != nil {
			return err
		}
		root.visit(func(n *docxNode) {
			if format == "docx" && n.name == "w:lvl" {
				pr := n.ensure("w:rPr")
				pr.remove("w:position")
				pr.ensure("w:sz").set("w:val", strconv.Itoa(int(style.Size*2+0.5)))
				if number := n.child("w:numFmt"); number != nil && number.get("w:val") != "bullet" {
					pr.ensure("w:rFonts").set("w:ascii", style.Font.Family)
					pr.ensure("w:rFonts").set("w:hAnsi", style.Font.Family)
				}
			}
			if format == "odt" && (n.name == "text:list-level-style-bullet" || n.name == "text:list-level-style-number") {
				pr := n.ensure("style:text-properties")
				font := style.odtFont()
				pr.set("style:text-position", font.get("style:text-position"))
				pr.set("fo:font-size", point(style.Size))
				if n.name == "text:list-level-style-number" {
					pr.set("style:font-name", style.Font.Family)
				}
			}
		})
		files[path], err = docxEncode(*root, DataLimit)
		if err != nil {
			return err
		}
	}
	return nil
}
