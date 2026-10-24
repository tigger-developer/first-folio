// ABOUTME: Verifies shared monospace typography in editable script and letter outputs.
// ABOUTME: Covers inline code through the public DOCX and ODT command paths.
package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRT068_1SharedMonospaceFontReachesEditableOutputs(t *testing.T) {
	for _, document := range []string{"script", "letter"} {
		for _, format := range []string{"docx", "odt"} {
			t.Run(document+"/"+format, func(t *testing.T) {
				t.Setenv("HOME", t.TempDir())
				dir := t.TempDir()
				writeAppFile(t, filepath.Join(dir, "script.yaml"), "folio:\n  mono:\n    font:\n      family: Iosevka Publish\n      size: 13pt\n      weight: bold\n      stretch: 125%\n      style: italic\n      letter-spacing: 0.1pt\n")
				var target string
				var args []string
				if document == "script" {
					source := filepath.Join(dir, "play.org")
					writeAppFile(t, source, "* ACT ONE\n**** CÁIT\nShared =MONO-OFFICE-MARKER=.\n")
					target = filepath.Join(dir, "play."+format)
					args = []string{"convert", source, target}
				} else {
					source := filepath.Join(dir, "letters.org")
					writeAppFile(t, source, "* Letters :letter:\n** Subject :subject:\nShared =MONO-OFFICE-MARKER=.\n*** Recipient :to:\n**** Theatre :org:\n")
					target = filepath.Join(dir, "letter-theatre."+format)
					args = []string{"letter", source, "--format", format, "--dir", dir}
				}
				status, stdout, stderr := runApp(t, args...)
				if status != 0 {
					t.Fatalf("status %d; stdout=%q stderr=%q", status, stdout, stderr)
				}
				raw, err := os.ReadFile(target)
				if err != nil {
					t.Fatal(err)
				}
				pkg := officePackage(t, raw)
				layers := proseInlineLayers(t, pkg, format, "MONO-OFFICE-MARKER")
				if format == "docx" {
					officeAssertEffective(t, layers, "rFonts", "ascii", "Iosevka Publish")
					officeAssertEffective(t, layers, "sz", "val", "26")
					officeAssertEffective(t, layers, "w", "val", "125")
				} else {
					officeAssertEffective(t, layers, "text-properties", "font-family", "Iosevka Publish")
					officeAssertEffective(t, layers, "text-properties", "font-size", "13pt")
					officeAssertEffective(t, layers, "text-properties", "text-scale", "125%")
				}
			})
		}
	}
}
