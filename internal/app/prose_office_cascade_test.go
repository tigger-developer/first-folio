// ABOUTME: Compares the public PDF and office routes under independently specified YAML layers.
// ABOUTME: Checks winner selection and inherited font size without coupling to loader internals.
package app

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestRT063_4_PDFCascadeParity(t *testing.T) {
	for _, mode := range []string{"manuscript", "letter"} {
		for layer := 0; layer < 5; layer++ {
			if mode == "letter" && layer == 2 {
				continue
			}
			t.Run(fmt.Sprintf("%s/layer-%d", mode, layer), func(t *testing.T) {
				home, dir := t.TempDir(), t.TempDir()
				t.Setenv("HOME", home)
				global := filepath.Join(home, ".config", "first-folio")
				if err := os.MkdirAll(global, 0700); err != nil {
					t.Fatal(err)
				}
				config := func(size int) string {
					return fmt.Sprintf("folio:\n  %s:\n    font:\n      family: Iosevka Publish\n      size: %dpt\n", mode, size)
				}
				proseSource(t, global, "script.yaml", config(9))
				if layer >= 1 {
					proseSource(t, global, "script-british.yaml", config(10))
				}
				want := 9 + layer
				if mode == "letter" {
					if layer >= 3 {
						want = 12
					}
					if layer == 4 {
						want = 13
					}
				}
				body := "CASCADE-SIZE-MARKER keeps its resolved typography.\n"
				if mode == "manuscript" && layer >= 2 {
					body = "---\n" + config(11) + "---\n" + body
				}
				local := "folio:\n  manuscript:\n    title-page:\n      enabled: false\n    toc:\n      enabled: false\n"
				if layer >= 3 {
					local = config(12)
				}
				proseSource(t, dir, "script.yaml", local)
				if layer == 4 {
					proseSource(t, dir, "script-british.yaml", config(13))
				}
				src := proseSource(t, dir, "source.md", body)
				if mode == "letter" {
					src = proseSource(t, dir, "source.org", "#+AUTHOR: Synthetic Author\n* Letters :letter:\n** Subject :subject:\n"+body+"*** Recipient :to:\n**** Theatre :org:\n")
				}
				for _, format := range []string{"pdf", "docx", "odt"} {
					target := filepath.Join(dir, "output."+format)
					args := []string{"manuscript", src, target}
					if mode == "letter" {
						target = filepath.Join(dir, "letter-theatre."+format)
						args = []string{"letter", src, "--dir", dir, "--format", format}
					}
					status, _, stderr := runApp(t, args...)
					if status != 0 {
						t.Fatalf("%s %s export: %s", mode, format, stderr)
					}
					if format == "pdf" {
						xmlPath := filepath.Join(dir, "pdf-spans.xml")
						proseTool(t, "pdftohtml", "-xml", "-zoom", "1", "-noroundcoord", target, xmlPath)
						raw, err := os.ReadFile(xmlPath)
						if err != nil {
							t.Fatal(err)
						}
						tree := officeParseXML(t, raw)
						fonts := map[string]float64{}
						for _, font := range tree.all("fontspec") {
							fonts[font.attr("id")], _ = strconv.ParseFloat(font.attr("size"), 64)
						}
						found := false
						for _, span := range tree.all("text") {
							if strings.Contains(span.text(), "CASCADE-SIZE-MARKER") {
								found = true
								if math.Abs(fonts[span.attr("font")]-float64(want)) > 0.25 {
									t.Errorf("PDF font size %.2f, want %d", fonts[span.attr("font")], want)
								}
							}
						}
						if !found {
							t.Error("PDF cascade specimen missing")
						}
					} else {
						pkg := prosePackage(t, target, format)
						layers := officeEffective(t, pkg, format, officeParagraph(t, officeBody(t, pkg, format), "CASCADE-SIZE-MARKER"))
						if format == "docx" {
							officeAssertEffective(t, layers, "sz", "val", strconv.Itoa(2*want))
						} else {
							officeAssertEffective(t, layers, "text-properties", "font-size", fmt.Sprintf("%dpt", want))
						}
					}
				}
			})
		}
	}
}
