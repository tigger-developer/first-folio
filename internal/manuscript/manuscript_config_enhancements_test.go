// ABOUTME: Regression tests for issue #15 manuscript config enhancements.
// ABOUTME: Covers RT-15.1..RT-15.47 across placeholders, footer, page, blank pages, alignment.
package manuscript

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

// -----------------------------------------------------------------------------
// AC15.1 -- page-header format placeholders [part] and [chapter]
// -----------------------------------------------------------------------------

// RT-15.1: format with [part] emits a part-state read in the header, and the template
// folio-part function updates the state at each part start so downstream headers pick it up.
func TestRT_15_1_HeaderPartPlaceholderTracksCurrentPart(t *testing.T) {
	typst := renderIssue15Manuscript(t, "folio:\n  manuscript:\n    page-header:\n      format: \"[part]\"\n")
	assertContains(t, typst, `state("folio-current-part").get()`)
	assertContains(t, typst, `state("folio-current-part").update`)
}

// RT-15.2: format with [chapter] emits a chapter-state read in the header, and the template
// folio-chapter function updates that state at each chapter start.
func TestRT_15_2_HeaderChapterPlaceholderTracksCurrentChapter(t *testing.T) {
	typst := renderIssue15Manuscript(t, "folio:\n  manuscript:\n    page-header:\n      format: \"[chapter]\"\n")
	assertContains(t, typst, `state("folio-current-chapter").get()`)
	assertContains(t, typst, `state("folio-current-chapter").update`)
}

// RT-15.3: default format [author] / [title] / [page] continues to substitute all three tokens.
func TestRT_15_3_HeaderAuthorTitlePagePlaceholdersStillSubstitute(t *testing.T) {
	typst := renderIssue15Manuscript(t, "")
	assertContains(t, typst, `Example Author`)
	assertContains(t, typst, `The Glass Orchard`)
	assertContains(t, typst, `#folio-display-page()`)
	assertNotContains(t, typst, "[author]")
	assertNotContains(t, typst, "[title]")
	assertNotContains(t, typst, "[page]")
}

// RT-15.4: unknown token in format renders as literal in generated Typst.
func TestRT_15_4_HeaderUnknownPlaceholderIsLiteral(t *testing.T) {
	typst := renderIssue15Manuscript(t, "folio:\n  manuscript:\n    page-header:\n      format: \"pre [unknown] post\"\n")
	assertContains(t, typst, `\[unknown\]`)
}

// RT-15.5: state defaults to empty, so pre-heading header renders empty for [part]/[chapter].
func TestRT_15_5_HeaderStateHasEmptyDefault(t *testing.T) {
	typst := renderIssue15Manuscript(t, "folio:\n  manuscript:\n    page-header:\n      format: \"[part]/[chapter]\"\n")
	assertContains(t, typst, `state("folio-current-part", "")`)
	assertContains(t, typst, `state("folio-current-chapter", "")`)
}

// -----------------------------------------------------------------------------
// AC15.2 -- page-footer block
// -----------------------------------------------------------------------------

// RT-15.6: page-footer.enabled: true renders a footer with the configured format on body pages.
func TestRT_15_6_PageFooterEnabledRendersFooter(t *testing.T) {
	typst := renderIssue15Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    page-footer:",
		"      enabled: true",
		"      format: \"footer-marker-abc\"",
		"",
	}, "\n"))
	assertContains(t, typst, `footer-marker-abc`)
	// The footer is emitted inside a context {} block that conditionally hides via skip-footer state.
	assertContains(t, typst, `footer: context {`)
}

// RT-15.7 [🚫 removed]: The default page-footer.enabled was originally documented as false. The
// amended Feature 2 defaults (see AC15.8) enable the footer with a centered [page] format. The ID
// is preserved here per SDLC immutability; the current default behaviour is covered by RT-15.54.
func TestRT_15_7_PageFooterDefaultDisabled(t *testing.T) {
	t.Skip("RT-15.7 removed: page-footer default changed to enabled with centered [page] -- see RT-15.54")
}

// RT-15.8: page-footer.font, font-size, font-weight propagate to footer typography.
func TestRT_15_8_PageFooterTypographyPropagates(t *testing.T) {
	typst := renderIssue15Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    page-footer:",
		"      enabled: true",
		"      format: \"footer-marker\"",
		"      font: Courier",
		"      font-size: 9pt",
		"      font-weight: bold",
		"",
	}, "\n"))
	body := extractBodyPageBlock(t, typst)
	assertContains(t, body, `font: "Courier"`)
	assertContains(t, body, `size: 9pt`)
	assertContains(t, body, `weight: "bold"`)
	assertContains(t, body, `footer-marker`)
}

// RT-15.9: page-footer.format supports the same placeholder set as page-header.
func TestRT_15_9_PageFooterFormatSupportsAllPlaceholders(t *testing.T) {
	typst := renderIssue15Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    page-footer:",
		"      enabled: true",
		"      format: \"[author] / [title] / [page] / [part] / [chapter]\"",
		"",
	}, "\n"))
	// The footer text substitutes author/title literally and part/chapter/page as state/counter reads.
	body := extractBodyPageBlock(t, typst)
	assertContains(t, body, `Example Author`)
	assertContains(t, body, `The Glass Orchard`)
	assertContains(t, body, `state("folio-current-part").get()`)
	assertContains(t, body, `state("folio-current-chapter").get()`)
	assertContains(t, body, `#folio-display-page()`)
}

