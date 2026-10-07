// ABOUTME: Verifies canonical manuscript running matter and migration diagnostics.
// ABOUTME: Exercises rendered output and injected warning streams without compiling PDFs.
package manuscript

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestManuscriptSharedRunningMatterUsesInjectedWarnings(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	input, output := filepath.Join(dir, "input.md"), filepath.Join(dir, "output.typ")
	writeFile(t, input, "---\ntitle: Shared title\n---\n# Chapter\n\nBody.\n")
	writeFile(t, filepath.Join(dir, "script.yaml"), `folio:
  page-header:
    enabled: true
    format: Shared [unknown-token]
    font:
      family: Shared Family
  manuscript:
    page-header:
      font:
        style: italic
    page-footer:
      format: Legacy footer [page]
`)
	var stdout, stderr bytes.Buffer
	if err := RunWithStreams([]string{input, output}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("warning leaked to stdout: %s", &stdout)
	}
	for _, role := range []string{"page-header", "page-footer"} {
		if strings.Count(stderr.String(), "folio.manuscript."+role+" is deprecated") != 1 {
			t.Fatalf("warnings = %s", &stderr)
		}
	}
	cfg, err := LoadConfig(dir, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Warnings) != 2 {
		t.Fatalf("config warnings = %v", cfg.Warnings)
	}
	rendered := readFile(t, output)
	assertContains(t, rendered, `Shared \[unknown-token\]`)
	assertContains(t, rendered, `font: "Shared Family", size: 10pt, weight: "regular", stretch: 100%, style: "italic"`)
	assertContains(t, rendered, "Legacy footer")
	stdout.Reset()
	stderr.Reset()
	if err := RunWithStreams([]string{input, output, "--dry-run"}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(stdout.String(), "deprecated") {
		t.Fatal("warnings leaked to dry-run stdout")
	}
	if strings.Count(stderr.String(), " is deprecated") != 2 {
		t.Fatalf("dry-run warnings = %s", &stderr)
	}
}
