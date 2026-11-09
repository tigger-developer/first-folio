// ABOUTME: Verifies Org underline syntax renders as yellow highlighting in PDF output.
// ABOUTME: Exercises script, letter, and manuscript conversion through the public CLI.
package app

import (
	"image"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestOrgUnderlineRendersAsYellowHighlightInPDF(t *testing.T) {
	requireW073Tools(t)
	const marker = "HIGHLIGHTMARKER"
	for _, tc := range []struct {
		name   string
		source string
		target func(string) string
		args   func(string, string) []string
		config string
	}{
		{
			name:   "script",
			source: "#+TITLE: Highlight test\n* ACT ONE\n**** CÁIT\nBefore _" + marker + "_ after.\n",
			target: func(dir string) string { return filepath.Join(dir, "script.pdf") },
			args: func(source, target string) []string {
				return []string{"convert", source, target}
			},
			config: "folio:\n  title-page:\n    enabled: false\n",
		},
		{
			name: "letter",
			source: "* Letters :letter:\n** Subject :subject:\nBefore _" + marker + "_ after.\n" +
				"*** Recipient :to:\n**** Example Theatre :org:\n",
			target: func(dir string) string { return filepath.Join(dir, "highlight-example-theatre.pdf") },
			args: func(source, _ string) []string {
				return []string{"letter", source, "--dir", filepath.Dir(source), "--prefix", "highlight"}
			},
		},
		{
			name:   "manuscript",
			source: "#+TITLE: Highlight test\n** Chapter One\nBefore _" + marker + "_ after.\n",
			target: func(dir string) string { return filepath.Join(dir, "manuscript.pdf") },
			args: func(source, target string) []string {
				return []string{"manuscript", source, target}
			},
			config: "folio:\n  manuscript:\n    title-page:\n      enabled: false\n    toc:\n      enabled: false\n",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			t.Setenv("HOME", t.TempDir())
			source := filepath.Join(dir, tc.name+".org")
			target := tc.target(dir)
			writeAppFile(t, source, tc.source)
			if tc.config != "" {
				writeAppFile(t, filepath.Join(dir, "script.yaml"), tc.config)
			}
			status, _, stderr := runApp(t, tc.args(source, target)...)
			if status != 0 {
				t.Fatalf("PDF conversion failed: %s", stderr)
			}
			text, err := exec.Command("pdftotext", target, "-").CombinedOutput()
			if err != nil {
				t.Fatalf("extracting PDF text: %v: %s", err, text)
			}
			if !strings.Contains(string(text), marker) {
				t.Fatalf("PDF lost highlighted text %q", marker)
			}
			if pixels := yellowHighlightPixels(rasterW073Page(t, target, 1)); pixels < 20 {
				t.Fatalf("PDF yellow highlight pixels = %d; want at least 20", pixels)
			}
		})
	}
}

func yellowHighlightPixels(page image.Image) int {
	count := 0
	for y := page.Bounds().Min.Y; y < page.Bounds().Max.Y; y++ {
		for x := page.Bounds().Min.X; x < page.Bounds().Max.X; x++ {
			r, g, b, _ := page.At(x, y).RGBA()
			red, green, blue := r>>8, g>>8, b>>8
			if red >= 248 && green >= 232 && green <= 250 && blue >= 130 && blue <= 210 {
				count++
			}
		}
	}
	return count
}
