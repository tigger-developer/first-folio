// ABOUTME: Maintains the legacy conversion and configuration parity corpus in Go.
// ABOUTME: Replaces shell regression loops with public application-boundary tables.
package app

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/tigger-developer/first-folio/internal/play"
)

var scriptRoundTripSources = map[play.Format]string{
	play.FormatOrg: `#+TITLE: Matrix Play
#+AUTHOR: Taḋg Paul
* ACT ONE
** SCENE ONE
*** Night.
**** CÁIT softly
Hello.
`,
	play.FormatMarkdown: `# Matrix Play

*by Taḋg Paul*

## ACT ONE

### SCENE ONE

*Night.*

**CÁIT:** *(softly)*
Hello.
`,
	play.FormatFountain: `Title: Matrix Play
Author: Taḋg Paul

> **ACT ONE** <

.SCENE ONE

Night.

CÁIT
(softly)
Hello.
`,
}

var scriptFormatExtensions = map[play.Format]string{
	play.FormatOrg:      ".org",
	play.FormatMarkdown: ".md",
	play.FormatFountain: ".fountain",
}

func TestConversionMatrixParity(t *testing.T) {
	sources := map[string]string{
		"org": `#+TITLE: Matrix Play
#+AUTHOR: Taḋg Paul
* ACT ONE
** Scene One
*** Night.
**** CÁIT softly
Hello.
***** CUT TO
`,
		"md": `# Matrix Play

*by Taḋg Paul*

## ACT ONE

### Scene One

*Night.*

**CÁIT:** *(softly)*
Hello.

> CUT TO
`,
		"fountain": `Title: Matrix Play
Author: Taḋg Paul

> **ACT ONE** <

.SCENE ONE

Night.

CÁIT
(softly)
Hello.

CUT TO:
`,
	}
	extensions := map[string]string{"org": ".org", "md": ".md", "fountain": ".fountain"}
	checks := map[string][]string{
		"org":      {"#+TITLE: Matrix Play", "* ACT ONE", "**** CÁIT softly", "CUT TO"},
		"md":       {"# Matrix Play", "## ACT ONE", "**CÁIT:** *(softly)*", "> CUT TO"},
		"fountain": {"Title: Matrix Play", "> **ACT ONE** <", "CÁIT", "CUT TO"},
	}
	for sourceName, sourceText := range sources {
		for targetName, extension := range extensions {
			t.Run(sourceName+" to "+targetName, func(t *testing.T) {
				dir := t.TempDir()
				t.Setenv("HOME", t.TempDir())
				source := filepath.Join(dir, "source"+extensions[sourceName])
				target := filepath.Join(dir, "target"+extension)
				writeAppFile(t, source, sourceText)
				status, stdout, stderr := runApp(t, "convert", source, target)
				if status != 0 {
					t.Fatalf("status %d\nstdout:%s\nstderr:%s", status, stdout, stderr)
				}
				output := readAppFile(t, target)
				for _, fragment := range checks[targetName] {
					if !strings.Contains(output, fragment) {
						t.Errorf("%s -> %s missing %q:\n%s", sourceName, targetName, fragment, output)
					}
				}
			})
		}
	}
}

func TestScriptFormatRoundTrips(t *testing.T) {
	routes := [][2]play.Format{
		{play.FormatOrg, play.FormatMarkdown},
		{play.FormatMarkdown, play.FormatOrg},
		{play.FormatOrg, play.FormatFountain},
		{play.FormatFountain, play.FormatOrg},
		{play.FormatMarkdown, play.FormatFountain},
		{play.FormatFountain, play.FormatMarkdown},
	}
	for _, route := range routes {
		name := string(route[0]) + " via " + string(route[1])
		t.Run(name, func(t *testing.T) {
			returned := runScriptRoundTrip(t, route[0], route[1])
			want := parseScriptDocument(t, route[0], scriptRoundTripSources[route[0]])
			got := parseScriptDocument(t, route[0], returned)
			if !reflect.DeepEqual(got, want) {
				t.Errorf("semantic document changed\nwant: %#v\n got: %#v\nreturned source:\n%s", want, got, returned)
			}
		})
	}
}

