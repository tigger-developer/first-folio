// ABOUTME: Verifies manuscript YAML cascade warnings through the command boundary.
// ABOUTME: Checks per-file diagnostics, stream separation, and supported outputs.
package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRT064_1ReportsEachLoadedYAMLFileOnceInTheCascade(t *testing.T) {
	dir := t.TempDir()
	home := t.TempDir()
	t.Setenv("HOME", home)

	globalDir := filepath.Join(home, ".config", "first-folio")
	localDir := dir
	globalPath := filepath.Join(globalDir, "script.yaml")
	globalStylePath := filepath.Join(globalDir, "script-us.yaml")
	localPath := filepath.Join(localDir, "script.yaml")
	localStylePath := filepath.Join(localDir, "script-us.yaml")
	writeAppFile(t, globalPath, manuscriptMarginYAML("20mm"))
	writeAppFile(t, globalStylePath, manuscriptMarginYAML("21mm"))
	writeAppFile(t, localPath, manuscriptMarginYAML("22mm"))
	writeAppFile(t, localStylePath, manuscriptMarginYAML("23mm"))
	writeAppFile(t, filepath.Join(localDir, "script-screenplay.yaml"), manuscriptMarginYAML("99mm"))

	source := filepath.Join(localDir, "chapter.md")
	writeAppFile(t, source, "---\ntitle: Cascade Example\nfolio:\n  manuscript:\n    margin: 19mm\n---\n\n# Chapter One\n\nBody text.\n")
	status, stdout, stderr := runApp(t, "manuscript", "--dry-run", "--style", "us", source, filepath.Join(dir, "out.typ"))
	if status != 0 {
		t.Fatalf("manuscript dry-run status %d\nstdout:\n%s\nstderr:\n%s", status, stdout, stderr)
	}
	assertLoadedYAMLWarnings(t, stderr, []string{globalPath, globalStylePath, localPath, localStylePath})
	if !strings.Contains(stdout, "margin: 23mm") {
		t.Fatalf("dry-run did not preserve the existing winning margin:\n%s", stdout)
	}
	if strings.Contains(stdout, "Warning:") {
		t.Fatalf("load warning leaked to stdout:\n%s", stdout)
	}
	for _, setting := range []string{"folio.manuscript.margin", "19mm", "20mm", "21mm", "22mm", "23mm"} {
		if strings.Contains(stderr, setting) {
			t.Errorf("load warning disclosed setting %q: %q", setting, stderr)
		}
	}
	if strings.Contains(stderr, "script-screenplay.yaml") {
		t.Fatalf("unselected style-specific YAML was reported: %q", stderr)
	}

	play := filepath.Join(localDir, "play.org")
	writeAppFile(t, play, "#+TITLE: Example\n* ACT ONE\n**** CÁIT\nHello.\n")
	status, _, stderr = runApp(t, "convert", play, filepath.Join(dir, "play.md"))
	if status != 0 {
		t.Fatalf("non-manuscript conversion status %d: %s", status, stderr)
	}
	if strings.Contains(stderr, "manuscript configuration cascade") {
		t.Fatalf("non-manuscript command emitted a cascade warning: %q", stderr)
	}
}

func TestRT064_2ReportsLoadedYAMLOnEveryManuscriptOutputRoute(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	source := filepath.Join(dir, "chapter.md")
	configPath := filepath.Join(dir, "script.yaml")
	writeAppFile(t, source, "---\ntitle: Cascade Example\n---\n\n# Chapter One\n\nBody text.\n")
	writeAppFile(t, configPath, manuscriptMarginYAML("24mm"))

	for _, extension := range []string{"typ", "pdf", "docx", "odt"} {
		t.Run(extension, func(t *testing.T) {
			target := filepath.Join(dir, "out."+extension)
			status, stdout, stderr := runApp(t, "manuscript", source, target)
			if status != 0 {
				t.Fatalf("manuscript %s status %d\nstdout:\n%s\nstderr:\n%s", extension, status, stdout, stderr)
			}
			if strings.Contains(stdout, "Warning:") {
				t.Fatalf("load warning leaked to stdout:\n%s", stdout)
			}
			info, err := os.Stat(target)
			if err != nil {
				t.Fatalf("expected %s output: %v", extension, err)
			}
			if info.Size() == 0 {
				t.Fatalf("generated %s output is empty", extension)
			}
			assertLoadedYAMLWarnings(t, stderr, []string{configPath})
		})
	}
}

func manuscriptMarginYAML(value string) string {
	return "folio:\n  manuscript:\n    margin: " + value + "\n"
}

func assertLoadedYAMLWarnings(t *testing.T, stderr string, paths []string) {
	t.Helper()
	var lines []string
	for _, line := range strings.Split(stderr, "\n") {
		if strings.Contains(line, "into manuscript configuration cascade") {
			lines = append(lines, line)
		}
	}
	if len(lines) != len(paths) {
		t.Fatalf("got %d stderr lines, want one per loaded YAML file (%d): %q", len(lines), len(paths), stderr)
	}
	for i, path := range paths {
		want := fmt.Sprintf("Warning: loaded YAML file %q into manuscript configuration cascade.", path)
		if lines[i] != want {
			t.Errorf("stderr line %d = %q, want %q", i+1, lines[i], want)
		}
	}
}
