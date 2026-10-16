// ABOUTME: Checks prose office input ordering, source boundaries, and prepublication errors.
// ABOUTME: Exercises public commands with independent source and destination sentinels.
package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRT063_3_OfficeSortedDeduplicatedInputs(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		for _, sourceFormat := range []string{"md", "org"} {
			t.Run(format+"/"+sourceFormat, func(t *testing.T) {
				t.Setenv("HOME", t.TempDir())
				dir := t.TempDir()
				first := proseSource(t, dir, "01-first."+sourceFormat, "FIRST-INPUT-SENTINEL\n")
				last := proseSource(t, dir, "02-last."+sourceFormat, "LAST-INPUT-SENTINEL\n")
				target := filepath.Join(dir, "output."+format)
				status, _, stderr := runApp(t, "manuscript", last, filepath.Join(dir, "*."+sourceFormat), first, target)
				if status != 0 {
					t.Fatalf("manuscript export: %s", stderr)
				}
				pkg := prosePackage(t, target, format)
				text := officeBody(t, pkg, format).text()
				for _, marker := range []string{"FIRST-INPUT-SENTINEL", "LAST-INPUT-SENTINEL"} {
					if strings.Count(text, marker) != 1 {
						t.Errorf("expected one copy of %s", marker)
					}
				}
				if strings.Index(text, "FIRST-INPUT-SENTINEL") > strings.Index(text, "LAST-INPUT-SENTINEL") {
					t.Error("source order was not lexical after glob expansion and deduplication")
				}
			})
		}
	}
}

func TestRT063_3_OfficeRejectsMixedInputFormats(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		t.Run(format, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			dir := t.TempDir()
			md := proseSource(t, dir, "first.md", "FIRST-INPUT\n")
			org := proseSource(t, dir, "second.org", "SECOND-INPUT\n")
			target := proseSource(t, dir, "output."+format, "EXISTING-DESTINATION")
			status, _, stderr := runApp(t, "manuscript", md, org, target)
			if status == 0 || !strings.Contains(stderr, "must not mix") {
				t.Errorf("mixed-format diagnostic missing: status=%d stderr=%s", status, stderr)
			}
			raw, err := os.ReadFile(target)
			if err != nil || string(raw) != "EXISTING-DESTINATION" {
				t.Errorf("mixed-format failure altered destination: %v", err)
			}
		})
	}
}

func TestRT063_4_OfficeRejectsLaterSourceConfiguration(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		t.Run(format, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			dir := t.TempDir()
			first := proseSource(t, dir, "01-first.md", "FIRST-INPUT\n")
			last := proseSource(t, dir, "02-last.md", "---\nfolio:\n  manuscript:\n    font:\n      size: 17pt\n---\nLAST-INPUT\n")
			target := proseSource(t, dir, "output."+format, "EXISTING-DESTINATION")
			status, _, stderr := runApp(t, "manuscript", first, last, target)
			if status == 0 || !strings.Contains(stderr, last) || !strings.Contains(stderr, "first resolved manuscript input") {
				t.Errorf("later configuration diagnostic missing: status=%d stderr=%s", status, stderr)
			}
			raw, err := os.ReadFile(target)
			if err != nil || string(raw) != "EXISTING-DESTINATION" {
				t.Errorf("configuration failure altered destination: %v", err)
			}
		})
	}
}
