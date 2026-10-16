// ABOUTME: Checks resolved prose typography against independently specified native properties.
// ABOUTME: Keeps paragraph, role font, and page geometry assertions separate from text presence.
package app

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

func TestRT063_4_OfficeManuscriptRoleFonts(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		for _, role := range []struct{ key, marker string }{
			{"font", "BODY-FONT-MARKER"},
			{"heading.font", "HEADING-FONT-MARKER"},
			{"quoted-block.font", "QUOTE-FONT-MARKER"},
			{"mono.font", "CODE-FONT-MARKER"},
		} {
			t.Run(format+"/"+role.key, func(t *testing.T) {
				t.Setenv("HOME", t.TempDir())
				dir := t.TempDir()
				src := proseSource(t, dir, "source.md", "## HEADING-FONT-MARKER\n\nBODY-FONT-MARKER\n\n> QUOTE-FONT-MARKER\n\n```\nCODE-FONT-MARKER\n```\n")
				parts := strings.Split(role.key, ".")
				cfg := "folio:\n  manuscript:\n    title-page:\n      enabled: false\n    toc:\n      enabled: false\n"
				indent := "    "
				for _, key := range parts {
					cfg += indent + key + ":\n"
					indent += "  "
				}
				for _, field := range []string{"family: Iosevka Publish", "size: 13pt", "weight: bold", "style: italic", "stretch: 90%", "letter-spacing: 1.25pt"} {
					cfg += indent + field + "\n"
				}
				proseSource(t, dir, "script.yaml", cfg)
				pkg := proseConvert(t, src, filepath.Join(dir, "output."+format))
				paragraph := officeParagraph(t, officeBody(t, pkg, format), role.marker)
				layers := officeEffective(t, pkg, format, paragraph)
				if format == "docx" {
					officeAssertEffective(t, layers, "rFonts", "ascii", "Iosevka Publish")
					officeAssertEffective(t, layers, "sz", "val", "26")
					officeAssertEffective(t, layers, "w", "val", "90")
					officeAssertEffective(t, layers, "spacing", "val", "25")
					for _, property := range []string{"b", "i"} {
						found := false
						for _, layer := range layers {
							for _, flag := range layer.all(property) {
								found = flag.attr("val") != "0" && flag.attr("val") != "false"
							}
						}
						if !found {
							t.Errorf("%s lost font flag %s", role.key, property)
						}
					}
				} else {
					officeAssertEffective(t, layers, "text-properties", "font-size", "13pt")
					officeAssertEffective(t, layers, "text-properties", "font-weight", "bold")
					officeAssertEffective(t, layers, "text-properties", "font-style", "italic")
					officeAssertEffective(t, layers, "text-properties", "text-scale", "90%")
					officeAssertEffective(t, layers, "text-properties", "letter-spacing", "1.25pt")
					name, exists := officeEffectiveAttr(layers, "text-properties", "font-name")
					matched := false
					for _, file := range []string{"styles.xml", "content.xml"} {
						for _, face := range pkg[file].all("font-face") {
							if face.attr("name") == name && strings.Trim(face.attr("font-family"), "'\"") == "Iosevka Publish" {
								matched = true
							}
						}
					}
					if !exists || !matched {
						t.Errorf("font reference %q does not resolve to configured family", name)
					}
				}
			})
		}
	}
}

func TestRT063_4_OfficeManuscriptParagraphGeometry(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		for _, indent := range []struct {
			value  string
			points float64
		}{{"10mm", 720 / 25.4}, {"2em", 24}} {
			t.Run(format+"/"+indent.value, func(t *testing.T) {
				t.Setenv("HOME", t.TempDir())
				dir := t.TempDir()
				cfg := fmt.Sprintf("folio:\n  manuscript:\n    font:\n      size: 12pt\n    title-page:\n      enabled: false\n    toc:\n      enabled: false\n    paragraph-indent: %s\n    quote-block-indent: 6mm\n    code-block-indent: 8mm\n", indent.value)
				proseSource(t, dir, "script.yaml", cfg)
				src := proseSource(t, dir, "source.md", "## First chapter\n\nOPENING-UNINDENTED\n\nBODY-INDENTED\n\n> QUOTE-INSET\n\n```\nCODE-INSET\n```\n\n## Second chapter\n\n```\nINTERVENING-CODE\n```\n\nAFTER-CODE-INDENTED\n")
				pkg := proseConvert(t, src, filepath.Join(dir, "output."+format))
				for _, tc := range []struct {
					marker string
					points float64
				}{{"OPENING-UNINDENTED", 0}, {"BODY-INDENTED", indent.points}, {"AFTER-CODE-INDENTED", indent.points}} {
					layers := officeEffective(t, pkg, format, officeParagraph(t, officeBody(t, pkg, format), tc.marker))
					officeAssertLength(t, layers, format, "ind", "firstLine", "paragraph-properties", "text-indent", tc.points)
				}
				for _, tc := range []struct {
					marker string
					points float64
				}{{"QUOTE-INSET", 6 * 72 / 25.4}, {"CODE-INSET", 8 * 72 / 25.4}} {
					layers := officeEffective(t, pkg, format, officeParagraph(t, officeBody(t, pkg, format), tc.marker))
					officeAssertLength(t, layers, format, "ind", "left", "paragraph-properties", "margin-left", tc.points)
				}
			})
		}
	}
}
