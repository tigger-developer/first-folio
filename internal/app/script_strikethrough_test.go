// ABOUTME: Verifies source-aware script strikethrough across rendered output paths.
// ABOUTME: Protects Org plus and Markdown double-tilde syntax without claiming single tilde.
package app

import (
	"path/filepath"
	"strings"
	"testing"
)

var scriptStrikethroughSources = map[string]string{
	"org": "* ACT ONE\n** SCENE ONE\n**** BOB\nKeep +deletedword+ and [fn:n].\n[fn:n] Note +deletednote+.\n",
	"md":  "# Strikethrough\n\n## ACT ONE\n### SCENE ONE\n**BOB:**\nKeep ~~deletedword~~ and ~singleliteral~ and [^n].\n[^n]: Note ~~deletednote~~.\n",
}

func TestRT070_1ScriptStrikethroughReachesTypst(t *testing.T) {
	for sourceFormat, sourceText := range scriptStrikethroughSources {
		t.Run(sourceFormat, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			dir := t.TempDir()
			source := filepath.Join(dir, "play."+sourceFormat)
			target := filepath.Join(dir, "play.typ")
			writeAppFile(t, source, sourceText)

			status, stdout, stderr := runApp(t, "convert", source, target)
			if status != 0 {
				t.Fatalf("status %d; stdout=%q stderr=%q", status, stdout, stderr)
			}
			typst := readAppFile(t, target)
			for _, marker := range []string{"deletedword", "deletednote"} {
				if !strings.Contains(typst, "#strike["+marker+"]") {
					t.Errorf("Typst lost strikethrough for %q: %s", marker, typst)
				}
			}
			if sourceFormat == "md" && !strings.Contains(typst, "~singleliteral~") {
				t.Errorf("Markdown single tildes did not remain literal: %s", typst)
			}
		})
	}
}

func TestRT070_2ScriptStrikethroughReachesOfficeRuns(t *testing.T) {
	binary := officeBuild(t)
	for sourceFormat, sourceText := range scriptStrikethroughSources {
		for _, format := range []string{"docx", "odt"} {
			t.Run(sourceFormat+"/"+format, func(t *testing.T) {
				dir := t.TempDir()
				source := filepath.Join(dir, "play."+sourceFormat)
				writeAppFile(t, source, sourceText)
				raw, diagnostic, err := officeRun(t, binary, dir, officeEnv(t, t.TempDir(), nil), "convert", source, "--to", format)
				if err != nil {
					t.Fatalf("%v: %s", err, diagnostic)
				}
				pkg := officePackage(t, raw)
				body := officeBody(t, pkg, format)
				officeAssertStrikethrough(t, pkg, format, body, "deletedword")
				notes := body
				if format == "docx" {
					notes = pkg["word/footnotes.xml"]
				}
				if notes == nil {
					t.Fatal("missing native footnotes")
				}
				officeAssertStrikethrough(t, pkg, format, notes, "deletednote")
				if sourceFormat == "md" && !strings.Contains(body.text(), "~singleliteral~") {
					t.Errorf("Markdown single tildes did not remain literal: %s", body.text())
				}
			})
		}
	}
}

func officeAssertStrikethrough(t *testing.T, pkg map[string]*officeXML, format string, root *officeXML, marker string) {
	t.Helper()
	kind := "r"
	if format == "odt" {
		kind = "span"
	}
	for _, n := range root.all(kind) {
		if n.text() != marker {
			continue
		}
		if format == "docx" {
			for _, strike := range n.all("strike") {
				if value := strike.attr("val"); value != "false" && value != "0" {
					return
				}
			}
		} else {
			value, exists := officeEffectiveAttr(officeEffective(t, pkg, format, n), "text-properties", "text-line-through-style")
			if exists && value != "none" {
				return
			}
		}
		t.Errorf("%s lacks native strikethrough in %s", marker, format)
		return
	}
	t.Errorf("missing separately styled text %q in %s", marker, strings.TrimSpace(root.text()))
}
