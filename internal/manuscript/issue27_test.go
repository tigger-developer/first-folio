// ABOUTME: Regression tests for issue #27 manuscript TOC PDF output.
// ABOUTME: Verifies printable text, visible entry content, styling, and internal links.
package manuscript

import (
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// RT-27.1: extracted TOC text contains no word joiners rejected by print validation.
func TestRT_27_1_TOCContainsNoWordJoiners(t *testing.T) {
	output := renderIssue27PDF(t)
	text := extractIssue27PDFText(t, output)

	if count := strings.Count(text, "\u2060"); count != 0 {
		t.Fatalf("extracted PDF contained %d U+2060 WORD JOINER characters", count)
	}
}

// RT-27.2: the custom entry retains ordered headings, leaders, and page numbers.
func TestRT_27_2_TOCRetainsVisibleEntryContent(t *testing.T) {
	output := renderIssue27PDF(t)
	text := extractIssue27PDFText(t, output)

	assertBefore(t, text, "PART ONE", "Chapter 1")
	assertBefore(t, text, "Chapter 1", "Chapter 2")
	entryPattern := regexp.MustCompile(`PART ONE[ .]+[0-9]+`)
	if !entryPattern.MatchString(text) {
		t.Fatalf("part TOC entry did not retain dot leaders and a page number:\n%s", text)
	}
}

// RT-27.3: PDF conversion exposes bold part text and internal TOC links.
func TestRT_27_3_TOCPartIsBoldAndEntriesAreLinked(t *testing.T) {
	requireTool(t, "pdftohtml")
	output := renderIssue27PDF(t)
	xmlPath := filepath.Join(t.TempDir(), "toc.xml")
	commandOutput(t, exec.Command("pdftohtml", "-xml", "-hidden", output, xmlPath))
	xmlOutput := readFile(t, xmlPath)

	assertContains(t, xmlOutput, "<b>PART ONE</b>")
	assertContains(t, xmlOutput, "<a href=")
	assertContains(t, xmlOutput, ">Chapter 1")
}

func renderIssue27PDF(t *testing.T) string {
	t.Helper()
	requireTool(t, "typst")
	requireTool(t, "pdftotext")
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "chapter.md"), strings.Join([]string{
		"---",
		"title: Printable Contents",
		"author: Example Author",
		"date: 2026-07-06",
		"---",
		"",
		"# PART ONE",
		"",
		"## Chapter 1",
		"",
		"The first chapter supplies the first table-of-contents entry.",
		"",
		"## Chapter 2",
		"",
		"The second chapter supplies the next entry.",
	}, "\n"))

	output := filepath.Join(dir, "manuscript.pdf")
	runManuscriptDirect(t, filepath.Join(dir, "chapter.md"), output)
	return output
}

func extractIssue27PDFText(t *testing.T, output string) string {
	t.Helper()
	textPath := filepath.Join(t.TempDir(), "manuscript.txt")
	commandOutput(t, exec.Command("pdftotext", "-layout", output, textPath))
	return readFile(t, textPath)
}
