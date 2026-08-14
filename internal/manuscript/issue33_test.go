// ABOUTME: Regression tests for issue #33 -- disabled manuscript header margins.
// ABOUTME: Verifies body pages fall back to the manuscript top margin after the TOC.

package manuscript

import (
	"encoding/xml"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// RT-33.1: a disabled header explicitly restores the manuscript top margin on body pages.
func TestRT_33_1_DisabledHeaderRestoresBodyTopMargin(t *testing.T) {
	typst := renderIssue15Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    margin: 13mm",
		"    gutter: 5mm",
		"    toc:",
		"      continuation-padding-before: 15mm",
		"    page-header:",
		"      enabled: false",
		"      distance-from-edge: 30mm",
		"      content-padding-after: 10mm",
		"",
	}, "\n"))
	body := extractBodyPageBlock(t, typst)

	assertContains(t, body, "top: 13mm,")
	assertContains(t, body, "inside: 13mm + 5mm,")
	assertContains(t, body, "outside: 13mm,")
	assertNotContains(t, body, "top: 30mm + 10mm,")
}

// RT-33.2: positioning values under a disabled header do not alter body-page geometry.
func TestRT_33_2_DisabledHeaderPositioningIsInert(t *testing.T) {
	config := func(distance, padding string) string {
		return strings.Join([]string{
			"folio:",
			"  manuscript:",
			"    margin: 13mm",
			"    gutter: 5mm",
			"    page-header:",
			"      enabled: false",
			"      distance-from-edge: " + distance,
			"      content-padding-after: " + padding,
			"",
		}, "\n")
	}

	first := extractBodyPageBlock(t, renderIssue15Manuscript(t, config("30mm", "10mm")))
	second := extractBodyPageBlock(t, renderIssue15Manuscript(t, config("60mm", "25mm")))
	if first != second {
		t.Fatal("disabled page-header positioning changed generated body-page geometry")
	}
}

// RT-33.3: an enabled header continues to reserve its distance and content padding.
func TestRT_33_3_EnabledHeaderRetainsConfiguredTopMargin(t *testing.T) {
	typst := renderIssue15Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    margin: 13mm",
		"    gutter: 5mm",
		"    page-header:",
		"      enabled: true",
		"      distance-from-edge: 30mm",
		"      content-padding-after: 10mm",
		"",
	}, "\n"))
	body := extractBodyPageBlock(t, typst)

	assertContains(t, body, "top: 30mm + 10mm,")
}

// RT-33.4: compiled continuity pages visibly begin at the configured manuscript margin.
func TestRT_33_4_DisabledHeaderPDFUsesBodyTopMargin(t *testing.T) {
	requireTool(t, "typst")
	requireTool(t, "pdftotext")
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "script.yaml"), strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    page: 5.5x8.5in",
		"    margin: 13mm",
		"    gutter: 5mm",
		"    toc:",
		"      continuation-padding-before: 15mm",
		"    page-header:",
		"      enabled: false",
		"      distance-from-edge: 30mm",
		"      content-padding-after: 10mm",
		"",
	}, "\n"))

	var manuscript strings.Builder
	manuscript.WriteString("---\ntitle: Margin Test\nauthor: Example Author\n---\n\n# PART ONE\n\n## Chapter 1\n")
	for i := 0; i < 180; i++ {
		fmt.Fprintf(&manuscript, "\nContinuityMarker paragraph %d fills the page with observable manuscript body text.\n", i)
	}
	writeFile(t, filepath.Join(dir, "manuscript.md"), manuscript.String())

	output := filepath.Join(dir, "manuscript.pdf")
	runManuscriptDirect(t, filepath.Join(dir, "manuscript.md"), output)
	bounds := commandOutput(t, exec.Command("pdftotext", "-bbox", output, "-"))

	var document struct {
		Pages []struct {
			Words []struct {
				YMin float64 `xml:"yMin,attr"`
				Text string  `xml:",chardata"`
			} `xml:"word"`
		} `xml:"body>doc>page"`
	}
	if err := xml.Unmarshal([]byte(bounds), &document); err != nil {
		t.Fatalf("parsing PDF bounding boxes: %v", err)
	}

	minimum := 0.0
	for _, page := range document.Pages {
		for _, word := range page.Words {
			if word.Text == "ContinuityMarker" && (minimum == 0 || word.YMin < minimum) {
				minimum = word.YMin
			}
		}
	}
	if minimum == 0 {
		t.Fatal("compiled PDF contains no ContinuityMarker text")
	}
	measuredMM := minimum / pointsPerMillimetre
	if measuredMM < 10 || measuredMM > 16 {
		t.Fatalf("continuity text begins %.2fmm from top, want approximately the configured 13mm margin", measuredMM)
	}
}
