// ABOUTME: Exercises the shared monospace migration warning through public commands.
// ABOUTME: Ensures the deprecated manuscript path reports on stderr without blocking output.
package app

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRT068_2LegacyMonospaceWarningUsesPublicCommandStderr(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	for _, command := range []string{"convert", "letter", "manuscript"} {
		t.Run(command, func(t *testing.T) {
			dir := t.TempDir()
			writeAppFile(t, filepath.Join(dir, "script.yaml"), "folio:\n  manuscript:\n    mono:\n      font:\n        family: Iosevka Publish\n        stretch: 125%\n")
			var args []string
			if command == "convert" {
				source := filepath.Join(dir, "play.org")
				writeAppFile(t, source, "* ACT ONE\n**** CÁIT\nLiteral =CODE=.\n")
				args = []string{"convert", source, filepath.Join(dir, "play.typ")}
			} else if command == "letter" {
				if _, err := exec.LookPath("typst"); err != nil {
					t.Skip("requires typst")
				}
				source := filepath.Join(dir, "letters.org")
				writeAppFile(t, source, "* Letters :letter:\n** Subject :subject:\nLiteral =CODE=.\n*** Recipient :to:\n**** Theatre :org:\n")
				args = []string{"letter", source, "--dir", dir}
			} else {
				source := filepath.Join(dir, "manuscript.md")
				writeAppFile(t, source, "---\ntitle: Manuscript\n---\n\n## Chapter One\n\nLiteral `CODE`.\n")
				args = []string{"manuscript", source, filepath.Join(dir, "manuscript.typ")}
			}
			status, stdout, stderr := runApp(t, args...)
			if status != 0 {
				t.Fatalf("status %d; stdout=%q stderr=%q", status, stdout, stderr)
			}
			if strings.Count(stderr, "folio.manuscript.mono.font is deprecated") != 1 || !strings.Contains(stderr, "folio.mono.font") {
				t.Fatalf("migration warning = %q", stderr)
			}
			if strings.Contains(stdout, "deprecated") {
				t.Fatalf("warning leaked to stdout: %q", stdout)
			}
		})
	}
}
