// ABOUTME: Regression tests for manuscript body and TOC line-spacing configuration.
// ABOUTME: Exercises public command output and measures rendered TOC entry positions.
package manuscript_test

import (
	"encoding/xml"
	"math"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// RT-30.1: a unit-bearing body spacing value compiles without a duplicated unit.
func TestRT_30_1_BodyLineSpacingLengthCompiles(t *testing.T) {
	requirePDFTools(t)
	dir := spacingProject(t, "2em", "1em", false)
	typst := filepath.Join(dir, "manuscript.typ")
	pdf := filepath.Join(dir, "manuscript.pdf")

	runFolio(t, "manuscript", filepath.Join(dir, "manuscript.md"), typst)
	runFolio(t, "manuscript", filepath.Join(dir, "manuscript.md"), pdf)

	assertNotContains(t, readTestFile(t, typst), "emem")
	if !strings.HasPrefix(readTestFile(t, pdf), "%PDF-") {
		t.Fatal("manuscript output is not a PDF")
	}
}

// RT-30.2: numeric body spacing retains baseline-multiplier behaviour.
func TestRT_30_2_NumericBodyLineSpacingRemainsMultiplier(t *testing.T) {
	dir := spacingProject(t, "2", "1em", false)
	typst := filepath.Join(dir, "manuscript.typ")

	runFolio(t, "manuscript", filepath.Join(dir, "manuscript.md"), typst)

	assertContains(t, readTestFile(t, typst), "leading: 1em,")
}

// RT-30.3: TOC spacing changes the measured separation of one-line entries.
func TestRT_30_3_TOCLineSpacingControlsEntrySeparation(t *testing.T) {
	requirePDFTools(t)
	compactDir := spacingProject(t, "1.5", "1em", true)
	wideDir := spacingProject(t, "1.5", "2em", true)
	compactPDF := filepath.Join(compactDir, "manuscript.pdf")
	widePDF := filepath.Join(wideDir, "manuscript.pdf")

	runFolio(t, "manuscript", filepath.Join(compactDir, "manuscript.md"), compactPDF)
	runFolio(t, "manuscript", filepath.Join(wideDir, "manuscript.md"), widePDF)

	compactGap := averageTOCEntryGap(t, compactPDF)
	wideGap := averageTOCEntryGap(t, widePDF)
	t.Logf("measured TOC entry gaps: 1em=%.2fpt, 2em=%.2fpt", compactGap, wideGap)
	if math.Abs(compactGap-10) > 1 {
		t.Fatalf("TOC 1em entry gap is %.2fpt, want approximately 10pt", compactGap)
	}
	if math.Abs(wideGap-20) > 1 {
		t.Fatalf("TOC 2em entry gap is %.2fpt, want approximately 20pt", wideGap)
	}
}

func spacingProject(t *testing.T, bodySpacing string, tocSpacing string, tocEnabled bool) string {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "manuscript.md"), strings.Join([]string{
		"---",
		"title: Spacing Test",
		"author: Example Author",
		"---",
		"",
		"## Alpha",
		"",
		"First body paragraph.",
		"",
		"## Beta",
		"",
		"Second body paragraph.",
		"",
		"## Gamma",
		"",
		"Third body paragraph.",
		"",
		"## Delta",
		"",
		"Fourth body paragraph.",
	}, "\n"))
	writeTestFile(t, filepath.Join(dir, "script.yaml"), strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    line-spacing: " + strconv.Quote(bodySpacing),
		"    toc:",
		"      enabled: " + strconv.FormatBool(tocEnabled),
		"      font-size: 10pt",
		"      line-spacing: " + strconv.Quote(tocSpacing),
		"",
	}, "\n"))
	return dir
}

func averageTOCEntryGap(t *testing.T, pdf string) float64 {
	t.Helper()
	cmd := exec.Command("pdftotext", "-bbox", pdf, "-")
	data := commandText(t, cmd)
	var document struct {
		Pages []struct {
			Words []struct {
				YMin float64 `xml:"yMin,attr"`
				Text string  `xml:",chardata"`
			} `xml:"word"`
		} `xml:"body>doc>page"`
	}
	if err := xml.Unmarshal([]byte(data), &document); err != nil {
		t.Fatalf("parsing pdftotext bbox XML: %v", err)
	}
	wanted := map[string]bool{"Alpha": true, "Beta": true, "Gamma": true, "Delta": true}
	for _, page := range document.Pages {
		positions := make([]float64, 0, len(wanted))
		for _, word := range page.Words {
			if wanted[word.Text] {
				positions = append(positions, word.YMin)
			}
		}
		if len(positions) != len(wanted) {
			continue
		}
		total := 0.0
		for i := 1; i < len(positions); i++ {
			total += positions[i] - positions[i-1]
		}
		return total / float64(len(positions)-1)
	}
	t.Fatalf("could not find all TOC entries in one PDF page")
	return 0
}
