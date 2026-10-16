// ABOUTME: Verifies native heading section assignment and configured physical parity.
// ABOUTME: Follows section/master references rather than counting unrelated page styles.
package app

import (
	"path/filepath"
	"strings"
	"testing"
)

func proseSectionForMarker(t *testing.T, pkg map[string]*officeXML, marker string) *officeXML {
	t.Helper()
	body := officeBody(t, pkg, "docx")
	found := false
	var section *officeXML
	var visit func(*officeXML)
	visit = func(n *officeXML) {
		if section != nil {
			return
		}
		if n.Name.Local == "p" && strings.Contains(n.text(), marker) {
			found = true
		}
		if found && n.Name.Local == "sectPr" {
			section = n
			return
		}
		for _, child := range n.Children {
			visit(child)
		}
	}
	visit(body)
	if section == nil {
		t.Fatalf("no section properties for %s", marker)
	}
	return section
}

func proseMasterForMarker(t *testing.T, pkg map[string]*officeXML, marker string) *officeXML {
	t.Helper()
	body := officeBody(t, pkg, "odt")
	name := ""
	found := false
	var visit func(*officeXML)
	visit = func(n *officeXML) {
		if found {
			return
		}
		if n.Name.Local == "p" || n.Name.Local == "h" {
			for _, layer := range officeEffective(t, pkg, "odt", n) {
				if value := layer.attr("master-page-name"); value != "" {
					name = value
				}
			}
			if strings.Contains(n.text(), marker) {
				found = true
			}
			return
		}
		for _, child := range n.Children {
			visit(child)
		}
	}
	visit(body)
	if !found {
		t.Fatalf("no paragraph containing %s", marker)
	}
	masters := pkg["styles.xml"].all("master-page")
	for _, master := range masters {
		if name == "" || master.attr("name") == name {
			return master
		}
	}
	t.Fatalf("no master %q for %s", name, marker)
	return nil
}

func TestRT063_6_OfficeHeadingParityAssignments(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		for _, kind := range []string{"part", "chapter"} {
			for _, side := range []string{"right", "left"} {
				t.Run(format+"/"+kind+"/"+side, func(t *testing.T) {
					t.Setenv("HOME", t.TempDir())
					dir := t.TempDir()
					cfg := "folio:\n  manuscript:\n    title-page:\n      enabled: false\n    toc:\n      enabled: false\n    page-numbering:\n      frontmatter-format: '1'\n      body-format: '1'\n      body-reset: never\n    " + kind + ":\n      page-break-before: true\n      blank-page-before: enforce-" + side + "\n      show-number: false\n"
					proseSource(t, dir, "script.yaml", cfg)
					prefix := "# "
					if kind == "chapter" {
						prefix = "## "
					}
					src := proseSource(t, dir, "source.md", prefix+"FIRST-HEADING\n\nFirst body.\n\n"+prefix+"PARITY-HEADING\n\nSecond body.\n")
					pkg := proseConvert(t, src, filepath.Join(dir, "output."+format))
					if format == "docx" {
						section := proseSectionForMarker(t, pkg, "PARITY-HEADING")
						want := "oddPage"
						if side == "left" {
							want = "evenPage"
						}
						officeAssertEffective(t, []*officeXML{section}, "type", "val", want)
					} else {
						master := proseMasterForMarker(t, pkg, "PARITY-HEADING")
						name := master.attr("page-layout-name")
						matched := false
						for _, layout := range pkg["styles.xml"].all("page-layout") {
							if layout.attr("name") != name {
								continue
							}
							matched = true
							officeAssertEffective(t, []*officeXML{layout}, "page-layout", "page-usage", side)
						}
						if !matched {
							t.Fatalf("master references missing layout %s", name)
						}
					}
				})
			}
		}
	}
}

