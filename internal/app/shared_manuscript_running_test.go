// ABOUTME: Exercises promoted running-matter settings through public manuscript PDF output.
// ABOUTME: Verifies canonical and legacy paths retain the same rendered manuscript text.
package app

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestSharedManuscriptRunningMatterMatchesLegacyPDF(t *testing.T) {
	for _, tool := range []string{"typst", "pdftotext"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("%s is not installed", tool)
		}
	}
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	source := filepath.Join(dir, "book.md")
	writeAppFile(t, source, "---\ntitle: Shared Book\nauthor: Example Author\n---\n\n## Chapter One\n\nBody one.\n\n## Chapter Two\n\nBody two.\n")
	blocks := "page-header:\n  enabled: true\n  format: 'MANUHEAD [title] [chapter] [unknown]'\n  distance-from-edge: 20mm\n  content-padding-after: 10mm\npage-footer:\n  enabled: true\n  format: 'MANUFOOT [page]/[total-pages]'\n"
	var baseline string
	for _, legacy := range []bool{false, true} {
		indent := "  "
		prefix := "folio:\n"
		if legacy {
			indent = "    "
			prefix += "  manuscript:\n"
		}
		var yaml strings.Builder
		yaml.WriteString(prefix)
		for _, line := range strings.Split(strings.TrimSuffix(blocks, "\n"), "\n") {
			yaml.WriteString(indent + line + "\n")
		}
		if !legacy {
			yaml.WriteString("  manuscript:\n")
		}
		yaml.WriteString("    toc:\n      enabled: false\n    chapter:\n      skip-header: false\n      skip-footer: false\n")
		writeAppFile(t, filepath.Join(dir, "script.yaml"), yaml.String())
		target := filepath.Join(dir, "out.pdf")
		status, _, stderr := runApp(t, "manuscript", source, target)
		if status != 0 {
			t.Fatalf("legacy=%v status %d: %s", legacy, status, stderr)
		}
		if strings.Contains(stderr, "deprecated") != legacy {
			t.Fatalf("legacy=%v unexpected diagnostics: %q", legacy, stderr)
		}
		raw, err := exec.Command("pdftotext", "-layout", target, "-").CombinedOutput()
		if err != nil {
			t.Fatalf("extracting manuscript PDF: %v: %s", err, raw)
		}
		text := string(raw)
		for _, marker := range []string{"MANUHEAD Shared Book", "MANUFOOT", "Chapter One", "Chapter Two", "[unknown]"} {
			if !strings.Contains(text, marker) {
				t.Errorf("legacy=%v PDF missing %q: %s", legacy, marker, text)
			}
		}
		if legacy && text != baseline {
			t.Fatal("moving running-matter settings to shared paths changed manuscript PDF text")
		}
		baseline = text
	}
}
