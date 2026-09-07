// ABOUTME: Renders the copyright frontmatter page (issue #21) from declarative
// ABOUTME: config into a single Typst content block for template insertion.
package manuscript

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// renderCopyrightPage composes the entire copyright-page emit as a single Typst
// content block: blank-page-before directive, skip-header/skip-footer state
// updates, alignment wrapper, credit blocks, body paragraphs, separator,
// publication lines, publisher, ISBN, and (optionally) barcode. Returns the
// empty string when the copyright block is not enabled.
func renderCopyrightPage(meta Metadata, cfg Config) string {
	c := cfg.Folio.Manuscript.Copyright
	if !c.Enabled {
		return ""
	}
	var b strings.Builder

	// Reset the title-page footer/numbering that would otherwise persist into the
	// copyright page (the British title-page grid footer for example). skip-header
	// / skip-footer are applied per-page via the state list below.
	b.WriteString("#set page(numbering: none, footer: none)\n")

	// Blank-page-before directive (enforce-left / enforce-right / true / false).
	if d := c.BlankPageBefore.TypstDirective(); d != "" {
		b.WriteString(d)
		b.WriteString("\n")
	}

	// Skip-header / skip-footer: record this page in the appropriate skip lists so
	// the running header/footer context suppresses them here. Runs inside a context
	// block because we need the current physical page number.
	if (c.SkipHeader != nil && *c.SkipHeader) || (c.SkipFooter != nil && *c.SkipFooter) {
		b.WriteString("#context {\n")
		b.WriteString("  let pg = counter(page).at(here()).first()\n")
		if c.SkipHeader != nil && *c.SkipHeader {
			b.WriteString(`  state("folio-skip-header-pages", ()).update(pages => pages + (pg,))` + "\n")
		}
		if c.SkipFooter != nil && *c.SkipFooter {
			b.WriteString(`  state("folio-skip-footer-pages", ()).update(pages => pages + (pg,))` + "\n")
		}
		b.WriteString("}\n")
	}

	// Content wrapper: font + alignment + block layout.
	b.WriteString(fmt.Sprintf("#text(%s)[\n", manuscriptFontArgs(c.Font)))
	b.WriteString(fmt.Sprintf("#set par(leading: %s)\n", copyrightLeading(c.LineSpacing)))
	b.WriteString(fmt.Sprintf("#align(%s)[\n", copyrightAlignExpr(c.Align)))

	// Split blocks into "top" (credits + body + separator) and "bottom"
	// (publication + publisher + ISBN + barcode), separated by #v(1fr) so
	// the bottom section pushes to the page bottom -- matching typical
	// publisher copyright-page layout.
	top, bottom := composeCopyrightBlocks(meta, cfg, c)
	for i, block := range top {
		if i > 0 {
			b.WriteString(fmt.Sprintf("#v(%s)\n", c.BlockSpacing))
		}
		b.WriteString(block)
		b.WriteString("\n")
	}
	if len(bottom) > 0 {
		b.WriteString("#v(1fr)\n")
		for i, block := range bottom {
			if i > 0 {
				b.WriteString(fmt.Sprintf("#v(%s)\n", c.BlockSpacing))
			}
			b.WriteString(block)
			b.WriteString("\n")
		}
	}

	b.WriteString("]\n") // close align
	b.WriteString("]\n") // close text

	// Blank-page-after directive.
	if d := c.BlankPageAfter.TypstDirective(); d != "" {
		b.WriteString(d)
		b.WriteString("\n")
	}

	// End of copyright page: hard pagebreak so the following block starts fresh.
	b.WriteString("#pagebreak()\n")

	return b.String()
}

