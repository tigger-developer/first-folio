// ABOUTME: Verifies manuscript line-height multipliers through the built CLI and PDF output.
// ABOUTME: Compares native font spacing and checks additive paragraph gaps and input errors.
package manuscript

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// RT048.1 and RT048.2: multipliers scale the native interval; paragraph gaps add lengths.
func TestRT048RenderedLineAndParagraphSpacing(t *testing.T) {
	binary := buildFontCLI(t)
	for _, font := range []string{"Libertinus Serif", "DejaVu Sans Mono"} {
		for _, size := range []int{10, 16} {
			t.Run(fmt.Sprintf("%s/%dpt", font, size), func(t *testing.T) {
				reference := t.TempDir()
				writeFile(t, filepath.Join(reference, "output.typ"), fmt.Sprintf("#set text(font: %q, size: %dpt)\nAlpha \\\nBeta\n", font, size))
				words := inlinePDFWords(t, compileFontPDF(t, reference))
				normal := words["Beta"].YMin - words["Alpha"].YMin
				if normal <= 0 {
					t.Fatal("native reference has no positive line interval")
				}
				t.Logf("native single-spacing interval: %.3fpt", normal)
				for _, multiplier := range []float64{0.75, 1, 1.5, 2} {
					for _, gap := range []struct {
						value  string
						points float64
					}{{"0", 0}, {"0.5em", float64(size) / 2}, {"5mm", 5 * 72 / 25.4}} {
						dir := lineSpacingFixture(t, font, size, strconv.FormatFloat(multiplier, 'f', -1, 64), gap.value)
						pdf := filepath.Join(dir, "output.pdf")
						commandOutput(t, exec.Command(binary, "manuscript", filepath.Join(dir, "input.md"), pdf))
						got := inlinePDFWords(t, pdf)
						for _, marker := range []string{"Alpha", "Beta", "Gamma"} {
							if _, ok := got[marker]; !ok {
								t.Fatalf("missing %s", marker)
							}
						}
						line := got["Beta"].YMin - got["Alpha"].YMin
						paragraph := got["Gamma"].YMin - got["Beta"].YMin
						want := normal * multiplier
						if math.Abs(line-want) > 0.03 {
							t.Errorf("multiplier %g gap %s: line %.3fpt, want %.3fpt", multiplier, gap.value, line, want)
						}
						if math.Abs(paragraph-want-gap.points) > 0.03 {
							t.Errorf("multiplier %g gap %s: paragraph %.3fpt, want %.3fpt", multiplier, gap.value, paragraph, want+gap.points)
						}
					}
				}
			})
		}
	}
}

// RT048.3: invalid spacing fails before writing output and gives migration guidance.
func TestRT048RejectInvalidLineSpacing(t *testing.T) {
	binary := buildFontCLI(t)
	for _, value := range []string{"1em", "5mm", "12pt", "0", "-1", "NaN", "+Inf", "-Inf", "", "abc", "true"} {
		t.Run(value, func(t *testing.T) {
			dir := lineSpacingFixture(t, "Libertinus Serif", 12, value, "0")
			for _, extension := range []string{"typ", "pdf"} {
				output := filepath.Join(dir, "output."+extension)
				data, err := exec.Command(binary, "manuscript", filepath.Join(dir, "input.md"), output).CombinedOutput()
				if err == nil {
					t.Errorf("accepted invalid line-spacing %q", value)
				}
				for _, required := range []string{"folio.manuscript.line-spacing", strconv.Quote(value), "positive finite multiplier", "1.0", "paragraph-spacing"} {
					if !strings.Contains(string(data), required) {
						t.Errorf("diagnostic missing %q: %s", required, data)
					}
				}
				if _, err := os.Stat(output); !os.IsNotExist(err) {
					t.Errorf("invalid spacing created output: %v", err)
				}
			}
		})
	}
}

func lineSpacingFixture(t *testing.T, font string, size int, multiplier, gap string) string {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "input.md"), "Alpha  \nBeta\n\nGamma\n")
	writeFile(t, filepath.Join(dir, "script.yaml"), fmt.Sprintf(`folio:
  manuscript:
    font:
      family: %q
      size: %dpt
    line-spacing: %q
    paragraph-spacing: %s
    title-page:
      enabled: false
    toc:
      enabled: false
    page-header:
      enabled: false
    page-footer:
      enabled: false
`, font, size, multiplier, gap))
	return dir
}

// RT048.4: omitted values use the embedded British base and US overlay.
func TestRT048PresetSpacingDefaults(t *testing.T) {
	binary := buildFontCLI(t)
	for _, style := range []struct{ name, multiplier string }{{"british", "1.5"}, {"us", "2"}} {
		t.Run(style.name, func(t *testing.T) {
			dir := lineSpacingFixture(t, "Libertinus Serif", 12, style.multiplier, "0pt")
			path := filepath.Join(dir, "script.yaml")
			config := readFile(t, path)
			writeFile(t, path, strings.ReplaceAll(strings.ReplaceAll(config, "    line-spacing: "+strconv.Quote(style.multiplier)+"\n", ""), "    paragraph-spacing: 0pt\n", ""))
			defaultPDF := filepath.Join(dir, "default.pdf")
			commandOutput(t, exec.Command(binary, "manuscript", filepath.Join(dir, "input.md"), defaultPDF, "--style", style.name))
			writeFile(t, path, config)
			explicitPDF := filepath.Join(dir, "explicit.pdf")
			commandOutput(t, exec.Command(binary, "manuscript", filepath.Join(dir, "input.md"), explicitPDF, "--style", style.name))
			got, want := inlinePDFWords(t, defaultPDF), inlinePDFWords(t, explicitPDF)
			for _, marker := range []string{"Alpha", "Beta", "Gamma"} {
				if _, ok := got[marker]; !ok {
					t.Fatalf("default output missing %s", marker)
				}
				if math.Abs(got[marker].YMin-want[marker].YMin) > 0.03 {
					t.Errorf("%s: default differs from explicit %sx/0pt", marker, style.multiplier)
				}
			}
		})
	}
}
