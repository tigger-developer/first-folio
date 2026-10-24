---
title: ODT Script Export
version: "1.9"
last-updated: 2026-10-24
---

# ODT script export

Export an Org, Markdown or Fountain script to editable ODT:

```sh
folio convert play.org play.odt
```

ODT uses Pandoc's JSON writer with a generated ODT reference containing the
resolved YAML cascade's typography and page settings. It does not convert a DOCX
file. DOCX retains its existing native writer; PDF retains its Typst renderer.

Pandoc must be installed and available on `PATH`. This route was verified with
Pandoc 3.9.0.2 and LibreOffice 26.2.3.2. It emits Pandoc JSON API version
1.23.1.1; incompatible Pandoc versions fail before publishing the destination.
No Typst executable is needed for ODT. No custom template, filter or extra
Pandoc argument interface is exposed.

Available fonts are [embedded as complete faces](font-embedding.md) after Pandoc
writes the document. Unsupported, missing or restricted fonts produce warnings.

The destination extension takes precedence over `--to`. With no target, use
`--to odt` or `folio.default-format: odt` and redirect stdout. Binary terminal
output requires `--force`. ODT is write-only. Manuscript and letter exports use
mode adapters to the shared configured ODT writer described below.

## Formatting and editing

| Area | Support and limits |
|------|--------------------|
| British scripts | One continuous table per scene, with two-cell dialogue rows and full-width merged direction, prop and transition rows. Headings remain outside the table. |
| US and screenplay | Separate cue, instruction and dialogue paragraphs with configured indentation. |
| Typography | Resolved role fonts, sizes, regular/bold weights, italic, small caps, tracking and integer stretch. Oblique uses italic with a warning. Unsupported weights fail. Available editable font faces are embedded; missing faces depend on the reading host. |
| British spacing | Single line spacing and the same compact vertical spacing as DOCX. Cast rows have no trailing paragraph gap. |
| Introduction paragraphs | Global and direct scene introductory paragraphs use all six font properties at `folio.positioning.frontmatter.paragraph.font`; omitted properties inherit `folio.font`. `render.frontmatter: false` hides both. Headings and speech keep their own roles. |
| Content | Ordered text, supported inline spans, cast tables and native footnotes. Configured visibility switches apply. Links and images are not fetched; the existing inline delimiter rules still apply to their text. |
| Pages | Configured A3/A4/A5/A6/Letter/Legal paper, margins and act breaks. Long table rows may split across pages. Pagination depends on the reading application. |
| Title | Dedicated or compact title. Vertical position is approximate; date/version follow the title rather than occupying bottom-of-page slots. |
| Running matter | First/odd/even page masters, title/author substitution and live page-number/page-count fields. Frontmatter uses body running matter, with a warning for configured frontmatter overrides. Header/footer clearance smaller than the font height is expanded with a warning. |

Inline emphasis follows the source format: Org `*bold*` is bold and `/italic/`
is italic; Markdown and Fountain retain `**bold**` and `*italic*`. These rules
also apply inside footnotes. Existing `**bold**`, `***bold italic***`, underline
and literal-code handling remain supported.

The shared inline parser can interpret slash pairs inside URLs as italic markers;
ODT also collapses repeated spaces in ordinary text. Use literal code spans where
inline delimiters must be retained. Neither office route is a general Org parser.

The ODT adapter shares the existing office layout calculations, whose internal
nodes currently use Word names. It projects those nodes to Pandoc JSON, then
supplies generated ODF styles. This preserves the current DOCX behaviour while
sharing the ODT writer and native reference-package generator with manuscript
and letter adapters. DOCX remains native and does not pass through Pandoc.

## Manuscripts and letters

Use `folio manuscript manuscript.org manuscript.odt` or
`folio letter letters.org --format odt`. Both use existing mode-specific YAML
and require Pandoc without Typst. Manuscript Markdown/Org and letter Org are parsed natively and
follow their PDF renderer's YAML cascade, including applicable manuscript source
frontmatter. Pandoc receives only structured JSON and the generated reference;
it does not read source Markdown or Org. They write ODT directly from structured prose;
no DOCX intermediate is created. The script approximations above do not define
the prose feature contract.

Ordinary prose uses the editor's natural character baseline. Native paragraph
line boxes control leading and spacing, keeping the caret and selection aligned
with body text. Footnote anchors retain genuine superscripts. Repeated headings
and contents entries share formatting styles by semantic level and effective
properties; bookmarks and page-boundary data remain specific to each instance.

Manuscripts use native paragraphs, lists, tables, footnotes, heading bookmarks,
editable positioned groups and contents page references. Update fields in the
editor before printing. An ODT frontmatter/body change of page-number format
starts a new page; configured displayed numbering and physical page-side rules
still apply. Letters retain recipient substitutions, Org emphasis and literal
spans, with their own font/margin/spacing configuration and no running matter.

### Frontmatter/body page-break limitation

The body starts on a new page when the frontmatter/body transition changes the
number format, requests a numbering reset, or changes the sequence of live
page/page-count fields in a header or footer. This can leave unused space at
the end of the introduction, even when no chapter page break was requested.
Configured numbering and page-side rules still apply. Wording changes around
the same page fields remain continuous when no reset or break is requested.

This is an accepted fidelity limitation for prose ODT exports.
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

The same resource and publication limits below apply. A manuscript's optional
barcode sidecar publishes after its document. Letter targets are preflighted
as a batch but published individually; later failure reports earlier outputs.

## Failure and resource handling

### Shared resource limits

Source input must be regular UTF-8 text representable in XML, at most 32 MiB.
The shared office layout bounds events to 50,000 and expanded inline segments
to 100,000. JSON, reference and output data are bounded at 128 MiB. Pandoc has
a 120-second deadline; captured diagnostics are bounded at 64 KiB.

The output must be an ODT ZIP containing well-formed XML. Each member is bounded
at 128 MiB, expanded contents at 256 MiB and entries at 256. Temporary references
are private and removed after conversion. Source/destination aliases,
nonregular destinations and destination symlinks are rejected. File output is
staged beside the destination and published by rename with mode 0600. A failed
conversion leaves an existing destination unchanged.

## History

- 1.9 (2026-10-24): Apply frontmatter visibility and typography to direct scene introductory paragraphs.
- 1.8 (2026-10-19): Preserve natural editable text baselines and reuse equivalent role styles.

- 1.7 (2026-10-19): Use native office source parsing and the PDF YAML cascade.

- 2026-10-19: Extend the shared native DOCX and configured ODT writers to manuscript and letter adapters.

- 1.5 (2026-10-17): Document accepted frontmatter/body reset and page-field transition breaks.

- 1.4 (2026-10-16): Document manuscript and letter office output.

- 1.3 (2026-10-12): Correct Org single-asterisk bold in body text and footnotes.

- 1.2 (2026-10-12): Documented independent script frontmatter paragraph fonts and body-font fallback.

- 1.1 (2026-10-12): Embed full font files with native ODF references and manifest entries.

- 1.0 (2026-10-11): Add configured Pandoc ODT export alongside native DOCX.
