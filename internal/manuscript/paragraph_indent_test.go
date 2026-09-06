// ABOUTME: Regression tests for manuscript first-line indentation around structural blocks.
// ABOUTME: Keeps only each chapter-opening paragraph flush while indenting prose after code and quotes.
package manuscript

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestOnlyChapterOpeningParagraphGetsFlushOverride(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "chapter.md"), strings.Join([]string{
		"## Chapter 1",
		"",
		"Opening prose is flush left.",
		"",
		"`BLOCK CODE`",
		"",
		"Prose after code is indented.",
		"",
		"> Quoted text.",
		"",
		"Prose after a quote is indented.",
		"",
		"## Chapter 2",
		"",
		"The next chapter also begins flush left.",
		"",
		"Its second paragraph is indented.",
		"",
		"## Chapter 3",
		"",
		"`CHAPTER OPENS WITH CODE`",
		"",
		"Prose after chapter-opening code is indented.",
		"",
		"## Chapter 4",
		"",
		"> Chapter-opening quoted text.",
		"",
		"Prose after a chapter-opening quote is indented.",
		"",
		"## Chapter 5",
		"",
		"```{=typst}",
		"#grid(columns: 2)[A][B]",
		"```",
		"",
		"Prose after arbitrary chapter-opening Typst is indented.",
		"",
	}, "\n"))

	output := filepath.Join(dir, "manuscript.typ")
	runManuscriptDirect(t, filepath.Join(dir, "chapter.md"), output)
	typst := readFile(t, output)
	normalized := strings.Join(strings.Fields(typst), " ")

	assertContains(t, typst, "first-line-indent: (amount: 10mm, all: true),")
	for _, opening := range []string{
		"Opening prose is flush left.",
		"The next chapter also begins flush left.",
	} {
		assertContains(t, normalized, "#par(first-line-indent: 0pt)[ "+opening+" ]")
	}
	if got := strings.Count(typst, "#par(first-line-indent: 0pt)["); got != 2 {
		t.Fatalf("chapter-opening flush overrides = %d, want 2", got)
	}
	for _, indented := range []string{
		"Prose after code is indented.",
		"Prose after a quote is indented.",
		"Its second paragraph is indented.",
		"Prose after chapter-opening code is indented.",
		"Prose after a chapter-opening quote is indented.",
		"Prose after arbitrary chapter-opening Typst is indented.",
	} {
		if strings.Contains(typst, "#par(first-line-indent: 0pt)[\n"+indented) {
			t.Fatalf("non-opening paragraph received a flush override: %q", indented)
		}
	}
}

func TestChapterOpeningFootnoteDoesNotMakeLaterProseFlush(t *testing.T) {
	var cfg Config
	normalizeConfig(&cfg)
	body, err := renderBlocks([]Block{
		{Kind: "chapter", Text: "Chapter 1", Name: "Chapter 1", Number: 1},
		{Kind: "footnote", Text: "Opening structural note."},
		{Kind: "paragraph", Text: "Prose after structural content is indented."},
	}, cfg)
	if err != nil {
		t.Fatalf("rendering blocks: %v", err)
	}
	if strings.Contains(body, "#par(first-line-indent: 0pt)[") {
		t.Fatalf("prose after a chapter-opening footnote received a flush override:\n%s", body)
	}
}
