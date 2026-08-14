// ABOUTME: Regression tests for manuscript, running-header, and running-footer letter spacing.
// ABOUTME: Verifies YAML inheritance and the generated Typst tracking values.

package manuscript

import (
	"encoding/xml"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

type issue34BBoxDocument struct {
	Pages []issue34BBoxPage `xml:"body>doc>page"`
}

type issue34BBoxPage struct {
	Flows []issue34BBoxFlow `xml:"flow"`
}

type issue34BBoxFlow struct {
	Blocks []issue34BBoxBlock `xml:"block"`
}

type issue34BBoxBlock struct {
	Lines []issue34BBoxLine `xml:"line"`
}

type issue34BBoxLine struct {
	Words []issue34BBoxWord `xml:"word"`
}

type issue34BBoxWord struct {
	XMin float64 `xml:"xMin,attr"`
	XMax float64 `xml:"xMax,attr"`
	Text string  `xml:",chardata"`
}

// RT-34.1: all manuscript letter spacing defaults to zero em.
func TestRT_34_1_LetterSpacingDefaultsToZeroEm(t *testing.T) {
	typst := renderIssue15Manuscript(t, "")
	body := extractBodyPageBlock(t, typst)

	assertContains(t, body, "tracking: 0em,")
	assertContains(t, extractHeaderBlock(t, typst), "tracking: 0em,")
	assertContains(t, extractFooterBlock(t, typst), "tracking: 0em,")
}

// RT-34.2: header and footer inherit manuscript letter spacing when unset.
func TestRT_34_2_HeaderAndFooterInheritManuscriptLetterSpacing(t *testing.T) {
	typst := renderIssue15Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    letter-spacing: 0.04em",
		"",
	}, "\n"))
	body := extractBodyPageBlock(t, typst)

	assertContains(t, body, "tracking: 0.04em,")
	assertContains(t, extractHeaderBlock(t, typst), "tracking: 0.04em,")
	assertContains(t, extractFooterBlock(t, typst), "tracking: 0.04em,")
}

// RT-34.3: header and footer can override manuscript letter spacing independently.
func TestRT_34_3_HeaderAndFooterLetterSpacingOverrides(t *testing.T) {
	typst := renderIssue15Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    letter-spacing: 0.01em",
		"    page-header:",
		"      letter-spacing: 0.08em",
		"    page-footer:",
		"      letter-spacing: -0.02em",
		"",
	}, "\n"))

	assertContains(t, extractHeaderBlock(t, typst), "tracking: 0.08em,")
	assertContains(t, extractFooterBlock(t, typst), "tracking: -0.02em,")
}

// RT-34.4: a header override does not implicitly change an unset footer override.
func TestRT_34_4_FooterInheritsManuscriptNotHeaderOverride(t *testing.T) {
	typst := renderIssue15Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    letter-spacing: 0.01em",
		"    page-header:",
		"      letter-spacing: 0.08em",
		"",
	}, "\n"))

	assertContains(t, extractHeaderBlock(t, typst), "tracking: 0.08em,")
	assertContains(t, extractFooterBlock(t, typst), "tracking: 0.01em,")
}

// RT-34.5: Typst compiles em-based and negative letter-spacing values in all three positions.
func TestRT_34_5_LetterSpacingCompiles(t *testing.T) {
	requireTool(t, "typst")
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "script.yaml"), strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    letter-spacing: 0.02em",
		"    page-header:",
		"      letter-spacing: 0.08em",
		"    page-footer:",
		"      letter-spacing: -0.01em",
		"",
	}, "\n"))
	writeFile(t, filepath.Join(dir, "manuscript.md"), strings.Join([]string{
		"---",
		"title: Letter Spacing Test",
		"author: Example Author",
		"---",
		"",
		"## Chapter 1",
		"",
		"Body text demonstrates configured letter spacing.",
	}, "\n"))

	runManuscriptDirect(t, filepath.Join(dir, "manuscript.md"), filepath.Join(dir, "manuscript.pdf"))
}

// RT-34.6: rendered body, header, and footer text visibly widens with positive letter spacing.
func TestRT_34_6_LetterSpacingChangesRenderedTextWidth(t *testing.T) {
	plain := renderLetterSpacingPDF(t, "0em")
	tracked := renderLetterSpacingPDF(t, "0.1em")

	for _, token := range []string{"BODYTRACK", "HEADERTRACK", "FOOTERTRACK"} {
		plainWidth := issue34WordWidth(t, plain, token)
		trackedWidth := issue34WordWidth(t, tracked, token)
		if trackedWidth <= plainWidth+4 {
			t.Fatalf("%s width with tracking = %.2fpt, without = %.2fpt; want a visible increase", token, trackedWidth, plainWidth)
		}
	}
}

func renderLetterSpacingPDF(t *testing.T, spacing string) string {
	t.Helper()
	requireTool(t, "typst")
	requireTool(t, "pdftotext")
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "script.yaml"), strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    letter-spacing: " + spacing,
		"    title-page:",
		"      enabled: false",
		"    toc:",
		"      enabled: false",
		"    page-header:",
		"      enabled: true",
		"      format: HEADERTRACK",
		"      letter-spacing: " + spacing,
		"    page-footer:",
		"      enabled: true",
		"      format: FOOTERTRACK",
		"      letter-spacing: " + spacing,
		"",
	}, "\n"))
	writeFile(t, filepath.Join(dir, "manuscript.md"), strings.Join([]string{
		"---",
		"title: Letter Spacing Test",
		"author: Example Author",
		"---",
		"",
		"## Chapter 1",
		"",
		"BODYTRACK demonstrates rendered letter spacing.",
	}, "\n"))
	output := filepath.Join(dir, "manuscript.pdf")
	runManuscriptDirect(t, filepath.Join(dir, "manuscript.md"), output)
	return output
}

func issue34WordWidth(t *testing.T, pdf, token string) float64 {
	t.Helper()
	bounds := commandOutput(t, exec.Command("pdftotext", "-bbox-layout", pdf, "-"))
	var document issue34BBoxDocument
	if err := xml.Unmarshal([]byte(bounds), &document); err != nil {
		t.Fatalf("parsing PDF bounding boxes: %v", err)
	}
	for _, line := range issue34Lines(document) {
		var text strings.Builder
		for _, word := range line.Words {
			text.WriteString(word.Text)
		}
		if !strings.HasPrefix(text.String(), token) || len(line.Words) == 0 {
			continue
		}
		characters := 0
		for _, word := range line.Words {
			characters += len(word.Text)
			if characters >= len(token) {
				return word.XMax - line.Words[0].XMin
			}
		}
	}
	t.Fatalf("rendered PDF contains no %s token", token)
	return 0
}

func issue34Lines(document issue34BBoxDocument) []issue34BBoxLine {
	var lines []issue34BBoxLine
	for _, page := range document.Pages {
		for _, flow := range page.Flows {
			for _, block := range flow.Blocks {
				lines = append(lines, block.Lines...)
			}
		}
	}
	return lines
}
