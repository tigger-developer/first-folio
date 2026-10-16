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
					cfg := "folio:\n  page-header:\n    enabled: true\n    format: VISIBLE-HEADER\n    alt-format: VISIBLE-HEADER\n  page-footer:\n    enabled: true\n    format: VISIBLE-FOOTER\n    alt-format: VISIBLE-FOOTER\n  manuscript:\n    title-page:\n      enabled: false\n    toc:\n      enabled: false\n    " + role + ":\n      skip-header: " + skip + "\n      skip-footer: " + skip + "\n"
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
									name := filepath.ToSlash(filepath.Clean(filepath.Join("word", rel.attr("Target"))))
									if strings.HasPrefix(rel.attr("Target"), "/") {
										name = strings.TrimPrefix(rel.attr("Target"), "/")
									}
									member := pkg[name]
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

func TestRT063_6_OfficeContinuousBoundaryAndNeverReset(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		for _, numbering := range []string{"1", "i", "I"} {
			t.Run(format+"/"+numbering, func(t *testing.T) {
				t.Setenv("HOME", t.TempDir())
				dir := t.TempDir()
				cfg := "folio:\n  page-header:\n    enabled: true\n    format: BODY [page]\n    frontmatter-format: FRONT [page]\n  manuscript:\n    title-page:\n      enabled: false\n    toc:\n      enabled: false\n    chapter:\n      page-break-before: false\n      blank-page-before: false\n      blank-page-after: false\n      position: 0mm\n      skip-header: false\n      skip-footer: false\n    page-numbering:\n      frontmatter-format: '" + numbering + "'\n      body-format: '" + numbering + "'\n      body-reset: never\n"
				proseSource(t, dir, "script.yaml", cfg)
				src := proseSource(t, dir, "source.md", "INTRO-CONTINUOUS\n\n## HEADING-CONTINUOUS\n\nBODY-CONTINUOUS\n")
				pkg := proseConvert(t, src, filepath.Join(dir, "output."+format))
				body := officeBody(t, pkg, format)
				heading := officeEffective(t, pkg, format, officeParagraph(t, body, "HEADING-CONTINUOUS"))
				if format == "docx" {
					sections := body.all("sectPr")
					for i, section := range sections {
						if i == 0 {
							continue
						}
						types := section.all("type")
						if len(types) == 0 || types[0].attr("val") != "continuous" {
							t.Error("unrequested section page break at continuous boundary")
						}
						for _, number := range section.all("pgNumType") {
							if number.attr("start") != "" {
								t.Error("body-reset never introduced native restart")
							}
						}
					}
					for _, br := range body.all("br") {
						if br.attr("type") == "page" {
							t.Error("continuous source acquired page break")
						}
					}
					for _, layer := range heading {
						for _, flag := range layer.all("pageBreakBefore") {
							if flag.attr("val") != "false" && flag.attr("val") != "0" {
								t.Error("heading has unrequested pageBreakBefore")
							}
						}
					}
				} else {
					intro := proseMasterForMarker(t, pkg, "INTRO-CONTINUOUS")
					chapter := proseMasterForMarker(t, pkg, "HEADING-CONTINUOUS")
					if intro.attr("name") != chapter.attr("name") {
						t.Error("same-format continuous ODT boundary changes master and therefore forces a page")
					}
					var assignments []string
					for _, field := range body.all("variable-set") {
						assignments = append(assignments, field.attr("string-value"))
					}
					if !strings.Contains(strings.Join(assignments, "|"), "BODY ") || len(intro.all("variable-get")) == 0 {
						t.Error("continuous boundary has no live frontmatter/body running-text transition")
					}
					for _, layer := range heading {
						for _, properties := range layer.all("paragraph-properties") {
							if properties.attr("break-before") == "page" {
								t.Error("continuous heading forces page break")
							}
							if number := properties.attr("page-number"); number != "" && number != "auto" {
								t.Error("body-reset never introduced native restart")
							}
						}
					}
				}
				proseAssertNumberFormat(t, pkg, format, numbering)
			})
		}
	}
}

