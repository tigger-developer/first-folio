// ABOUTME: Checks native contents inclusion, links, leaders and live page-reference switches.
// ABOUTME: Uses unique visible heading phrases to distinguish body headings from generated entries.
package app

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestRT063_5_OfficeContentsSelection(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		for _, excluded := range []string{"none", "parts", "chapters", "sections"} {
			t.Run(format+"/exclude-"+excluded, func(t *testing.T) {
				t.Setenv("HOME", t.TempDir())
				dir := t.TempDir()
				cfg := "folio:\n  manuscript:\n    title-page:\n      enabled: false\n    toc:\n      enabled: true\n      title: Selection contents\n"
				for _, level := range []string{"parts", "chapters", "sections"} {
					value := "true"
					if level == excluded {
						value = "false"
					}
					cfg += "      include-" + level + ": " + value + "\n"
				}
				proseSource(t, dir, "script.yaml", cfg)
				src := proseSource(t, dir, "source.md", "# Selected part phrase\n\n## Selected chapter phrase\n\n### Selected section phrase\n\nBODY-SENTINEL\n")
				pkg := proseConvert(t, src, filepath.Join(dir, "output."+format))
				text := officeBody(t, pkg, format).text()
				for _, heading := range []struct{ level, text string }{
					{"parts", "Selected part phrase"}, {"chapters", "Selected chapter phrase"}, {"sections", "Selected section phrase"},
				} {
					want := 2 // Body heading plus its editable contents entry.
					if heading.level == excluded {
						want = 1
					}
					if got := strings.Count(text, heading.text); got != want {
						t.Errorf("%s occurs %d times; want %d for excluded=%s", heading.text, got, want, excluded)
					}
				}
			})
		}
	}
}

func TestRT063_5_OfficeContentsPresentationSwitches(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		for _, disabled := range []string{"none", "links", "dot-leaders", "page-numbers", "all"} {
			t.Run(format+"/disable-"+disabled, func(t *testing.T) {
				t.Setenv("HOME", t.TempDir())
				dir := t.TempDir()
				cfg := "folio:\n  manuscript:\n    title-page:\n      enabled: false\n    toc:\n      enabled: true\n      include-parts: false\n      include-chapters: true\n"
				values := map[string]bool{}
				for _, key := range []string{"links", "dot-leaders", "page-numbers"} {
					values[key] = disabled != key && disabled != "all"
					value := "false"
					if values[key] {
						value = "true"
					}
					cfg += "      " + key + ": " + value + "\n"
				}
				proseSource(t, dir, "script.yaml", cfg)
				src := proseSource(t, dir, "source.md", "## Contents entry phrase\n\nBODY-SENTINEL\n")
				pkg := proseConvert(t, src, filepath.Join(dir, "output."+format))
				body := officeBody(t, pkg, format)
				linkKind := "hyperlink"
				if format == "odt" {
					linkKind = "a"
				}
				if (len(body.all(linkKind)) > 0) != values["links"] {
					t.Errorf("contents links do not match links=%v", values["links"])
				}
				entry := officeParagraph(t, body, "Contents entry phrase")
				layers := officeEffective(t, pkg, format, entry)
				dots := false
				if format == "docx" {
					leader, _ := officeEffectiveAttr(layers, "tab", "leader")
					dots = leader == "dot"
				} else {
					leader, _ := officeEffectiveAttr(layers, "tab-stop", "leader-style")
					text, _ := officeEffectiveAttr(layers, "tab-stop", "leader-text")
					dots = leader == "dotted" || text == "."
				}
				if dots != values["dot-leaders"] {
					t.Errorf("contents leaders do not match dot-leaders=%v", values["dot-leaders"])
				}
				if proseContentsHasPageFields(body, format) != values["page-numbers"] {
					t.Errorf("contents live page fields do not match page-numbers=%v", values["page-numbers"])
				}
			})
		}
	}
}

func proseContentsHasPageFields(body *officeXML, format string) bool {
	if format == "odt" {
		if len(body.all("index-entry-page-number")) > 0 {
			return true
		}
		for _, field := range body.all("bookmark-ref") {
			if field.attr("reference-format") == "page" {
				return true
			}
		}
		return false
	}
	for _, kind := range []string{"instrText", "fldSimple"} {
		for _, field := range body.all(kind) {
			instruction := field.text()
			if kind == "fldSimple" {
				instruction = field.attr("instr")
			}
			instruction = strings.TrimSpace(instruction)
			if strings.HasPrefix(instruction, "PAGEREF ") || (strings.HasPrefix(instruction, "TOC ") && !strings.Contains(instruction, `\n`)) {
				return true
			}
		}
	}
	return false
}
