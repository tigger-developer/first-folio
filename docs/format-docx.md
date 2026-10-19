---
title: DOCX Script Export
version: "1.7"
last-updated: 2026-10-19
---

# DOCX script export

Export Org, Markdown or Fountain scripts to editable DOCX:

```sh
folio convert play.org play.docx
```

The destination extension selects DOCX. For redirected output, use
`folio convert play.org --to docx > play.docx`. Binary terminal output requires
`--force`. DOCX is write-only. Manuscript and letter commands also export DOCX
through mode adapters to the same native writer, as described below.
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

## Manuscripts and letters

Use `folio manuscript chapter.md manuscript.docx` or
`folio letter letters.org --format docx`. Both extend the existing native
DOCX writer with manuscript or letter configuration. Manuscripts use Pandoc
only for source parsing; letters need no Pandoc. They do not invoke Typst. Manuscript source syntax and letter recipient
selection remain unchanged.

Manuscripts retain editable headings, paragraphs, lists, tables, links and
footnotes, with generated title, contents and copyright matter. Contents use
live page references: update fields in the editor before printing. Native
sections separate numbering and running matter. Letters use recipient-specific
paragraphs and never inherit manuscript running matter. Reader pagination may
differ from PDF; editing and save/reopen need validation in the intended editor.

### Opening-page header/footer fidelity limitation

A manuscript section with special first-page handling uses the odd-page version
of any retained opening header or footer, even when that opening falls on an
even physical page. This is an accepted DOCX fidelity limitation. Special
handling includes opening-header/footer suppression, title sections and contents
continuation padding. Subsequent pages select their odd/even variants normally.

For an ordinary chapter or part opening:

| Opening-page settings | Physical page 43 | Physical page 44 |
|-----------------------|------------------|------------------|
| Keep both bands | Odd header and footer | Even header and footer |
| Hide header, keep footer | No header; odd footer | No header; odd footer instead of even |
| Keep header, hide footer | Odd header; no footer | Odd header instead of even; no footer |
| Hide both bands | Neither appears | Neither appears |

The chapter or part heading in the document body is unaffected. The limitation
changes running text or its alignment, not the displayed page counter. Identical
odd/even variants have no visible difference. These examples assume no numbering
restart. Suppression remains effective; a hidden header or footer is not restored.

Nonstandard weight handling has a separate
[documented limitation](font-embedding.md#prose-weight-fidelity).

### Font fidelity limitation

Earlier manuscript specimens rendered by this host's LibreOffice substituted
ordinary-width Iosevka Publish for Iosevka Publish Extended Bold and Extended
Italic in both DOCX and ODT. The exports retain the configured family references
and embed the Extended faces, but embedding does not guarantee that an editor
selects them. The resulting glyph widths can alter wrapping and pagination.
Use the ordinary family with the YAML `stretch` percentage when scaling is wanted.
The selection of an Extended face remains reader-dependent; its cause has
not been established. Microsoft Word behaviour remains unverified. Inspect the
result in the intended editor before relying on exact typography or pagination.

Input is limited to 32 MiB across manuscript files; Pandoc JSON and output are
bounded at 128 MiB. Pandoc calls have a 120-second deadline and 64 KiB diagnostic
limit. Package and font limits above also apply. Destination aliases, symlinks
and nonregular files are rejected before publication. A manuscript barcode
sidecar publishes after the document. Letter files publish individually; later
failure reports previously published paths rather than rolling them back.

## ODT

ODT is now a separate [configured Pandoc export](format-odt.md). The earlier
DOCX-to-ODT experiment preserved text and merged tables but lost role typography,
title composition, page settings and spacing. That generic conversion is not
the product ODT route. The native DOCX writer serves scripts, manuscripts and letters.

## History

- 2026-10-19: Extend the shared native DOCX and configured ODT writers to manuscript and letter adapters.

- 1.6 (2026-10-17): Document the accepted manuscript opening-page band limitation.

- 1.5 (2026-10-16): Document manuscript and letter office output.

- 1.4 (2026-10-12): Correct Org single-asterisk bold in body text and footnotes.

- 1.3 (2026-10-12): Documented independent script frontmatter paragraph fonts and body-font fallback.

- 1.2 (2026-10-12): Add complete font embedding and permission-aware warnings.

- 1.1 (2026-10-11): Link the separate configured ODT exporter; retain DOCX behaviour.

- 1.0 (2026-10-11): Document native DOCX adoption and its supported subset.
