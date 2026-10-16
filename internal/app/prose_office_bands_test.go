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
	for _, kind := range []string{band, band + "-left", band + "-first"} {
		result = append(result, master.all(kind)...)
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