func runScriptRoundTrip(t *testing.T, sourceFormat play.Format, intermediateFormat play.Format) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	source := filepath.Join(dir, "source"+scriptFormatExtensions[sourceFormat])
	intermediate := filepath.Join(dir, "intermediate"+scriptFormatExtensions[intermediateFormat])
	returned := filepath.Join(dir, "returned"+scriptFormatExtensions[sourceFormat])
	writeAppFile(t, source, scriptRoundTripSources[sourceFormat])
	for _, conversion := range [][2]string{{source, intermediate}, {intermediate, returned}} {
		status, stdout, stderr := runApp(t, "convert", conversion[0], conversion[1])
		if status != 0 {
			t.Fatalf("converting %s to %s: status %d\nstdout:%s\nstderr:%s", conversion[0], conversion[1], status, stdout, stderr)
		}
	}
	return readAppFile(t, returned)
}

func parseScriptDocument(t *testing.T, format play.Format, source string) play.Document {
	t.Helper()
	doc, _, err := play.Parse(format, source, "round-trip"+scriptFormatExtensions[format])
	if err != nil {
		t.Fatal(err)
	}
	return doc
}

func TestScriptLayoutConfigurationParity(t *testing.T) {
	sourceText := `#+TITLE: Config Play
#+SUBTITLE: A Subtitle
#+AUTHOR: Example Author
#+DATE: 2026-01-01
#+VERSION: Draft 1
* Introduction
Intro text.
* ACT ONE
** Scene One
*** Night.
**** CÁIT softly
Hello.[fn:note]
***** CUT TO
[fn:note] Note text.
`
	tests := []struct {
		name   string
		yaml   string
		want   []string
		absent []string
	}{
		{"root layout", "folio:\n  font: Test Font\n  font-size: 10pt\n  page: a5\n  margin: 18mm\n", []string{`font: "Test Font"`, "size: 10pt", `paper: "a5"`, "margin: 18mm"}, nil},
		{"speaker", "folio:\n  positioning:\n    speech:\n      speaker:\n        bold: false\n        suffix: \"\"\n", []string{`weight: "regular"`, "#upper[#name]"}, []string{"#name:"}},
		{"direction", "folio:\n  positioning:\n    stage-direction:\n      italic: false\n      align: center\n      space-before: 3em\n", []string{"#align(center)[#body]", "above: 3em"}, nil},
		{"headers", "folio:\n  positioning:\n    act-header:\n      align: left\n      font-size: 18pt\n      bold: false\n      case-transform: upper\n    scene-header:\n      align: center\n      font-size: 14pt\n      case-transform: upper\n", []string{"align(left)", "size: 18pt", `weight: "regular"`, "#upper[#title]", "align(center)", "size: 14pt"}, nil},
		{"title", "folio:\n  title-page:\n    title:\n      font-size: 18pt\n      bold: false\n      italic: true\n    subtitle:\n      italic: false\n    author:\n      prefix: \"\"\n", []string{"size: 18pt", `style: "italic"`, "A Subtitle", "Example Author"}, []string{"by Example Author"}},
		{"filtered", "render:\n  stage-directions: false\n  frontmatter: false\n  footnotes: false\n  transitions: false\n", []string{"#dialogue"}, []string{"Night.", "Intro text.", "#footnote", "CUT TO"}},
		{"us style", "folio:\n  style: us\n", []string{"spacing: 0em", "#align(left)"}, []string{"columns: (7em, 1fr)"}},
		{"screenplay style", "folio:\n  style: screenplay\n", []string{`font: "Courier Prime"`, "#v(40%)", `weight: "regular"`, "pad(left: 25.4mm, right: 25.4mm)", "#align(center)[_(#direction)_]", "#pad(left: 101.6mm)[#align(right)", "#align(left)[#body]", "#upper[#title]", ")[Written by]", "#v(0.3em)", ")[Example Author]"}, []string{"columns: (7em, 1fr)", "#align(left)[_#body _]", "Written by\\nExample Author"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			t.Setenv("HOME", t.TempDir())
			source := filepath.Join(dir, "play.org")
			target := filepath.Join(dir, "play.typ")
			writeAppFile(t, source, sourceText)
			writeAppFile(t, filepath.Join(dir, "script.yaml"), tt.yaml)
			status, stdout, stderr := runApp(t, "convert", source, target)
			if status != 0 {
				t.Fatalf("status %d\nstdout:%s\nstderr:%s", status, stdout, stderr)
			}
			output := readAppFile(t, target)
			for _, fragment := range tt.want {
				if !strings.Contains(output, fragment) {
					t.Errorf("missing %q:\n%s", fragment, output)
				}
			}
			for _, fragment := range tt.absent {
				if strings.Contains(output, fragment) {
					t.Errorf("unexpected %q:\n%s", fragment, output)
				}
			}
		})
	}
}