// RT-15.10: page-footer.align, distance-from-edge, content-padding-after position the footer.
func TestRT_15_10_PageFooterPositioningPropagates(t *testing.T) {
	typst := renderIssue15Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    page-footer:",
		"      enabled: true",
		"      format: \"footer-marker\"",
		"      align: center",
		"      distance-from-edge: 18mm",
		"      content-padding-after: 12mm",
		"",
	}, "\n"))
	body := extractBodyPageBlock(t, typst)
	// The footer align expression is wrapped in a context block for skip-footer support.
	assertContains(t, body, `align(center)[`)
	assertContains(t, body, `bottom: 18mm + 12mm`)
	assertContains(t, body, `footer-descent: 12mm`)
}

// RT-15.11: page-footer is absent from the title page and TOC page.
func TestRT_15_11_PageFooterAbsentFromTitleAndTOC(t *testing.T) {
	typst := renderIssue15Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    page-footer:",
		"      enabled: true",
		"      format: \"footer-marker-should-not-be-on-title-or-toc\"",
		"",
	}, "\n"))
	titleTOC := extractTitleAndTOCBlock(t, typst)
	assertNotContains(t, titleTOC, `footer-marker-should-not-be-on-title-or-toc`)
}

// RT-15.12: missing page-footer typography inherits from page-header (which inherits from root).
func TestRT_15_12_PageFooterInheritsFromHeader(t *testing.T) {
	typst := renderIssue15Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    page-header:",
		"      font: \"HeaderFont\"",
		"      font-size: 8pt",
		"      font-weight: light",
		"    page-footer:",
		"      enabled: true",
		"      format: \"footer-marker\"",
		"",
	}, "\n"))
	body := extractBodyPageBlock(t, typst)
	// Footer text() should use the header's font/size/weight because page-footer didn't set them.
	assertContains(t, body, `font: "HeaderFont"`)
	assertContains(t, body, `size: 8pt`)
	assertContains(t, body, `weight: "light"`)
}

// -----------------------------------------------------------------------------
// AC15.3 -- custom WxH page dimensions
// -----------------------------------------------------------------------------

// RT-15.13: page: a4 continues to render at A4 dimensions (named preset passthrough).
func TestRT_15_13_PageA4Preset(t *testing.T) {
	typst := renderIssue15Manuscript(t, "folio:\n  manuscript:\n    page: a4\n")
	assertContains(t, typst, `paper: "a4"`)
}

// RT-15.14: page: 200x300mm renders at 200mm x 300mm.
func TestRT_15_14_PageCustomMillimetres(t *testing.T) {
	typst := renderIssue15Manuscript(t, "folio:\n  manuscript:\n    page: 200x300mm\n")
	assertContains(t, typst, `width: 200mm`)
	assertContains(t, typst, `height: 300mm`)
	assertNotContains(t, typst, `paper: "200x300mm"`)
}

// RT-15.15: page: 5.5x8.5in renders at 5.5in x 8.5in.
func TestRT_15_15_PageCustomInches(t *testing.T) {
	typst := renderIssue15Manuscript(t, "folio:\n  manuscript:\n    page: 5.5x8.5in\n")
	assertContains(t, typst, `width: 5.5in`)
	assertContains(t, typst, `height: 8.5in`)
	assertNotContains(t, typst, `paper: "5.5x8.5in"`)
}

// RT-15.16: page: uk-book-b continues to render at the named preset.
func TestRT_15_16_PageUKBookBPreset(t *testing.T) {
	typst := renderIssue15Manuscript(t, "folio:\n  manuscript:\n    page: uk-book-b\n")
	assertContains(t, typst, `paper: "uk-book-b"`)
}

// RT-15.17: page: bogus exits non-zero with a diagnostic naming the offending value.
func TestRT_15_17_PageBogusRejected(t *testing.T) {
	assertIssue15ConfigRejected(t, "folio:\n  manuscript:\n    page: bogus\n", "bogus")
}

// RT-15.18: page: 200x300xy (unknown unit) exits non-zero with diagnostic.
func TestRT_15_18_PageUnknownUnitRejected(t *testing.T) {
	assertIssue15ConfigRejected(t, "folio:\n  manuscript:\n    page: 200x300xy\n", "200x300xy")
}

// RT-15.19: page: 200mm (missing height) exits non-zero with diagnostic.
func TestRT_15_19_PageMissingHeightRejected(t *testing.T) {
	assertIssue15ConfigRejected(t, "folio:\n  manuscript:\n    page: 200mm\n", "200mm")
}

// RT-15.20: page: 5.5inx200mm (mixed unit) exits non-zero with diagnostic.
func TestRT_15_20_PageMixedUnitRejected(t *testing.T) {
	assertIssue15ConfigRejected(t, "folio:\n  manuscript:\n    page: 5.5inx200mm\n", "5.5inx200mm")
}

// -----------------------------------------------------------------------------
// AC15.4 -- blank-page-before / blank-page-after
// -----------------------------------------------------------------------------

