// ABOUTME: W042 regressions for configured fonts at manuscript content boundaries.
// ABOUTME: Exercises the built CLI and unmodified Typst/PDF output, preserving W039 role coverage.
package manuscript

import (
	"encoding/xml"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// RT042.1: the affected spans must use the full mono rule, not merely emit its values elsewhere.
func TestRT042_1InlineFontPropertiesReachContent(t *testing.T) {
	binary := buildFontCLI(t)
	cases := []struct {
		name      string
		overrides map[string]any
		args      []string
	}{
		{"family", map[string]any{"family": "New Computer Modern"}, []string{`font: "New Computer Modern"`}},
		{"size", map[string]any{"size": "13.5pt"}, []string{"size: 13.5pt"}},
		{"numeric-weight", map[string]any{"weight": 650}, []string{"weight: 650"}},
		{"named-weight", map[string]any{"weight": "bold"}, []string{`weight: "bold"`}},
		{"stretch", map[string]any{"stretch": "75%"}, []string{"stretch: 75%"}},
		{"regular-style", map[string]any{"style": "regular"}, []string{`style: "normal"`}},
		{"italic-style", map[string]any{"style": "italic"}, []string{`style: "italic"`}},
		{"oblique-style", map[string]any{"style": "oblique"}, []string{`style: "oblique"`}},
		{"negative-spacing", map[string]any{"letter-spacing": "-0.03em"}, []string{"tracking: -0.03em"}},
		{"positive-spacing", map[string]any{"letter-spacing": "0.1pt"}, []string{"tracking: 0.1pt"}},
		{"combined", map[string]any{"family": "New Computer Modern", "size": "13.5pt", "weight": 650, "stretch": "75%", "style": "italic", "letter-spacing": "-0.03em"}, []string{`font: "New Computer Modern"`, "size: 13.5pt", "weight: 650", "stretch: 75%", `style: "italic"`, "tracking: -0.03em"}},
	}
	for _, format := range []string{"md", "org"} {
		for _, tc := range cases {
			t.Run(format+"/"+tc.name, func(t *testing.T) {
				cfg := fontFixtureConfig()
				setFontTestPath(cfg, "folio.manuscript.mono.font", tc.overrides)
				dir, source := fontFixture(t, format, false, cfg)
				output := renderFontCLI(t, binary, dir, source)
				markers := []string{"section_probe"}
				if format == "md" {
					markers = append(markers, "deep_probe")
				} else {
					// Pandoc's Org reader treats level four as a list; preserve that existing raw route.
					assertContains(t, output, "`deep_probe`")
				}
				for _, marker := range markers {
					if !strings.Contains(output, `#raw("`+marker+`", block: false)`) {
						t.Errorf("%s bypasses the complete mono rule", marker)
					}
				}
				rule := fontOutputRule(t, output, "#show raw.where(block: false):")
				for _, arg := range tc.args {
					assertContains(t, rule, arg)
				}
				// Partial overrides inherit the same role's lower layers, including an embedded font family.
				defaults := map[string]string{"family": `font: "DejaVu Sans Mono"`, "size": "size: 10pt", "weight": `weight: "regular"`, "stretch": "stretch: 100%", "style": `style: "normal"`, "letter-spacing": "tracking: 0em"}
				for key, arg := range defaults {
					if _, changed := tc.overrides[key]; !changed {
						assertContains(t, rule, arg)
					}
				}
				assertContains(t, output, "`para_probe`")
				assertContains(t, output, "block_probe")
				compileFontPDF(t, dir)
			})
		}
	}
}

// RT042.2: preserve literals, nested emphasis and title/outline casing through the public CLI.
func TestRT042_2InlineLiteralsAndTitleCasing(t *testing.T) {
	binary := buildFontCLI(t)
	for _, format := range []string{"md", "org"} {
		for _, mode := range []string{"as-written", "upper"} {
			t.Run(format+"/"+mode, func(t *testing.T) {
				cfg := fontFixtureConfig()
				setFontTestPath(cfg, "folio.manuscript.part.case-transform", mode)
				setFontTestPath(cfg, "folio.manuscript.chapter.case-transform", mode)
				setFontTestPath(cfg, "folio.manuscript.mono.font.weight", 650)
				dir, source := fontFixture(t, format, true, cfg)
				output := renderFontCLI(t, binary, dir, source)
				for _, marker := range []string{"part_Code", "chapter_Code"} {
					if mode == "upper" {
						assertContains(t, output, "`"+strings.ReplaceAll(strings.ToUpper(marker), "_", `\_`)+"`")
					} else {
						assertContains(t, output, `#raw("`+marker+`", block: false)`)
					}
				}
				literal := `#panic("unsafe")_\[]λ`
				if format == "org" {
					// Preserve the existing Org canonicalizer's treatment of this escaped bracket.
					literal = `#panic("unsafe")_[]λ`
					assertContains(t, output, `#raw("#panic(\"unsafe\")_[]λ", block: false)`)
				} else {
					assertContains(t, output, `#raw("#panic(\"unsafe\")_\\[]λ", block: false)`)
				}
				assertContains(t, output, `#raw("bold_probe", block: false)`)
				assertContains(t, output, `#raw("italic_probe", block: false)`)
				pdf := compileFontPDF(t, dir)
				requireTool(t, "pdftotext")
				text := commandOutput(t, exec.Command("pdftotext", "-layout", pdf, "-"))
				for _, marker := range []string{"part_Code", "chapter_Code"} {
					if mode == "upper" {
						marker = strings.ReplaceAll(strings.ToUpper(marker), "_", `\_`)
					}
					if strings.Count(text, marker) != 2 {
						t.Errorf("want %q once in body and once in TOC; got:\n%s", marker, text)
					}
				}
				for _, expected := range []string{literal, "bold_probe", "italic_probe", "para_probe", "block_probe"} {
					assertContains(t, text, expected)
				}
			})
		}
	}
}

// RT042.3: compare heading font records and bold markup with plain reference text in the actual PDF.
func TestRT042_3HeadingFontAndOutlineIsolation(t *testing.T) {
	binary := buildFontCLI(t)
	for _, format := range []string{"md", "org"} {
		for _, weight := range []any{"regular", 650} {
			t.Run(fmt.Sprint(format, "/", weight), func(t *testing.T) {
				cfg := fontFixtureConfig()
				font := map[string]any{"family": "Libertinus Serif", "size": "13.5pt", "weight": weight, "stretch": "117.5%", "style": "oblique", "letter-spacing": "0.03em"}
				setFontTestPath(cfg, "folio.manuscript.heading.font", font)
				setFontTestPath(cfg, "folio.manuscript.font", font)
				dir, source := fontFixture(t, format, false, cfg)
				output := renderFontCLI(t, binary, dir, source)
				pdf := compileFontPDF(t, dir)
				entries := extractFontPDF(t, pdf)
				reference := findFontEntry(t, entries, "ReferenceProbe")
				for _, marker := range []string{"PartProbe", "ChapterProbe", "SectionProbe"} {
					var matching, outline int
					for _, entry := range entries {
						if strings.Contains(entry.Text, marker) {
							if entry.Size == reference.Size && entry.Family == reference.Family && entry.Bold == reference.Bold {
								matching++
							}
							if entry.Size == 15 && strings.Contains(entry.Family, "Sans") && !entry.Bold {
								outline++
							}
						}
					}
					if matching != 1 || outline != 1 {
						t.Errorf("%s: want one reference-font heading and one separate TOC-font entry; got %d/%d", marker, matching, outline)
					}
				}
				for marker, size := range map[string]int{"HeaderProbe": 14, "FooterProbe": 12} {
					entry := findFontEntry(t, entries, marker)
					if entry.Size != size || !strings.Contains(entry.Family, "Sans") || entry.Bold {
						t.Errorf("running band %s lost its own font: %+v", marker, entry)
					}
				}
				rule := fontOutputRule(t, output, "#show heading: set text(")
				for _, arg := range []string{`font: "Libertinus Serif"`, "size: 13.5pt", "stretch: 117.5%", `style: "oblique"`, "tracking: 0.03em"} {
					assertContains(t, rule, arg)
				}
				if weight == "regular" {
					assertContains(t, rule, `weight: "regular"`)
				} else {
					assertContains(t, rule, "weight: 650")
				}

			})
		}
	}
}

func buildFontCLI(t *testing.T) string {
	t.Helper()
	requireTool(t, "pandoc")
	binary := filepath.Join(t.TempDir(), "folio")
	cmd := exec.Command("go", "build", "-o", binary, "./cmd/folio")
	cmd.Dir = testProjectRoot(t)
	commandOutput(t, cmd)
	return binary
}

func fontFixtureConfig() map[string]any {
	cfg := map[string]any{}
	for path, value := range map[string]any{
		"folio.manuscript.title-page.enabled":    false,
		"folio.manuscript.toc.include-sections":  true,
		"folio.manuscript.toc.part-bold":         false,
		"folio.manuscript.page-header.format":    "HeaderProbe",
		"folio.manuscript.page-header.font.size": "9pt",
		"folio.manuscript.page-footer.format":    "FooterProbe",
		"folio.manuscript.page-footer.font.size": "8pt",
	} {
		setFontTestPath(cfg, path, value)
	}
	return cfg
}

func fontFixture(t *testing.T, format string, titleCode bool, cfg map[string]any) (string, string) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	// A real lower configuration layer supplies portable font families; production defaults remain untouched.
	lower := map[string]any{}
	for _, role := range []string{"mono", "heading", "toc", "toc.heading", "page-header", "page-footer"} {
		setFontTestPath(lower, "folio.manuscript."+role+".font.family", "DejaVu Sans Mono")
	}
	lowerYAML, err := yaml.Marshal(lower)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(home, ".config", "first-folio", "script.yaml"), string(lowerYAML))
	dir := t.TempDir()
	raw, err := yaml.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(dir, "script.yaml"), string(raw))
	part, chapter := "PartProbe", "ChapterProbe"
	if titleCode {
		part += " `part_Code`"
		chapter += " `chapter_Code`"
	}
	source := "# " + part + "\n\n## " + chapter + "\n\n### SectionProbe `section_probe`\n\n#### Deep `deep_probe`\n\nReferenceProbe and `para_probe`.\n\n```\nblock_probe\n```\n"
	if titleCode {
		source += "\n### Literal `#panic(\"unsafe\")_\\[]λ` with **bold `bold_probe`** and *italic `italic_probe`*\n"
	}
	if format == "org" {
		// Use Pandoc's public conversion to obtain equivalent Org without duplicating its inline grammar.
		cmd := exec.Command("pandoc", "--from=markdown", "--to=org")
		cmd.Stdin = strings.NewReader(source)
		source = commandOutput(t, cmd)
	}
	path := filepath.Join(dir, "input."+format)
	writeFile(t, path, source)
	return dir, path
}

