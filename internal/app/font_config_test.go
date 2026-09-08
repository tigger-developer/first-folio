// ABOUTME: Public-command regressions for W039 font validation and rendering.
// ABOUTME: Verifies role-scoped Typst and the no-output failure boundary.
package app

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	folio "github.com/tigger-developer/first-folio"
)

var scriptFontRolePaths = []string{
	"folio.font",
	"folio.heading.font",
	"folio.title-page.title.font",
	"folio.title-page.subtitle.font",
	"folio.title-page.author.font",
	"folio.title-page.date.font",
	"folio.title-page.version.font",
	"folio.positioning.speech.speaker.font",
	"folio.positioning.speech.speech-instruction.font",
	"folio.positioning.speech.dialogue.font",
	"folio.positioning.stage-direction.font",
	"folio.positioning.transition.font",
	"folio.positioning.frontmatter.header.font",
	"folio.positioning.act-header.font",
	"folio.positioning.scene-header.font",
}

var renderedFontProperties = []struct {
	name, value, marker string
}{
	{"family", "W039 Family", `"W039 Family"`},
	{"size", "13.5pt", "13.5pt"},
	{"weight", "643", "643"},
	{"stretch", "117.5%", "117.5%"},
	{"style", "oblique", `"oblique"`},
	{"letter-spacing", "-0.03em", "-0.03em"},
}

func TestRT039_1EveryScriptFontPropertyReachesOnlyItsRole(t *testing.T) {
	for _, path := range scriptFontRolePaths {
		for _, property := range renderedFontProperties {
			t.Run(strings.TrimPrefix(path, "folio.")+"/"+property.name, func(t *testing.T) {
				t.Setenv("HOME", t.TempDir())
				dir := t.TempDir()
				source := filepath.Join(dir, "play.org")
				target := filepath.Join(dir, "play.typ")
				writeAppFile(t, source, strings.Join([]string{
					"#+TITLE: Font Roles",
					"#+SUBTITLE: Every Property",
					"#+AUTHOR: Example Author",
					"#+DATE: 7 September 2026",
					"#+VERSION: Draft 1",
					"* ACT ONE",
					"** Scene One",
					"*** A prop :prop:",
					"*** CUT TO: :transition:",
					"**** CÁIT quietly",
					"Hello.",
					"",
				}, "\n"))
				writeAppFile(t, filepath.Join(dir, "script.yaml"), oneFontPropertyYAML(path, property.name, property.value))

				status, _, stderr := runApp(t, "convert", source, target)
				if status != 0 {
					t.Fatal(stderr)
				}
				typst := readAppFile(t, target)
				if count := strings.Count(typst, property.marker); count != 1 {
					t.Fatalf("%s marker %q occurs %d times, want exactly once in its role", path, property.marker, count)
				}
			})
		}
	}
}

func TestRT039_1ScriptFontBlockEmitsAllPropertiesForSelectedRole(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	source := filepath.Join(dir, "play.org")
	target := filepath.Join(dir, "play.typ")
	writeAppFile(t, source, "#+TITLE: Font Roles\n* ACT ONE\n** Scene One\n**** CÁIT\nHello.\n")
	writeAppFile(t, filepath.Join(dir, "script.yaml"), `folio:
  positioning:
    act-header:
      font:
        family: W039 Act
        size: 13.5pt
        weight: 650
        stretch: 117.5%
        style: oblique
        letter-spacing: -0.03em
`)
	status, _, stderr := runApp(t, "convert", source, target)
	if status != 0 {
		t.Fatal(stderr)
	}
	typst := readAppFile(t, target)
	assertAppContains(t, typst, `#text(font: "W039 Act", size: 13.5pt, weight: 650, stretch: 117.5%, style: "oblique", tracking: -0.03em)`)
	assertAppContains(t, typst, `font: "Libertinus Serif", size: 12pt, weight: "bold"`)
}