// RT-15.21: chapter.blank-page-before: true inserts an unnumbered blank page before each chapter.
func TestRT_15_21_ChapterBlankPageBefore(t *testing.T) {
	typst := renderIssue15Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    chapter:",
		"      blank-page-before: true",
		"",
	}, "\n"))
	assertContains(t, typst, `#folio-blank-page()`)
	// Blank page markers must appear BEFORE the chapter emit.
	assertBefore(t, typst, `#folio-blank-page()`, `#folio-chapter(first: false,`)
	assertContains(t, typst, `)[Chapter 1]`)
}

// RT-15.22: chapter.blank-page-after: true inserts a blank page after each chapter.
func TestRT_15_22_ChapterBlankPageAfter(t *testing.T) {
	typst := renderIssue15Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    chapter:",
		"      blank-page-after: true",
		"",
	}, "\n"))
	assertContains(t, typst, `#folio-blank-page()`)
	assertBefore(t, typst, `#folio-chapter(first: false,`, `#folio-blank-page()`)
	assertContains(t, typst, `)[Chapter 1]`)
}

// RT-15.23: part.blank-page-before: true inserts a blank page before each part.
func TestRT_15_23_PartBlankPageBefore(t *testing.T) {
	typst := renderIssue15Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    part:",
		"      blank-page-before: true",
		"",
	}, "\n"))
	assertContains(t, typst, `#folio-blank-page()`)
	assertBefore(t, typst, `#folio-blank-page()`, `#folio-part(first: true,`)
	assertContains(t, typst, `)[PART ONE]`)
}

// RT-15.24: part.blank-page-after: true inserts a blank page after each part.
func TestRT_15_24_PartBlankPageAfter(t *testing.T) {
	typst := renderIssue15Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    part:",
		"      blank-page-after: true",
		"",
	}, "\n"))
	assertContains(t, typst, `#folio-blank-page()`)
	assertBefore(t, typst, `#folio-part(first: true,`, `#folio-blank-page()`)
	assertContains(t, typst, `)[PART ONE]`)
}

// RT-15.25: blank-page-before + page-break-before both set produces one blank + one heading page.
func TestRT_15_25_BlankPageBeforeWithPageBreakBefore(t *testing.T) {
	typst := renderIssue15Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    chapter:",
		"      blank-page-before: true",
		"      page-break-before: true",
		"",
	}, "\n"))
	// The emitter must emit exactly one blank-page marker per chapter (not two).
	countBlankMarkers := strings.Count(typst, `#folio-blank-page()`)
	chapterCount := strings.Count(typst, `#folio-chapter(first: `)
	if countBlankMarkers != chapterCount {
		t.Fatalf("expected exactly one blank-page marker per chapter (chapters=%d, markers=%d)\n%s", chapterCount, countBlankMarkers, typst)
	}
}

// RT-15.26: blank-page-before with page-break-before: false still produces one blank + heading.
func TestRT_15_26_BlankPageBeforeWithoutPageBreakBefore(t *testing.T) {
	typst := renderIssue15Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    chapter:",
		"      blank-page-before: true",
		"      page-break-before: false",
		"",
	}, "\n"))
	assertContains(t, typst, `#folio-blank-page()`)
}

// RT-15.27: folio-blank-page suppresses numbering, header, and footer for the blank page.
func TestRT_15_27_BlankPageIsUnnumberedAndBlank(t *testing.T) {
	typst := renderIssue15Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    chapter:",
		"      blank-page-before: true",
		"",
	}, "\n"))
	// The template macro must set numbering: none, header: none, footer: none for the blank page.
	assertContains(t, typst, `#let folio-blank-page()`)
	assertContains(t, typst, `numbering: none`)
	assertContains(t, typst, `header: none`)
	assertContains(t, typst, `footer: none`)
}

// RT-15.28: all four blank-page-* defaults are false, producing no blank pages.
func TestRT_15_28_BlankPageDefaultsAreFalse(t *testing.T) {
	typst := renderIssue15Manuscript(t, "")
	assertNotContains(t, typst, `#folio-blank-page()`)
}

// -----------------------------------------------------------------------------
// AC15.5 -- per-item title-page compound alignment
// -----------------------------------------------------------------------------

// RT-15.29: title-page.title.align: bottom-center places the title at bottom-center.
func TestRT_15_29_TitleAlignBottomCenter(t *testing.T) {
	typst := renderIssue15Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    title-page:",
		"      title:",
		"        align: bottom-center",
		"",
	}, "\n"))
	assertContains(t, typst, `#place(bottom + center`)
	// The placed content includes the title text.
	assertPlacedItemContains(t, typst, "bottom + center", "The Glass Orchard")
}

// RT-15.30: subtitle.align: top-center places the subtitle at top-center.
func TestRT_15_30_SubtitleAlignTopCenter(t *testing.T) {
	typst := renderIssue15Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    title-page:",
		"      subtitle:",
		"        align: top-center",
		"",
	}, "\n"))
	assertPlacedItemContains(t, typst, "top + center", "A Novel")
}

// RT-15.31: author.align: top-right places the author at top-right.
func TestRT_15_31_AuthorAlignTopRight(t *testing.T) {
	typst := renderIssue15Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    title-page:",
		"      author:",
		"        align: top-right",
		"",
	}, "\n"))
	assertPlacedItemContains(t, typst, "top + right", "Example Author")
}

