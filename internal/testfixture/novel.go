// ABOUTME: Generates deterministic equivalent Markdown and Org novels for W063 tests.
// ABOUTME: Supplies an independent ordered paragraph oracle without reading exported output.
package testfixture

import (
	"fmt"
	"strings"
)

const NovelParagraphs = 10000
const NovelChapters = 50

// Novel returns synthetic source and its expected plain body paragraphs.
func Novel(format string) (string, []string) {
	var out strings.Builder
	if format == "org" {
		out.WriteString("#+TITLE: The Quiet Courtyard\n#+AUTHOR: Synthetic Writer\n")
	} else {
		out.WriteString("---\ntitle: The Quiet Courtyard\nauthor: Synthetic Writer\n---\n")
	}
	paragraphs := make([]string, NovelParagraphs)
	for i := range paragraphs {
		if i%200 == 0 {
			prefix := "##"
			if format == "org" {
				prefix = "**"
			}
			fmt.Fprintf(&out, "\n%s CH%03d\n\n", prefix, i/200+1)
		}
		paragraphs[i] = fmt.Sprintf("P%05d Alice carried twelve letters through the quiet courtyard before dawn each morning END%05d.", i, i)
		out.WriteString(paragraphs[i] + "\n\n")
		if i == 0 || i == 5000 || i == 9999 {
			if format == "org" {
				fmt.Fprintf(&out, "*BOLD%05d* /ITALIC%05d/ =CODE%05d= [[https://example.invalid/novel][LINK%05d]] [fn:n%d]\n\n#+begin_quote\nQUOTE%05d\n#+end_quote\n\n- LIST%05d\n  - NESTED%05d\n\n| Column | Value |\n|--------+-------|\n| TABLE%05d | CELL%05d |\n\n[fn:n%d] NOTE%05d\n\n", i, i, i, i, i, i, i, i, i, i, i, i)
			} else {
				fmt.Fprintf(&out, "**BOLD%05d** *ITALIC%05d* `CODE%05d` [LINK%05d](https://example.invalid/novel) [^n%d]\n\n> QUOTE%05d\n\n- LIST%05d\n  - NESTED%05d\n\n| Column | Value |\n|--------|-------|\n| TABLE%05d | CELL%05d |\n\n[^n%d]: NOTE%05d\n\n", i, i, i, i, i, i, i, i, i, i, i, i)
			}
		}
	}
	if format == "org" {
		out.WriteString("\n** Private :noexport:\nPRIVATE-NOVEL-MARKER\n")
	} else {
		out.WriteString("\n## Private <!-- noexport -->\nPRIVATE-NOVEL-MARKER\n")
	}
	return out.String(), paragraphs
}

const NovelProfile = `folio:
  manuscript:
    margin: 20mm
    paragraph-indent: 0mm
    justify: false
    hyphenation: false
    line-spacing: 1.4
    font:
      family: Iosevka Publish
      size: 10pt
    title-page:
      enabled: true
    toc:
      enabled: true
      include-chapters: true
      page-numbers: true
      links: true
      line-spacing: 1.4
    chapter:
      page-break-before: true
      show-number: false
      show-name: true
      skip-header: false
      skip-footer: false
    copyright:
      enabled: true
      body:
        - Synthetic novel-scale copyright specimen.
    page-numbering:
      frontmatter-format: i
      body-format: '1'
      body-reset: first-part-or-chapter
  page-header:
    enabled: true
    format: NOVEL-EVEN [page]
    alt-format: NOVEL-ODD [page]
    frontmatter-format: NOVEL-FRONT-EVEN [page]
    alt-frontmatter-format: NOVEL-FRONT-ODD [page]
  page-footer:
    enabled: true
    format: NOVEL-PAGE [page]
`
