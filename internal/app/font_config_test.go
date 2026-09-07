// ABOUTME: Public-command regressions for W039 font validation and rendering.
// ABOUTME: Verifies role-scoped Typst and the no-output failure boundary.
package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
