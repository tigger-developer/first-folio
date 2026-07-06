<!-- Version: 0.1 | Last updated: 2026-07-06 -->

# Markdown Manuscript Format

Markdown manuscript input is a prose contract, separate from the Markdown stage-play contract.

## Metadata Contract

All YAML frontmatter values are treated as manuscript strings. Quote values when that keeps the intent clearest, but the parser also accepts YAML scalars and converts them to strings, so `wordcount: about 90,000 words`, `wordcount: approx 100k words`, `wordcount: 20.000 mots`, and `wordcount: 90000` are all valid. Dates should be written as ISO strings such as `2026-07-06`; rendered output uses `folio.manuscript.date-format`.

Supported frontmatter fields are `title`, `subtitle`, `author`, `attribution`, `date`, `version`, `wordcount`, `contact-name`, `address`, `phone`, `email`, and `website`. `attribution` is optional and defaults to empty; when set, it prefixes the author name with a space, so `attribution: by` and `author: Taḋg Paul` render as `by Taḋg Paul`. `author-attribution` is accepted as a compatibility alias. `contact-name` is optional and is used only for the title-page contact block; it does not default to the manuscript author.

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
| `` `code` `` and fenced code blocks | Monospace text |
| `--` and `---` | En dash and em dash |
| `[^name]` and `[^name]: text` | Footnote reference and definition |
| Blockquotes, links, lists, and tables | Standard Markdown document elements |
| HTML comments | Private notes, excluded |
| Heading ending `<!-- noexport -->` | Private section excluded until the next same-or-higher heading |

Setext headings are not part of the manuscript contract; use ATX headings (`#`, `##`, `###`) only. HTML blocks are not supported.

Section breaks default to a centred `#` marker in rendered manuscripts. Override `folio.manuscript.scene-break.marker` in YAML config to use another marker.

Lists, tables, and fenced code blocks render with `0.5em` vertical padding before and after by default. Override `folio.manuscript.list.space-before`, `folio.manuscript.list.space-after`, `folio.manuscript.table.space-before`, `folio.manuscript.table.space-after`, `folio.manuscript.code-block.space-before`, and `folio.manuscript.code-block.space-after` to adjust this spacing.

Fountain is not accepted by manuscript mode.

## Example

```markdown
---
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

### Notes <!-- noexport -->

This planning note is excluded.
```
