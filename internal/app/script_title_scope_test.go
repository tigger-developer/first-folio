// ABOUTME: Protects manuscript independence from the script compact-title switch.
// ABOUTME: Verifies root title-page settings do not alter manuscript title-page output.
package app

import (
	"path/filepath"
	"testing"
)

func TestScriptCompactTitleDoesNotChangeManuscript(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	source, target := filepath.Join(dir, "chapter.md"), filepath.Join(dir, "book.typ")
	writeAppFile(t, source, "---\ntitle: Independent manuscript\nauthor: Example Author\n---\n\n## Chapter One\n\nBody.\n")
	status, _, stderr := runApp(t, "manuscript", source, target)
	if status != 0 {
		t.Fatalf("baseline status %d: %s", status, stderr)
	}
	want := readAppFile(t, target)
	writeAppFile(t, filepath.Join(dir, "script.yaml"), "folio:\n  title-page:\n    enabled: false\n")
	status, _, stderr = runApp(t, "manuscript", source, target)
	if status != 0 {
		t.Fatalf("compact script setting status %d: %s", status, stderr)
	}
	if got := readAppFile(t, target); got != want {
		t.Fatal("script compact-title setting changed manuscript output")
	}
}
