// ABOUTME: Checks proportional monospace stretch for fenced code blocks through the manuscript CLI.
// ABOUTME: Measures actual PDF glyph bounds for word width, block alignment and retained pagination.
package manuscript

import (
	"fmt"
	"math"
	"os/exec"
	"strings"
	"testing"
)

// blockStretchSource returns an equivalent two-line indented code block for each input format.
func blockStretchSource(format string) string {
	if format == "org" {
		return strings.Join([]string{
			"Before the block.",
			"",
			"#+begin_example",
			"codeA codeB",
			"  codeC codeD",
			"#+end_example",
			"",
			"After the block.",
			"",
		}, "\n")
	}
	return strings.Join([]string{
		"Before the block.",
		"",
		"```",
		"codeA codeB",
		"  codeC codeD",
		"```",
		"",
		"After the block.",
		"",
	}, "\n")
}

// RT043.1: the configured mono stretch scales fenced code-block width proportionally,
// exactly as it already scales paragraph inline monospace, preserving the monospace grid.
func TestRT043_1ProportionalBlockStretch(t *testing.T) {
	binary := buildFontCLI(t)
	codeWords := []string{"codeA", "codeB", "codeC", "codeD"}
	for _, format := range []string{"md", "org"} {
		var baseline map[string]inlinePDFWord
		for _, percent := range []int{100, 150, 200, 50} {
			t.Run(fmt.Sprintf("%s/%d", format, percent), func(t *testing.T) {
				cfg := fontFixtureConfig()
				setFontTestPath(cfg, "folio.manuscript.mono.font.stretch", fmt.Sprintf("%d%%", percent))
				setFontTestPath(cfg, "folio.manuscript.toc.enabled", false)
				setFontTestPath(cfg, "folio.manuscript.font.size", "18pt")
				setFontTestPath(cfg, "folio.manuscript.mono.font.size", "18pt")
				dir, path := fontFixture(t, format, false, cfg)
				writeFile(t, path, blockStretchSource(format))
				renderFontCLI(t, binary, dir, path)
				words := inlinePDFWords(t, compileFontPDF(t, dir))
				for _, marker := range append(codeWords, "Before", "After") {
					if _, ok := words[marker]; !ok {
						t.Fatalf("PDF lacks %q", marker)
					}
				}
				if percent == 100 {
					baseline = words
					return
				}
				ratio := float64(percent) / 100

				// Every code glyph run widens by the configured ratio without changing height.
				for _, marker := range codeWords {
					word, base := words[marker], baseline[marker]
					wantWidth := (base.XMax - base.XMin) * ratio
					if got := word.XMax - word.XMin; math.Abs(got-wantWidth) > 0.02 {
						t.Errorf("%s width %.3fpt, want %.3fpt at %d%%", marker, got, wantWidth, percent)
					}
					if math.Abs(word.YMin-base.YMin) > 0.02 || math.Abs(word.YMax-base.YMax) > 0.02 {
						t.Errorf("%s height or vertical placement changed: %+v versus %+v", marker, word, base)
					}
				}

				// The block keeps its left edge; inner spacing scales so the monospace grid stays aligned.
				if math.Abs(words["codeA"].XMin-baseline["codeA"].XMin) > 0.02 {
					t.Errorf("block left edge moved to %.3fpt, want %.3fpt", words["codeA"].XMin, baseline["codeA"].XMin)
				}
				for _, marker := range []string{"codeB", "codeC", "codeD"} {
					wantOffset := (baseline[marker].XMin - baseline["codeA"].XMin) * ratio
					if got := words[marker].XMin - words["codeA"].XMin; math.Abs(got-wantOffset) > 0.02 {
						t.Errorf("%s offset from the block edge %.3fpt, want %.3fpt at %d%%", marker, got, wantOffset, percent)
					}
				}

				// The two source lines remain two rendered lines, and surrounding prose is untouched.
				if math.Abs(words["codeA"].YMin-words["codeB"].YMin) > 0.02 {
					t.Error("the first code line was split across rendered lines")
				}
				if words["codeC"].YMin <= words["codeA"].YMin {
					t.Error("the second code line did not follow the first")
				}
				for _, marker := range []string{"Before", "After"} {
					word, base := words[marker], baseline[marker]
					if math.Abs((word.XMax-word.XMin)-(base.XMax-base.XMin)) > 0.02 {
						t.Errorf("surrounding prose word %s changed width: %+v versus %+v", marker, word, base)
					}
				}
			})
		}
	}
}

// RT043.2: a stretched code block still breaks across pages, so no source line is dropped.
func TestRT043_2StretchedBlockRetainsEveryLine(t *testing.T) {
	binary := buildFontCLI(t)
	const lines = 80
	for _, format := range []string{"md", "org"} {
		t.Run(format, func(t *testing.T) {
			cfg := fontFixtureConfig()
			setFontTestPath(cfg, "folio.manuscript.mono.font.stretch", "150%")
			setFontTestPath(cfg, "folio.manuscript.toc.enabled", false)
			dir, path := fontFixture(t, format, false, cfg)
			body := make([]string, 0, lines)
			for i := 1; i <= lines; i++ {
				body = append(body, fmt.Sprintf("codeline%03d", i))
			}
			opener, closer := "```", "```"
			if format == "org" {
				opener, closer = "#+begin_example", "#+end_example"
			}
			source := strings.Join(append(append([]string{opener}, body...), closer, ""), "\n")
			writeFile(t, path, source)
			renderFontCLI(t, binary, dir, path)
			pdf := compileFontPDF(t, dir)
			requireTool(t, "pdftotext")
			text := commandOutput(t, exec.Command("pdftotext", "-layout", pdf, "-"))
			for _, marker := range body {
				if !strings.Contains(text, marker) {
					t.Errorf("stretched code block dropped %s", marker)
				}
			}
		})
	}
}

// RT043.3: the block monospace rule carries every configured font property, as the inline rule does.
func TestRT043_3BlockFontPropertiesReachContent(t *testing.T) {
	binary := buildFontCLI(t)
	overrides := map[string]any{
		"family":         "New Computer Modern",
		"size":           "13.5pt",
		"weight":         650,
		"stretch":        "75%",
		"style":          "italic",
		"letter-spacing": "-0.03em",
	}
	want := []string{
		`font: "New Computer Modern"`,
		"size: 13.5pt",
		"weight: 650",
		"stretch: 75%",
		`style: "italic"`,
		"tracking: -0.03em",
	}
	for _, format := range []string{"md", "org"} {
		t.Run(format, func(t *testing.T) {
			cfg := fontFixtureConfig()
			setFontTestPath(cfg, "folio.manuscript.mono.font", overrides)
			setFontTestPath(cfg, "folio.manuscript.toc.enabled", false)
			dir, path := fontFixture(t, format, false, cfg)
			writeFile(t, path, blockStretchSource(format))
			output := renderFontCLI(t, binary, dir, path)
			rule := fontOutputRule(t, output, "#let folio-code(body)")
			for _, arg := range want {
				assertContains(t, rule, arg)
			}
			compileFontPDF(t, dir)
		})
	}
}
