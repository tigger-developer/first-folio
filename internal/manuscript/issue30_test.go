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

// RT-30.1 (length acceptance) and RT-30.2 (font-size-based intervals) are retired
// by W048 - Font-relative line spacing. RT048.1 and RT048.3 now verify native
// font intervals and rejection of lengths through the executable and PDF output.

// RT-30.3: TOC spacing changes the measured separation of one-line entries.
func TestRT_30_3_TOCLineSpacingControlsEntrySeparation(t *testing.T) {
	requirePDFTools(t)
	compactDir := spacingProject(t, "1.5", "1", true)
	wideDir := spacingProject(t, "1.5", "2", true)
	compactPDF := filepath.Join(compactDir, "manuscript.pdf")
	widePDF := filepath.Join(wideDir, "manuscript.pdf")

	runFolio(t, "manuscript", filepath.Join(compactDir, "manuscript.md"), compactPDF)
	runFolio(t, "manuscript", filepath.Join(wideDir, "manuscript.md"), widePDF)

	compactGap := averageTOCEntryGap(t, compactPDF)
	wideGap := averageTOCEntryGap(t, widePDF)
	t.Logf("measured TOC entry gaps: 1x=%.2fpt, 2x=%.2fpt", compactGap, wideGap)
	if compactGap <= 0 || math.Abs(wideGap-2*compactGap) > 0.03 {
		t.Fatalf("TOC 2x gap %.3fpt, want twice 1x gap %.3fpt", wideGap, compactGap)
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
		"      font:",
		"        size: 10pt",
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