func proseAssertNumberFormat(t *testing.T, pkg map[string]*officeXML, format, numbering string) {
	t.Helper()
	for _, member := range pkg {
		if format == "odt" {
			for _, kind := range []string{"page-number", "page-variable-get"} {
				for _, field := range member.all(kind) {
					if field.attr("num-format") == numbering {
						return
					}
				}
			}
		} else {
			want := map[string]string{"1": "decimal", "i": "lowerRoman", "I": "upperRoman"}[numbering]
			for _, property := range member.all("pgNumType") {
				if property.attr("fmt") == want {
					return
				}
			}
			for _, kind := range []string{"instrText", "fldSimple"} {
				for _, field := range member.all(kind) {
					instruction := field.text()
					if kind == "fldSimple" {
						instruction = field.attr("instr")
					}
					if !strings.Contains(instruction, "PAGE") {
						continue
					}
					if numbering == "1" && !strings.Contains(instruction, "roman") && !strings.Contains(instruction, "ROMAN") {
						return
					}
					if numbering == "i" && strings.Contains(instruction, `\* roman`) {
						return
					}
					if numbering == "I" && strings.Contains(instruction, `\* ROMAN`) {
						return
					}
				}
			}
		}
	}
	t.Errorf("native running fields do not express page-number format %s", numbering)
}

func proseNativeBreakCount(t *testing.T, pkg map[string]*officeXML, format string) int {
	t.Helper()
	body := officeBody(t, pkg, format)
	count := 0
	if format == "docx" {
		for _, br := range body.all("br") {
			if br.attr("type") == "page" {
				count++
			}
		}
		for _, section := range body.all("sectPr") {
			types := section.all("type")
			if len(types) == 0 || types[0].attr("val") != "continuous" {
				count++
			}
		}
	}
	previousMaster := ""
	var visit func(*officeXML)
	visit = func(node *officeXML) {
		if node.Name.Local != "p" && node.Name.Local != "h" {
			for _, child := range node.Children {
				visit(child)
			}
			return
		}
		layers := officeEffective(t, pkg, format, node)
		before, after := false, false
		for _, layer := range layers {
			if format == "docx" {
				for _, flag := range layer.all("pageBreakBefore") {
					before = flag.attr("val") != "false" && flag.attr("val") != "0"
				}
			} else {
				for _, prop := range layer.all("paragraph-properties") {
					if prop.attr("break-before") == "page" {
						before = true
					}
					if prop.attr("break-after") == "page" {
						after = true
					}
				}
				if name := layer.attr("master-page-name"); name != "" {
					if previousMaster != "" && name != previousMaster {
						before = true
					}
					previousMaster = name
				}
			}
		}
		if before {
			count++
		}
		if after {
			count++
		}
	}
	visit(body)
	return count
}

func TestRT063_6_OfficeExplicitBlankPageSwitches(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		for _, role := range []string{"part", "chapter", "toc", "copyright"} {
			for _, side := range []string{"before", "after"} {
				t.Run(format+"/"+role+"/"+side, func(t *testing.T) {
					t.Setenv("HOME", t.TempDir())
					dir := t.TempDir()
					src := proseSource(t, dir, "source.md", "INTRO-MARKER\n\n# Part\n\n## Chapter\n\nBODY-MARKER\n")
					var counts [2]int
					for i, enabled := range []string{"false", "true"} {
						cfg := "folio:\n  manuscript:\n    title-page:\n      enabled: false\n"
						for _, kind := range []string{"part", "chapter", "toc", "copyright"} {
							cfg += "    " + kind + ":\n"
							if kind == "toc" || kind == "copyright" {
								cfg += "      enabled: true\n"
							} else {
								cfg += "      page-break-before: false\n"
							}
							for _, direction := range []string{"before", "after"} {
								value := "false"
								if kind == role && direction == side {
									value = enabled
								}
								cfg += "      blank-page-" + direction + ": " + value + "\n"
							}
						}
						proseSource(t, dir, "script.yaml", cfg)
						pkg := proseConvert(t, src, filepath.Join(dir, "output."+format))
						counts[i] = proseNativeBreakCount(t, pkg, format)
						for _, marker := range []string{"INTRO-MARKER", "BODY-MARKER"} {
							if strings.Count(officeBody(t, pkg, format).text(), marker) != 1 {
								t.Errorf("blank-page policy lost or duplicated %s", marker)
							}
						}
					}
					if counts[1] <= counts[0] {
						t.Errorf("blank-page-%s true adds no native page transition: false=%d true=%d", side, counts[0], counts[1])
					}
				})
			}
		}
	}
}

