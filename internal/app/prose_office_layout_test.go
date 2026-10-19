// ABOUTME: Checks resolved prose typography against independently specified native properties.
// ABOUTME: Keeps paragraph, role font, and page geometry assertions separate from text presence.
package app

import (
	"fmt"
	"math"
	"path/filepath"
	"strconv"
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
			{"title-page.title.font", "TITLE-FONT-MARKER"},
			{"title-page.subtitle.font", "SUBTITLE-FONT-MARKER"},
			{"title-page.author.font", "AUTHOR-FONT-MARKER"},
			{"title-page.date.font", "DATE-FONT-MARKER"},
			{"title-page.wordcount.font", "WORDCOUNT-FONT-MARKER"},
			{"title-page.version.font", "VERSION-FONT-MARKER"},
			{"title-page.contact.font", "CONTACT-FONT-MARKER"},
			{"copyright.font", "COPYRIGHT-FONT-MARKER"},
			{"copyright.label.font", "LABEL-FONT-MARKER"},
			{"toc.font", "HEADING-FONT-MARKER"},
			{"toc.heading.font", "CONTENTS-FONT-MARKER"},
		} {
			t.Run(format+"/"+role.key, func(t *testing.T) {
				t.Setenv("HOME", t.TempDir())
				dir := t.TempDir()
				src := proseSource(t, dir, "source.md", "---\ntitle: TITLE-FONT-MARKER\nsubtitle: SUBTITLE-FONT-MARKER\nauthor: AUTHOR-FONT-MARKER\ndate: DATE-FONT-MARKER\nwordcount: WORDCOUNT-FONT-MARKER\nversion: VERSION-FONT-MARKER\ncontact-name: CONTACT-FONT-MARKER\n---\n\n## HEADING-FONT-MARKER\n\nBODY-FONT-MARKER\n\n> QUOTE-FONT-MARKER\n\n```\nCODE-FONT-MARKER\n```\n")
				parts := strings.Split(role.key, ".")
				cfg := "folio:\n  manuscript:\n"
				if !strings.HasPrefix(role.key, "title-page.") {
					cfg += "    title-page:\n      enabled: false\n"
				}
				if !strings.HasPrefix(role.key, "toc.") {
					cfg += "    toc:\n      enabled: false\n"
				}
				indent := "    "
				for _, key := range parts {
					cfg += indent + key + ":\n"
					indent += "  "
					if key == "copyright" {
						cfg += indent + "enabled: true\n" + indent + "credits: [COPYRIGHT-FONT-MARKER]\n" + indent + "publisher: LABEL-FONT-MARKER\n"
					}
					if key == "toc" {
						cfg += indent + "title: CONTENTS-FONT-MARKER\n"
					}
				}
				for _, field := range []string{"family: Iosevka Publish", "size: 13pt", "weight: bold", "style: italic", "stretch: 90%", "letter-spacing: 1.25pt"} {
					cfg += indent + field + "\n"
				}
				proseSource(t, dir, "script.yaml", cfg)
				pkg := proseConvert(t, src, filepath.Join(dir, "output."+format))
				paragraph := officeParagraph(t, officeBody(t, pkg, format), role.marker)
				layers := officeEffective(t, pkg, format, paragraph)
				if role.key == "copyright.label.font" {
					layers = proseInlineLayers(t, pkg, format, role.marker)
				}
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

func TestRT063_4_OfficeLetterTypographyAndGeometry(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		t.Run(format, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			dir := t.TempDir()
			src := proseSource(t, dir, "letters.org", "#+AUTHOR: Author\n* Letters :letter:\n** Subject :subject:\nLETTER-FONT-MARKER\n*** Recipient :to:\n**** Theatre :org:\n")
			proseSource(t, dir, "script.yaml", "folio:\n  letter:\n    page: a4\n    margin-top: 21mm\n    margin-bottom: 22mm\n    margin-left: 23mm\n    margin-right: 24mm\n    font:\n      family: Iosevka Publish\n      size: 13pt\n      weight: bold\n      style: italic\n      stretch: 90%\n      letter-spacing: 1.25pt\n")
			status, _, stderr := runApp(t, "letter", src, "--format", format)
			if status != 0 {
				t.Fatalf("letter export: %s", stderr)
			}
			pkg := prosePackage(t, filepath.Join(dir, "letter-theatre."+format), format)
			body := officeBody(t, pkg, format)
			layers := officeEffective(t, pkg, format, officeParagraph(t, body, "LETTER-FONT-MARKER"))
			if format == "docx" {
				officeAssertEffective(t, layers, "rFonts", "ascii", "Iosevka Publish")
				officeAssertEffective(t, layers, "sz", "val", "26")
				officeAssertEffective(t, layers, "w", "val", "90")
				officeAssertEffective(t, layers, "spacing", "val", "25")
				for _, property := range []string{"b", "i"} {
					active := false
					for _, layer := range layers {
						for _, flag := range layer.all(property) {
							active = flag.attr("val") != "0" && flag.attr("val") != "false"
						}
					}
					if !active {
						t.Errorf("letter font lost %s", property)
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
					t.Errorf("letter font reference %q does not resolve to configured family", name)
				}
			}
			pageLayers := []*officeXML{body}
			if format == "odt" {
				pageLayers = []*officeXML{pkg["styles.xml"]}
			}
			for _, side := range []struct {
				name string
				mm   float64
			}{{"top", 21}, {"bottom", 22}, {"left", 23}, {"right", 24}} {
				officeAssertLength(t, pageLayers, format, "pgMar", side.name, "page-layout-properties", "margin-"+side.name, side.mm*72/25.4)
			}
			officeAssertLength(t, pageLayers, format, "pgSz", "w", "page-layout-properties", "page-width", 210*72/25.4)
			officeAssertLength(t, pageLayers, format, "pgSz", "h", "page-layout-properties", "page-height", 297*72/25.4)
		})
	}
}

func TestRT063_4_OfficeLineIntervalUsesFontMetrics(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		for _, tc := range []struct {
			multiplier string
			points     float64
		}{{"1", 16.62}, {"1.5", 24.93}, {"2", 33.24}} {
			t.Run(format+"/"+tc.multiplier, func(t *testing.T) {
				t.Setenv("HOME", t.TempDir())
				dir := t.TempDir()
				proseSource(t, dir, "script.yaml", "folio:\n  manuscript:\n    font:\n      family: Iosevka Publish\n      size: 12pt\n    line-spacing: "+tc.multiplier+"\n    title-page:\n      enabled: false\n    toc:\n      enabled: false\n")
				src := proseSource(t, dir, "source.md", "LINE-INTERVAL-MARKER: ordinary prose using the bundled regular face.\n")
				pkg := proseConvert(t, src, filepath.Join(dir, "output."+format))
				layers := officeEffective(t, pkg, format, officeParagraph(t, officeBody(t, pkg, format), "LINE-INTERVAL-MARKER"))
				// Independent PDF baseline: 12pt Iosevka Publish has a 16.62pt
				// single interval under the existing metrics plus 0.65em rule.
				officeAssertLength(t, layers, format, "spacing", "line", "paragraph-properties", "line-height", tc.points)
				if format == "docx" {
					officeAssertEffective(t, layers, "spacing", "lineRule", "exact")
				}
			})
		}
	}
}

func TestRT063_3_OfficeFootnoteTypography(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		t.Run(format, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			dir := t.TempDir()
			proseSource(t, dir, "script.yaml", "folio:\n  manuscript:\n    font:\n      family: Iosevka Publish\n      size: 12pt\n    title-page:\n      enabled: false\n    toc:\n      enabled: false\n")
			src := proseSource(t, dir, "source.md", "NATIVE-NOTE-ANCHOR[^note].\n\n[^note]: NOTE-FONT-MARKER\n")
			pkg := proseConvert(t, src, filepath.Join(dir, "output."+format))
			root := officeBody(t, pkg, format)
			if format == "docx" {
				root = pkg["word/footnotes.xml"]
				if root == nil {
					t.Fatal("native footnote part missing")
				}
			}
			layers := officeEffective(t, pkg, format, officeParagraph(t, root, "NOTE-FONT-MARKER"))
			if format == "docx" {
				officeAssertEffective(t, layers, "rFonts", "ascii", "Iosevka Publish")
				officeAssertEffective(t, layers, "sz", "val", "20")
			} else {
				officeAssertEffective(t, layers, "text-properties", "font-size", "10.2pt")
			}
		})
	}
}

