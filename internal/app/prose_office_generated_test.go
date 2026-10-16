// ABOUTME: Checks independent title alignment and copyright ordering through native paragraphs.
// ABOUTME: Uses labelled fields and contrasting settings without relying on generated style names.
package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func proseAssertFile(t *testing.T, path, want string) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil || string(raw) != want {
		t.Errorf("%s changed: %v", path, err)
	}
}

func proseAssertAlignment(t *testing.T, pkg map[string]*officeXML, format, marker, want string) {
	t.Helper()
	layers := officeEffective(t, pkg, format, officeParagraph(t, officeBody(t, pkg, format), marker))
	element, attr := "jc", "val"
	if format == "odt" {
		element, attr = "paragraph-properties", "text-align"
	}
	got, exists := officeEffectiveAttr(layers, element, attr)
	if got == "start" {
		got = "left"
	}
	if got == "end" {
		got = "right"
	}
	if !exists || got != want {
		t.Errorf("%s alignment=%q present=%v; want %s", marker, got, exists, want)
	}
}

func TestRT063_5_OfficeIndependentTitleAlignment(t *testing.T) {
	metadata := "---\ntitle: TITLE-ALIGN\nsubtitle: SUBTITLE-ALIGN\nauthor: AUTHOR-ALIGN\ndate: DATE-ALIGN\nwordcount: WORDCOUNT-ALIGN\nversion: VERSION-ALIGN\ncontact-name: CONTACT-ALIGN\n---\n\nBODY-ALIGN\n"
	for _, format := range []string{"docx", "odt"} {
		for _, align := range []string{"left", "center", "right"} {
			t.Run(format+"/"+align, func(t *testing.T) {
				t.Setenv("HOME", t.TempDir())
				dir := t.TempDir()
				cfg := "folio:\n  manuscript:\n    toc:\n      enabled: false\n    title-page:\n      enabled: true\n      title-block-align: left\n      footer-align: right\n"
				roles := []string{"title", "subtitle", "author", "date", "wordcount", "version", "contact"}
				for _, role := range roles {
					cfg += "      " + role + ":\n        align: " + align + "\n"
				}
				proseSource(t, dir, "script.yaml", cfg)
				src := proseSource(t, dir, "source.md", metadata)
				pkg := proseConvert(t, src, filepath.Join(dir, "output."+format))
				for _, role := range roles {
					proseAssertAlignment(t, pkg, format, strings.ToUpper(role)+"-ALIGN", align)
				}
			})
		}
	}
}

func TestRT063_5_OfficeCopyrightOrder(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		for _, position := range []string{"after-title", "after-toc", "after-frontmatter"} {
			t.Run(format+"/"+position, func(t *testing.T) {
				t.Setenv("HOME", t.TempDir())
				dir := t.TempDir()
				cfg := "folio:\n  manuscript:\n    title-page:\n      enabled: true\n    toc:\n      enabled: true\n      title: CONTENTS-ORDER\n    copyright:\n      enabled: true\n      position: " + position + "\n      credits: [CREDIT-ORDER]\n      body: [LEGAL-ORDER]\n      separator: SEPARATOR-ORDER\n      publication: [PUBLICATION-ORDER]\n      publisher-preposition: Printed by\n      publisher: PUBLISHER-ORDER\n      isbn-label: NUMBER-ORDER\n      isbn: '978-0-000000-00-2'\n"
				proseSource(t, dir, "script.yaml", cfg)
				src := proseSource(t, dir, "source.md", "---\ntitle: TITLE-ORDER\n---\n\n## Chapter\n\nBODY-ORDER\n")
				pkg := proseConvert(t, src, filepath.Join(dir, "output."+format))
				text := officeBody(t, pkg, format).text()
				sequence := []string{"TITLE-ORDER", "CREDIT-ORDER", "LEGAL-ORDER", "SEPARATOR-ORDER", "PUBLICATION-ORDER", "Printed by", "PUBLISHER-ORDER", "NUMBER-ORDER", "978-0-000000-00-2", "BODY-ORDER"}
				previous := -1
				for _, marker := range sequence {
					at := strings.Index(text, marker)
					if at < 0 || at <= previous {
						t.Errorf("%s missing or out of order", marker)
					}
					previous = at
				}
				copyright, contents := strings.Index(text, "CREDIT-ORDER"), strings.Index(text, "CONTENTS-ORDER")
				if contents < 0 || (copyright < contents) != (position == "after-title") {
					t.Errorf("copyright position %s not retained relative to contents", position)
				}
			})
		}
	}
}

func TestRT063_5_OfficeInvalidISBNPreservesBothTargets(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		t.Run(format, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			dir := t.TempDir()
			proseSource(t, dir, "script.yaml", "folio:\n  manuscript:\n    copyright:\n      enabled: true\n      isbn: '978-0-000000-00-3'\n      isbn-barcode: render-and-file\n")
			src := proseSource(t, dir, "source.md", "BODY\n")
			target := proseSource(t, dir, "output."+format, "PRIOR-DOCUMENT")
			sidecar := proseSource(t, dir, "output.barcode.svg", "PRIOR-SIDECAR")
			status, _, stderr := runApp(t, "manuscript", src, target)
			if status == 0 || !strings.Contains(stderr, "isbn") {
				t.Errorf("invalid ISBN accepted or not identified: %d %s", status, stderr)
			}
			proseAssertFile(t, target, "PRIOR-DOCUMENT")
			proseAssertFile(t, sidecar, "PRIOR-SIDECAR")
		})
	}
}
