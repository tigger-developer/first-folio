---
schema: https://github.com/tigger-developer/first-folio/blob/master/schema/manuscript.md
title: First Folio Markdown Manuscript Schema
version: "0.3"
updated: "2026-09-10"
---

# Markdown Manuscript Schema

This is First Folio's **limited Markdown prose format**, separate from its
script format. Only the structures described here belong to the manuscript
contract. `folio manuscript` reads this source and renders Typst or PDF.

**Reference example:** [The Glass Orchard](../examples/dummy-manuscript.md).

## Frontmatter

Document metadata is a YAML block at the beginning of the file:

```yaml
---
schema: https://github.com/tigger-developer/first-folio/blob/master/schema/manuscript.md
title: The Glass Orchard
subtitle: A Novel
author: Example Author
attribution: by
date: "2026-09-10"
version: Draft 1
wordcount: about 90,000 words
contact-name: Example Agent
address: 100 Example Street / Sample City
phone: +353 1 000 0000
email: author@example.invalid
website: https://example.invalid
---
```

- `schema` identifies this source format.
- `title`, `subtitle`, and `author` identify the manuscript.
- `attribution` is an optional author prefix; it defaults to empty.
  `author-attribution` is a compatibility alias.
- `date` and `version` provide the manuscript date and draft label.
- `wordcount` is display text, not a computed count. Numeric input is also
  accepted and converted to text.
- `contact-name`, `address`, `phone`, `email`, and `website` supply optional
  contact details. Contact name does not default to the author.

Metadata is interpreted as strings. ISO date strings preserve the date's
meaning; manuscript configuration controls the rendered date format.
The schema document's own YAML `version` describes this definition, while a
manuscript's `version` describes its draft.

## Prose structure

| Source form | Meaning |
|---|---|
| `# PART ONE` | Part division |
| `## Chapter 1` | Chapter |
| `### Section` and deeper headings | Local section |
| Plain paragraphs separated by blank lines | Prose |
| `***` or `---` on its own line, with blank lines around it | Scene break |

Only hash-prefixed headings are supported. Underlined headings are not.
The scene-break marker defaults to a centred `#` in rendered output.
Part and chapter numbering and layout are controlled by manuscript configuration.

Only a prose paragraph immediately following a chapter heading is flush left.
Other prose uses the configured paragraph indent, including prose after a
blockquote or code block that opens a chapter.

## Inline syntax and blocks

| Source form | Meaning |
|---|---|
| `**bold**` | Bold |
| `*italic*` | Italic |
| `~~deleted~~` | Strikethrough |
| A backtick code span within prose | Inline monospace text |
| A backtick code span alone on a line | Monospace block |
| A fenced code block | Monospace block |
| `--` and `---` within text | En dash and em dash |
| `[^note]` with a `[^note]: Explanation.` definition | Footnote |
| `> Quoted prose.` | Blockquote |
| `[label](URL)` | Link |
| Bulleted and numbered lists | Lists |
| Pipe tables | Tables |

Blockquotes are prose quotations in this format. The linked example shows
supported quotations, links, lists, tables, and emphasis together.

## Limits and private material

HTML blocks, source-document images, and Fountain input are outside the
manuscript contract. Support for the listed constructs does not imply support
for a general Markdown standard or every extension available in the parser.

The existing parser excludes private sections marked by a heading followed by
an HTML comment containing the word `noexport`, ending at the next heading of
the same or a higher level. It also excludes standalone HTML-comment notes.
This legacy convention is described here without embedding HTML comments in
the schema document or its examples.

## Output boundary

The existing Markdown serializer adds this schema URL even when metadata is
empty. The public manuscript command currently exposes Typst/PDF output only.
The schema declaration does not itself introduce Markdown or Org output modes.
Typst and PDF output do not receive a schema declaration.

## Related documentation

- [Markdown manuscript format reference](../docs/format-manuscript-markdown.md).
- [Org manuscript schema](manuscript.org).
- [Manuscript configuration](../docs/config.md).

## Revision history

- 0.3, 2026-09-10: Rewrite around the limited prose contract; use YAML version
  metadata and remove literal HTML comments from the schema document.
- 0.2, 2026-09-10: Record serializer and output boundaries.
- 0.1, 2026-09-10: Initial definition, superseded by this rewrite.
