// ABOUTME: Guards editable prose baselines and reuse of native paragraph styles.
// ABOUTME: Exercises both manuscript sources and outputs through public commands.
package app

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

func TestRT063_4_OfficeNaturalBodyBaseline(t *testing.T) {
	for _, input := range []string{"md", "org", "letter"} {
		for _, format := range []string{"docx", "odt"} {
			t.Run(input+"/"+format, func(t *testing.T) {
				t.Setenv("HOME", t.TempDir())
				dir := t.TempDir()
				proseSource(t, dir, "script.yaml", "folio:\n  manuscript:\n    line-spacing: 2\n    font:\n      size: 10pt\n    title-page:\n      enabled: false\n    toc:\n      enabled: false\n")
				text := "FIRST-BODY-PARAGRAPH\n\nSECOND-BODY-PARAGRAPH\n\nThird body has a note[^n].\n\n[^n]: NOTE-SPECIMEN\n"
				if input == "org" {
					text = strings.ReplaceAll(strings.ReplaceAll(text, "[^n]:", "[fn:n]"), "[^n]", "[fn:n]")
				}
				path := proseSource(t, dir, "source."+input, text)
				target := filepath.Join(dir, "output."+format)
				args := []string{"manuscript", path, target}
				if input == "letter" {
					path = proseSource(t, dir, "source.org", "#+AUTHOR: Writer\n* Letters :letter:\n** Subject :subject:\n"+text+"*** Recipient :to:\n**** Theatre :org:\n")
					target = filepath.Join(dir, "letter-theatre."+format)
					args = []string{"letter", path, "--format", format, "--dir", dir}
				}
				status, _, stderr := runApp(t, args...)
				if status != 0 {
					t.Fatal(stderr)
				}
				pkg := prosePackage(t, target, format)
				for _, marker := range []string{"FIRST-BODY-PARAGRAPH", "SECOND-BODY-PARAGRAPH"} {
					paragraph := officeParagraph(t, officeBody(t, pkg, format), marker)
					layers := append(officeEffective(t, pkg, format, paragraph), paragraph)
					for _, layer := range layers {
						for _, n := range layer.all("position") {
							if n.attr("val") != "0" {
								t.Errorf("%s body has character offset %s", marker, n.attr("val"))
							}
						}
						for _, n := range layer.all("vertAlign") {
							if n.attr("val") != "baseline" {
								t.Errorf("ordinary body has vertical alignment %s", n.attr("val"))
							}
						}
						for _, n := range layer.all("text-properties") {
							if value := n.attr("text-position"); value != "" && value != "0% 100%" {
								t.Errorf("%s body has character offset %s", marker, value)
							}
						}
					}
				}
			})
		}
	}
}

func TestRT063_13_OfficeStylesDoNotGrowWithChapters(t *testing.T) {
	for _, input := range []string{"md", "org"} {
		for _, format := range []string{"docx", "odt"} {
			t.Run(input+"/"+format, func(t *testing.T) {
				t.Setenv("HOME", t.TempDir())
				counts := []int{}
				for _, chapters := range []int{4, 40} {
					dir := t.TempDir()
					proseSource(t, dir, "script.yaml", "folio:\n  manuscript:\n    title-page:\n      enabled: false\n    toc:\n      enabled: true\n      include-chapters: true\n      page-numbers: true\n      links: true\n")
					var text strings.Builder
					prefix := "##"
					if input == "org" {
						prefix = "**"
					}
					for i := 1; i <= chapters; i++ {
						fmt.Fprintf(&text, "%s CH%03d\n\nBODY%03d\n\n", prefix, i, i)
					}
					pkg := proseConvert(t, proseSource(t, dir, "source."+input, text.String()), filepath.Join(dir, "output."+format))
					styles := pkg["word/styles.xml"]
					if format == "odt" {
						styles = pkg["styles.xml"].all("styles")[0]
					}
					count := 0
					for _, style := range styles.all("style") {
						name := style.attr("styleId")
						if format == "odt" {
							name = style.attr("name")
						}
						if strings.HasPrefix(name, "FolioHeading") || strings.HasPrefix(name, "FolioContentsEntry") {
							count++
						}
					}
					counts = append(counts, count)
					// Every heading's bookmark and live contents reference must remain distinct.
					body := officeBody(t, pkg, format)
					refKind, refAttr := "fldSimple", "instr"
					if format == "odt" {
						refKind, refAttr = "bookmark-ref", "ref-name"
					}
					refs := map[string]bool{}
					for _, ref := range body.all(refKind) {
						v := ref.attr(refAttr)
						if format == "odt" || strings.HasPrefix(v, "PAGEREF ") {
							refs[v] = true
						}
					}
					if len(refs) != chapters {
						t.Errorf("got %d distinct contents references, want %d", len(refs), chapters)
					}
					for i := 1; i <= chapters; i++ {
						if !strings.Contains(body.text(), fmt.Sprintf("BODY%03d", i)) {
							t.Errorf("lost chapter %d", i)
						}
					}
				}
				if counts[0] != counts[1] || counts[1] > 6 {
					t.Errorf("heading/contents styles grow with instances: %v; want bounded role styles", counts)
				}
			})
		}
	}
}