// RT-15.32: date.align: bottom-left places the date at bottom-left.
func TestRT_15_32_DateAlignBottomLeft(t *testing.T) {
	typst := renderIssue15Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    title-page:",
		"      date:",
		"        align: bottom-left",
		"      include-date: true",
		"",
	}, "\n"))
	assertPlacedItemContains(t, typst, "bottom + left", "6 July 2026")
}

// RT-15.33: wordcount.align: bottom-right places the word count at bottom-right.
func TestRT_15_33_WordCountAlignBottomRight(t *testing.T) {
	typst := renderIssue15Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    title-page:",
		"      wordcount:",
		"        align: bottom-right",
		"      include-wordcount: true",
		"",
	}, "\n"))
	assertPlacedItemContains(t, typst, "bottom + right", "90000")
}

// RT-15.34: version.align: top-left places the version at top-left.
func TestRT_15_34_VersionAlignTopLeft(t *testing.T) {
	typst := renderIssue15Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    title-page:",
		"      version:",
		"        align: top-left",
		"      include-version: true",
		"",
	}, "\n"))
	assertPlacedItemContains(t, typst, "top + left", "Draft 4")
}

// RT-15.35: contact.align: top-right moves the contact block from the hardcoded top-left.
func TestRT_15_35_ContactAlignTopRight(t *testing.T) {
	typst := renderIssue15Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    title-page:",
		"      contact:",
		"        align: top-right",
		"",
	}, "\n"))
	assertPlacedItemContains(t, typst, "top + right", "Example Agent")
}

// RT-15.36: bare compass value 'center' is accepted and treated as center-center.
func TestRT_15_36_BareCompassCenterAccepted(t *testing.T) {
	typst := renderIssue15Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    title-page:",
		"      title:",
		"        align: center",
		"",
	}, "\n"))
	// center becomes center + horizon in Typst (horizon is Typst's vertical center token).
	assertPlacedItemContains(t, typst, "center + horizon", "The Glass Orchard")
}

// RT-15.37: title.align: middle-middle is rejected with a diagnostic naming the offending value.
func TestRT_15_37_InvalidVerticalAlignRejected(t *testing.T) {
	assertIssue15ConfigRejected(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    title-page:",
		"      title:",
		"        align: middle-middle",
		"",
	}, "\n"), "middle-middle")
}

// RT-15.38: author.align: bottom-diagonal is rejected with a diagnostic.
func TestRT_15_38_InvalidHorizontalAlignRejected(t *testing.T) {
	assertIssue15ConfigRejected(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    title-page:",
		"      author:",
		"        align: bottom-diagonal",
		"",
	}, "\n"), "bottom-diagonal")
}

// Supplement to RT-15.33/34: an item with a per-item align must not also render in the pre-existing
// British title-page grid footer or the US bottom-center wordcount placement. This prevents the
// duplication a human reviewer flagged during UT-15.4 presentation.
func TestRT_15_5_PerItemAlignSuppressesLegacyGridDuplication(t *testing.T) {
	typst := renderIssue15Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    title-page:",
		"      version:",
		"        align: top-left",
		"      wordcount:",
		"        align: bottom-right",
		"      date:",
		"        align: bottom-left",
		"",
	}, "\n"))
	titleTOC := extractTitleAndTOCBlock(t, typst)
	// Each item text should appear at its per-item place block, never inside the grid footer.
	// The grid columns are marked by `columns: (1fr, 1fr, 1fr)`; ensure each item text only
	// appears once in the title/TOC region.
	countVersion := strings.Count(titleTOC, "Draft 4")
	countWordCount := strings.Count(titleTOC, "90000")
	countDate := strings.Count(titleTOC, "6 July 2026")
	if countVersion > 1 {
		t.Fatalf("version text appears %d times; expected once (no grid-footer duplication)\n%s", countVersion, titleTOC)
	}
	if countWordCount > 1 {
		t.Fatalf("word-count text appears %d times; expected once (no grid-footer duplication)\n%s", countWordCount, titleTOC)
	}
	if countDate > 1 {
		t.Fatalf("date text appears %d times; expected once (no grid-footer duplication)\n%s", countDate, titleTOC)
	}
}

// RT-15.39: legacy title-block-align and footer-align continue to work when per-item align is unset.
func TestRT_15_39_LegacyTitleBlockAlignFallback(t *testing.T) {
	typst := renderIssue15Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    title-page:",
		"      title-block-align: left",
		"      footer-align: right",
		"",
	}, "\n"))
	// Legacy title-block-align: the title/subtitle/author group aligns left when no per-item override.
	assertContains(t, typst, `left + horizon`)
}

// -----------------------------------------------------------------------------
// AC15.6 -- page-header / page-footer compound page-pair alignment
// -----------------------------------------------------------------------------

// RT-15.40: page-header.align: left-right yields left on odd pages and right on even pages.
func TestRT_15_40_HeaderPairLeftRight(t *testing.T) {
	typst := renderIssue15Manuscript(t, "folio:\n  manuscript:\n    page-header:\n      align: left-right\n")
	body := extractBodyPageBlock(t, typst)
	assertContains(t, body, `calc.odd(counter(page).at(here()).first())`)
	// Both arms present with the correct compass tokens.
	assertContains(t, body, `left`)
	assertContains(t, body, `right`)
}

