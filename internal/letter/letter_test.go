// ABOUTME: Characterizes cover-letter parsing, substitution, and Typst formatting.
// ABOUTME: Covers multiple recipients, metadata fallback, dates, and inline Org markup.
package letter

import (
	"strings"
	"testing"

	"github.com/tadg-paul/first-folio/internal/config"
)

const letterFixture = `#+TITLE: Test Play
#+AUTHOR: Taḋg Paul
#+EMAIL: author@example.invalid
* Cover letters :letter:
** Sender / Dublin :sender:
** Submission :subject:
Hello [moniker] at [org].

This has =inline code=, *bold*, /italic/, and _underline_.
*** Yours sincerely :closing:
*** 2026-07-13
*** Recipient One / Cork :to:
**** One Theatre :org:
**** Aoife :moniker:
*** Recipient Two / Galway :to:
**** Two Theatre :org:
**** Niamh :moniker:
`

func TestParseLetters(t *testing.T) {
	letters, err := ParseOrg(letterFixture)
	if err != nil {
		t.Fatal(err)
	}
	if len(letters) != 2 {
		t.Fatalf("letter count = %d, want 2", len(letters))
	}
	if letters[0].Organization != "One Theatre" || !strings.Contains(letters[0].Body, "Hello Aoife at One Theatre.") {
		t.Fatalf("first letter = %#v", letters[0])
	}
	if letters[1].Recipient != "Recipient Two / Galway" || letters[1].Date != "2026-07-13" {
		t.Fatalf("second letter = %#v", letters[1])
	}
}

func TestRenderTypstInlineOrgMarkup(t *testing.T) {
	letters, err := ParseOrg(letterFixture)
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(config.Options{Mode: config.ModeLetter, Home: t.TempDir(), LocalDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	typst, err := RenderTypst(letters[0], cfg)
	if err != nil {
		t.Fatal(err)
	}
	for _, fragment := range []string{
		`#text(font: "Libertinus Mono")[inline code]`,
		`*bold*`,
		`_italic_`,
		`#underline[underline]`,
		`13 July 2026`,
		`Hello Aoife at One Theatre.`,
	} {
		if !strings.Contains(typst, fragment) {
			t.Errorf("Typst missing %q:\n%s", fragment, typst)
		}
	}
	if strings.Contains(typst, "=inline code=") {
		t.Error("literal Org code delimiters remain")
	}
}
