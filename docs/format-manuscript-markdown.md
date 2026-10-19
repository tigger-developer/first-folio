---
title: Markdown Manuscript Format
version: "0.7"
last-updated: 2026-09-27
---

# Markdown Manuscript Format

Markdown manuscript input is a limited prose contract, separate from the Markdown stage-play contract. The [Markdown manuscript schema](../schema/manuscript.md) defines its source declaration and supported structures.

## Metadata Contract

Examples declare `schema: https://github.com/tigger-developer/first-folio/blob/master/schema/manuscript.md` in YAML frontmatter. The existing Markdown serializer emits that declaration even with empty metadata. The public manuscript command continues to write Typst/PDF only, without a schema declaration in those outputs.

All YAML frontmatter values are treated as manuscript strings. Quote values when that keeps the intent clearest, but the parser also accepts YAML scalars and converts them to strings, so `wordcount: about 90,000 words`, `wordcount: approx 100k words`, `wordcount: 20.000 mots`, and `wordcount: 90000` are all valid. Dates should be written as ISO strings such as `2026-07-06`; rendered output uses `folio.manuscript.date-format`.

Supported frontmatter fields are `title`, `subtitle`, `author`, `attribution`, `date`, `version`, `wordcount`, `contact-name`, `address`, `phone`, `email`, and `website`. `attribution` is optional and defaults to empty; when set, it prefixes the author name with a space, so `attribution: by` and `author: Example Author` render as `by Example Author`. `author-attribution` is accepted as a compatibility alias. `contact-name` is optional and is used only for the title-page contact block; it does not default to the manuscript author.

## Element Schema

| Markdown syntax | Manuscript meaning |
|---|---|
| YAML frontmatter bounded by `---` | Manuscript metadata |
| `# PART ONE` | Part divider page |
| `## Chapter 1` | Chapter start page |
| `### Section` and deeper | Local section heading |
| Plain paragraphs | Body prose |
| `***` or `---` on its own line, surrounded by blank lines | Section break, rendered as the configured manuscript scene-break marker |
| `**bold**` | Bold text |
| `*italic*` | Italic text |
| `~~deleted~~` | Strikethrough text |
| `` `code` `` within a mixed-content line | Inline monospace text |
| `` `code` `` as the only non-whitespace content on a line | Monospace code block |
| Fenced code blocks | Monospace code block |
| `--` and `---` | En dash and em dash |
| `[^name]` and `[^name]: text` | Footnote reference and definition |
| Blockquotes, links, lists, and tables | Supporting prose elements illustrated by the manuscript example |
| HTML comments | Private notes, excluded |
| Heading ending with an HTML comment containing `noexport` | Private section excluded until the next same-or-higher heading |

Setext headings are not part of the manuscript contract; use ATX headings (`#`, `##`, `###`) only. HTML blocks and source-document images are not supported.

Section breaks default to a centred `#` marker in rendered manuscripts. Override `folio.manuscript.scene-break.marker` in YAML config to use another marker.

Lists and tables default to `0.5em` vertical clearance; quoted and code blocks default to `6mm` in the British base, inherited by US. Set `folio.manuscript.quoted-block-spacing` or `folio.manuscript.code-block-spacing` to change both sides of those blocks equally. The existing `folio.manuscript.code-block.space-before` and `folio.manuscript.code-block.space-after` properties override the equal code-block spacing on their respective sides. Set `folio.manuscript.quote-block-indent` or `folio.manuscript.code-block-indent` to inset every line of the corresponding block from the left; both default to `0em` and are independent of prose first-line indentation. Configure blockquote typography through `folio.manuscript.quoted-block.font`; its `family`, `size`, `weight`, `stretch`, `style`, and `letter-spacing` properties inherit independently from the same quote-font role in lower configuration layers, not from the body font.

Body and quoted text support geometric font stretch and `folio.manuscript.hyphenation: auto|true|false`. Code remains unhyphenated. See [hyphenation modes and stretched-text limits](config.md#manuscript-hyphenation).

Quoted blocks add `quoted-block-spacing` to normal body paragraph separation. Their `quote-block-indent` is the complete left inset, without additional native quote or prose first-line indentation. Both settings use the body font size for `em` units; quote typography remains controlled by `quoted-block.font`.

Code-block clearance adds to normal body paragraph separation, including between consecutive blocks. Vertical `em` clearances use the body font size. Complete-line code spans use this same block layout; code embedded within prose remains inline.

Fenced code blocks take their typography from `folio.manuscript.mono.font`, the same role as inline code; `folio.manuscript.code-block` carries layout only. A configured `stretch` scales the block horizontally by that percentage, matching inline code, while the block keeps its glyph height, its column alignment and its page breaks.

A backtick code span that is the only non-whitespace content on its source line uses the fenced-code block layout, including configured spacing and indentation. Any non-whitespace content before the opening backtick or after the closing backtick keeps the code span inline. For example, in ``but `echo "text" #because this is all code` ``, `but ` is prose and everything between the backticks, including `#because`, is inline code. Existing fenced blocks are unchanged. Org manuscript verbatim spans follow the same distinction through canonical Markdown conversion.

Only a prose paragraph directly opening a chapter is flush left. Every other prose paragraph uses `folio.manuscript.paragraph-indent`, including prose after code blocks and blockquotes. When a chapter begins with a structural block, the prose following that block is indented normally.

Fountain is not accepted by manuscript mode.

## Example

```markdown
---
schema: https://github.com/tigger-developer/first-folio/blob/master/schema/manuscript.md
title: The Glass Orchard
subtitle: A Novel
author: Example Author
attribution: by
date: 2026-07-06
version: Draft 2
wordcount: about 90,000 words
contact-name: Example Agent
address: 100 Example Street / Sample City / Exampleland
phone: +353 1 000 0000
email: author@example.invalid
website: https://example.invalid
---

# PART ONE

## Chapter 1

The rain had been falling since Tuesday. The ledger flashed **WAIT** -- then the latch answered --- and Mira typed `nine-bell`.

***

By noon, the hands had moved backwards twice.

```

The previous example also included a private Notes section containing the text
"This planning note is excluded." That illustrates the legacy exclusion rule
described in the table; the example no longer embeds HTML-comment syntax.

## Revision History

- 0.7, 2026-09-27: Linked stretch and hyphenation controls and reconciled current block-clearance defaults.
- 0.6, 2026-09-26: Document exact quote indentation and additive quote clearance.
- 0.5, 2026-09-25: Clarified additive code-block clearance and complete-line code spans.
- 0.4, 2026-09-17: Record that fenced code blocks take their typography from the monospace font role, that configured stretch scales them proportionally, and that `code-block` carries layout only.
- 0.3, 2026-09-10: Add the manuscript schema declaration, use YAML document-version metadata, and describe private-section syntax without embedding HTML comments.
- 0.2, 2026-08-10: Previous manuscript-reference revision.
