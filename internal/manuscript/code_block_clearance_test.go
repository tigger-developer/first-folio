// ABOUTME: Verifies paragraph separation around standalone and fenced manuscript code.
// ABOUTME: Measures PDF transitions and additive block clearance through the public CLI.
package manuscript

import (
	"fmt"
	"math"
	"os/exec"
	"path/filepath"
	"testing"
)

// RT050.1: code clearances add to normal paragraph separation on every boundary.
func TestRT050CodeBlockClearance(t *testing.T) {
	binary := buildFontCLI(t)
	for _, fenced := range []bool{false, true} {
		for _, size := range []int{9, 12} {
			for _, multiplier := range []string{"1", "1.5", "2"} {
				t.Run(fmt.Sprintf("fenced=%t/mono=%d/spacing=%s", fenced, size, multiplier), func(t *testing.T) {
					var baseline []float64
					for _, scenario := range []struct {
						gap, overrides, paragraph string
						increments                [3]float64
					}{
						{"0.5em", "", "0pt", [3]float64{}},
						{"5mm", "", "0pt", [3]float64{5*72/25.4 - 6, 5*72/25.4 - 6, 5*72/25.4 - 6}},
						{"5mm", "    code-block:\n      space-before: 0.5em\n", "0pt", [3]float64{0, 5*72/25.4 - 6, 5*72/25.4 - 6}},
						{"5mm", "    code-block:\n      space-after: 0.5em\n", "0pt", [3]float64{5*72/25.4 - 6, 5*72/25.4 - 6, 0}},
						{"5mm", "    code-block:\n      space-before: 0.5em\n      space-after: 0.5em\n", "0pt", [3]float64{}},
						{"0.5em", "", "5mm", [3]float64{5 * 72 / 25.4, 5 * 72 / 25.4, 5 * 72 / 25.4}},
					} {
						gap := scenario.gap
						t.Setenv("HOME", t.TempDir())
						dir := t.TempDir()
						code := func(s string) string {
							if fenced {
								return "```\n" + s + "\n```"
							}
							return "`" + s + "`"
						}
						writeFile(t, filepath.Join(dir, "input.md"), "Alpha\n\nBeta\n\n"+code("Gamma")+"\n\n"+code("Delta")+"\n\nEpsilon `inline` suffix.\n\nZeta\n")
						writeFile(t, filepath.Join(dir, "script.yaml"), fmt.Sprintf(`folio:
  manuscript:
    font:
      family: DejaVu Sans Mono
      size: 12pt
    mono:
      font:
        family: DejaVu Sans Mono
        size: %dpt
    line-spacing: %s
    paragraph-spacing: %s
    code-block-spacing: %s
%s    title-page:
      enabled: false
    toc:
      enabled: false
    page-header:
      enabled: false
    page-footer:
      enabled: false
`, size, multiplier, scenario.paragraph, gap, scenario.overrides))
						pdf := filepath.Join(dir, "output.pdf")
						commandOutput(t, exec.Command(binary, "manuscript", filepath.Join(dir, "input.md"), pdf))
						words := inlinePDFWords(t, pdf)
						for _, marker := range []string{"Alpha", "Beta", "Gamma", "Delta", "Epsilon", "inline", "suffix.", "Zeta"} {
							if _, ok := words[marker]; !ok {
								t.Fatalf("missing %s", marker)
							}
						}
						prose := words["Beta"].YMin - words["Alpha"].YMin
						// Compare lower font-box edges across mixed font sizes.
						gaps := []float64{words["Gamma"].YMax - words["Beta"].YMax, words["Delta"].YMax - words["Gamma"].YMax, words["Epsilon"].YMax - words["Delta"].YMax}
						for i, got := range gaps {
							if got < prose-0.03 {
								t.Errorf("%s boundary %d: %.3fpt below ordinary paragraph interval %.3fpt", gap, i, got, prose)
							}
							if baseline != nil && math.Abs(got-baseline[i]-scenario.increments[i]) > 0.03 {
								t.Errorf("boundary %d: clearance increment %.3fpt, want %.3fpt", i, got-baseline[i], scenario.increments[i])
							}
						}
						if baseline == nil {
							baseline = gaps
						}
						if math.Abs(words["suffix."].YMin-words["Epsilon"].YMin) > 0.03 {
							t.Error("mixed inline code split its prose line")
						}
						if math.Abs(words["Zeta"].YMin-words["Epsilon"].YMin-prose) > 0.03 {
							t.Error("mixed inline code changed following paragraph spacing")
						}
					}
				})
			}
		}
	}
}
