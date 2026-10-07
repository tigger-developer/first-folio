// ABOUTME: Exercises manuscript running-matter migration diagnostics through the public CLI.
// ABOUTME: Ensures deprecation warnings use the injected stderr stream, never stdout.
package app

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestManuscriptLegacyRunningMatterWarningUsesStderr(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	source := filepath.Join(dir, "chapter.md")
	writeAppFile(t, source, "---\ntitle: Migration\n---\n\n# Chapter One\n\nBody.\n")
	writeAppFile(t, filepath.Join(dir, "script.yaml"), "folio:\n  manuscript:\n    page-header:\n      format: Legacy header\n")
	status, stdout, stderr := runApp(t, "manuscript", source, filepath.Join(dir, "out.typ"))
	if status != 0 {
		t.Fatalf("status %d: %s", status, stderr)
	}
	if !strings.Contains(stderr, "deprecated") || !strings.Contains(stderr, "folio.manuscript.page-header") {
		t.Fatalf("missing migration warning on stderr: %q", stderr)
	}
	if strings.Contains(stdout, "deprecated") {
		t.Fatalf("migration warning leaked to stdout: %q", stdout)
	}
}

func TestLetterSharedRunningMatterProducesNoWarnings(t *testing.T) {
	if _, err := exec.LookPath("typst"); err != nil {
		t.Skip("typst is not installed")
	}
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	source := filepath.Join(dir, "letters.org")
	writeAppFile(t, source, "#+AUTHOR: Example Author\n* Letters :letter:\n** Sender / Dublin :sender:\n** Submission :subject:\nHello [org].\n*** Recipient / Cork :to:\n**** Theatre :org:\n")
	writeAppFile(t, filepath.Join(dir, "script.yaml"), "folio:\n  page-header:\n    enabled: true\n    format: Letter must ignore this\n  page-footer:\n    enabled: true\n  manuscript:\n    page-header:\n      format: Legacy manuscript header\n")
	status, _, stderr := runApp(t, "letter", source, "--dir", dir)
	if status != 0 || stderr != "" {
		t.Fatalf("status %d, stderr %q; letters must ignore running matter silently", status, stderr)
	}
}

func TestLetterIgnoresMalformedRunningMatter(t *testing.T) {
	if _, err := exec.LookPath("typst"); err != nil {
		t.Skip("typst is not installed")
	}
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	source := filepath.Join(dir, "letters.org")
	writeAppFile(t, source, "#+AUTHOR: Example Author\n* Letters :letter:\n** Sender / Dublin :sender:\n** Submission :subject:\nHello [org].\n*** Recipient / Cork :to:\n**** Theatre :org:\n")
	for _, block := range []string{
		"  page-header:\n    enabled: nope\n",
		"  page-footer:\n    align: diagonal\n",
		"  page-header:\n    font:\n      size: invalid\n",
		"  page-header:\n    font-size: retired\n",
		"  page-header: not-a-mapping\n",
		"  manuscript:\n    page-header: not-a-mapping\n",
	} {
		writeAppFile(t, filepath.Join(dir, "script.yaml"), "folio:\n"+block)
		status, _, stderr := runApp(t, "letter", source, "--dir", dir)
		if status != 0 || stderr != "" {
			t.Errorf("irrelevant block %q blocked letter: status %d stderr %q", block, status, stderr)
		}
	}
}