func TestRT063_6_OfficeHeadingPageBreakSwitch(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		for _, role := range []string{"part", "chapter", "toc"} {
			t.Run(format+"/"+role, func(t *testing.T) {
				t.Setenv("HOME", t.TempDir())
				dir := t.TempDir()
				prefix := "# "
				if role == "chapter" {
					prefix = "## "
				}
				src := proseSource(t, dir, "source.md", "---\ntitle: BREAK-SWITCH\n---\n"+prefix+"First\n\nFirst body.\n\n"+prefix+"Second\n\nSecond body.\n")
				var breaks [2]int
				for i, enabled := range []string{"false", "true"} {
					cfg := "folio:\n  manuscript:\n    title-page:\n      enabled: false\n    toc:\n      enabled: false\n"
					if role == "toc" {
						cfg = "folio:\n  manuscript:\n    title-page:\n      enabled: true\n"
					}
					cfg += "    " + role + ":\n      page-break-before: " + enabled + "\n      blank-page-before: false\n      blank-page-after: false\n"
					if role == "toc" {
						cfg += "      enabled: true\n"
					}
					proseSource(t, dir, "script.yaml", cfg)
					pkg := proseConvert(t, src, filepath.Join(dir, "output."+format))
					breaks[i] = proseNativeBreakCount(t, pkg, format)
				}
				if breaks[1] <= breaks[0] {
					t.Errorf("second %s page-break-before has no native effect: %v", role, breaks)
				}
			})
		}
	}
}

func TestRT063_6_OfficeTitlePageNumberSwitch(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		for _, enabled := range []string{"false", "true"} {
			t.Run(format+"/"+enabled, func(t *testing.T) {
				t.Setenv("HOME", t.TempDir())
				dir := t.TempDir()
				cfg := "folio:\n  manuscript:\n    style: us\n    title-page:\n      enabled: true\n      page-number: " + enabled + "\n    toc:\n      enabled: false\n      page-break-before: true\n"
				proseSource(t, dir, "script.yaml", cfg)
				src := proseSource(t, dir, "source.md", "---\ntitle: TITLE-PAGE-SWITCH\n---\n\n## Chapter\n\nBody.\n")
				pkg := proseConvert(t, src, filepath.Join(dir, "output."+format))
				var footers []*officeXML
				if format == "docx" {
					section := proseSectionForMarker(t, pkg, "TITLE-PAGE-SWITCH")
					kind := "default"
					if len(section.all("titlePg")) > 0 {
						kind = "first"
					}
					for _, ref := range section.all("footerReference") {
						if ref.attr("type") != kind {
							continue
						}
						for _, rel := range pkg["word/_rels/document.xml.rels"].all("Relationship") {
							if rel.attr("Id") != ref.attr("id") {
								continue
							}
							name := filepath.ToSlash(filepath.Clean(filepath.Join("word", rel.attr("Target"))))
							if strings.HasPrefix(rel.attr("Target"), "/") {
								name = strings.TrimPrefix(rel.attr("Target"), "/")
							}
							if pkg[name] == nil {
								t.Fatal("title footer reference is broken")
							}
							footers = append(footers, pkg[name])
						}
					}
				} else {
					master := proseMasterForMarker(t, pkg, "TITLE-PAGE-SWITCH")
					footers = master.all("footer-first")
					if len(footers) == 0 {
						footers = master.all("footer")
					}
				}
				pageField := false
				for _, footer := range footers {
					if footer.attr("display") == "false" {
						continue
					}
					if len(footer.all("page-number"))+len(footer.all("page-variable-get")) > 0 {
						pageField = true
					}
					for _, field := range footer.all("instrText") {
						if strings.HasPrefix(strings.TrimSpace(field.text()), "PAGE") {
							pageField = true
						}
					}
					for _, field := range footer.all("fldSimple") {
						if strings.HasPrefix(strings.TrimSpace(field.attr("instr")), "PAGE") {
							pageField = true
						}
					}
				}
				if pageField != (enabled == "true") {
					t.Errorf("title page native page field=%v for page-number=%s", pageField, enabled)
				}
			})
		}
	}
}
