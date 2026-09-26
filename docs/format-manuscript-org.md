---
title: Org-mode Manuscript Format
version: "0.6"
last-updated: 2026-09-26
---

# Org-mode Manuscript Format

Org-mode manuscript input uses org front matter and headings for prose manuscript structure. It is separate from the org-mode stage-play contract.

The [Org manuscript schema](../schema/manuscript.org) defines its declaration and source structure. Examples include the full schema URL. The public manuscript command writes Typst/PDF; there is currently no manuscript Org writer.

## Metadata Contract

All org front matter values are treated as manuscript strings. `#+WORDCOUNT: about 90,000 words`, `#+WORDCOUNT: approx 100k words`, `#+WORDCOUNT: 20.000 mots`, and `#+WORDCOUNT: 90000` are all valid because the field is rendered as entered. Dates should be written as ISO strings such as `2026-07-06`; rendered output uses `folio.manuscript.date-format`.

Supported front matter fields are `TITLE`, `SUBTITLE`, `AUTHOR`, `ATTRIBUTION`, `DATE`, `VERSION`, `WORDCOUNT`, `CONTACT-NAME`, `ADDRESS`, `PHONE`, `EMAIL`, and `WEBSITE`. `ATTRIBUTION` is optional and defaults to empty; when set, it prefixes the author name with a space, so `#+ATTRIBUTION: by` and `#+AUTHOR: Example Author` render as `by Example Author`. `AUTHOR-ATTRIBUTION` is accepted as a compatibility alias. `CONTACT-NAME` is optional and is used only for the title-page contact block; it does not default to the manuscript author.

## Element Schema

| Org syntax | Manuscript meaning |
|---|---|
| `#+TITLE: The Glass Orchard` | Manuscript title |
| `#+SUBTITLE: A Novel` | Subtitle |
| `#+AUTHOR: Example Author` | Author name |
| `#+ATTRIBUTION: by` | Optional author attribution prefix |
| `#+DATE: 2026-07-06` | Manuscript date |
| `#+VERSION: Draft 4` | Draft/version marker |
| `#+WORDCOUNT: about 90,000 words` | Approximate word count |
| `#+CONTACT-NAME: Example Agent` | Optional title-page contact name |
| `#+ADDRESS: ...` | Postal address |
| `#+PHONE: ...` | Phone number |
| `#+EMAIL: ...` | Email address |
| `#+WEBSITE: ...` | Website |
| `* PART ONE` | Part divider page |
| `** Chapter 1` | Chapter start page |
| `*** Section` and deeper | Local section heading |
| Plain paragraphs | Body prose |
| `-----` or `_____` on its own line | Section break, rendered as the configured manuscript scene-break marker |
| `*bold*` | Bold text |
| `/italic/` | Italic text |
| `+deleted+` | Strikethrough text |
| `=code=` and source blocks | Monospace text |
| `--` and `---` | En dash and em dash |
| `[fn:name]` and `[fn:name] Text` | Footnote reference and definition |
| Quotes, links, lists, and tables | Standard org-mode document elements |
| Heading tagged `:noexport:` | Private section excluded with children |

Fountain is not accepted by manuscript mode. Source-document images are not supported.

Section breaks default to a centred `#` marker in rendered manuscripts. Override `folio.manuscript.scene-break.marker` in YAML config to use another marker.

Lists, tables, and source blocks render with `0.5em` vertical padding before and after by default. Override `folio.manuscript.list.space-before`, `folio.manuscript.list.space-after`, `folio.manuscript.table.space-before`, `folio.manuscript.table.space-after`, `folio.manuscript.code-block.space-before`, and `folio.manuscript.code-block.space-after` to adjust this spacing.

Quoted blocks add `quoted-block-spacing` to normal body paragraph separation. Their `quote-block-indent` is the complete left inset, without additional native quote or prose first-line indentation. Both settings use the body font size for `em` units; quote typography remains controlled by `quoted-block.font`.

Code-block clearance adds to normal body paragraph separation, including between consecutive blocks. Vertical `em` clearances use the body font size. Complete-line code spans use this same block layout; code embedded within prose remains inline.

Source blocks take their typography from `folio.manuscript.mono.font`, the same role as verbatim spans; `folio.manuscript.code-block` carries layout only. A configured `stretch` scales the block horizontally by that percentage, matching inline verbatim text, while the block keeps its glyph height, its column alignment and its page breaks.

## Example

```org
#+SCHEMA: https://github.com/tigger-developer/first-folio/blob/master/schema/manuscript.org
#+TITLE: The Glass Orchard
#+SUBTITLE: A Novel
#+AUTHOR: Example Author
#+ATTRIBUTION: by
#+DATE: 2026-07-06
#+VERSION: Draft 4
#+WORDCOUNT: about 90,000 words
#+CONTACT-NAME: Example Agent
#+ADDRESS: 100 Example Street / Sample City / Exampleland
#+PHONE: +353 1 000 0000
#+EMAIL: author@example.invalid
#+WEBSITE: https://example.invalid

* PART ONE
** Chapter 1
The rain had been falling since Tuesday. The ledger flashed *WAIT* -- then the latch answered --- and Mira typed =nine-bell=.

-----

By noon, the hands had moved backwards twice.

*** Notes :noexport:
This planning note is excluded.
```

## Revision History

- 0.6, 2026-09-26: Document exact quote indentation and additive quote clearance.
- 0.5, 2026-09-25: Clarified additive code-block clearance and complete-line code spans.
- 0.4, 2026-09-17: Record that source blocks take their typography from the monospace font role, that configured stretch scales them proportionally, and that `code-block` carries layout only.
- 0.3, 2026-09-10: Add the Org manuscript schema declaration and use YAML document-version metadata.
- 0.2, 2026-08-10: Previous manuscript-reference revision.
