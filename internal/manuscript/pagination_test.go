// ABOUTME: Regression tests for manuscript widow and orphan pagination control.
// ABOUTME: Verifies the YAML switch maps directly to Typst text pagination costs.

package manuscript

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestWidowOrphanControlDefaultsEnabled(t *testing.T) {
	typst := renderIssue15Manuscript(t, "")
	if !strings.Contains(typst, "costs: (widow: 100%, orphan: 100%)") {
		t.Fatalf("expected widow/orphan control to be enabled by default")
	}
}

func TestWidowOrphanControlCanBeDisabled(t *testing.T) {
	config := "folio:\n  manuscript:\n    widow-orphan-control: false\n"
	typst := renderIssue15Manuscript(t, config)
	if !strings.Contains(typst, "costs: (widow: 0%, orphan: 0%)") {
		t.Fatalf("expected disabled widow/orphan control to emit zero pagination costs")
	}

	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "script.yaml"), config)
	writeFile(t, filepath.Join(dir, "chapter.md"), markdownChapterOne())
	runManuscriptDirect(t, filepath.Join(dir, "chapter.md"), filepath.Join(dir, "manuscript.pdf"))
}