// RT-15.41: page-header.align: right-left -- left-page (verso, even) right, right-page (recto, odd) left.
// Template emits `if calc.odd(...) { <recto> } else { <verso> }` = `{ left } else { right }`.
func TestRT_15_41_HeaderPairRightLeft(t *testing.T) {
	typst := renderIssue15Manuscript(t, "folio:\n  manuscript:\n    page-header:\n      align: right-left\n")
	body := extractBodyPageBlock(t, typst)
	assertContains(t, body, `calc.odd(counter(page).at(here()).first())`)
	assertContains(t, body, `{ left } else { right }`)
}

// RT-15.42: page-header.align: center-left -- left-page (verso) center, right-page (recto) left.
func TestRT_15_42_HeaderPairCenterLeft(t *testing.T) {
	typst := renderIssue15Manuscript(t, "folio:\n  manuscript:\n    page-header:\n      align: center-left\n")
	body := extractBodyPageBlock(t, typst)
	assertContains(t, body, `{ left } else { center }`)
}

// RT-15.43: page-footer.align: right-center -- left-page (verso) right, right-page (recto) center.
func TestRT_15_43_FooterPairRightCenter(t *testing.T) {
	typst := renderIssue15Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    page-footer:",
		"      enabled: true",
		"      format: \"pair-footer\"",
		"      align: right-center",
		"",
	}, "\n"))
	body := extractBodyPageBlock(t, typst)
	assertContains(t, body, `calc.odd(counter(page).at(here()).first())`)
	assertContains(t, body, `{ center } else { right }`)
}

// RT-15.44: scalar align: right continues to apply uniformly (backwards compatible).
// Under the state-based skip-header wrapper, the header line becomes:
//   header: context { if state("folio-skip-header").get() { none } else { align(right)[...] } },
// which no longer contains calc.odd (the pair conditional).
func TestRT_15_44_ScalarAlignAppliesUniformly(t *testing.T) {
	typst := renderIssue15Manuscript(t, "folio:\n  manuscript:\n    page-header:\n      align: right\n")
	body := extractBodyPageBlock(t, typst)
	assertContains(t, body, `align(right)[`)
	assertNotContains(t, body, `calc.odd(`)
}

// RT-15.45: page-header.align: left-bogus is rejected with a diagnostic.
func TestRT_15_45_HeaderPairInvalidRightRejected(t *testing.T) {
	assertIssue15ConfigRejected(t, "folio:\n  manuscript:\n    page-header:\n      align: left-bogus\n", "left-bogus")
}

// RT-15.46: page-header.align: left-right-center (three tokens) is rejected.
func TestRT_15_46_HeaderPairThreeTokensRejected(t *testing.T) {
	assertIssue15ConfigRejected(t, "folio:\n  manuscript:\n    page-header:\n      align: left-right-center\n", "left-right-center")
}

// RT-15.47: on a single-page manuscript, the odd-page rule applies. The odd branch corresponds
// to the SECOND compound token (right-page). `right-left` -> right-page = left, so odd branch = left.
func TestRT_15_47_SinglePageUsesOddRule(t *testing.T) {
	typst := renderIssue15Manuscript(t, "folio:\n  manuscript:\n    page-header:\n      align: right-left\n")
	body := extractBodyPageBlock(t, typst)
	assertContains(t, body, `{ left } else { right }`)
}

// -----------------------------------------------------------------------------
// AC15.11 -- blank-page-before / blank-page-after enum on part / chapter / toc
// -----------------------------------------------------------------------------

// RT-15.63: chapter.blank-page-before: enforce-right emits a Typst pagebreak(to: "odd") directive.
func TestRT_15_63_ChapterBlankPageEnforceRight(t *testing.T) {
	typst := renderIssue15Manuscript(t, "folio:\n  manuscript:\n    chapter:\n      blank-page-before: enforce-right\n")
	assertContains(t, typst, `#pagebreak(weak: true, to: "odd")`)
	assertBefore(t, typst, `#pagebreak(weak: true, to: "odd")`, `#folio-chapter(first: false,`)
}

// RT-15.64: chapter.blank-page-after: enforce-left emits a Typst pagebreak(to: "even") directive after chapter.
func TestRT_15_64_ChapterBlankPageEnforceLeft(t *testing.T) {
	typst := renderIssue15Manuscript(t, "folio:\n  manuscript:\n    chapter:\n      blank-page-after: enforce-left\n")
	assertContains(t, typst, `#pagebreak(weak: true, to: "even")`)
	assertBefore(t, typst, `#folio-chapter(first: false,`, `#pagebreak(weak: true, to: "even")`)
}

// RT-15.65: part.blank-page-before: enforce-right emits pagebreak(to: "odd") before part.
func TestRT_15_65_PartBlankPageEnforceRight(t *testing.T) {
	typst := renderIssue15Manuscript(t, "folio:\n  manuscript:\n    part:\n      blank-page-before: enforce-right\n")
	assertContains(t, typst, `#pagebreak(weak: true, to: "odd")`)
	assertBefore(t, typst, `#pagebreak(weak: true, to: "odd")`, `#folio-part(first: true,`)
}

