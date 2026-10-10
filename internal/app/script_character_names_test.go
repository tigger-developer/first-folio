// ABOUTME: Verifies cast-based name capitalisation through public script rendering.
// ABOUTME: Keeps dialogue intact and exercises visible and hidden character tables.
package app

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

func TestScriptCharacterNamesFromTable(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	source := filepath.Join(dir, "play.org")
	target := filepath.Join(dir, "play.typ")
	writeAppFile(t, source, `#+TITLE: Names
* CHARACTERS
| Tom | Inpatient |
* Act One
** Scene One
*** Tom closes his eyes.
**** Tom
Good morning, Tom.
`)
	status, _, stderr := runApp(t, "convert", source, target)
	if status != 0 {
		t.Fatalf("status %d: %s", status, stderr)
	}
	output := readAppFile(t, target)
	if !strings.Contains(output, "#stage-direction[TOM closes his eyes.]") {
		t.Fatalf("character name was not capitalized:\n%s", output)
	}
	if !strings.Contains(output, `#dialogue("Tom")[Good morning, Tom.]`) {
		t.Fatal("dialogue or speaker cue changed")
	}
}

func TestScriptCharacterAliases(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	for _, style := range []string{"british", "us", "screenplay"} {
		for _, visible := range []bool{true, false} {
			t.Run(fmt.Sprintf("%s/table=%t", style, visible), func(t *testing.T) {
				dir := t.TempDir()
				source := filepath.Join(dir, "play.org")
				target := filepath.Join(dir, "play.typ")
				writeAppFile(t, filepath.Join(dir, "script.yaml"), fmt.Sprintf("render:\n  character-table: %t\n", visible))
				writeAppFile(t, source, `* CHARACTERS
| ALISON (NURSE ALISON) | A nurse |
| THOMAS (TOM) | Main character |
| MARGARET (MAGS, MARG.) | A visitor |
| Mairéad | A patient |
* Act One
** Scene One
*** Nurse Alison greets Tom, Mags and Marg. Mairéad enters. _Alison exits_.
**** Alison
Nurse Alison greets Tom, Mags and Marg. Mairéad enters.
`)
				status, _, stderr := runApp(t, "convert", source, target, "--style", style)
				if status != 0 {
					t.Fatalf("status %d: %s", status, stderr)
				}
				output := readAppFile(t, target)
				want := "#stage-direction[NURSE ALISON greets TOM, MAGS and MARG. MAIRÉAD enters. #underline[ALISON exits].]"
				if !strings.Contains(output, want) {
					t.Fatalf("missing %q", want)
				}
				if !strings.Contains(output, `#dialogue("Alison")[Nurse Alison greets Tom, Mags and Marg. Mairéad enters.]`) {
					t.Fatal("dialogue or speaker cue changed")
				}
				if strings.Contains(output, "(NURSE ALISON)") || strings.Contains(output, "(MAGS, MARG.)") {
					t.Fatal("aliases leaked into rendered cast")
				}
				if strings.Contains(output, "#table(columns:") != visible {
					t.Fatal("table visibility not respected")
				}
				if visible && !strings.Contains(output, "[ALISON], [A nurse],") {
					t.Fatal("primary name missing from cast")
				}
			})
		}
	}
}

func TestMarkdownCastAliasesConvert(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	source := filepath.Join(dir, "play.md")
	writeAppFile(t, source, `# Names

| Character | Description |
|-----------|-------------|
| ALISON (NURSE ALISON) | A nurse |

## Act One

*Nurse Alison greets Alison.*

**ALISON:**
Nurse Alison greets Alison.
`)
	for _, format := range []string{"org", "md", "fountain"} {
		status, output, stderr := runApp(t, "convert", source, "--to", format)
		if status != 0 {
			t.Fatalf("%s: status %d: %s", format, status, stderr)
		}
		if !strings.Contains(output, "NURSE ALISON greets ALISON.") {
			t.Fatalf("%s: missing capitalized direction", format)
		}
		if !strings.Contains(output, "Nurse Alison greets Alison.") {
			t.Fatalf("%s: dialogue changed", format)
		}
		if strings.Contains(output, "ALISON (NURSE ALISON)") != (format != "fountain") {
			t.Fatalf("%s: incorrect alias preservation", format)
		}
	}
	target := filepath.Join(dir, "play.typ")
	status, _, stderr := runApp(t, "convert", source, target)
	if status != 0 {
		t.Fatalf("status %d: %s", status, stderr)
	}
	if !strings.Contains(readAppFile(t, target), "#stage-direction[NURSE ALISON greets ALISON.]") {
		t.Fatal("Markdown render missing capitalized direction")
	}
}
