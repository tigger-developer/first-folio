// ABOUTME: Regression tests for configurable padding on continued TOC pages.
// ABOUTME: Measures public PDF entry positions so page-one layout stays stable.
package manuscript_test

import (
	"encoding/xml"
	"fmt"
	"math"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// RT-31.1: configured continuation padding moves the first entry on page two.
func TestRT_31_1_TOCContinuationPaddingMovesLaterPage(t *testing.T) {
	requirePDFTools(t)
	_, unpaddedPDF := renderContinuationTOC(t, "")
	_, paddedPDF := renderContinuationTOC(t, "15mm")
	_, unpaddedContinuation := tocEntryStarts(t, unpaddedPDF)
	_, paddedContinuation := tocEntryStarts(t, paddedPDF)
	shift := paddedContinuation - unpaddedContinuation

	t.Logf("measured continuation shift: %.2fpt", shift)
	if math.Abs(shift-42.52) > 1 {
		t.Fatalf("15mm continuation padding shifted first entry by %.2fpt, want approximately 42.52pt", shift)
	}
}

// RT-31.2: an omitted key retains the zero-padding default in generated output.
func TestRT_31_2_TOCContinuationPaddingDefaultsToZero(t *testing.T) {
	dir, _ := renderContinuationTOC(t, "")
	typst := filepath.Join(dir, "manuscript.typ")

	runFolio(t, "manuscript", filepath.Join(dir, "manuscript.md"), typst)

	assertContains(t, readTestFile(t, typst), "#let folio-toc-continuation-padding = 0mm")
}

// RT-31.3: continuation padding does not move the first entry on TOC page one.
func TestRT_31_3_TOCContinuationPaddingLeavesFirstPageUnchanged(t *testing.T) {
	requirePDFTools(t)
	_, unpaddedPDF := renderContinuationTOC(t, "")
	_, paddedPDF := renderContinuationTOC(t, "15mm")
	unpaddedFirst, _ := tocEntryStarts(t, unpaddedPDF)
	paddedFirst, _ := tocEntryStarts(t, paddedPDF)

	if math.Abs(paddedFirst-unpaddedFirst) > 0.5 {
		t.Fatalf("continuation padding moved first TOC entry from %.2fpt to %.2fpt", unpaddedFirst, paddedFirst)
	}
}

func renderContinuationTOC(t *testing.T, padding string) (string, string) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	var manuscript strings.Builder
	manuscript.WriteString("---\ntitle: Continued TOC Test\nauthor: Example Author\n---\n")
	for i := 1; i <= 70; i++ {
		fmt.Fprintf(&manuscript, "\n## Entry%02d\n\nBody text.\n", i)
	}
	writeTestFile(t, filepath.Join(dir, "manuscript.md"), manuscript.String())
	config := []string{
		"folio:",
		"  manuscript:",
		"    toc:",
		"      font-size: 10pt",
		"      line-spacing: 1.15em",
	}
	if padding != "" {
		config = append(config, "      continuation-padding-before: "+padding)
	}
	writeTestFile(t, filepath.Join(dir, "script.yaml"), strings.Join(config, "\n")+"\n")
	pdf := filepath.Join(dir, "manuscript.pdf")
	runFolio(t, "manuscript", filepath.Join(dir, "manuscript.md"), pdf)
	return dir, pdf
}

func tocEntryStarts(t *testing.T, pdf string) (float64, float64) {
	t.Helper()
	data := commandText(t, exec.Command("pdftotext", "-bbox", pdf, "-"))
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
	entry := regexp.MustCompile(`^Entry[0-9]{2}$`)
	starts := make([]float64, 0, 2)
	for _, page := range document.Pages {
		first := 0.0
		for _, word := range page.Words {
			if entry.MatchString(word.Text) && (first == 0 || word.YMin < first) {
				first = word.YMin
			}
		}
		if first > 0 {
			starts = append(starts, first)
			if len(starts) == 2 {
				return starts[0], starts[1]
			}
		}
	}
	t.Fatalf("could not find TOC entries on two PDF pages")
	return 0, 0
}
