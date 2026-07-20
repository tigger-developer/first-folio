// ABOUTME: Regression tests for issue #22 -- Typst enum-marker escape in heading body.
// ABOUTME: A composed heading like `1. Foo` (empty prefix + ". " separator) would
// ABOUTME: otherwise be re-parsed by Typst as a numbered-list enum item.
package manuscript

import (
	"strings"
	"testing"
)

// The shared fixture emits parts like "PART ONE" and chapters like "Chapter 1"
// (see markdownChapterOne / markdownChapterTwo). After the parser strips the
// "Part " / "Chapter " prefix, the semantic name is empty, so composeHeadingParts
// falls back to block.Text -- yielding chapter body "Chapter 1" and part body
// "PART ONE". With chapter.prefix: "" and separator ". ", the composed body
// becomes `1. Chapter 1`, which starts with the Typst enum-marker pattern.

// RT-22.1: chapter with empty prefix + ". " separator emits `1\. ...` (backslash-escaped).
func TestRT_22_1_ChapterEmptyPrefixDotSeparatorEscapesEnumMarker(t *testing.T) {
	typst := renderIssue15Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    chapter:",
		"      prefix: \"\"",
		"      separator: \". \"",
		"      show-name: true",
		"      show-number: true",
		"",
	}, "\n"))
	// Body must contain the escaped form, not the raw enum trigger.
	assertContains(t, typst, `[1\. Chapter 1]`)
	assertNotContains(t, typst, "[1. Chapter 1]")
}

// RT-22.2: chapter with empty prefix + ") " separator emits `1\) ...`.
func TestRT_22_2_ChapterEmptyPrefixParenSeparatorEscapesEnumMarker(t *testing.T) {
	typst := renderIssue15Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    chapter:",
		"      prefix: \"\"",
		"      separator: \") \"",
		"      show-name: true",
		"      show-number: true",
		"",
	}, "\n"))
	assertContains(t, typst, `[1\) Chapter 1]`)
	assertNotContains(t, typst, "[1) Chapter 1]")
}

// RT-22.3: chapter with non-empty prefix + ". " separator is UNCHANGED (no leading digit).
func TestRT_22_3_ChapterNonEmptyPrefixNoEscape(t *testing.T) {
	typst := renderIssue15Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    chapter:",
		"      prefix: \"Chapter \"",
		"      separator: \". \"",
		"      show-name: true",
		"      show-number: true",
		"",
	}, "\n"))
	// Composed body is "Chapter 1. Chapter 1" -- starts with letter, no enum trigger,
	// no backslash should be inserted before the mid-string dot.
	assertContains(t, typst, "[Chapter 1. Chapter 1]")
	assertNotContains(t, typst, `Chapter 1\. Chapter 1`)
}

// RT-22.4: chapter with show-number: false emits just the name, no escape.
func TestRT_22_4_ChapterNoNumberNoEscape(t *testing.T) {
	typst := renderIssue15Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    chapter:",
		"      prefix: \"\"",
		"      separator: \". \"",
		"      show-name: true",
		"      show-number: false",
		"",
	}, "\n"))
	assertContains(t, typst, "[Chapter 1]")
	assertNotContains(t, typst, `\. `)
}

// RT-22.5: the `full: %q` argument (state-storage form) is NOT escaped.
// State values are rendered as string content (not re-parsed as markup) so a
// visible backslash there would be a rendering regression.
func TestRT_22_5_StateFullArgIsNotEscaped(t *testing.T) {
	typst := renderIssue15Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    chapter:",
		"      prefix: \"\"",
		"      separator: \". \"",
		"      show-name: true",
		"      show-number: true",
		"",
	}, "\n"))
	// The `full:` named arg carries the state-storage form. It must not contain
	// a backslash-escaped enum marker (Go's %q would render `\.` as `\\.` in the
	// emitted Typst string literal, which would be visible on rendering).
	assertContains(t, typst, `full: "1. Chapter 1"`)
	assertNotContains(t, typst, `full: "1\\. Chapter 1"`)
}

// RT-22.6: part with empty prefix + ". " separator emits `1\. ...` (same escape as chapter).
func TestRT_22_6_PartEmptyPrefixDotSeparatorEscapesEnumMarker(t *testing.T) {
	typst := renderIssue15Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    part:",
		"      prefix: \"\"",
		"      separator: \". \"",
		"      show-name: true",
		"      show-number: true",
		"",
	}, "\n"))
	// The fixture's part heading is "PART ONE"; after prefix-strip the semantic
	// name is empty and composeHeadingParts falls back to block.Text.
	assertContains(t, typst, `[1\. PART ONE]`)
	assertNotContains(t, typst, "[1. PART ONE]")
}
