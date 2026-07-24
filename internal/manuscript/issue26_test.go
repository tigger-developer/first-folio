// ABOUTME: Regression tests for issue #26 title-page footer placement.
// ABOUTME: Compiles British manuscripts and inspects their observable PDF geometry and text.
package manuscript

import (
	"encoding/xml"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const pointsPerMillimetre = 72.0 / 25.4

type pdfBBoxDocument struct {
	Pages []pdfBBoxPage `xml:"body>doc>page"`
}

type pdfBBoxPage struct {
	Height float64       `xml:"height,attr"`
	Words  []pdfBBoxWord `xml:"word"`
}

type pdfBBoxWord struct {
	YMax float64 `xml:"yMax,attr"`
	Text string  `xml:",chardata"`
}

// RT-26.1: the reported 13mm margin keeps title metadata within KDP's 6.4mm boundary.
func TestRT_26_1_TitleFooterClearsPrintableBoundary(t *testing.T) {
	output := renderIssue26PDF(t, "")
	bounds := commandOutput(t, exec.Command("pdftotext", "-f", "1", "-l", "1", "-bbox", output, "-"))
	page := firstPDFBBoxPage(t, bounds)

	maxY := 0.0
	for _, word := range page.Words {
		if word.YMax > maxY {
			maxY = word.YMax
		}
	}
	if maxY == 0 {
		t.Fatal("title page contained no extractable text")
	}

	clearance := page.Height - maxY
	minimum := 6.4 * pointsPerMillimetre
	if clearance < minimum {
		t.Fatalf("title-page text clearance = %.2fpt (%.2fmm), want at least %.2fpt (6.4mm)", clearance, clearance/pointsPerMillimetre, minimum)
	}
}

// RT-26.2: the corrected PDF retains every configured footer metadata value.
func TestRT_26_2_TitleFooterRetainsMetadata(t *testing.T) {
	output := renderIssue26PDF(t, "")
	text := commandOutput(t, exec.Command("pdftotext", "-f", "1", "-l", "1", "-layout", output, "-"))

	assertContains(t, text, "Draft 4")
	assertContains(t, text, "about 90,000 words")
	assertContains(t, text, "6 July 2026")
}

// RT-26.3: an individually aligned item appears once rather than in both footer paths.
func TestRT_26_3_AlignedTitleMetadataIsNotDuplicated(t *testing.T) {
	output := renderIssue26PDF(t, strings.Join([]string{
		"      date:",
		"        align: bottom-right",
	}, "\n"))
	text := commandOutput(t, exec.Command("pdftotext", "-f", "1", "-l", "1", "-layout", output, "-"))

	if count := strings.Count(text, "6 July 2026"); count != 1 {
		t.Fatalf("title-page date appeared %d times, want once:\n%s", count, text)
	}
}

func renderIssue26PDF(t *testing.T, titlePageConfig string) string {
	t.Helper()
	requireTool(t, "typst")
	requireTool(t, "pdftotext")
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "script.yaml"), strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    page: 6x9in",
		"    margin: 13mm",
		"    title-page:",
		"      enabled: true",
		"      include-version: true",
		"      include-wordcount: true",
		"      include-date: true",
		titlePageConfig,
		"",
	}, "\n"))
	writeFile(t, filepath.Join(dir, "chapter.md"), issue26Markdown())

	output := filepath.Join(dir, "manuscript.pdf")
	runManuscriptDirect(t, filepath.Join(dir, "chapter.md"), output)
	return output
}

func firstPDFBBoxPage(t *testing.T, content string) pdfBBoxPage {
	t.Helper()
	var document pdfBBoxDocument
	if err := xml.Unmarshal([]byte(content), &document); err != nil {
		t.Fatalf("parsing PDF bounding boxes: %v", err)
	}
	if len(document.Pages) != 1 {
		t.Fatalf("bounding-box output contained %d pages, want 1", len(document.Pages))
	}
	return document.Pages[0]
}

func issue26Markdown() string {
	return strings.Join([]string{
		"---",
		"title: Boundary House",
		"author: Example Author",
		"date: 2026-07-06",
		"version: Draft 4",
		"wordcount: about 90,000 words",
		"---",
		"",
		"# PART ONE",
		"",
		"## Chapter 1",
		"",
		"The first page began inside the printable boundary.",
	}, "\n")
}
