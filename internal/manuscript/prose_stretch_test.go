// ABOUTME: Measures independent manuscript font stretch through CLI-generated PDFs.
// ABOUTME: Covers prose, quotations, code, emphasis and paragraph flow.
package manuscript

import (
	"fmt"
	"math"
	"os/exec"
	"strings"
	"testing"
)

func TestManuscriptProportionalRoleStretch(t *testing.T) {
	binary := buildFontCLI(t)
	for _, format := range []string{"md", "org"} {
		var baseline map[string]inlinePDFWord
		for _, role := range []string{"baseline", "font", "quoted-block.font", "mono.font", "combined"} {
			for _, percent := range []int{100, 50, 160, 200} {
				if (role == "baseline" || role == "combined") && percent != 100 {
					continue
				}
				t.Run(fmt.Sprintf("%s/%s/%d", format, role, percent), func(t *testing.T) {
					cfg := proseStretchConfig()
					if role == "combined" {
						setFontTestPath(cfg, "folio.manuscript.font.stretch", "160%")
						setFontTestPath(cfg, "folio.manuscript.quoted-block.font.stretch", "130%")
						setFontTestPath(cfg, "folio.manuscript.mono.font.stretch", "120%")
					} else if role != "baseline" {
						setFontTestPath(cfg, "folio.manuscript."+role+".stretch", fmt.Sprintf("%d%%", percent))
					}
					source := "Plainword **Boldword** *Italicword* `Inlineword` Afterword.\n\n> Quoteword **Quotebold** *Quoteitalic* `Quotecode`\n\n```\nBlockword\n```\n\n`Standaloneword`\n"
					words := renderStretchWords(t, binary, format, cfg, source)
					if role == "baseline" {
						baseline = words
					}
					for _, group := range []struct{ role, markers string }{
						{"font", "Plainword Boldword Italicword Afterword."},
						{"quoted-block.font", "Quoteword Quotebold Quoteitalic"},
						{"mono.font", "Inlineword Quotecode Blockword Standaloneword"},
					} {
						for _, marker := range strings.Fields(group.markers) {
							got, ok := words[marker]
							if !ok {
								t.Fatalf("PDF lost %s", marker)
							}
							base := baseline[marker]
							factor := 1.0
							if role == "combined" {
								factor = map[string]float64{"font": 1.6, "quoted-block.font": 1.3, "mono.font": 1.2}[group.role]
							}
							if role == group.role {
								factor = float64(percent) / 100
							}
							want := (base.XMax - base.XMin) * factor
							if math.Abs(got.XMax-got.XMin-want) > 0.03 {
								t.Errorf("%s width %.3fpt, want %.3fpt", marker, got.XMax-got.XMin, want)
							}
							if math.Abs((got.YMax-got.YMin)-(base.YMax-base.YMin)) > 0.03 {
								t.Errorf("%s glyph height changed", marker)
							}
						}
					}
				})
			}
		}
	}
}

func proseStretchConfig() map[string]any {
	cfg := fontFixtureConfig()
	for path, value := range map[string]any{
		"font.family": "Libertinus Serif", "font.size": "10pt", "font.stretch": "100%",
		"quoted-block.font.family": "Libertinus Serif", "quoted-block.font.size": "10pt", "quoted-block.font.stretch": "100%",
		"mono.font.size": "10pt", "mono.font.stretch": "100%",
		"toc.enabled": false, "page-header.enabled": false, "page-footer.enabled": false,
		"justify": false, "paragraph-indent": "0pt", "margin": "20mm",
	} {
		setFontTestPath(cfg, "folio.manuscript."+path, value)
	}
	return cfg
}

func renderStretchWords(t *testing.T, binary, format string, cfg map[string]any, source string) map[string]inlinePDFWord {
	t.Helper()
	dir, path := fontFixture(t, format, false, cfg)
	if format == "org" {
		cmd := exec.Command("pandoc", "--from=markdown", "--to=org")
		cmd.Stdin = strings.NewReader(source)
		source = commandOutput(t, cmd)
	}
	writeFile(t, path, source)
	renderFontCLI(t, binary, dir, path)
	return inlinePDFWords(t, compileFontPDF(t, dir))
}

func TestManuscriptStretchWrapsAndJustifies(t *testing.T) {
	binary := buildFontCLI(t)
	for _, justify := range []bool{false, true} {
		for _, percent := range []int{100, 160} {
			t.Run(fmt.Sprintf("justify=%t/%d", justify, percent), func(t *testing.T) {
				cfg := proseStretchConfig()
				setFontTestPath(cfg, "folio.manuscript.justify", justify)
				setFontTestPath(cfg, "folio.manuscript.page", "6x9in")
				setFontTestPath(cfg, "folio.manuscript.font.stretch", fmt.Sprintf("%d%%", percent))
				var source strings.Builder
				for i := 0; i < 100; i++ {
					fmt.Fprintf(&source, "word%03d ", i)
				}
				words := renderStretchWords(t, binary, "md", cfg, source.String())
				left, right := 20*72/25.4, 6*72-20*72/25.4
				lines := map[float64][]inlinePDFWord{}
				for i := 0; i < 100; i++ {
					marker := fmt.Sprintf("word%03d", i)
					word, ok := words[marker]
					if !ok {
						t.Fatalf("missing %s", marker)
					}
					if word.XMin < left-0.05 || word.XMax > right+0.05 {
						t.Errorf("%s outside text margins: %+v", marker, word)
					}
					lines[word.YMin] = append(lines[word.YMin], word)
				}
				if len(lines) < 3 {
					t.Fatal("paragraph failed to wrap")
				}
				for y, line := range lines {
					for i := 1; i < len(line); i++ {
						if line[i].XMin < line[i-1].XMax {
							t.Errorf("overlapping words at y=%f", y)
						}
					}
					last := line[len(line)-1]
					if justify && last.Text != "word099" && math.Abs(last.XMax-right) > 0.2 {
						t.Errorf("justified line ends %.3f, want %.3f", last.XMax, right)
					}
				}
			})
		}
	}
}