// composeCopyrightBlocks returns two ordered lists of non-empty content blocks:
// top (credits + body + separator) rendered from the top of the page, and bottom
// (publication + publisher + ISBN + barcode) pushed to the page bottom by a
// #v(1fr) spacer in the calling emit. Empty blocks are omitted so block-spacing
// collapses.
func composeCopyrightBlocks(meta Metadata, cfg Config, c CopyrightConfig) (top, bottom []string) {
	// ---- Top: credits + body + separator ----
	for _, credit := range effectiveCredits(meta, c) {
		trimmed := strings.TrimSpace(credit)
		if trimmed == "" {
			continue
		}
		top = append(top, renderBodyParagraph(trimmed, cfg))
	}
	for _, para := range c.Body {
		trimmed := strings.TrimSpace(para)
		if trimmed == "" {
			continue
		}
		top = append(top, renderBodyParagraph(trimmed, cfg))
	}
	if c.Separator != "" {
		top = append(top, fmt.Sprintf("#v(%s)\n#align(center)[%s]\n#v(%s)",
			c.SeparatorSpaceBefore,
			escapeTypst(c.Separator),
			c.SeparatorSpaceAfter,
		))
	}

	// ---- Bottom: publication + publisher + ISBN + barcode ----
	if len(c.Publication) > 0 {
		var pub strings.Builder
		for i, line := range c.Publication {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				continue
			}
			if i > 0 {
				pub.WriteString(" \\\n")
			}
			pub.WriteString(escapeTypst(trimmed))
		}
		if pub.Len() > 0 {
			bottom = append(bottom, pub.String())
		}
	}
	if c.Publisher != "" {
		bottom = append(bottom, fmt.Sprintf("%s #text(%s)[%s]",
			escapeTypst(c.PublisherPreposition),
			manuscriptFontArgs(c.Label.Font),
			escapeTypst(c.Publisher),
		))
	}
	if c.ISBN != "" {
		bottom = append(bottom, fmt.Sprintf("#text(%s)[%s]: %s",
			manuscriptFontArgs(c.Label.Font),
			escapeTypst(c.ISBNLabel),
			escapeTypst(c.ISBN),
		))
	}
	if c.ISBNBarcode == "render" || c.ISBNBarcode == "render-and-file" {
		if svg, err := renderEAN13SVG(c.ISBN); err == nil {
			bottom = append(bottom, fmt.Sprintf("#image(bytes(%q), format: \"svg\", width: 40mm)", svg))
		}
	}
	return top, bottom
}

// effectiveCredits returns the user-configured credits list, or a single default
// line ("Copyright © YEAR Author Name.") if the list is empty and folio.author
// is set. Returns nil (no credit block) if the user set an explicit empty list
// AND folio.author is empty.
func effectiveCredits(meta Metadata, c CopyrightConfig) []string {
	if len(c.Credits) > 0 {
		return c.Credits
	}
	if meta.Author == "" {
		return nil
	}
	year := deriveYear(meta.Date)
	if year == "" {
		return []string{fmt.Sprintf("Copyright © %s.", meta.Author)}
	}
	return []string{fmt.Sprintf("Copyright © %s %s.", year, meta.Author)}
}

// deriveYear returns the 4-digit year parsed from meta.Date (YYYY-MM-DD form).
// meta.Date is guaranteed non-empty by applyMetadataOverrides (defaults to today).
func deriveYear(date string) string {
	if len(date) >= 4 {
		return date[:4]
	}
	return strconv.Itoa(time.Now().Year())
}

// renderBodyParagraph converts a markdown-mini body string to a Typst content
// paragraph, translating **bold** -> *bold*, *italic* -> _italic_, --- -> em-dash,
// -- -> en-dash. Emitted as a single paragraph -- callers apply block-spacing.
//
// Markdown rule syntax: when the entry is ONLY the horizontal-rule form
// (`---`, `***`, or `___` optionally with trailing whitespace), render as a
// scene-break line using the configured scene-break marker (centred).
func renderBodyParagraph(md string, cfg Config) string {
	trimmed := strings.TrimSpace(md)
	if isMarkdownRule(trimmed) {
		marker := cfg.Folio.Manuscript.SceneBreak.Marker
		if marker == "" {
			marker = "#"
		}
		return fmt.Sprintf("#align(center)[%s]", escapeTypst(marker))
	}
	return markdownMiniToTypst(md)
}

