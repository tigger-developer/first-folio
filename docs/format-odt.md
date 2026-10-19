---
title: ODT Script Export
version: "1.3"
last-updated: 2026-10-12
---

# ODT script export

Export an Org, Markdown or Fountain script to editable ODT:

```sh
folio convert play.org play.odt
```

ODT uses Pandoc's JSON writer with a generated ODT reference containing the
resolved `script.yaml` typography and page settings. It does not convert a DOCX
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
output requires `--force`. ODT is write-only and available only to `convert`,
not the manuscript or letter commands.

## Formatting and editing

| Area | Support and limits |
|------|--------------------|
| British scripts | One continuous table per scene, with two-cell dialogue rows and full-width merged direction, prop and transition rows. Headings remain outside the table. |
| US and screenplay | Separate cue, instruction and dialogue paragraphs with configured indentation. |
| Typography | Resolved role fonts, sizes, regular/bold weights, italic, small caps, tracking and integer stretch. Oblique uses italic with a warning. Unsupported weights fail. Available editable font faces are embedded; missing faces depend on the reading host. |
| British spacing | Single line spacing and the same compact vertical spacing as DOCX. Cast rows have no trailing paragraph gap. |
| Introduction paragraphs | All six font properties at `folio.positioning.frontmatter.paragraph.font`; omitted properties inherit `folio.font`. Headings and speech keep their own roles. |
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
leaving a later decision about a common export engine open.

## Failure and resource handling

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

- 1.3 (2026-10-12): Correct Org single-asterisk bold in body text and footnotes.

- 1.2 (2026-10-12): Documented independent script frontmatter paragraph fonts and body-font fallback.

- 1.1 (2026-10-12): Embed full font files with native ODF references and manifest entries.

- 1.0 (2026-10-11): Add configured Pandoc ODT export alongside native DOCX.
