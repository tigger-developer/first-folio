// ABOUTME: Verifies manuscript dispatch through the single in-process Go application.
// ABOUTME: Covers help, rendering classification, configuration failure, and body justification.
package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestManuscriptDispatchInProcess(t *testing.T) {
	status, stdout, stderr := runApp(t, "manuscript", "--help")
	if status != 0 || stderr != "" || !strings.Contains(stdout, "Usage: folio manuscript") {
		t.Fatalf("help status %d\nstdout:%s\nstderr:%s", status, stdout, stderr)
	}

	dir := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	source := filepath.Join(dir, "chapter.md")
	writeAppFile(t, source, "---\ntitle: Test Manuscript\nauthor: Example Author\n---\n\n## Chapter 1\n\nBody.\n")
	status, stdout, stderr = runApp(t, "manuscript", "--dry-run", source, filepath.Join(dir, "out.pdf"))
	if status != 0 || stderr != "" {
		t.Fatalf("dry-run status %d\nstdout:%s\nstderr:%s", status, stdout, stderr)
	}
	for _, fragment := range []string{"format: markdown", "style: british", "page: a4"} {
		if !strings.Contains(stdout, fragment) {
			t.Errorf("dry-run missing %q:\n%s", fragment, stdout)
		}
	}
}

func TestManuscriptDispatchClassifiesCompleteLineCodeSpans(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	source := filepath.Join(dir, "chapter.md")
	target := filepath.Join(dir, "manuscript.typ")
	writeAppFile(t, source, strings.Join([]string{
		"## Chapter 1",
		"",
		"This is `inline_code` here.",
		"",
		"`echo from a complete line`",
		"",
		"`echo with trailing content` # remains inline",
		"",
	}, "\n"))

	status, stdout, stderr := runApp(t, "manuscript", source, target)
	if status != 0 {
		t.Fatalf("status %d\nstdout:%s\nstderr:%s", status, stdout, stderr)
	}
	typst := readAppFile(t, target)
	for _, fragment := range []string{
		"This is `inline_code` here.",
		"```\necho from a complete line\n```",
		"`echo with trailing content` \\# remains inline",
	} {
		if !strings.Contains(typst, fragment) {
			t.Errorf("generated Typst missing classification evidence %q:\n%s", fragment, typst)
		}
	}
}

func TestManuscriptDispatchRejectsInvalidBlockIndentBeforeOutput(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	source := filepath.Join(dir, "chapter.md")
	target := filepath.Join(dir, "manuscript.typ")
	writeAppFile(t, source, "## Chapter 1\n\n`echo from a complete line`\n")
	writeAppFile(t, filepath.Join(dir, "script.yaml"), "folio:\n  manuscript:\n    code-block-indent: inward\n")

	status, _, stderr := runApp(t, "manuscript", source, target)
	if status == 0 {
		t.Fatalf("invalid code-block indent returned status zero")
	}
	if !strings.Contains(stderr, "folio.manuscript.code-block-indent") {
		t.Fatalf("invalid code-block indent diagnostic omitted its path: %s", stderr)
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("invalid code-block indent wrote output %s", target)
	}
}

func TestRT_25_1_ManuscriptDefaultsToJustifiedBodyText(t *testing.T) {
	typst := renderJustificationManuscript(t, "")
	if !strings.Contains(typst, "justify: true") {
		t.Errorf("generated Typst does not justify body paragraphs by default:\n%s", typst)
	}
}

func TestRT_25_2_ManuscriptHonoursRaggedRightOverride(t *testing.T) {
	typst := renderJustificationManuscript(t, "folio:\n  manuscript:\n    justify: false\n")
	if !strings.Contains(typst, "justify: false") {
		t.Errorf("generated Typst does not apply the ragged-right override:\n%s", typst)
	}
}

func TestRT_25_3_ManuscriptJustificationIsScopedAfterFrontmatter(t *testing.T) {
	typst := renderJustificationManuscript(t, "")
	outline := strings.Index(typst, "#outline(title: none)")
	justify := strings.Index(typst, "justify: true")
	if outline < 0 || justify < 0 || justify < outline {
		t.Errorf("body justification must follow title-page and TOC composition: outline=%d justify=%d", outline, justify)
	}
	if strings.Count(typst, "justify:") != 1 {
		t.Errorf("generated Typst contains %d justification rules, want 1", strings.Count(typst, "justify:"))
	}
}

func renderJustificationManuscript(t *testing.T, config string) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	source := filepath.Join(dir, "chapter.md")
	target := filepath.Join(dir, "manuscript.typ")
	writeAppFile(t, source, "---\ntitle: Test Manuscript\nauthor: Example Author\n---\n\n## Chapter 1\n\nA paragraph long enough to expose its alignment in the rendered manuscript.\n")
	if config != "" {
		writeAppFile(t, filepath.Join(dir, "script.yaml"), config)
	}
	status, stdout, stderr := runApp(t, "manuscript", source, target)
	if status != 0 {
		t.Fatalf("status %d\nstdout:%s\nstderr:%s", status, stdout, stderr)
	}
	return readAppFile(t, target)
}
