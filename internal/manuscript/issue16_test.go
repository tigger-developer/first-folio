// ABOUTME: Regression tests for issue #16 -- page-numbering + chapter number-reset.
// ABOUTME: Covers frontmatter/body format style ("1" / "I" / "i"), body-reset
// ABOUTME: enum, and chapter.number-reset per-part / never.
package manuscript

import (
	"strings"
	"testing"
)

// RT-16.1: default page-numbering emits arabic ("1") on both branches.
func TestRT_16_1_DefaultPageNumberingIsArabic(t *testing.T) {
	typst := renderIssue15Manuscript(t, "")
	// folio-display-page selects "1" for both frontmatter and body branches by default.
	assertContains(t, typst, `if is-body { "1" } else { "1" }`)
}

// RT-16.2: frontmatter-format: "i" produces lowercase roman numbering call for frontmatter branch.
func TestRT_16_2_FrontmatterFormatRomanLower(t *testing.T) {
	typst := renderIssue15Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    page-numbering:",
		"      frontmatter-format: \"i\"",
		"",
	}, "\n"))
	// The display macro selects "i" for the frontmatter branch.
	assertContains(t, typst, `if is-body { "1" } else { "i" }`)
}

// RT-16.3: body-format: "I" produces uppercase roman for body branch.
func TestRT_16_3_BodyFormatRomanUpper(t *testing.T) {
	typst := renderIssue15Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    page-numbering:",
		"      body-format: \"I\"",
		"",
	}, "\n"))
	assertContains(t, typst, `if is-body { "I" } else { "1" }`)
}

// RT-16.4: invalid frontmatter-format rejected with diagnostic.
func TestRT_16_4_InvalidFrontmatterFormatRejected(t *testing.T) {
	assertIssue15ConfigRejected(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    page-numbering:",
		"      frontmatter-format: \"iv\"",
		"",
	}, "\n"), "frontmatter-format")
}

// RT-16.5: body-reset: never omits the offset-seed update inside folio-part /
// folio-chapter. The reader of the offset in folio-display-page still appears
// (it initializes to 0), but the writer is not conditionally emitted.
func TestRT_16_5_BodyResetNeverOmitsOffsetSeed(t *testing.T) {
	typst := renderIssue15Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    page-numbering:",
		"      body-reset: never",
		"",
	}, "\n"))
	// The offset seed WRITE (inside folio-part / folio-chapter) must not appear.
	// The reader in folio-display-page still references the state (initialized to 0).
	assertNotContains(t, typst, `state("folio-page-offset", 0).update(counter(page).at(here()).first() - 1)`)
}

// RT-16.6: default body-reset ("first-part-or-chapter") retains the offset seed.
func TestRT_16_6_DefaultBodyResetSeedsOffset(t *testing.T) {
	typst := renderIssue15Manuscript(t, "")
	assertContains(t, typst, `state("folio-page-offset", 0).update(counter(page).at(here()).first() - 1)`)
}

// RT-16.7: invalid body-reset value rejected.
func TestRT_16_7_InvalidBodyResetRejected(t *testing.T) {
	assertIssue15ConfigRejected(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    page-numbering:",
		"      body-reset: sometimes",
		"",
	}, "\n"), "body-reset")
}

// RT-16.8: chapter.number-reset defaults to per-part (backwards compatible).
func TestRT_16_8_ChapterNumberResetDefaultsPerPart(t *testing.T) {
	// Fixture has: PART ONE > Chapter 1 > Chapter 2. Default per-part resets on
	// each new part (which the parser does), so chapters count 1, 2 within the
	// single part. Assert the second chapter emits number: "2".
	typst := renderIssue15Manuscript(t, "")
	assertContains(t, typst, `number: "2"`)
}

// RT-16.9: chapter.number-reset: never numbers chapters continuously across parts.
// Since the shared fixture only has one part, verify the emit path -- the config
// value threads through and the reset behaviour would apply if multi-part input
// were present.
func TestRT_16_9_ChapterNumberResetNeverAccepted(t *testing.T) {
	typst := renderIssue15Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    chapter:",
		"      number-reset: never",
		"",
	}, "\n"))
	// The rendering path accepted the value without error; chapters still emit.
	assertContains(t, typst, "#folio-chapter")
}

// RT-16.10: invalid chapter.number-reset value rejected.
func TestRT_16_10_InvalidChapterNumberResetRejected(t *testing.T) {
	assertIssue15ConfigRejected(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    chapter:",
		"      number-reset: sometimes",
		"",
	}, "\n"), "number-reset")
}