// RT-15.66: toc.blank-page-before: enforce-right emits pagebreak(to: "odd") before the TOC section.
func TestRT_15_66_TOCBlankPageEnforceRight(t *testing.T) {
	typst := renderIssue15Manuscript(t, "folio:\n  manuscript:\n    toc:\n      blank-page-before: enforce-right\n")
	assertContains(t, typst, `#pagebreak(weak: true, to: "odd")`)
	// The directive must appear in the title/TOC region and precede the TOC heading.
	titleTOC := extractTitleAndTOCBlock(t, typst)
	assertContains(t, titleTOC, `#pagebreak(weak: true, to: "odd")`)
}

// RT-15.67: toc.blank-page-after: true emits an unconditional blank page after the TOC.
func TestRT_15_67_TOCBlankPageAfterTrue(t *testing.T) {
	typst := renderIssue15Manuscript(t, "folio:\n  manuscript:\n    toc:\n      blank-page-after: true\n")
	assertContains(t, typst, `#folio-blank-page()`)
	// It should appear in the TOC region (between title-page area and the body counter reset).
	titleTOC := extractTitleAndTOCBlock(t, typst)
	assertContains(t, titleTOC, `#folio-blank-page()`)
}

// RT-15.68: invalid blank-page value fails config load with a diagnostic naming the offending value.
func TestRT_15_68_InvalidBlankPageValueRejected(t *testing.T) {
	assertIssue15ConfigRejected(t, "folio:\n  manuscript:\n    chapter:\n      blank-page-before: enforce-diagonal\n", "enforce-diagonal")
}

// RT-15.69: unconditional bool true still works alongside the enum values (backwards compat).
func TestRT_15_69_BoolTrueStillEmitsUnconditionalBlankPage(t *testing.T) {
	typst := renderIssue15Manuscript(t, "folio:\n  manuscript:\n    chapter:\n      blank-page-before: true\n")
	assertContains(t, typst, `#folio-blank-page()`)
	assertNotContains(t, typst, `#pagebreak(weak: true, to: "odd")`)
}

// -----------------------------------------------------------------------------
// AC15.12 -- publisher-ready title-page corner defaults in the British and US presets
// -----------------------------------------------------------------------------

// RT-15.70: an unconfigured British manuscript places each title-page item at the shipped
// corner default. Verified via #place() calls in generated Typst.
func TestRT_15_70_BritishPresetShipsTitleCornerDefaults(t *testing.T) {
	typst := renderIssue15Manuscript(t, "")
	// Title at top-center, author at top-right, version at top-left,
	// date at bottom-left, wordcount at bottom-right, contact at bottom-center.
	assertContains(t, typst, `#place(top + center`)
	assertContains(t, typst, `#place(top + right`)
	assertContains(t, typst, `#place(top + left`)
	assertContains(t, typst, `#place(bottom + left`)
	assertContains(t, typst, `#place(bottom + right`)
	assertContains(t, typst, `#place(bottom + center`)
}

// RT-15.71: a user override that empties per-item aligns reverts to the legacy centred group.
func TestRT_15_71_ExplicitEmptyPerItemAlignRevertsToLegacyGroup(t *testing.T) {
	typst := renderIssue15Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    title-page:",
		"      title-block-align: center",
		"      title:",
		"        align: \"\"",
		"      subtitle:",
		"        align: \"\"",
		"      author:",
		"        align: \"\"",
		"      date:",
		"        align: \"\"",
		"      wordcount:",
		"        align: \"\"",
		"      version:",
		"        align: \"\"",
		"      contact:",
		"        align: \"\"",
		"",
	}, "\n"))
	// With per-item aligns cleared and title-block-align set to center, the legacy centred
	// group is used for title/subtitle/author. No per-item #place(top + right)[...] etc.
	// (Contact still uses #place(top + left, float: true) as its default legacy placement.)
	titleTOC := extractTitleAndTOCBlock(t, typst)
	// The title group is aligned via #align(center + horizon) legacy path.
	assertContains(t, titleTOC, `center + horizon`)
	// The stack under the legacy align contains the title text.
	assertContains(t, titleTOC, `The Glass Orchard`)
}

// -----------------------------------------------------------------------------
// AC15.9 -- skip-header / skip-footer on part and chapter blocks
// -----------------------------------------------------------------------------

// RT-15.56: chapter.skip-header: true is threaded to the folio-chapter macro.
func TestRT_15_56_ChapterSkipHeader(t *testing.T) {
	typst := renderIssue15Manuscript(t, "folio:\n  manuscript:\n    chapter:\n      skip-header: true\n")
	assertContains(t, typst, `skip-header: true`)
	assertContains(t, typst, `#folio-chapter(first: false, skip-header: true`)
}

// RT-15.57: chapter.skip-footer: true is threaded to the folio-chapter macro.
func TestRT_15_57_ChapterSkipFooter(t *testing.T) {
	typst := renderIssue15Manuscript(t, "folio:\n  manuscript:\n    chapter:\n      skip-footer: true\n")
	assertContains(t, typst, `skip-footer: true`)
	assertContains(t, typst, `#folio-chapter(first: false, `)
}

// RT-15.58: part.skip-header: true is threaded to the folio-part macro.
func TestRT_15_58_PartSkipHeader(t *testing.T) {
	typst := renderIssue15Manuscript(t, "folio:\n  manuscript:\n    part:\n      skip-header: true\n")
	assertContains(t, typst, `#folio-part(first: true, skip-header: true`)
}

