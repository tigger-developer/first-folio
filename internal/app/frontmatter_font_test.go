// ABOUTME: Tests independent frontmatter paragraph fonts through conversion output.
// ABOUTME: Covers six native font properties and preserves dialogue and headings.
package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFrontmatterParagraphExports(t *testing.T) {
	binary := officeBuild(t)
	dir := t.TempDir()
	source := filepath.Join(dir, "play.org")
	writeAppFile(t, source, "#+TITLE: Font scope\n* Introduction\nDetailed introduction text.\n\nAnother introduction paragraph.\n* ACT ONE\n** SCENE ONE\n**** BOB\nSpoken dialogue.\n")
	writeAppFile(t, filepath.Join(dir, "script.yaml"), "folio:\n  positioning:\n    frontmatter:\n      paragraph:\n        font:\n          family: Iosevka Publish\n          size: 9pt\n          weight: bold\n          style: italic\n          stretch: 110%\n          letter-spacing: 0.1pt\n")
	for _, format := range []string{"typst", "docx", "odt"} {
		t.Run(format, func(t *testing.T) {
			args := []string{"convert", source, "--to", format}
			typstPath := filepath.Join(dir, "play.typ")
			if format == "typst" {
				args = []string{"convert", source, typstPath}
			}
			raw, stderr, err := officeRun(t, binary, dir, officeEnv(t, t.TempDir(), nil), args...)
			if err != nil {
				t.Fatalf("%v: %s", err, stderr)
			}
			if format == "typst" {
				raw, err = os.ReadFile(typstPath)
				if err != nil {
					t.Fatal(err)
				}
				for _, fragment := range []string{`font: "Iosevka Publish", size: 9pt, weight: "bold", stretch: 110%, style: "italic", tracking: 0.1pt`, `#frontmatter-paragraph[Detailed introduction text.]`, `#frontmatter-paragraph[Another introduction paragraph.]`} {
					if !strings.Contains(string(raw), fragment) {
						t.Errorf("missing %s", fragment)
					}
				}
				return
			}
			pkg := officePackage(t, raw)
			body := officeBody(t, pkg, format)
			for _, text := range []string{"Detailed introduction text.", "Another introduction paragraph."} {
				layers := officeEffective(t, pkg, format, officeParagraph(t, body, text))
				if format == "docx" {
					for _, p := range [][3]string{{"rFonts", "ascii", "Iosevka Publish"}, {"sz", "val", "18"}, {"b", "val", "true"}, {"i", "val", "true"}, {"w", "val", "110"}, {"spacing", "val", "2"}} {
						officeAssertEffective(t, layers, p[0], p[1], p[2])
					}
				} else {
					for _, p := range [][2]string{{"font-family", "Iosevka Publish"}, {"font-size", "9pt"}, {"font-weight", "bold"}, {"font-style", "italic"}, {"text-scale", "110%"}, {"letter-spacing", "0.1pt"}} {
						officeAssertEffective(t, layers, "text-properties", p[0], p[1])
					}
				}
			}
			for _, text := range []string{"Spoken dialogue.", "Introduction"} {
				layers := officeEffective(t, pkg, format, officeParagraph(t, body, text))
				if format == "docx" {
					officeAssertEffective(t, layers, "rFonts", "ascii", "Libertinus Serif")
				} else {
					officeAssertEffective(t, layers, "text-properties", "font-family", "Libertinus Serif")
				}
			}
		})
	}
}
