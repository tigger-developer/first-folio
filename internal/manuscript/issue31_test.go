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

// RT-31.1: the default reserved band aligns first entries across TOC pages.
func TestRT_31_1_DefaultTOCContinuationPaddingAlignsPages(t *testing.T) {
	requirePDFTools(t)
	_, pdf := renderContinuationTOC(t, "")
	firstPage, continuation := tocEntryStarts(t, pdf)

	t.Logf("measured first-entry positions: page one=%.2fpt, continuation=%.2fpt", firstPage, continuation)
	if math.Abs(firstPage-continuation) > 1 {
		t.Fatalf("TOC first entries are not aligned: page one %.2fpt, continuation %.2fpt", firstPage, continuation)
	}
}

// RT-31.2: an omitted key uses the British base preset's reserved band.
func TestRT_31_2_TOCContinuationPaddingDefaultsToFifteenMillimetres(t *testing.T) {
	dir, _ := renderContinuationTOC(t, "")
	typst := filepath.Join(dir, "manuscript.typ")

	runFolio(t, "manuscript", filepath.Join(dir, "manuscript.md"), typst)

	assertContains(t, readTestFile(t, typst), "#let folio-toc-continuation-padding = 15mm")
}

// RT-31.3: overriding the reserved band moves both TOC page starts equally.
func TestRT_31_3_TOCContinuationPaddingOverrideMovesBothPages(t *testing.T) {
	requirePDFTools(t)
	_, defaultPDF := renderContinuationTOC(t, "")
	_, customPDF := renderContinuationTOC(t, "20mm")
	defaultFirst, defaultContinuation := tocEntryStarts(t, defaultPDF)
	customFirst, customContinuation := tocEntryStarts(t, customPDF)
	firstShift := customFirst - defaultFirst
	continuationShift := customContinuation - defaultContinuation

	t.Logf("measured 5mm override shifts: page one=%.2fpt, continuation=%.2fpt", firstShift, continuationShift)
	if math.Abs(firstShift-14.17) > 1 || math.Abs(continuationShift-14.17) > 1 {
		t.Fatalf("20mm override shifts are %.2fpt and %.2fpt, want approximately 14.17pt", firstShift, continuationShift)
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