func TestRT063_4_OfficeHeadingGeometry(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		for _, align := range []string{"left", "center", "right"} {
			t.Run(format+"/"+align, func(t *testing.T) {
				t.Setenv("HOME", t.TempDir())
				dir := t.TempDir()
				cfg := "folio:\n  manuscript:\n    title-page:\n      enabled: false\n    toc:\n      enabled: false\n    part:\n      align: " + align + "\n      show-number: false\n    chapter:\n      align: " + align + "\n      show-number: false\n      position: 0mm\n      space-after: 17pt\n"
				proseSource(t, dir, "script.yaml", cfg)
				src := proseSource(t, dir, "source.md", "# PART-GEOMETRY\n\n## CHAPTER-GEOMETRY\n\nBody.\n")
				pkg := proseConvert(t, src, filepath.Join(dir, "output."+format))
				for _, marker := range []string{"PART-GEOMETRY", "CHAPTER-GEOMETRY"} {
					proseAssertAlignment(t, pkg, format, marker, align)
				}
				layers := officeEffective(t, pkg, format, officeParagraph(t, officeBody(t, pkg, format), "CHAPTER-GEOMETRY"))
				officeAssertLength(t, layers, format, "spacing", "after", "paragraph-properties", "margin-bottom", 17)
			})
		}
	}
}

func TestRT063_6_OfficeOpeningBandSuppression(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		for _, role := range []string{"part", "chapter", "copyright"} {
			for _, skip := range []string{"true", "false"} {
				t.Run(format+"/"+role+"/"+skip, func(t *testing.T) {
					t.Setenv("HOME", t.TempDir())
					dir := t.TempDir()
					cfg := "folio:\n  page-header:\n    enabled: true\n    format: VISIBLE-HEADER\n  page-footer:\n    enabled: true\n    format: VISIBLE-FOOTER\n  manuscript:\n    title-page:\n      enabled: false\n    toc:\n      enabled: false\n    " + role + ":\n      skip-header: " + skip + "\n      skip-footer: " + skip + "\n"
					source, marker := "## OPENING-MARKER\n\nBody.\n", "OPENING-MARKER"
					if role == "part" {
						source = "# OPENING-MARKER\n\nBody.\n"
					}
					if role == "copyright" {
						cfg += "      enabled: true\n      credits: [OPENING-MARKER]\n"
						source = "## Chapter\n\nBody.\n"
					}
					proseSource(t, dir, "script.yaml", cfg)
					src := proseSource(t, dir, "source.md", source)
					pkg := proseConvert(t, src, filepath.Join(dir, "output."+format))
					for _, band := range []string{"header", "footer"} {
						visible := false
						if format == "docx" {
							section := proseSectionForMarker(t, pkg, marker)
							kind := "default"
							if len(section.all("titlePg")) > 0 {
								kind = "first"
							}
							for _, ref := range section.all(band + "Reference") {
								if ref.attr("type") != kind {
									continue
								}
								for _, rel := range pkg["word/_rels/document.xml.rels"].all("Relationship") {
									if rel.attr("Id") != ref.attr("id") {
										continue
									}
									member := pkg[filepath.ToSlash(filepath.Clean(filepath.Join("word", rel.attr("Target"))))]
									if member == nil {
										t.Fatal("opening running band relationship is broken")
									}
									visible = strings.Contains(member.text(), "VISIBLE-"+strings.ToUpper(band))
								}
							}
						} else {
							master := proseMasterForMarker(t, pkg, marker)
							parts := master.all(band + "-first")
							if len(parts) == 0 {
								parts = master.all(band)
							}
							for _, part := range parts {
								if part.attr("display") != "false" && strings.Contains(part.text(), "VISIBLE-"+strings.ToUpper(band)) {
									visible = true
								}
							}
						}
						if visible != (skip == "false") {
							t.Errorf("%s opening %s visibility=%v for skip-%s=%s", role, band, visible, band, skip)
						}
					}
				})
			}
		}
	}
}
