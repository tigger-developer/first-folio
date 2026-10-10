// ABOUTME: Covers character-name aliases and Unicode-aware stage-direction matching.
// ABOUTME: Verifies that undeclared names and non-direction events remain untouched.
package play

import (
	"strings"
	"testing"
)

func TestCapitalizeStageNamesBoundaries(t *testing.T) {
	tests := []struct{ name, cell, text, want string }{
		{"words", "Tom", "Tom's tom, Tommy atom Tom2 2Tom", "TOM's TOM, Tommy atom Tom2 2Tom"},
		{"unicode", "Mairéad", "Mairéad's mairéad Mairéadín", "MAIRÉAD's MAIRÉAD Mairéadín"},
		{"unicode boundaries", "Tom", "éTom Tomé Tom\u0301", "éTom Tomé Tom\u0301"},
		{"markup", "Tom", "_Tom_ /Tom/ *Tom*", "_TOM_ /TOM/ *TOM*"},
		{"aliases", "MARGARET (MAGS, MARG.)", "Margaret Mags Marg. Marg margin", "MARGARET MAGS MARG. MARG margin"},
		{"phrase", "ALISON (NURSE ALISON)", "Nurse Alison greets Alison", "NURSE ALISON greets ALISON"},
		{"literal metacharacters", "A+B (C[1])", "A+B meets C[1].", "A+B meets C[1]."},
		{"overlapping phrase boundary", "Tom (Tom Smith)", "Tom Smithson enters", "TOM Smithson enters"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := Document{Events: []Event{{Kind: EventCharacterTableRow, Name: tt.cell}, {Kind: EventStageDirection, Text: tt.text}, {Kind: EventDialogue, Text: tt.text}}}
			CapitalizeStageNames(&doc)
			if got := doc.Events[1].Text; got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
			if doc.Events[2].Text != tt.text {
				t.Fatal("dialogue changed")
			}
			CapitalizeStageNames(&doc)
			if doc.Events[1].Text != tt.want {
				t.Fatal("capitalisation is not idempotent")
			}
		})
	}
}

func TestStageNamesRequireCastTable(t *testing.T) {
	doc := Document{Events: []Event{{Kind: EventCharacter, Name: "Tom"}, {Kind: EventStageDirection, Text: "Tom enters."}}}
	CapitalizeStageNames(&doc)
	if doc.Events[1].Text != "Tom enters." {
		t.Fatal("speaker cue incorrectly used as lookup")
	}
}

func TestCastAliasRoundTrip(t *testing.T) {
	source := "* CHARACTERS\n| Alison (Nurse Alison) | A nurse |\n* Act One\n**** Alison\nHello, Tom.\n"
	doc, _, err := Parse(FormatOrg, source, "play.org")
	if err != nil {
		t.Fatal(err)
	}
	output, err := Emit(doc, FormatMarkdown)
	if err != nil {
		t.Fatal(err)
	}
	restored, _, err := Parse(FormatMarkdown, output.Text, "play.md")
	if err != nil {
		t.Fatal(err)
	}
	org, err := Emit(restored, FormatOrg)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(org.Text, "| Alison (Nurse Alison) | A nurse |") {
		t.Fatal("alias declaration lost on round trip")
	}
	fountain, err := Emit(doc, FormatFountain)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(fountain.Text, "Alison - A nurse") || strings.Contains(fountain.Text, "(Nurse Alison)") {
		t.Fatal("Fountain cast should show only primary name")
	}
}

func TestParseCastName(t *testing.T) {
	for _, tt := range []struct {
		cell, primary string
		aliases       []string
	}{
		{" Tom ", "Tom", nil},
		{"MARGARET ( MAGS, MARG. )", "MARGARET", []string{"MAGS", "MARG."}},
		{"Tom (", "Tom (", nil},
		{"Tom (TOM,, )", "Tom", []string{"TOM"}},
	} {
		primary, aliases := ParseCastName(tt.cell)
		if primary != tt.primary || strings.Join(aliases, "|") != strings.Join(tt.aliases, "|") {
			t.Errorf("%q: got %q %v", tt.cell, primary, aliases)
		}
	}
}
