<!-- Version: 0.1 | Last updated: 2026-07-06 -->

# Org-mode Manuscript Format

Org-mode manuscript input uses org front matter and headings for prose manuscript structure. It is separate from the org-mode stage-play contract.

## Metadata Contract

All org front matter values are treated as manuscript strings. `#+WORDCOUNT: about 90,000 words`, `#+WORDCOUNT: approx 100k words`, `#+WORDCOUNT: 20.000 mots`, and `#+WORDCOUNT: 90000` are all valid because the field is rendered as entered. Dates should be written as ISO strings such as `2026-07-06`; rendered output uses `folio.manuscript.date-format`.

Supported front matter fields are `TITLE`, `SUBTITLE`, `AUTHOR`, `ATTRIBUTION`, `DATE`, `VERSION`, `WORDCOUNT`, `CONTACT-NAME`, `ADDRESS`, `PHONE`, `EMAIL`, and `WEBSITE`. `ATTRIBUTION` is optional and defaults to empty; when set, it prefixes the author name with a space, so `#+ATTRIBUTION: by` and `#+AUTHOR: Taḋg Paul` render as `by Taḋg Paul`. `AUTHOR-ATTRIBUTION` is accepted as a compatibility alias. `CONTACT-NAME` is optional and is used only for the title-page contact block; it does not default to the manuscript author.

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

Fountain is not accepted by manuscript mode.

Section breaks default to a centred `#` marker in rendered manuscripts. Override `folio.manuscript.scene-break.marker` in YAML config to use another marker.

Lists, tables, and source blocks render with `0.5em` vertical padding before and after by default. Override `folio.manuscript.list.space-before`, `folio.manuscript.list.space-after`, `folio.manuscript.table.space-before`, `folio.manuscript.table.space-after`, `folio.manuscript.code-block.space-before`, and `folio.manuscript.code-block.space-after` to adjust this spacing.

## Example

```org
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
