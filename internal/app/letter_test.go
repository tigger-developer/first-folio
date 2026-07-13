// ABOUTME: Characterizes cover-letter generation through the public Go application.
// ABOUTME: Covers recipient filtering, output naming, PDF compilation, and diagnostics.
package app

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestLetterCommandGeneratesFilteredPDF(t *testing.T) {
	if _, err := exec.LookPath("typst"); err != nil {
		t.Skip("typst is not installed")
	}
	dir := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	source := filepath.Join(dir, "letters.org")
	writeAppFile(t, source, `#+AUTHOR: Example Author
* Letters :letter:
** Sender / Dublin :sender:
** Submission :subject:
Hello [org].
*** Recipient One / Cork :to:
**** One Theatre :org:
*** Recipient Two / Galway :to:
**** Two Theatre :org:
`)
	status, stdout, stderr := runApp(t, "letter", source, "--to", "Two", "--dir", dir, "--prefix", "submission")
	if status != 0 {
		t.Fatalf("status %d\nstdout:%s\nstderr:%s", status, stdout, stderr)
	}
	if !strings.Contains(stdout, "Generated:") {
		t.Fatalf("missing generated message: %s", stdout)
	}
	target := filepath.Join(dir, "submission-two-theatre.pdf")
	raw, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(raw), "%PDF") {
		t.Fatalf("invalid PDF: %d bytes", len(raw))
	}
}

func TestLetterCommandRejectsMissingSections(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	source := filepath.Join(dir, "plain.org")
	writeAppFile(t, source, "#+TITLE: No letters\n")
	status, _, stderr := runApp(t, "letter", source)
	if status == 0 || !strings.Contains(stderr, "no :letter: tagged sections") {
		t.Fatalf("status %d stderr %q", status, stderr)
	}
}