func TestRT039_2CopiedBritishBaseLeavesPublicOutputsUnchanged(t *testing.T) {
	toolHome := os.Getenv("HOME")
	t.Setenv("HOME", t.TempDir())
	base, err := folio.Assets.ReadFile("presets/british.yaml")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("script", func(t *testing.T) {
		plainDir, copiedDir := t.TempDir(), t.TempDir()
		source := "#+TITLE: W039\n* ACT ONE\n**** CÁIT\nHello.\n"
		plainSource := filepath.Join(plainDir, "play.org")
		copiedSource := filepath.Join(copiedDir, "play.org")
		writeAppFile(t, plainSource, source)
		writeAppFile(t, copiedSource, source)
		writeAppFile(t, filepath.Join(copiedDir, "script.yaml"), string(base))
		plainOutput := filepath.Join(plainDir, "play.typ")
		copiedOutput := filepath.Join(copiedDir, "play.typ")
		for _, args := range [][]string{{"convert", plainSource, plainOutput}, {"convert", copiedSource, copiedOutput}} {
			status, _, stderr := runApp(t, args...)
			if status != 0 {
				t.Fatal(stderr)
			}
		}
		if plain, copied := readAppFile(t, plainOutput), readAppFile(t, copiedOutput); plain != copied {
			t.Fatal("script output changed when the British base was copied locally")
		}
	})

	t.Run("manuscript", func(t *testing.T) {
		plainDir, copiedDir := t.TempDir(), t.TempDir()
		source := "---\ntitle: W039\nauthor: Example Author\n---\n\n## Chapter One\n\nBody.\n"
		plainSource := filepath.Join(plainDir, "chapter.md")
		copiedSource := filepath.Join(copiedDir, "chapter.md")
		writeAppFile(t, plainSource, source)
		writeAppFile(t, copiedSource, source)
		writeAppFile(t, filepath.Join(copiedDir, "script.yaml"), string(base))
		plainOutput := filepath.Join(plainDir, "manuscript.typ")
		copiedOutput := filepath.Join(copiedDir, "manuscript.typ")
		for _, args := range [][]string{{"manuscript", plainSource, plainOutput}, {"manuscript", copiedSource, copiedOutput}} {
			status, _, stderr := runApp(t, args...)
			if status != 0 {
				t.Fatal(stderr)
			}
		}
		if plain, copied := readAppFile(t, plainOutput), readAppFile(t, copiedOutput); plain != copied {
			t.Fatal("manuscript output changed when the British base was copied locally")
		}
	})

	t.Run("letter", func(t *testing.T) {
		for _, tool := range []string{"typst", "pdf-to-png"} {
			if _, err := exec.LookPath(tool); err != nil {
				t.Skipf("%s is not installed", tool)
			}
		}
		plainDir, copiedDir := t.TempDir(), t.TempDir()
		source := "#+AUTHOR: Example Author\n* Letters :letter:\n** Subject :subject:\nBody.\n*** Recipient :to:\n**** Example :org:\n"
		plainSource := filepath.Join(plainDir, "letters.org")
		copiedSource := filepath.Join(copiedDir, "letters.org")
		writeAppFile(t, plainSource, source)
		writeAppFile(t, copiedSource, source)
		writeAppFile(t, filepath.Join(copiedDir, "script.yaml"), string(base))
		for _, args := range [][]string{{"letter", plainSource, "--dir", plainDir, "--prefix", "w039"}, {"letter", copiedSource, "--dir", copiedDir, "--prefix", "w039"}} {
			status, _, stderr := runApp(t, args...)
			if status != 0 {
				t.Fatal(stderr)
			}
		}
		plainPDF := singlePDF(t, plainDir)
		copiedPDF := singlePDF(t, copiedDir)
		if !rasterPixelsEqual(rasterizePDF(t, plainPDF, toolHome), rasterizePDF(t, copiedPDF, toolHome)) {
			t.Fatal("letter PDF changed when the British base was copied locally")
		}
	})
}

func TestRT039_3ManuscriptFontMergesSameRoleAcrossPublicLayers(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	t.Setenv("HOME", home)
	writeAppFile(t, filepath.Join(home, ".config", "first-folio", "script.yaml"), `folio:
  manuscript:
    page-header:
      font:
        family: W039 Layered Header
        weight: 643
`)
	writeAppFile(t, filepath.Join(project, "script.yaml"), `folio:
  style: us
  manuscript:
    font:
      family: W039 Body Must Not Leak
    page-header:
      font:
        size: 13.5pt
`)
	writeAppFile(t, filepath.Join(project, "script-us.yaml"), `folio:
  manuscript:
    page-header:
      font:
        stretch: 117.5%
`)
	source := filepath.Join(project, "chapter.md")
	target := filepath.Join(project, "manuscript.typ")
	writeAppFile(t, source, "---\ntitle: W039\nauthor: Example Author\n---\n\n## Chapter One\n\nBody.\n")
	status, _, stderr := runApp(t, "manuscript", source, target)
	if status != 0 {
		t.Fatal(stderr)
	}
	typst := readAppFile(t, target)
	assertAppContains(t, typst, `font: "W039 Layered Header", size: 13.5pt, weight: 643, stretch: 117.5%, style: "normal", tracking: 0em`)
	assertAppContains(t, typst, `font: "W039 Body Must Not Leak", size: 10pt`)
	if strings.Contains(typst, `font: "W039 Body Must Not Leak", size: 13.5pt`) {
		t.Fatal("manuscript body font leaked into the page-header role")
	}
}