func TestConfigLayerMatrixParity(t *testing.T) {
	home := t.TempDir()
	dir := t.TempDir()
	t.Setenv("HOME", home)
	source := filepath.Join(dir, "play.org")
	writeAppFile(t, source, "#+TITLE: Source\n* ACT ONE\n**** CÁIT\nHello.\n")
	writeAppFile(t, filepath.Join(home, ".config", "first-folio", "script.yaml"), "title: Global\nfolio:\n  font: Global Font\n  margin: 30mm\n")
	writeAppFile(t, filepath.Join(home, ".config", "first-folio", "script-us.yaml"), "folio:\n  margin: 28mm\n")
	writeAppFile(t, filepath.Join(dir, "script.yaml"), "title: Local\nfolio:\n  style: us\n  font: Local Font\n")
	writeAppFile(t, filepath.Join(dir, "script-us.yaml"), "folio:\n  page: a5\n")
	target := filepath.Join(dir, "out.typ")
	status, _, stderr := runApp(t, "convert", source, target, "--font", "CLI Font")
	if status != 0 {
		t.Fatal(stderr)
	}
	output := readAppFile(t, target)
	for _, fragment := range []string{"Local", `font: "CLI Font"`, "margin: 28mm", `paper: "a5"`} {
		if !strings.Contains(output, fragment) {
			t.Errorf("layered output missing %q:\n%s", fragment, output)
		}
	}
}

func TestConversionDiagnosticsParity(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	fountain := filepath.Join(dir, "lossy.fountain")
	writeAppFile(t, fountain, "Title: Test\n\n= synopsis\n===\n\n.CUT\n")
	status, _, stderr := runApp(t, "convert", fountain, filepath.Join(dir, "out.md"))
	if status != 0 || !strings.Contains(stderr, "dropping synopsis") || !strings.Contains(stderr, "dropping page break") {
		t.Fatalf("status %d stderr %q", status, stderr)
	}
	invalid := filepath.Join(dir, "invalid.org")
	if err := os.WriteFile(invalid, []byte{0xff, 0xfe}, 0o644); err != nil {
		t.Fatal(err)
	}
	status, _, stderr = runApp(t, "convert", invalid, filepath.Join(dir, "invalid.md"))
	if status == 0 || !strings.Contains(stderr, "invalid encoding") {
		t.Fatalf("status %d stderr %q", status, stderr)
	}
	for _, extension := range []string{".pdf", ".typ"} {
		path := filepath.Join(dir, "input"+extension)
		writeAppFile(t, path, "not actually binary")
		status, _, stderr = runApp(t, "convert", path, filepath.Join(dir, fmt.Sprintf("from-%s.md", extension[1:])))
		if status == 0 || !strings.Contains(stderr, "write-only") {
			t.Errorf("%s status %d stderr %q", extension, status, stderr)
		}
	}
}