// RT-15.59: part.skip-footer: true is threaded to the folio-part macro.
func TestRT_15_59_PartSkipFooter(t *testing.T) {
	typst := renderIssue15Manuscript(t, "folio:\n  manuscript:\n    part:\n      skip-footer: true\n")
	assertContains(t, typst, `#folio-part(first: true, `)
	assertContains(t, typst, `skip-footer: true`)
}

// RT-15.60: skip-header/skip-footer default false; folio-part/folio-chapter calls omit the
// skip params (their template defaults are false).
func TestRT_15_60_SkipDefaultsFalse(t *testing.T) {
	typst := renderIssue15Manuscript(t, "")
	// Default calls carry no skip-header or skip-footer keyword arguments.
	assertNotContains(t, typst, `skip-header: true`)
	assertNotContains(t, typst, `skip-footer: true`)
}

// RT-15.61: template folio-part macro accepts skip-header and skip-footer parameters. The
// suppression itself is done via the folio-skip-header-pages / folio-skip-footer-pages
// state lists: folio-part / folio-chapter push their current page number to the appropriate
// list when their flag is true, and the header/footer context checks list membership so only
// the specific heading page is hidden (not subsequent multi-page body pages).
func TestRT_15_61_TemplateFolioPartMacroSupportsSkip(t *testing.T) {
	typst := renderIssue15Manuscript(t, "")
	assertContains(t, typst, `#let folio-part(first: false, skip-header: false, skip-footer: false`)
	assertContains(t, typst, `state("folio-skip-header-pages"`)
	assertContains(t, typst, `state("folio-skip-footer-pages"`)
	// Header/footer contexts check membership of the current page in the skip list.
	assertContains(t, typst, `pg in state("folio-skip-header-pages", ()).final()`)
	assertContains(t, typst, `pg in state("folio-skip-footer-pages", ()).final()`)
}

// RT-15.62: template folio-chapter macro accepts skip-header and skip-footer parameters.
func TestRT_15_62_TemplateFolioChapterMacroSupportsSkip(t *testing.T) {
	typst := renderIssue15Manuscript(t, "")
	assertContains(t, typst, `#let folio-chapter(first: false, skip-header: false, skip-footer: false`)
}

// -----------------------------------------------------------------------------
// AC15.7 -- binding gutter
// -----------------------------------------------------------------------------

// RT-15.48: default gutter is 0mm and the running-page margin keeps the pre-gutter rest: <base> shape.
func TestRT_15_48_GutterDefaultsToZero(t *testing.T) {
	typst := renderIssue15Manuscript(t, "")
	body := extractBodyPageBlock(t, typst)
	assertContains(t, body, `rest: 20mm`)
	assertNotContains(t, body, `inside:`)
	assertNotContains(t, body, `outside:`)
}

// RT-15.49: gutter: 15mm emits an inside margin of base + 15mm.
func TestRT_15_49_GutterInsideMarginAdded(t *testing.T) {
	typst := renderIssue15Manuscript(t, "folio:\n  manuscript:\n    gutter: 15mm\n")
	body := extractBodyPageBlock(t, typst)
	assertContains(t, body, `inside: 20mm + 15mm`)
}

// RT-15.50: gutter: 15mm emits outside margin of base only (asymmetric binding-side padding).
func TestRT_15_50_GutterOutsideMarginBase(t *testing.T) {
	typst := renderIssue15Manuscript(t, "folio:\n  manuscript:\n    gutter: 15mm\n")
	body := extractBodyPageBlock(t, typst)
	assertContains(t, body, `outside: 20mm`)
}

// RT-15.51: gutter: 0mm explicitly configured produces the same shape as unconfigured.
func TestRT_15_51_ExplicitZeroGutterMatchesUnconfigured(t *testing.T) {
	baseline := renderIssue15Manuscript(t, "")
	explicit := renderIssue15Manuscript(t, "folio:\n  manuscript:\n    gutter: 0mm\n")
	baselineBody := extractBodyPageBlock(t, baseline)
	explicitBody := extractBodyPageBlock(t, explicit)
	if baselineBody != explicitBody {
		t.Fatalf("explicit gutter: 0mm should match unconfigured baseline\n--- baseline ---\n%s\n--- explicit ---\n%s", baselineBody, explicitBody)
	}
}

// -----------------------------------------------------------------------------
// AC15.8 -- amended default header format/align and default footer
// -----------------------------------------------------------------------------

// RT-15.52: the default page-header.format is "[title] • [chapter] • [author]".
// The generated Typst substitutes [title]/[author] literally and [chapter] via state.
func TestRT_15_52_DefaultHeaderFormatIsTitleChapterAuthor(t *testing.T) {
	typst := renderIssue15Manuscript(t, "")
	body := extractBodyPageBlock(t, typst)
	// Header text contains the bullet separator and both literal substitutions.
	assertContains(t, body, `The Glass Orchard • `)
	assertContains(t, body, ` • Example Author`)
	// The chapter placeholder is preserved as a state read (no chapter yet on this header line).
	assertContains(t, body, `state("folio-current-chapter").get()`)
}