func TestRT039_5PublicCommandsRejectFontErrorsBeforeOutput(t *testing.T) {
	tests := []struct {
		name, config, wantPath string
		args                   func(string, string) []string
	}{
		{
			name: "script retired scalar", config: "folio:\n  font: Old Family\n", wantPath: "folio.font",
			args: func(source, output string) []string { return []string{"convert", source, output} },
		},
		{
			name: "script retired heading", config: "folio:\n  heading-font: Old Heading\n", wantPath: "folio.heading-font",
			args: func(source, output string) []string { return []string{"convert", source, output} },
		},
		{
			name: "script non-scalar block", config: "folio:\n  font: [Serif]\n", wantPath: "folio.font",
			args: func(source, output string) []string { return []string{"convert", source, output} },
		},
		{
			name: "script empty family", config: "folio:\n  font:\n    family: '   '\n", wantPath: "folio.font.family",
			args: func(source, output string) []string { return []string{"convert", source, output} },
		},
		{
			name: "script null family", config: "folio:\n  font:\n    family: null\n", wantPath: "folio.font.family",
			args: func(source, output string) []string { return []string{"convert", source, output} },
		},
		{
			name: "script invalid weight", config: "folio:\n  font:\n    weight: 900.5\n", wantPath: "folio.font.weight",
			args: func(source, output string) []string { return []string{"convert", source, output} },
		},
		{
			name: "script invalid stretch", config: "folio:\n  font:\n    stretch: -1\n", wantPath: "folio.font.stretch",
			args: func(source, output string) []string { return []string{"convert", source, output} },
		},
		{
			name: "script invalid style", config: "folio:\n  font:\n    style: bold\n", wantPath: "folio.font.style",
			args: func(source, output string) []string { return []string{"convert", source, output} },
		},
		{
			name: "script invalid letter spacing", config: "folio:\n  font:\n    letter-spacing: 2px\n", wantPath: "folio.font.letter-spacing",
			args: func(source, output string) []string { return []string{"convert", source, output} },
		},
		{
			name: "manuscript invalid size", config: "folio:\n  manuscript:\n    page-header:\n      font:\n        size: 0pt\n", wantPath: "folio.manuscript.page-header.font.size",
			args: func(source, output string) []string { return []string{"manuscript", source, output} },
		},
		{
			name: "letter unknown property", config: "folio:\n  letter:\n    font:\n      decoration: underline\n", wantPath: "folio.letter.font.decoration",
			args: func(source, output string) []string {
				return []string{"letter", source, "--dir", output, "--prefix", "w039"}
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			dir := t.TempDir()
			source := filepath.Join(dir, "source.org")
			output := filepath.Join(dir, "output")
			if strings.HasPrefix(test.name, "manuscript") {
				source = filepath.Join(dir, "source.md")
				output += ".typ"
				writeAppFile(t, source, "## Chapter One\n\nBody.\n")
			} else if strings.HasPrefix(test.name, "letter") {
				writeAppFile(t, source, "#+AUTHOR: Example Author\n* Letters :letter:\n** Subject :subject:\nBody.\n*** Recipient :to:\n**** Example :org:\n")
			} else {
				output += ".typ"
				writeAppFile(t, source, "* ACT ONE\n**** CÁIT\nHello.\n")
			}
			writeAppFile(t, filepath.Join(dir, "script.yaml"), test.config)
			status, _, stderr := runApp(t, test.args(source, output)...)
			if status == 0 || !strings.Contains(stderr, test.wantPath) {
				t.Fatalf("status %d stderr %q, want path %s", status, stderr, test.wantPath)
			}
			if _, err := os.Stat(output); !os.IsNotExist(err) {
				t.Fatalf("invalid configuration created output %s", output)
			}
		})
	}
}

func assertAppContains(t *testing.T, text, fragment string) {
	t.Helper()
	if !strings.Contains(text, fragment) {
		t.Fatalf("missing %q in:\n%s", fragment, text)
	}
}

func oneFontPropertyYAML(path, property, value string) string {
	var output strings.Builder
	for depth, part := range strings.Split(path, ".") {
		fmt.Fprintf(&output, "%s%s:\n", strings.Repeat("  ", depth), part)
	}
	fmt.Fprintf(&output, "%s%s: %s\n", strings.Repeat("  ", len(strings.Split(path, "."))), property, value)
	return output.String()
}

func singlePDF(t *testing.T, dir string) string {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(dir, "*.pdf"))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 1 {
		t.Fatalf("found %d PDF outputs in %s, want one", len(paths), dir)
	}
	raw, err := os.ReadFile(paths[0])
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(raw, []byte("%PDF")) {
		t.Fatalf("output %s is not a PDF", paths[0])
	}
	return paths[0]
}

func rasterizePDF(t *testing.T, path string, toolHome string) image.Image {
	t.Helper()
	cmd := exec.Command("pdf-to-png", filepath.Base(path), "120")
	cmd.Dir = filepath.Dir(path)
	cmd.Env = []string{"HOME=" + toolHome}
	for _, value := range os.Environ() {
		if !strings.HasPrefix(value, "HOME=") {
			cmd.Env = append(cmd.Env, value)
		}
	}
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("rasterizing %s: %v\n%s", path, err, output)
	}
	name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	paths, err := filepath.Glob(filepath.Join(filepath.Dir(path), name+"-*.png"))
	if err != nil || len(paths) != 1 {
		t.Fatalf("raster output for %s = %v, %v; want one page", path, paths, err)
	}
	file, err := os.Open(paths[0])
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	raster, err := png.Decode(file)
	if err != nil {
		t.Fatal(err)
	}
	return raster
}
