// ABOUTME: Checks script footnote visibility through public conversion outputs.
// ABOUTME: Keeps disabled note references out of Typst and native DOCX text.
package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRT065_1DisabledFootnotesRemoveReferences(t *testing.T) {
	for _, fixture := range []struct {
		extension, source, reference, unknown string
	}{
		{"org", "#+TITLE: Notes\n* ACT ONE\n*** Before[fn:note] after[fn:note]. Escaped \\[fn:note]. Unknown [fn:missing].\n[fn:note] Defined note.\n", "[fn:note]", "[fn:missing]"},
		{"md", "# Notes\n\n## ACT ONE\n\nBefore[^note] after[^note]. Escaped \\[^note]. Unknown [^missing].\n\n[^note]: Defined note.\n", "[^note]", "[^missing]"},
	} {
		for _, output := range []string{"typ", "docx"} {
			t.Run(fixture.extension+"/"+output, func(t *testing.T) {
				dir := t.TempDir()
				t.Setenv("HOME", t.TempDir())
				writeAppFile(t, filepath.Join(dir, "script.yaml"), "render:\n  footnotes: false\n")
				source := filepath.Join(dir, "play."+fixture.extension)
				writeAppFile(t, source, fixture.source)
				target := filepath.Join(dir, "play."+output)
				status, stdout, stderr := runApp(t, "convert", source, target)
				if status != 0 {
					t.Fatalf("status %d\nstdout:%s\nstderr:%s", status, stdout, stderr)
				}
				visible := ""
				if output == "docx" {
					raw, err := os.ReadFile(target)
					if err != nil {
						t.Fatal(err)
					}
					pkg := officePackage(t, raw)
					visible = pkg["word/document.xml"].text()
					if note := pkg["word/footnotes.xml"]; note != nil && strings.Contains(note.text(), "Defined note.") {
						t.Fatal("disabled footnote body remains in DOCX")
					}
				} else {
					visible = readAppFile(t, target)
				}
				if got := strings.Count(visible, strings.Trim(fixture.reference, "[]")); got != 1 || strings.Contains(visible, "Defined note.") {
					t.Fatalf("%s has %d reference markers; want only the escaped literal", output, got)
				}
				if !strings.Contains(visible, strings.Trim(fixture.unknown, "[]")) {
					t.Fatalf("unresolved reference changed in %s", output)
				}
			})
		}
	}
}