func TestRT063_4_OfficeRelativeRoleFontLengths(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		t.Run(format, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			dir := t.TempDir()
			proseSource(t, dir, "script.yaml", "folio:\n  manuscript:\n    font:\n      family: Iosevka Publish\n      size: 12pt\n    title-page:\n      enabled: false\n    toc:\n      enabled: false\n    heading:\n      font:\n        size: 1.5em\n        letter-spacing: 0.1em\n    quoted-block:\n      font:\n        size: 0.5em\n        letter-spacing: 0.1em\n")
			src := proseSource(t, dir, "source.md", "## RELATIVE-HEADING\n\nBODY-SIZE-STABLE\n\n> RELATIVE-QUOTE\n")
			pkg := proseConvert(t, src, filepath.Join(dir, "output."+format))
			for _, role := range []struct {
				marker, sizeWord, sizeODT string
				tracking                  float64
			}{
				{"RELATIVE-HEADING", "36", "18pt", 1.8}, {"RELATIVE-QUOTE", "12", "6pt", 0.6},
			} {
				layers := officeEffective(t, pkg, format, officeParagraph(t, officeBody(t, pkg, format), role.marker))
				if format == "docx" {
					officeAssertEffective(t, layers, "sz", "val", role.sizeWord)
				} else {
					officeAssertEffective(t, layers, "text-properties", "font-size", role.sizeODT)
				}
				officeAssertLength(t, layers, format, "spacing", "val", "text-properties", "letter-spacing", role.tracking)
			}
			body := officeEffective(t, pkg, format, officeParagraph(t, officeBody(t, pkg, format), "BODY-SIZE-STABLE"))
			if format == "docx" {
				officeAssertEffective(t, body, "sz", "val", "24")
			} else {
				officeAssertEffective(t, body, "text-properties", "font-size", "12pt")
			}
		})
	}
}

