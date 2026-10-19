---
title: DOCX Script Export
version: "1.4"
last-updated: 2026-10-12
---

# DOCX script export

Export Org, Markdown or Fountain scripts to editable DOCX:

```sh
folio convert play.org play.docx
```

The destination extension selects DOCX. For redirected output, use
`folio convert play.org --to docx > play.docx`. Binary terminal output requires
`--force`. DOCX is write-only; manuscript and letter commands do not export DOCX.
The writer uses Go's XML and ZIP libraries and needs neither Pandoc nor Typst.

Available fonts are [embedded as complete faces](font-embedding.md), including
bundled Iosevka Publish. Unsupported, missing or restricted fonts produce warnings.

## Formatting and editing

| Area | Support and limits |
|------|--------------------|
| British scripts | One continuous table per scene: two-cell dialogue rows and full-width merged direction, prop and transition rows. Headings remain outside the table. |
| US and screenplay | Separate cue and dialogue paragraphs using resolved style geometry. |
| Typography | Role fonts, sizes, regular/bold weights, italic, tracking and integer stretch. Oblique uses italic; unsupported weights fail. |
| British spacing | Single line spacing; speech, direction, prop and transition vertical gaps are half the resolved spacing. Cast entries have no trailing paragraph gap. |
| Introduction paragraphs | All six font properties at `folio.positioning.frontmatter.paragraph.font`; omitted properties inherit `folio.font`. Headings and speech keep their own roles. |
| Content | Ordered text, inline formatting, cast tables and native footnotes; configured visibility switches apply. |
| Pages | Configured paper, margins and act breaks. Pagination depends on the reading application. Long rows can split across pages. |
| Title | Dedicated or compact title. Vertical positioning is approximate; date/version are in-flow rather than bottom-of-page slots. |
| Running matter | First/odd/even parts and native page fields. A single section uses body running matter for intro pages, with a warning for frontmatter overrides. |

Inline emphasis follows the source format: Org `*bold*` is bold and `/italic/`
is italic; Markdown and Fountain retain `**bold**` and `*italic*`. These rules
also apply inside footnotes. Existing `**bold**`, `***bold italic***`, underline
and literal-code handling remain supported.

The writer rejects source/destination identity, destination symlinks and invalid
XML text. Source size is limited to 32 MiB; generated XML and package sizes are
bounded at 128 MiB. Embedded fonts have a separate 64 MiB budget; the final
expanded package is limited to 256 MiB and 256 entries. Validated file output is staged beside its destination and
published by rename. Existing destination permissions are replaced by mode 0600.

## ODT

ODT is now a separate [configured Pandoc export](format-odt.md). The earlier
DOCX-to-ODT experiment preserved text and merged tables but lost role typography,
title composition, page settings and spacing. That generic conversion is not
the product ODT route. The native DOCX writer remains unchanged.

## History

- 1.4 (2026-10-12): Correct Org single-asterisk bold in body text and footnotes.

- 1.3 (2026-10-12): Documented independent script frontmatter paragraph fonts and body-font fallback.

- 1.2 (2026-10-12): Add complete font embedding and permission-aware warnings.

- 1.1 (2026-10-11): Link the separate configured ODT exporter; retain DOCX behaviour.

- 1.0 (2026-10-11): Document native DOCX adoption and its supported subset.