// RT-15.53: the default page-header.align is left-right -- left-page (verso) left, right-page
// (recto) right. Template emits `{ right } else { left }` (odd branch = second token = right).
// This is the classical outer-edge running-head convention.
func TestRT_15_53_DefaultHeaderAlignIsLeftRight(t *testing.T) {
	typst := renderIssue15Manuscript(t, "")
	body := extractBodyPageBlock(t, typst)
	assertContains(t, body, `calc.odd(counter(page).at(here()).first())`)
	assertContains(t, body, `{ right } else { left }`)
}

// RT-15.54: the default page-footer is enabled with format "[page]" and align center.
func TestRT_15_54_DefaultFooterIsCenteredPageNumber(t *testing.T) {
	typst := renderIssue15Manuscript(t, "")
	body := extractBodyPageBlock(t, typst)
	assertContains(t, body, `footer: context {`)
	assertContains(t, body, `align(center)[`)
	assertContains(t, body, `#folio-display-page()`)
}

// RT-15.55: page-footer.enabled: false explicitly configured omits the running footer.
func TestRT_15_55_ExplicitFooterDisableOmitsFooter(t *testing.T) {
	typst := renderIssue15Manuscript(t, "folio:\n  manuscript:\n    page-footer:\n      enabled: false\n")
	body := extractBodyPageBlock(t, typst)
	if strings.Contains(body, `footer: context {`) {
		t.Fatalf("expected explicit page-footer.enabled: false to omit footer, got:\n%s", body)
	}
	assertContains(t, body, `footer: none`)
}

// -----------------------------------------------------------------------------
// Shared helpers
// -----------------------------------------------------------------------------

func renderIssue15Manuscript(t *testing.T, scriptYAML string) string {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	// The manuscript loader reads script.yaml from the source directory of the first input file.
	// Put the config file alongside the chapters so overrides apply.
	if scriptYAML != "" {
		writeFile(t, filepath.Join(dir, "part1", "script.yaml"), scriptYAML)
	}
	writeFile(t, filepath.Join(dir, "part1", "ch01.md"), markdownChapterOne())
	writeFile(t, filepath.Join(dir, "part1", "ch02.md"), markdownChapterTwo())
	output := filepath.Join(dir, "out.typ")
	runManuscriptDirect(t, filepath.Join(dir, "part?", "ch??.md"), output)
	return readFile(t, output)
}

func assertIssue15ConfigRejected(t *testing.T, scriptYAML string, wantInError string) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "script.yaml"), scriptYAML)
	writeFile(t, filepath.Join(dir, "ch01.md"), markdownChapterOne())
	output := filepath.Join(dir, "out.typ")
	var out bytes.Buffer
	err := RunWithIO([]string{filepath.Join(dir, "ch01.md"), output}, &out)
	if err == nil {
		t.Fatalf("expected config to be rejected; instead got success. yaml:\n%s", scriptYAML)
	}
	assertContains(t, err.Error()+out.String(), wantInError)
}

// extractBodyPageBlock returns the substring of the generated Typst that configures the running
// body pages -- the region between the `#counter(page).update(1)` line (which marks the start of
// the body) and the end of the document. This is the only region where page-header/page-footer
// customisation appears.
func extractBodyPageBlock(t *testing.T, typst string) string {
	t.Helper()
	// The body-page setup begins with this comment (added when the counter reset was removed).
	marker := `// counter(page) is intentionally NOT reset here.`
	idx := strings.Index(typst, marker)
	if idx < 0 {
		t.Fatalf("body-page marker %q not found in Typst output:\n%s", marker, typst)
	}
	return typst[idx:]
}

// extractTitleAndTOCBlock returns the substring covering the title page and TOC region, i.e.
// everything before the body-page marker.
func extractTitleAndTOCBlock(t *testing.T, typst string) string {
	t.Helper()
	// The body-page setup begins with this comment (added when the counter reset was removed).
	marker := `// counter(page) is intentionally NOT reset here.`
	idx := strings.Index(typst, marker)
	if idx < 0 {
		t.Fatalf("body-page marker %q not found in Typst output:\n%s", marker, typst)
	}
	return typst[:idx]
}

// assertPlacedItemContains verifies that a `#place(<alignSpec>...)[...]` block in the Typst source
// contains the given item text within its content.
func assertPlacedItemContains(t *testing.T, typst string, alignSpec string, itemText string) {
	t.Helper()
	needle := `#place(` + alignSpec
	start := strings.Index(typst, needle)
	if start < 0 {
		t.Fatalf("place block with align spec %q not found in Typst output:\n%s", alignSpec, typst)
	}
	// Find the matching content brackets after the place() call. Scan from the first `[` past `)`.
	rest := typst[start:]
	openBracket := strings.Index(rest, "[")
	if openBracket < 0 {
		snippet := rest
		if len(snippet) > 200 {
			snippet = snippet[:200]
		}
		t.Fatalf("place block %q has no content bracket:\n%s", alignSpec, snippet)
	}
	// Take up to the next place() or 2000 chars, whichever is shorter -- enough to see the content.
	end := min(openBracket+2000, len(rest))
	region := rest[openBracket:end]
	if !strings.Contains(region, itemText) {
		t.Fatalf("place(%s) region does not contain %q; region:\n%s", alignSpec, itemText, region)
	}
}
