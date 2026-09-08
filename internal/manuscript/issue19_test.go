// ABOUTME: Regression tests for issue #19 -- page-header/page-footer font style.
// ABOUTME: Verifies nested style values land in generated Typst without cross-role inheritance.
package manuscript

import (
	"strings"
	"testing"
)

// RT-19.1: page-header.font.style: italic emits style: "italic" inside the header text() call.
func TestRT_19_1_PageHeaderFontStyleItalic(t *testing.T) {
	typst := renderIssue15Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    page-header:",
		"      enabled: true",
		"      font:",
		"        style: italic",
		"",
	}, "\n"))
	header := extractHeaderBlock(t, typst)
	assertContains(t, header, `style: "italic"`)
}

// RT-19.2: page-header.font.style: oblique emits style: "oblique".
func TestRT_19_2_PageHeaderFontStyleOblique(t *testing.T) {
	typst := renderIssue15Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    page-header:",
		"      enabled: true",
		"      font:",
		"        style: oblique",
		"",
	}, "\n"))
	header := extractHeaderBlock(t, typst)
	assertContains(t, header, `style: "oblique"`)
}

// RT-19.3: the complete British block emits its explicit regular style as Typst normal.
func TestRT_19_3_PageHeaderFontStyleUnsetEmitsNoStyle(t *testing.T) {
	typst := renderIssue15Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    page-header:",
		"      enabled: true",
		"",
	}, "\n"))
	header := extractHeaderBlock(t, typst)
	assertContains(t, header, `style: "normal"`)
}

// RT-19.4: page-footer.font.style: italic emits style: "italic" inside the footer text() call.
func TestRT_19_4_PageFooterFontStyleItalic(t *testing.T) {
	typst := renderIssue15Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    page-footer:",
		"      enabled: true",
		"      format: \"[page]\"",
		"      font:",
		"        style: italic",
		"",
	}, "\n"))
	footer := extractFooterBlock(t, typst)
	assertContains(t, footer, `style: "italic"`)
}

// RT-19.5: page-footer.font.style remains role-local when page-header changes.
func TestRT_19_5_PageFooterFontStyleInheritsFromHeader(t *testing.T) {
	typst := renderIssue15Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    page-header:",
		"      font:",
		"        style: italic",
		"    page-footer:",
		"      enabled: true",
		"      format: \"[page]\"",
		"",
	}, "\n"))
	footer := extractFooterBlock(t, typst)
	assertContains(t, footer, `style: "normal"`)
}

// extractHeaderBlock returns the `header: context { ... }` region of the running-page
// set-page call, so header-specific assertions don't accidentally match the footer.
func extractHeaderBlock(t *testing.T, typst string) string {
	t.Helper()
	idx := strings.Index(typst, "header: context {")
	if idx == -1 {
		t.Fatalf("header: context block not found in generated Typst")
	}
	rest := typst[idx:]
	// Read until the matching closing `},` for this header block.
	end := strings.Index(rest, "},")
	if end == -1 {
		t.Fatalf("could not find end of header context block")
	}
	return rest[:end]
}

// extractFooterBlock returns the `footer: context { ... }` region so footer-specific
// assertions don't match the header block.
func extractFooterBlock(t *testing.T, typst string) string {
	t.Helper()
	idx := strings.Index(typst, "footer: context {")
	if idx == -1 {
		t.Fatalf("footer: context block not found in generated Typst")
	}
	rest := typst[idx:]
	end := strings.Index(rest, "},")
	if end == -1 {
		t.Fatalf("could not find end of footer context block")
	}
	return rest[:end]
}
