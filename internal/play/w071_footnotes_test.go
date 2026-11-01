// ABOUTME: Defines W071's structured Org footnote and round-trip regression boundary.
// ABOUTME: Prevents blank-separated note paragraphs from leaking into play dialogue.
package play

import (
	"slices"
	"strings"
	"testing"
)

func TestRT071_2_OrgFootnoteParagraphsStayStructured(t *testing.T) {
	source := "* ACT ONE\n** SCENE ONE\n**** CÁIT\nLine.[fn:note]\n\n[fn:note] First *bold* paragraph.\n\nSecond /italic/ paragraph.\n\nThird paragraph.\n"
	doc, _, err := Parse(FormatOrg, source, "fixture.org")
	if err != nil {
		t.Fatal(err)
	}
	var note Event
	for _, event := range doc.Events {
		if event.Kind == EventFootnote {
			note = event
		}
		if event.Kind != EventFootnote && (strings.Contains(event.Text, "Second") || strings.Contains(event.Text, "Third")) {
			t.Errorf("footnote continuation leaked into %s: %#v", event.Kind, event)
		}
	}
	want := []string{"First *bold* paragraph.", "Second /italic/ paragraph.", "Third paragraph."}
	if !slices.Equal(note.Paragraphs, want) {
		t.Fatalf("note paragraphs = %#v, want %#v", note.Paragraphs, want)
	}

	emitted, err := Emit(doc, FormatOrg)
	if err != nil {
		t.Fatal(err)
	}
	roundTrip, _, err := Parse(FormatOrg, emitted.Text, "roundtrip.org")
	if err != nil {
		t.Fatal(err)
	}
	for _, event := range roundTrip.Events {
		if event.Kind == EventFootnote && slices.Equal(event.Paragraphs, want) {
			return
		}
	}
	t.Fatalf("round-trip lost note paragraphs:\n%s\n%#v", emitted.Text, roundTrip.Events)
}

func TestRT071_2_OrgFootnoteTerminationBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name          string
		suffix        string
		followingKind EventKind
		followingName string
		followingText string
	}{
		{"next definition", "\n[fn:other] Other note.\n", EventFootnote, "other", "Other note."},
		{"headline", "\n*** Direction\n", EventStageDirection, "", "Direction"},
		{"two blank lines", "\n\nAfter note.\n", EventDialogue, "", "After note."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := "* ACT ONE\n** SCENE ONE\n**** CÁIT\nLine.[fn:note]\n\n[fn:note] First paragraph.\n\nSecond paragraph.\n" + tc.suffix
			doc, _, err := Parse(FormatOrg, source, "boundary.org")
			if err != nil {
				t.Fatal(err)
			}
			foundNote, foundFollowing := false, false
			for _, event := range doc.Events {
				if event.Kind == EventFootnote && event.Name == "note" {
					want := []string{"First paragraph.", "Second paragraph."}
					if !slices.Equal(event.Paragraphs, want) {
						t.Fatalf("paragraphs = %#v, want %#v", event.Paragraphs, want)
					}
					foundNote = true
				}
				if event.Kind == tc.followingKind && event.Name == tc.followingName && event.Text == tc.followingText {
					foundFollowing = true
				}
			}
			if !foundNote {
				t.Error("primary note missing")
			}
			if !foundFollowing {
				t.Errorf("content after boundary was lost: want kind=%s name=%q text=%q in %#v", tc.followingKind, tc.followingName, tc.followingText, doc.Events)
			}
		})
	}
}