func TestRT063_4_OfficeMirroredGutter(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		t.Run(format, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			dir := t.TempDir()
			proseSource(t, dir, "script.yaml", "folio:\n  page-header:\n    enabled: false\n  page-footer:\n    enabled: false\n  manuscript:\n    page: a4\n    margin: 20mm\n    gutter: 5mm\n    title-page:\n      enabled: false\n    toc:\n      enabled: false\n")
			src := proseSource(t, dir, "source.md", "GUTTER-BODY\n")
			pkg := proseConvert(t, src, filepath.Join(dir, "output."+format))
			var layers []*officeXML
			if format == "docx" {
				layers = []*officeXML{proseSectionForMarker(t, pkg, "GUTTER-BODY")}
				settings := pkg["word/settings.xml"]
				if settings == nil || len(settings.all("mirrorMargins")) == 0 {
					t.Fatal("DOCX gutter lacks mirrored margins")
				}
				for _, flag := range settings.all("mirrorMargins") {
					if flag.attr("val") == "false" || flag.attr("val") == "0" {
						t.Error("mirrored margins disabled")
					}
				}
				points := func(attr string) float64 {
					raw, exists := officeEffectiveAttr(layers, "pgMar", attr)
					if !exists {
						return 0
					}
					n, err := strconv.ParseFloat(raw, 64)
					if err != nil {
						t.Fatal(err)
					}
					return n / 20
				}
				if math.Abs(points("left")+points("gutter")-25*72/25.4) > 0.051 {
					t.Error("inner margin does not include 5mm gutter in addition to 20mm margin")
				}
				officeAssertLength(t, layers, format, "pgMar", "right", "", "", 20*72/25.4)
			} else {
				master := proseMasterForMarker(t, pkg, "GUTTER-BODY")
				for _, layout := range pkg["styles.xml"].all("page-layout") {
					if layout.attr("name") == master.attr("page-layout-name") {
						layers = []*officeXML{layout}
					}
				}
				if len(layers) == 0 {
					t.Fatal("missing assigned page layout")
				}
				officeAssertEffective(t, layers, "page-layout", "page-usage", "mirrored")
				officeAssertLength(t, layers, format, "", "", "page-layout-properties", "margin-left", 25*72/25.4)
				officeAssertLength(t, layers, format, "", "", "page-layout-properties", "margin-right", 20*72/25.4)
			}
			for _, side := range []string{"top", "bottom"} {
				officeAssertLength(t, layers, format, "pgMar", side, "page-layout-properties", "margin-"+side, 20*72/25.4)
			}
		})
	}
}
