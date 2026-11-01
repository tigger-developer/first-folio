// ABOUTME: Characterizes stage-play parsing, intro classification, and text emitters.
// ABOUTME: Covers the shared semantic events across Org, Markdown, and Fountain.
package play

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseOrgPlayEvents(t *testing.T) {
	source := `#+TITLE: Samhain
#+AUTHOR: Tadhg O'Brien
* CHARACTERS
| CÁIT | A doctor |
* Synopsis
A short introduction.
* ACT ONE
** Scene One
*** Night falls.
**** CÁIT quietly
Hello [fn:door].
*** CUT TO: :transition:
[fn:door] The old door.
`
	doc, warnings, err := Parse(FormatOrg, source, "play.org")
	if err != nil {
		t.Fatal(err)
	}
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}
	want := []Event{
		{Kind: EventFrontMatter, Key: "title", Text: "Samhain"},
		{Kind: EventFrontMatter, Key: "author", Text: "Tadhg O'Brien"},
		{Kind: EventCharacterTableStart, Text: "CHARACTERS"},
		{Kind: EventCharacterTableRow, Name: "CÁIT", Text: "A doctor"},
		{Kind: EventCharacterTableEnd},
		{Kind: EventIntroHeader, Text: "Synopsis"},
		{Kind: EventIntroText, Text: "A short introduction."},
		{Kind: EventActHeader, Text: "ACT ONE"},
		{Kind: EventSceneHeader, Text: "Scene One"},
		{Kind: EventStageDirection, Text: "Night falls."},
		{Kind: EventCharacter, Name: "CÁIT", Direction: "quietly"},
		{Kind: EventDialogue, Text: "Hello [fn:door]."},
		{Kind: EventTransition, Text: "CUT TO:"},
		{Kind: EventFootnote, Name: "door", Text: "The old door."},
	}
	assertEvents(t, doc.Events, want)
}

func TestParseMarkdownPlayEvents(t *testing.T) {
	source := `# Samhain

**A Play**

*by Tadhg O'Brien*

## ACT ONE

### Scene One

*Night falls.*

**CÁIT:** *(quietly)*
Hello.[^door]

> CUT TO:

[^door]: The old door.
`
	doc, _, err := Parse(FormatMarkdown, source, "play.md")
	if err != nil {
		t.Fatal(err)
	}
	if got := doc.Metadata["title"]; got != "Samhain" {
		t.Fatalf("title = %q", got)
	}
	assertHasEvent(t, doc.Events, Event{Kind: EventCharacter, Name: "CÁIT", Direction: "quietly"})
	assertHasEvent(t, doc.Events, Event{Kind: EventFootnote, Name: "door", Text: "The old door."})
}

func TestParseFountainWarningsAndEvents(t *testing.T) {
	source := `Title: Samhain
Author: Tadhg O'Brien

= dropped synopsis
===

> **ACT ONE** <

.SCENE ONE

CÁIT
(quietly)
Hello.

CUT TO:
`
	doc, warnings, err := Parse(FormatFountain, source, "play.fountain")
	if err != nil {
		t.Fatal(err)
	}
	if len(warnings) != 2 || !strings.Contains(warnings[0], "dropping synopsis") || !strings.Contains(warnings[1], "dropping page break") {
		t.Fatalf("warnings = %#v", warnings)
	}
	assertHasEvent(t, doc.Events, Event{Kind: EventCharacter, Name: "CÁIT", Direction: "quietly"})
	assertHasEvent(t, doc.Events, Event{Kind: EventTransition, Text: "CUT TO:"})
}

func TestTextEmittersPreserveCoreContract(t *testing.T) {
	doc := Document{
		Metadata: map[string]string{"title": "Samhain", "subtitle": "A Play", "author": "Tadhg O'Brien", "version": "Draft 2", "date": "2026-07-13"},
		Events: []Event{
			{Kind: EventActHeader, Text: "ACT ONE"},
			{Kind: EventSceneHeader, Text: "Scene One"},
			{Kind: EventStageDirection, Text: "Night falls."},
			{Kind: EventCharacter, Name: "CÁIT", Direction: "quietly"},
			{Kind: EventDialogue, Text: "Hello [fn:door]."},
			{Kind: EventTransition, Text: "CUT TO:"},
			{Kind: EventFootnote, Name: "door", Text: "The old door."},
		},
	}

	org, err := Emit(doc, FormatOrg)
	if err != nil {
		t.Fatal(err)
	}
	for _, fragment := range []string{"#+TITLE: Samhain", "* ACT ONE", "**** CÁIT quietly", "[fn:door] The old door."} {
		if !strings.Contains(org.Text, fragment) {
			t.Errorf("org missing %q:\n%s", fragment, org.Text)
		}
	}

	markdown, err := Emit(doc, FormatMarkdown)
	if err != nil {
		t.Fatal(err)
	}
	for _, fragment := range []string{"# Samhain", "## ACT ONE", "**CÁIT:** *(quietly)*", "[^door]: The old door."} {
		if !strings.Contains(markdown.Text, fragment) {
			t.Errorf("markdown missing %q:\n%s", fragment, markdown.Text)
		}
	}

	fountain, err := Emit(doc, FormatFountain)
	if err != nil {
		t.Fatal(err)
	}
	for _, fragment := range []string{"Title: Samhain", "> **ACT ONE** <", ".SCENE ONE", "CÁIT", "[[The old door.]]"} {
		if !strings.Contains(fountain.Text, fragment) {
			t.Errorf("fountain missing %q:\n%s", fragment, fountain.Text)
		}
	}
}

func assertEvents(t *testing.T, got []Event, want []Event) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("event count = %d, want %d\ngot: %#v", len(got), len(want), got)
	}
	for i := range want {
		if !reflect.DeepEqual(got[i], want[i]) {
			t.Errorf("event %d = %#v, want %#v", i, got[i], want[i])
		}
	}
}

func assertHasEvent(t *testing.T, events []Event, want Event) {
	t.Helper()
	for _, event := range events {
		if reflect.DeepEqual(event, want) {
			return
		}
	}
	t.Errorf("missing event %#v in %#v", want, events)
}
