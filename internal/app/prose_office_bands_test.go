// ABOUTME: Resolves actual first-section running bands to test frontmatter inheritance.
// ABOUTME: Checks assigned parts and master pages rather than unused package strings.
package app

import (
	"path"
	"path/filepath"
	"strings"
	"testing"
)

func proseFirstBands(t *testing.T, pkg map[string]*officeXML, format, band string) []*officeXML {
	t.Helper()
	body := officeBody(t, pkg, format)
	if format == "docx" {
		sections := body.all("sectPr")
		if len(sections) == 0 {
			t.Fatal("first section has no native page properties")
		}
		rels := pkg["word/_rels/document.xml.rels"]
		var result []*officeXML
		for _, ref := range sections[0].all(band + "Reference") {
			if rels == nil {
				t.Fatal("running band has no relationship part")
			}
			resolved := false
			for _, rel := range rels.all("Relationship") {
				if rel.attr("Id") != ref.attr("id") {
					continue
				}
				name := path.Clean(path.Join("word", rel.attr("Target")))
				if strings.HasPrefix(rel.attr("Target"), "/") {
					name = strings.TrimPrefix(rel.attr("Target"), "/")
				}
				part := pkg[name]
				if part == nil {
					t.Fatalf("band reference resolves to missing part %s", name)
				}
				result = append(result, part)
				resolved = true
			}
			if !resolved {
				t.Fatalf("unresolved %s reference %s", band, ref.attr("id"))
			}
		}
		return result
	}
	styles := pkg["styles.xml"]
	if styles == nil {
		t.Fatal("ODT has no styles part")
	}
	masters := styles.all("master-page")
	if len(masters) == 0 {
		t.Fatal("ODT has no page master")
	}
	master := masters[0]
	// The first explicit paragraph master, when present, overrides the
	// document's default first master. Resolve inherited paragraph styles.
	var first *officeXML
	var visit func(*officeXML)
	visit = func(n *officeXML) {
		if first != nil {
			return
		}
		if n.Name.Local == "p" || n.Name.Local == "h" {
			first = n
			return
		}
		for _, child := range n.Children {
			visit(child)
		}
	}
	visit(body)
	if first != nil {
		for _, layer := range officeEffective(t, pkg, format, first) {
			name := layer.attr("master-page-name")
			if name == "" {
				continue
			}
			found := false
			for _, candidate := range masters {
				if candidate.attr("name") == name {
					master, found = candidate, true
				}
			}
			if !found {
				t.Fatalf("unresolved page master %s", name)
			}
		}
	}
	var result []*officeXML
	seen := map[string]bool{}
	for master != nil && !seen[master.attr("name")] {
		seen[master.attr("name")] = true
		for _, kind := range []string{band, band + "-left", band + "-first"} {
			result = append(result, master.all(kind)...)
		}
		next := master.attr("next-style-name")
		master = nil
		for _, candidate := range masters {
			if candidate.attr("name") == next {
				master = candidate
				break
			}
		}
	}
	return result
}

