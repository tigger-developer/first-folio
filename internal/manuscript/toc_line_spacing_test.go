// ABOUTME: Verifies TOC font-relative line spacing through rendered PDF geometry.
// ABOUTME: Covers wrapped entries, entry gaps, preset defaults and invalid multipliers.
package manuscript

import (
	"encoding/xml"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// RT037.5: TOC lines and adjacent entries scale the TOC font's native interval.
func TestRT037TOCLineSpacing(t *testing.T) {
	binary := buildFontCLI(t)
	for _, font := range []string{"Libertinus Serif", "DejaVu Sans Mono"} {
		for _, size := range []int{10, 16} {
			t.Run(fmt.Sprintf("%s/%d", font, size), func(t *testing.T) {
				ref := t.TempDir()
				writeFile(t, filepath.Join(ref, "output.typ"), fmt.Sprintf("#set text(font: %q, size: %dpt)\nAlpha \\\nBeta\n", font, size))
				words := inlinePDFWords(t, compileFontPDF(t, ref))
				native := words["Beta"].YMin - words["Alpha"].YMin
				for _, multiplier := range []float64{0.75, 1, 1.5, 2} {
					dir := tocSpacingFixture(t, font, size, strconv.FormatFloat(multiplier, 'f', -1, 64))
					pdf := filepath.Join(dir, "output.pdf")
					commandOutput(t, exec.Command(binary, "manuscript", filepath.Join(dir, "input.md"), pdf))
					positions := tocLinePositions(t, pdf)
					if len(positions) < 4 {
						t.Fatal("fixture did not wrap the first entry")
					}
					for i := 1; i < len(positions); i++ {
						got := positions[i] - positions[i-1]
						if math.Abs(got-native*multiplier) > 0.03 {
							t.Errorf("%gx interval %d: %.3fpt, want %.3fpt", multiplier, i, got, native*multiplier)
						}
					}
				}
			})
		}
	}
}

func tocSpacingFixture(t *testing.T, font string, size int, value string) string {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "input.md"), "## Alpha "+strings.Repeat("filler ", 50)+"Omega\n\nBody.\n\n## Beta\n\nBody.\n\n## Gamma\n\nBody.\n")
	writeFile(t, filepath.Join(dir, "script.yaml"), fmt.Sprintf(`folio:
  manuscript:
    title-page:
      enabled: false
    toc:
      font:
        family: %q
        size: %dpt
      line-spacing: %q
`, font, size, value))
	return dir
}

func tocLinePositions(t *testing.T, pdf string) []float64 {
	t.Helper()
	data := commandOutput(t, exec.Command("pdftotext", "-bbox", pdf, "-"))
	var doc struct {
		Pages []struct {
			Words []inlinePDFWord `xml:"word"`
		} `xml:"body>doc>page"`
	}
	if err := xml.Unmarshal([]byte(data), &doc); err != nil {
		t.Fatal(err)
	}
	for _, page := range doc.Pages {
		hasBeta, hasGamma := false, false
		positions := map[float64]bool{}
		for _, word := range page.Words {
			switch word.Text {
			case "Beta":
				hasBeta = true
			case "Gamma":
				hasGamma = true
			}
			switch word.Text {
			case "Alpha", "filler", "Omega", "Beta", "Gamma":
				positions[word.YMin] = true
			}
		}
		if hasBeta && hasGamma {
			result := make([]float64, 0, len(positions))
			for y := range positions {
				result = append(result, y)
			}
			sort.Float64s(result)
			return result
		}
	}
	t.Fatal("TOC page missing")
	return nil
}

// RT037.6: invalid TOC values fail before output with field-specific migration guidance.
func TestRT037RejectInvalidTOCLineSpacing(t *testing.T) {
	binary := buildFontCLI(t)
	for _, value := range []string{"1em", "5mm", "12pt", "0", "-1", "NaN", "+Inf", "-Inf", "", "abc", "true"} {
		t.Run(value, func(t *testing.T) {
			dir := tocSpacingFixture(t, "Libertinus Serif", 10, value)
			cfgPath := filepath.Join(dir, "script.yaml")
			cfg := readFile(t, cfgPath)
			for _, enabled := range []bool{true, false} {
				writeFile(t, cfgPath, strings.Replace(cfg, "    toc:\n", fmt.Sprintf("    toc:\n      enabled: %t\n", enabled), 1))
				for _, ext := range []string{"typ", "pdf"} {
					target := filepath.Join(dir, "output."+ext)
					data, err := exec.Command(binary, "manuscript", filepath.Join(dir, "input.md"), target).CombinedOutput()
					if err == nil {
						t.Errorf("accepted %q", value)
					}
					for _, want := range []string{"folio.manuscript.toc.line-spacing", strconv.Quote(value), "positive finite multiplier", "1.0", "part-gap-before"} {
						if !strings.Contains(string(data), want) {
							t.Errorf("missing %q: %s", want, data)
						}
					}
					if _, err := os.Stat(target); !os.IsNotExist(err) {
						t.Errorf("invalid value created output: %v", err)
					}
				}
			}
		})
	}
}

// RT037.5: both styles inherit a 1.15x TOC default independently of body spacing.
func TestRT037TOCSpacingDefaults(t *testing.T) {
	binary := buildFontCLI(t)
	for _, style := range []string{"british", "us"} {
		dir := tocSpacingFixture(t, "Libertinus Serif", 10, "1.15")
		cfgPath := filepath.Join(dir, "script.yaml")
		cfg := readFile(t, cfgPath)
		var positions [][]float64
		for _, config := range []string{cfg, strings.ReplaceAll(cfg, "      line-spacing: \"1.15\"\n", "")} {
			writeFile(t, cfgPath, config)
			pdf := filepath.Join(dir, "output.pdf")
			commandOutput(t, exec.Command(binary, "manuscript", filepath.Join(dir, "input.md"), pdf, "--style", style))
			positions = append(positions, tocLinePositions(t, pdf))
		}
		if fmt.Sprint(positions[0]) != fmt.Sprint(positions[1]) {
			t.Errorf("%s default differs from explicit 1.15", style)
		}
	}
}
