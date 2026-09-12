// ABOUTME: Checks proportional inline monospace stretch through the manuscript CLI.
// ABOUTME: Measures actual PDF glyph bounds, including surrounding text and vertical alignment.
package manuscript

import (
	"encoding/xml"
	"fmt"
	"io"
	"math"
	"os/exec"
	"strings"
	"testing"
)

// RT042.4: percentages change actual paragraph code width even with a single-width font.
func TestRT042_4ProportionalInlineStretch(t *testing.T) {
	binary := buildFontCLI(t)
	for _, format := range []string{"md", "org"} {
		var baseline map[string]inlinePDFWord
		for _, percent := range []int{100, 150, 200, 50} {
			t.Run(fmt.Sprintf("%s/%d", format, percent), func(t *testing.T) {
				cfg := fontFixtureConfig()
				setFontTestPath(cfg, "folio.manuscript.mono.font.stretch", fmt.Sprintf("%d%%", percent))
				setFontTestPath(cfg, "folio.manuscript.toc.enabled", false)
				dir, path := fontFixture(t, format, false, cfg)
				source := "Before `monospace_0123` after.\n"
				if format == "org" {
					source = "Before =monospace_0123= after.\n"
				}
				writeFile(t, path, source)
				renderFontCLI(t, binary, dir, path)
				words := inlinePDFWords(t, compileFontPDF(t, dir))
				if percent == 100 {
					baseline = words
				}
				for _, marker := range []string{"Before", "monospace_0123", "after."} {
					word, ok := words[marker]
					if !ok {
						t.Fatalf("PDF lacks %q", marker)
					}
					base := baseline[marker]
					wantWidth := base.XMax - base.XMin
					if marker == "monospace_0123" {
						wantWidth *= float64(percent) / 100
					}
					if got := word.XMax - word.XMin; math.Abs(got-wantWidth) > 0.02 {
						t.Errorf("%s width %.3fpt, want %.3fpt at %d%%", marker, got, wantWidth, percent)
					}
					if math.Abs(word.YMin-base.YMin) > 0.02 || math.Abs(word.YMax-base.YMax) > 0.02 {
						t.Errorf("%s height or vertical placement changed: %+v versus %+v", marker, word, base)
					}
				}
				if words["monospace_0123"].XMax >= words["after."].XMin {
					t.Error("scaled code overlaps following prose")
				}
			})
		}
	}
}

type inlinePDFWord struct {
	XMin float64 `xml:"xMin,attr"`
	XMax float64 `xml:"xMax,attr"`
	YMin float64 `xml:"yMin,attr"`
	YMax float64 `xml:"yMax,attr"`
	Text string  `xml:",chardata"`
}

func inlinePDFWords(t *testing.T, pdf string) map[string]inlinePDFWord {
	t.Helper()
	requireTool(t, "pdftotext")
	output := commandOutput(t, exec.Command("pdftotext", "-bbox", pdf, "-"))
	decoder := xml.NewDecoder(strings.NewReader(output))
	words := map[string]inlinePDFWord{}
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			return words
		}
		if err != nil {
			t.Fatal(err)
		}
		if start, ok := token.(xml.StartElement); ok && start.Name.Local == "word" {
			var word inlinePDFWord
			if err := decoder.DecodeElement(&word, &start); err != nil {
				t.Fatal(err)
			}
			words[word.Text] = word
		}
	}
}