func TestRT063_6_OfficeFrontmatterBandInheritance(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		for _, band := range []string{"header", "footer"} {
			for _, mode := range []string{"absent", "empty", "distinct", "disabled"} {
				t.Run(format+"/"+band+"/"+mode, func(t *testing.T) {
					t.Setenv("HOME", t.TempDir())
					dir := t.TempDir()
					enabled := "true"
					if mode == "disabled" {
						enabled = "false"
					}
					cfg := "folio:\n  page-" + band + ":\n    enabled: " + enabled + "\n    format: BODY-INHERITED\n    alt-format: BODY-INHERITED\n"
					want := "BODY-INHERITED"
					if mode == "empty" {
						cfg += "    frontmatter-format: ''\n    alt-frontmatter-format: ''\n"
						want = ""
					} else if mode == "distinct" {
						cfg += "    frontmatter-format: FRONT-ONLY\n    alt-frontmatter-format: FRONT-ONLY\n"
						want = "FRONT-ONLY"
					} else if mode == "disabled" {
						want = ""
					}
					cfg += "  manuscript:\n    title-page:\n      enabled: false\n    toc:\n      enabled: false\n"
					proseSource(t, dir, "script.yaml", cfg)
					src := proseSource(t, dir, "source.md", "No body heading: this remains frontmatter.\n")
					pkg := proseConvert(t, src, filepath.Join(dir, "output."+format))
					var texts []string
					for _, part := range proseFirstBands(t, pkg, format, band) {
						texts = append(texts, part.text())
					}
					joined := strings.Join(texts, " ")
					if want == "" {
						if strings.Contains(joined, "BODY-INHERITED") || strings.Contains(joined, "FRONT-ONLY") {
							t.Errorf("blank or disabled band contains text %q", joined)
						}
					} else if !strings.Contains(joined, want) {
						t.Errorf("assigned frontmatter band %q does not contain %s", joined, want)
					}
					if mode == "distinct" && strings.Contains(joined, "BODY-INHERITED") {
						t.Errorf("body format leaked into frontmatter: %s", joined)
					}
				})
			}
		}
	}
}

func TestRT063_4_OfficeRunningBandFontsAndAlignment(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		for _, band := range []string{"header", "footer"} {
			t.Run(format+"/"+band, func(t *testing.T) {
				t.Setenv("HOME", t.TempDir())
				dir := t.TempDir()
				cfg := "folio:\n  page-" + band + ":\n    enabled: true\n    format: BAND-EVEN\n    alt-format: BAND-ODD\n    align: left-right\n    font:\n      family: Iosevka Publish\n      size: 13pt\n      weight: bold\n      style: italic\n      stretch: 90%\n      letter-spacing: 1.25pt\n  manuscript:\n    title-page:\n      enabled: false\n    toc:\n      enabled: false\n    font:\n      size: 12pt\n"
				proseSource(t, dir, "script.yaml", cfg)
				src := proseSource(t, dir, "source.md", "BODY-NO-FONT-LEAK\n")
				pkg := proseConvert(t, src, filepath.Join(dir, "output."+format))
				seen := map[string]bool{}
				for _, part := range proseFirstBands(t, pkg, format, band) {
					for _, paragraph := range part.all("p") {
						want := ""
						if strings.Contains(paragraph.text(), "BAND-EVEN") {
							want = "left"
							seen["even"] = true
						}
						if strings.Contains(paragraph.text(), "BAND-ODD") {
							want = "right"
							seen["odd"] = true
						}
						if want == "" {
							continue
						}
						layers := officeEffective(t, pkg, format, paragraph)
						if format == "docx" {
							officeAssertEffective(t, layers, "jc", "val", want)
							officeAssertEffective(t, layers, "rFonts", "ascii", "Iosevka Publish")
							officeAssertEffective(t, layers, "sz", "val", "26")
							officeAssertEffective(t, layers, "w", "val", "90")
							officeAssertEffective(t, layers, "spacing", "val", "25")
							for _, flag := range []string{"b", "i"} {
								active := false
								for _, layer := range layers {
									for _, node := range layer.all(flag) {
										active = node.attr("val") != "false" && node.attr("val") != "0"
									}
								}
								if !active {
									t.Errorf("running %s lost %s font property", band, flag)
								}
							}
						} else {
							got, _ := officeEffectiveAttr(layers, "paragraph-properties", "text-align")
							if got == "start" {
								got = "left"
							}
							if got == "end" {
								got = "right"
							}
							if got != want {
								t.Errorf("running band alignment %s; want %s", got, want)
							}
							for attr, value := range map[string]string{"font-size": "13pt", "font-weight": "bold", "font-style": "italic", "text-scale": "90%", "letter-spacing": "1.25pt"} {
								officeAssertEffective(t, layers, "text-properties", attr, value)
							}
							name, _ := officeEffectiveAttr(layers, "text-properties", "font-name")
							family := ""
							for _, file := range []string{"styles.xml", "content.xml"} {
								for _, face := range pkg[file].all("font-face") {
									if face.attr("name") == name {
										family = strings.Trim(face.attr("font-family"), "'\"")
									}
								}
							}
							if family != "Iosevka Publish" {
								t.Errorf("running band family=%q", family)
							}
						}
					}
				}
				if !seen["even"] || !seen["odd"] {
					t.Errorf("assigned odd/even running bands missing: %v", seen)
				}
				body := officeEffective(t, pkg, format, officeParagraph(t, officeBody(t, pkg, format), "BODY-NO-FONT-LEAK"))
				if format == "docx" {
					officeAssertEffective(t, body, "sz", "val", "24")
				} else {
					officeAssertEffective(t, body, "text-properties", "font-size", "12pt")
				}
			})
		}
	}
}