func renderFontCLI(t *testing.T, binary, dir, source string) string {
	t.Helper()
	cmd := exec.Command(binary, "manuscript", source, filepath.Join(dir, "output.typ"))
	cmd.Dir = dir
	commandOutput(t, cmd)
	return readFile(t, filepath.Join(dir, "output.typ"))
}

func fontOutputRule(t *testing.T, output, start string) string {
	t.Helper()
	i := strings.Index(output, start)
	if i < 0 {
		t.Fatalf("missing governing output rule %q", start)
	}
	end := strings.Index(output[i:], "\n\n")
	if end < 0 {
		t.Fatalf("unterminated output rule %q", start)
	}
	return output[i : i+end]
}

func compileFontPDF(t *testing.T, dir string) string {
	t.Helper()
	requireTool(t, "typst")
	pdf := filepath.Join(dir, "output.pdf")
	commandOutput(t, exec.Command("typst", "compile", "--ignore-system-fonts", filepath.Join(dir, "output.typ"), pdf))
	return pdf
}

type fontPDFEntry struct {
	Text, Family string
	Size         int
	Bold         bool
}

func extractFontPDF(t *testing.T, pdf string) []fontPDFEntry {
	t.Helper()
	requireTool(t, "pdftohtml")
	path := filepath.Join(filepath.Dir(pdf), "output.xml")
	commandOutput(t, exec.Command("pdftohtml", "-xml", "-hidden", pdf, path))
	var doc struct {
		Pages []struct {
			Fonts []struct {
				ID     string `xml:"id,attr"`
				Family string `xml:"family,attr"`
				Size   int    `xml:"size,attr"`
			} `xml:"fontspec"`
			Texts []struct {
				Font  string `xml:"font,attr"`
				Inner string `xml:",innerxml"`
			} `xml:"text"`
		} `xml:"page"`
	}
	if err := xml.Unmarshal([]byte(readFile(t, path)), &doc); err != nil {
		t.Fatal(err)
	}
	fonts := map[string]fontPDFEntry{}
	var entries []fontPDFEntry
	for _, page := range doc.Pages {
		for _, font := range page.Fonts {
			fonts[font.ID] = fontPDFEntry{Family: font.Family, Size: font.Size}
		}
		for _, text := range page.Texts {
			entry, ok := fonts[text.Font]
			if !ok {
				t.Fatalf("unknown PDF font %s", text.Font)
			}
			entry.Text = text.Inner
			entry.Bold = strings.Contains(text.Inner, "<b>")
			entries = append(entries, entry)
		}
	}
	return entries
}
func findFontEntry(t *testing.T, entries []fontPDFEntry, marker string) fontPDFEntry {
	t.Helper()
	for _, entry := range entries {
		if strings.Contains(entry.Text, marker) {
			return entry
		}
	}
	t.Fatalf("PDF lacks %s", marker)
	return fontPDFEntry{}
}
