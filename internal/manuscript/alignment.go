// ABOUTME: Manuscript alignment parser for title-page items and header/footer.
// ABOUTME: Handles compass keywords, compound V-H pairs, and book-layout page-pair aliases.
package manuscript

import (
	"fmt"
	"strings"
)

// TitleItemAlign returns the Typst alignment expression for a title-page item align value.
// Accepted forms:
//   - group: retain the standard title-page group layout
//   - bare compass keyword: left | center | right -> horizontal + horizon (Typst's vertical centre)
//   - compound V-H: V in {top, center, bottom}, H in {left, center, right}
//
// Returns an error whose message names the offending value if the input is not recognized.
func TitleItemAlign(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "group" {
		return "", nil
	}
	switch value {
	case "left":
		return "left + horizon", nil
	case "center":
		return "center + horizon", nil
	case "right":
		return "right + horizon", nil
	}
	parts := strings.Split(value, "-")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid title-page alignment %q: expected group, compass keyword (left|center|right) or compound V-H (e.g. top-left, bottom-center)", value)
	}
	vert, ok := titleVertical[parts[0]]
	if !ok {
		return "", fmt.Errorf("invalid title-page alignment %q: unknown vertical token %q (expected top, center, or bottom)", value, parts[0])
	}
	horiz, ok := titleHorizontal[parts[1]]
	if !ok {
		return "", fmt.Errorf("invalid title-page alignment %q: unknown horizontal token %q (expected left, center, or right)", value, parts[1])
	}
	return vert + " + " + horiz, nil
}

// TitleItemFloatable reports whether an item alignment is safe to place with `float: true`.
// Typst's `#place(alignment, float: true)` requires the vertical component to be `top` or
// `bottom`; anything else (including the compass center's `horizon` mapping) rejects float.
func TitleItemFloatable(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return false
	}
	// Bare compass keyword maps to `X + horizon` -- not floatable.
	if _, ok := horizontalCompass[value]; ok {
		return false
	}
	parts := strings.Split(value, "-")
	if len(parts) != 2 {
		return false
	}
	return parts[0] == "top" || parts[0] == "bottom"
}

var (
	titleVertical = map[string]string{
		"top":    "top",
		"center": "horizon",
		"bottom": "bottom",
	}
	titleHorizontal = map[string]string{
		"left":   "left",
		"center": "center",
		"right":  "right",
	}
	horizontalCompass = map[string]struct{}{
		"left":   {},
		"center": {},
		"right":  {},
	}
)

// HeaderFooterAlignSpec is the parsed shape of a page-header or page-footer align value.
// When IsPair is true, EvenArm applies on left-hand (verso, even) pages and OddArm on right-hand
// (recto, odd) pages. In the source syntax the FIRST token is the LEFT (verso, even) page's
// alignment and the SECOND token is the RIGHT (recto, odd) page's alignment -- matching the
// visual layout of an open book (left of reader = even, right of reader = odd), per issue #15.
// When IsPair is false, Scalar applies on every page (uniform alignment).
type HeaderFooterAlignSpec struct {
	IsPair  bool
	Scalar  string
	OddArm  string
	EvenArm string
}

// ParseHeaderFooterAlign parses a page-header or page-footer align value.
// Accepted forms:
//   - scalar compass keyword: left | center | right -> uniform alignment
//   - compound page-pair: A-B where both tokens are in {left, center, right}, treated as
//     (left-page, right-page) alignments -- i.e. (even-page, odd-page). The compound `left-right`
//     therefore places left-alignment on verso pages and right-alignment on recto pages, which is
//     the classical outer-edge running-head convention.
//
// Returns an error naming the offending value if the input is not recognized.
func ParseHeaderFooterAlign(value string) (HeaderFooterAlignSpec, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return HeaderFooterAlignSpec{}, nil
	}
	if _, ok := horizontalCompass[value]; ok {
		return HeaderFooterAlignSpec{Scalar: value}, nil
	}
	parts := strings.Split(value, "-")
	if len(parts) != 2 {
		return HeaderFooterAlignSpec{}, fmt.Errorf("invalid page-header/page-footer alignment %q: expected compass keyword (left|center|right) or compound page-pair (e.g. left-right, center-left)", value)
	}
	if _, ok := horizontalCompass[parts[0]]; !ok {
		return HeaderFooterAlignSpec{}, fmt.Errorf("invalid page-header/page-footer alignment %q: unknown left-page token %q (expected left, center, or right)", value, parts[0])
	}
	if _, ok := horizontalCompass[parts[1]]; !ok {
		return HeaderFooterAlignSpec{}, fmt.Errorf("invalid page-header/page-footer alignment %q: unknown right-page token %q (expected left, center, or right)", value, parts[1])
	}
	// First token = LEFT page (verso, even). Second token = RIGHT page (recto, odd).
	return HeaderFooterAlignSpec{IsPair: true, EvenArm: parts[0], OddArm: parts[1]}, nil
}

// TypstAlignExpression returns the Typst expression to use inside `align(...)` for a
// HeaderFooterAlignSpec. When scalar, it's the compass keyword. When paired, it's a
// `#context if calc.odd(...) { odd } else { even }`-style conditional expression.
func (s HeaderFooterAlignSpec) TypstAlignExpression() string {
	if s.IsPair {
		return "if calc.odd(counter(page).at(here()).first()) { " + s.OddArm + " } else { " + s.EvenArm + " }"
	}
	return s.Scalar
}
