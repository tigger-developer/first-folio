// ABOUTME: Regression tests for the manuscript total-pages placeholder.
// ABOUTME: Verifies header and footer totals against public PDF output.
package manuscript_test

import (
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// RT-29.1: page headers resolve total-pages to the physical PDF page count.
func TestRT_29_1_HeaderResolvesTotalPages(t *testing.T) {
	dir, pdf := renderTotalPagesPDF(t, "HEADER [total-pages]", "", false)
	total := pdfPageCount(t, pdf)
	text := pdfText(t, dir, pdf)

	assertContains(t, text, "HEADER "+strconv.Itoa(total))
}

// RT-29.2: combined header format retains current and total page values.
func TestRT_29_2_HeaderCombinesCurrentAndTotalPages(t *testing.T) {
	dir, pdf := renderTotalPagesPDF(t, "HEADER [page]/[total-pages]", "", false)
	total := pdfPageCount(t, pdf)
	text := pdfText(t, dir, pdf)
	pattern := regexp.MustCompile(`HEADER [0-9]+/` + strconv.Itoa(total))

	if !pattern.MatchString(text) {
		t.Fatalf("header does not contain current/total page format:\n%s", text)
	}
}

// RT-29.3: page footers resolve total-pages to the physical PDF page count.
func TestRT_29_3_FooterResolvesTotalPages(t *testing.T) {
	dir, pdf := renderTotalPagesPDF(t, "", "FOOTER [total-pages]", false)
	total := pdfPageCount(t, pdf)
	text := pdfText(t, dir, pdf)

	assertContains(t, text, "FOOTER "+strconv.Itoa(total))
}

// RT-29.4: physical totals include intentional blanks despite body page reset.
func TestRT_29_4_TotalIncludesBlankAndIgnoresDisplayReset(t *testing.T) {
	_, basePDF := renderTotalPagesPDF(t, "", "FOOTER [page]/[total-pages]", false)
	blankDir, blankPDF := renderTotalPagesPDF(t, "", "FOOTER [page]/[total-pages]", true)
	baseTotal := pdfPageCount(t, basePDF)
	blankTotal := pdfPageCount(t, blankPDF)
	blankText := pdfText(t, blankDir, blankPDF)

	if blankTotal != baseTotal+1 {
		t.Fatalf("blank-page PDF has %d pages, want %d", blankTotal, baseTotal+1)
	}
	assertContains(t, blankText, "FOOTER 1/"+strconv.Itoa(blankTotal))
}

// RT-29.5: existing and unknown placeholders retain their public behaviour.
func TestRT_29_5_ExistingAndUnknownPlaceholdersRemainStable(t *testing.T) {
	dir, pdf := renderTotalPagesPDF(t, "META [title] | [chapter] | [unknown] | [page]", "", false)
	text := pdfText(t, dir, pdf)

	assertContains(t, text, "META Total Pages Test")
	assertContains(t, text, "Chapter One")
	assertContains(t, text, "[unknown]")
}

// RT-29.6: a standalone total-pages placeholder renders without decoration.
func TestRT_29_6_StandaloneTotalPagesHasNoSeparatorArtefact(t *testing.T) {
	dir, pdf := renderTotalPagesPDF(t, "", "[total-pages]", false)
	total := pdfPageCount(t, pdf)
	text := pdfText(t, dir, pdf)
	line := regexp.MustCompile(`(?m)^\s*` + strconv.Itoa(total) + `\s*$`)

	if !line.MatchString(text) {
		t.Fatalf("standalone total page count %d not found:\n%s", total, text)
	}
}

func renderTotalPagesPDF(t *testing.T, header string, footer string, blankBefore bool) (string, string) {
	t.Helper()
	requirePDFTools(t)
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "chapter.md"), totalPagesManuscript())
	writeTestFile(t, filepath.Join(dir, "script.yaml"), totalPagesConfig(header, footer, blankBefore))
	output := filepath.Join(dir, "manuscript.pdf")
	runFolio(t, "manuscript", filepath.Join(dir, "chapter.md"), output)
	return dir, output
}

func totalPagesManuscript() string {
	return strings.Join([]string{
		"---",
		"title: Total Pages Test",
		"author: Example Author",
		"---",
		"",
		"## Chapter One",
		"",
		"A paragraph long enough to establish the manuscript body.",
	}, "\n")
}

func totalPagesConfig(header string, footer string, blankBefore bool) string {
	lines := []string{
		"folio:",
		"  manuscript:",
		"    toc:",
		"      enabled: false",
		"    page-header:",
		"      format: " + strconv.Quote(header),
		"    page-footer:",
		"      format: " + strconv.Quote(footer),
		"    chapter:",
		"      blank-page-before: " + strconv.FormatBool(blankBefore),
		"",
	}
	return strings.Join(lines, "\n")
}

func requirePDFTools(t *testing.T) {
	t.Helper()
	for _, name := range []string{"typst", "pdfinfo", "pdftotext"} {
		if _, err := exec.LookPath(name); err != nil {
			t.Skipf("%s is not installed", name)
		}
	}
}

func pdfPageCount(t *testing.T, pdf string) int {
	t.Helper()
	output := commandText(t, exec.Command("pdfinfo", pdf))
	for _, line := range strings.Split(output, "\n") {
		if !strings.HasPrefix(line, "Pages:") {
			continue
		}
		count, err := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(line, "Pages:")))
		if err != nil {
			t.Fatalf("parsing PDF page count: %v", err)
		}
		return count
	}
	t.Fatalf("pdfinfo did not report a page count:\n%s", output)
	return 0
}

func pdfText(t *testing.T, dir string, pdf string) string {
	t.Helper()
	cmd := exec.Command("pdftotext", "-layout", pdf, "-")
	cmd.Dir = dir
	return commandText(t, cmd)
}

func commandText(t *testing.T, cmd *exec.Cmd) string {
	t.Helper()
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s failed: %v\n%s", cmd.String(), err, output)
	}
	return string(output)
}