// isMarkdownRule returns true if s is exactly a Markdown thematic-break token
// (3+ hyphens, asterisks, or underscores, all matching, no other characters).
func isMarkdownRule(s string) bool {
	if len(s) < 3 {
		return false
	}
	ch := s[0]
	if ch != '-' && ch != '*' && ch != '_' {
		return false
	}
	for i := 1; i < len(s); i++ {
		if s[i] != ch {
			return false
		}
	}
	return true
}

var (
	markdownBoldRE   = regexp.MustCompile(`\*\*([^*]+)\*\*`)
	markdownItalicRE = regexp.MustCompile(`(?:^|[^*])(\*[^*]+\*)(?:[^*]|$)`)
)

// markdownMiniToTypst converts a small subset of Markdown inline markup to Typst
// syntax. This intentionally keeps the surface small and predictable: bold,
// italic, en-dash, em-dash. Existing Typst inline syntax in the input passes
// through unchanged (double-conversion is idempotent for this subset).
func markdownMiniToTypst(md string) string {
	s := md
	// Em-dash / en-dash first so they don't get eaten by italic detection.
	s = strings.ReplaceAll(s, "---", "—")
	s = strings.ReplaceAll(s, "--", "–")
	// **bold** -> sentinel form so the italic pass below doesn't re-interpret the
	// single asterisks we just produced. We restore to Typst *bold* at the end.
	const (
		boldOpen  = "\x00BOLD_OPEN\x00"
		boldClose = "\x00BOLD_CLOSE\x00"
	)
	s = markdownBoldRE.ReplaceAllString(s, boldOpen+`$1`+boldClose)
	// Remaining single-* pairs are Markdown italic. Convert to Typst _italic_.
	s = convertMarkdownItalic(s)
	// Restore bold sentinels to Typst *bold*.
	s = strings.ReplaceAll(s, boldOpen, "*")
	s = strings.ReplaceAll(s, boldClose, "*")
	return s
}

// convertMarkdownItalic pairs up single-asterisks left after **bold** conversion
// and rewrites them as _italic_ Typst emphasis. Single-asterisk sequences at word
// boundaries only (avoids catching e.g. arithmetic 2*3 in copyright body).
func convertMarkdownItalic(s string) string {
	var out strings.Builder
	inItalic := false
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if ch == '*' {
			// Word-boundary check: opening italic requires a non-word char before
			// (or start), and non-space after; closing italic requires non-space
			// before and non-word after (or end).
			if !inItalic {
				// Opening: prev is start / non-word, next is non-space.
				prevOK := i == 0 || !isWordChar(s[i-1])
				nextOK := i+1 < len(s) && s[i+1] != ' '
				if prevOK && nextOK {
					out.WriteByte('_')
					inItalic = true
					continue
				}
			} else {
				// Closing: prev is non-space, next is end / non-word.
				prevOK := i > 0 && s[i-1] != ' '
				nextOK := i+1 == len(s) || !isWordChar(s[i+1])
				if prevOK && nextOK {
					out.WriteByte('_')
					inItalic = false
					continue
				}
			}
		}
		out.WriteByte(ch)
	}
	return out.String()
}

func isWordChar(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || b == '_'
}

// copyrightAlignExpr converts a compass keyword to a Typst align expression.
func copyrightAlignExpr(align string) string {
	switch strings.ToLower(strings.TrimSpace(align)) {
	case "left":
		return "left"
	case "right":
		return "right"
	case "top-center", "top-centre":
		return "top + center"
	case "bottom-center", "bottom-centre":
		return "bottom + center"
	default:
		return "center"
	}
}

func copyrightLeading(spacing string) string {
	if strings.TrimSpace(spacing) == "" {
		return "0.65em"
	}
	// spacing here is a Typst-compatible length (e.g. "1.15em"); convert to a
	// leading value (multiplier - 1)em when we detect a bare multiplier.
	if v, err := strconv.ParseFloat(strings.TrimSpace(spacing), 64); err == nil {
		return fmt.Sprintf("%gem", v-1)
	}
	// Otherwise pass through (already an em/pt length).
	return spacing
}

func orFallback(v, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return v
}