func TestRT063_6_OfficeRunningPlaceholders(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		t.Run(format, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			dir := t.TempDir()
			cfg := "folio:\n  page-header:\n    enabled: true\n    format: '[title] | [author] | [part] | [part-number] | [part-prefix] | [part-full] | [chapter] | [chapter-number] | [chapter-prefix] | [chapter-full] | [page] | [total-pages] | [unknown]'\n  manuscript:\n    title-page:\n      enabled: false\n    toc:\n      enabled: false\n    part:\n      prefix: 'PART-PREFIX '\n      separator: ': '\n      show-number: true\n      number-format: I\n      skip-header: false\n    chapter:\n      prefix: 'CHAPTER-PREFIX '\n      separator: ': '\n      show-number: true\n      number-format: '1'\n      skip-header: false\n"
			proseSource(t, dir, "script.yaml", cfg)
			src := proseSource(t, dir, "source.md", "---\ntitle: HEADER-TITLE\nauthor: HEADER-AUTHOR\n---\n\n# PART-NAME\n\n## CHAPTER-NAME\n\nBODY-SENTINEL\n")
			pkg := proseConvert(t, src, filepath.Join(dir, "output."+format))
			var texts []string
			pageCount := false
			for name, member := range pkg {
				var bands []*officeXML
				if format == "docx" && strings.HasPrefix(name, "word/header") {
					bands = append(bands, member)
				}
				if format == "odt" && name == "styles.xml" {
					for _, kind := range []string{"header", "header-left", "header-first"} {
						bands = append(bands, member.all(kind)...)
					}
				}
				for _, band := range bands {
					texts = append(texts, band.text())
					if len(band.all("page-count")) > 0 {
						pageCount = true
					}
					for _, field := range band.all("instrText") {
						if strings.HasPrefix(strings.TrimSpace(field.text())+" ", "NUMPAGES ") {
							pageCount = true
						}
					}
					for _, field := range band.all("fldSimple") {
						if strings.HasPrefix(strings.TrimSpace(field.attr("instr"))+" ", "NUMPAGES ") {
							pageCount = true
						}
					}
				}
			}
			text := strings.Join(texts, " ")
			for _, value := range []string{"HEADER-TITLE", "HEADER-AUTHOR", "PART-NAME", "CHAPTER-NAME", "PART-PREFIX", "CHAPTER-PREFIX", "PART-PREFIX I: PART-NAME", "CHAPTER-PREFIX 1: CHAPTER-NAME", "[unknown]"} {
				if !strings.Contains(text, value) {
					t.Errorf("running matter lacks substituted %q", value)
				}
			}
			if !pageCount {
				t.Error("total-pages is not a native page-count field")
			}
			for _, token := range []string{"[title]", "[author]", "[part]", "[chapter]", "[part-number]", "[chapter-number]", "[part-full]", "[chapter-full]", "[page]", "[total-pages]"} {
				if strings.Contains(text, token) {
					t.Errorf("unresolved running placeholder %s", token)
				}
			}
		})
	}
}
