// ABOUTME: Verifies source-aware emphasis in native DOCX and ODT output.
// ABOUTME: Covers Org, Markdown and Fountain text plus native footnotes.
package app

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestOfficeSourceEmphasis(t *testing.T) {
	binary := officeBuild(t)
	sources := map[string]string{
		".org":      "#+TITLE: Emphasis\n* Introduction\nIntro *boldintro* /italicintro/.\n* ACT ONE\n** SCENE ONE\n**** BOB\nSay *boldword* /italicword/ **doubleword** ***combinedword***.[fn:n]\n[fn:n] Note *boldnote* /italicnote/.\n",
		".md":       "# Emphasis\n\n## ACT ONE\n### SCENE ONE\n**BOB:**\nSay **boldword** *italicword* **doubleword** ***combinedword***.[^n]\n[^n]: Note **boldnote** *italicnote*.\n",
		".fountain": "Title: Emphasis\n\n# ACT ONE\n\nINT. ROOM - DAY\n\nBOB\nSay **boldword** *italicword* **doubleword** ***combinedword***.\n",
	}
	for extension, source := range sources {
		for _, format := range []string{"docx", "odt"} {
			t.Run(extension+"/"+format, func(t *testing.T) {
				dir := t.TempDir()
				src := filepath.Join(dir, "play"+extension)
				writeAppFile(t, src, source)
				raw, diagnostic, err := officeRun(t, binary, dir, officeEnv(t, t.TempDir(), nil), "convert", src, "--to", format)
				if err != nil {
					t.Fatalf("%v: %s", err, diagnostic)
				}
				pkg := officePackage(t, raw)
				body := officeBody(t, pkg, format)
				checks := []struct {
					text         string
					bold, italic bool
				}{
					{"boldword", true, false}, {"italicword", false, true},
					{"doubleword", true, false}, {"combinedword", true, true},
				}
				if extension == ".org" {
					officeAssertEmphasis(t, pkg, format, body, "boldintro", true, false)
					officeAssertEmphasis(t, pkg, format, body, "italicintro", false, true)
				}
				for _, check := range checks {
					officeAssertEmphasis(t, pkg, format, body, check.text, check.bold, check.italic)
				}
				if extension != ".fountain" {
					notes := body
					if format == "docx" {
						notes = pkg["word/footnotes.xml"]
					}
					if notes == nil {
						t.Fatal("missing native footnotes")
					}
					officeAssertEmphasis(t, pkg, format, notes, "boldnote", true, false)
					officeAssertEmphasis(t, pkg, format, notes, "italicnote", false, true)
				}
			})
		}
	}
}

func officeAssertEmphasis(t *testing.T, pkg map[string]*officeXML, format string, root *officeXML, text string, bold, italic bool) {
	t.Helper()
	kind := "r"
	if format == "odt" {
		kind = "span"
	}
	for _, n := range root.all(kind) {
		if n.text() != text {
			continue
		}
		gotBold, gotItalic := false, false
		if format == "docx" {
			for _, p := range n.all("b") {
				gotBold = p.attr("val") != "false" && p.attr("val") != "0"
			}
			for _, p := range n.all("i") {
				gotItalic = p.attr("val") != "false" && p.attr("val") != "0"
			}
		} else {
			layers := officeEffective(t, pkg, format, n)
			weight, _ := officeEffectiveAttr(layers, "text-properties", "font-weight")
			style, _ := officeEffectiveAttr(layers, "text-properties", "font-style")
			gotBold, gotItalic = weight == "bold", style == "italic"
		}
		if gotBold != bold || gotItalic != italic {
			t.Errorf("%s: bold=%v italic=%v; want bold=%v italic=%v", text, gotBold, gotItalic, bold, italic)
		}
		return
	}
	t.Errorf("missing separately styled text %q in %s", text, strings.TrimSpace(root.text()))
}
