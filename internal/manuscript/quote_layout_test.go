// ABOUTME: Verifies manuscript quote indentation and additive vertical clearances.
// ABOUTME: Measures public CLI PDF output for Markdown and Org quote blocks.
package manuscript

import (
	"fmt"
	"math"
	"os/exec"
	"path/filepath"
	"testing"
)

// RT051.1: configured indentation is the entire quote inset, without a native extra inset.
// RT051.2: quote clearance adds to paragraph separation in the body font context.
func TestRT051QuoteLayout(t *testing.T) {
	binary := buildFontCLI(t)
	for _, format := range []string{"md", "org"} {
		for _, size := range []int{10, 14} {
			for _, indent := range []struct {
				value  string
				points float64
			}{{"0mm", 0}, {"6mm", 6 * 72 / 25.4}, {"2em", 24}} {
				t.Run(fmt.Sprintf("%s/%d/%s", format, size, indent.value), func(t *testing.T) {
					var baseline []float64
					for _, scenario := range []struct {
						gap, paragraph string
						increment      float64
					}{{"0.5em", "0pt", 0}, {"5mm", "0pt", 5*72/25.4 - 6}, {"0.5em", "5mm", 5 * 72 / 25.4}} {
						t.Setenv("HOME", t.TempDir())
						dir := t.TempDir()
						quote := "> Gamma\n>\n> Theta\n"
						if format == "org" {
							quote = "#+begin_quote\nGamma\n\nTheta\n#+end_quote\n"
						}
						source := filepath.Join(dir, "input."+format)
						writeFile(t, source, "Alpha\n\nBeta\n\n"+quote+"\nDelta\n\nEpsilon\n")
						writeFile(t, filepath.Join(dir, "script.yaml"), fmt.Sprintf(`folio:
  manuscript:
    margin: 20mm
    font:
      family: DejaVu Sans Mono
      size: 12pt
    line-spacing: 1.5
    paragraph-indent: 5mm
    paragraph-spacing: %s
    quoted-block-spacing: %s
    quote-block-indent: %s
    quoted-block:
      font:
        family: DejaVu Sans Mono
        size: %dpt
    title-page:
      enabled: false
    toc:
      enabled: false
    page-header:
      enabled: false
    page-footer:
      enabled: false
`, scenario.paragraph, scenario.gap, indent.value, size))
						pdf := filepath.Join(dir, "output.pdf")
						commandOutput(t, exec.Command(binary, "manuscript", source, pdf))
						words := inlinePDFWords(t, pdf)
						for _, marker := range []string{"Alpha", "Beta", "Gamma", "Theta", "Delta", "Epsilon"} {
							if _, ok := words[marker]; !ok {
								t.Fatalf("missing %s", marker)
							}
						}
						for _, marker := range []string{"Gamma", "Theta"} {
							if got, want := words[marker].XMin, 20*72/25.4+indent.points; math.Abs(got-want) > 0.03 {
								t.Errorf("%s left %.3fpt, want %.3fpt", marker, got, want)
							}
						}
						if got, want := words["Beta"].XMin, 25*72/25.4; math.Abs(got-want) > 0.03 {
							t.Errorf("prose indent %.3fpt, want %.3fpt", got, want)
						}
						prose := words["Beta"].YMin - words["Alpha"].YMin
						gaps := []float64{words["Gamma"].YMax - words["Beta"].YMax, words["Delta"].YMax - words["Theta"].YMax}
						for i, got := range gaps {
							if got < prose-0.03 {
								t.Errorf("quote boundary %d: %.3fpt below paragraph interval %.3fpt", i, got, prose)
							}
							if baseline != nil && math.Abs(got-baseline[i]-scenario.increment) > 0.03 {
								t.Errorf("boundary %d increment %.3fpt, want %.3fpt", i, got-baseline[i], scenario.increment)
							}
						}
						if baseline == nil {
							baseline = gaps
						}
						if math.Abs(words["Epsilon"].YMin-words["Delta"].YMin-prose) > 0.03 {
							t.Error("quote changed following prose separation")
						}
					}
				})
			}
		}
	}
}
