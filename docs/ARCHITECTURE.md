---
title: Architecture
version: "0.12"
last-updated: 2026-10-12
---

# Architecture

First Folio is a CLI-first Go application. One `folio` executable owns stage-play conversion, cover-letter generation, and prose-manuscript rendering.

## Runtime

`cmd/folio` is the sole product entry point. It delegates process-independent command handling to `internal/app`, which returns an exit status and writes through injected streams.

The subcommands are:

- `folio convert`: parses Org, Markdown, or Fountain stage plays into typed events and emits Org, Markdown, Fountain, Typst, PDF, DOCX, or ODT.
- `folio letter`: parses Org `:letter:` sections and renders recipient-specific PDFs.
- `folio manuscript`: parses Markdown or Org prose manuscripts and renders Typst or PDF.

Typst and Pandoc remain explicit external tools where their documented features require them. No command invokes Perl or a project-owned shell script.

## Packages

| Path | Responsibility |
|---|---|
| `cmd/folio/` | Process entry point for the single executable |
| `internal/app/` | Public CLI dispatch, conversion orchestration, script rendering, and process-level integration tests |
| `internal/config/` | Shared embedded-preset and YAML configuration loading |
| `internal/play/` | Typed stage-play model plus Org, Markdown, and Fountain parsers/emitters |
| `internal/letter/` | Cover-letter model, Org parser, Typst renderer, and compiler |
| `internal/manuscript/` | Manuscript input, parser, canonicalization, rendering, and compilation |
| `templates/` | File-backed Typst layouts for scripts, letters, and manuscripts |
| `presets/` | British base and explicit style override YAML |
| `cmd/update-homebrew/` | Checked release-formula update and publication tooling |

The root `assets.go` embeds help documents, presets, Typst templates, and licensed font assets. The installed binary therefore does not discover a project root, depend on its working directory, or require a root environment variable.

`internal/typst.Compile` supplies the bundled Iosevka Publish fonts to script,
manuscript and letter PDF compilation. Each invocation extracts the embedded
faces into a private temporary directory, passes it through Typst's
`--font-path` alongside existing `TYPST_FONT_PATHS`, and removes it on success
or failure. It contains no style-specific behaviour. All font selection stays
in layered YAML. Exported Typst files require fonts on the compiling host.
Office exports use a shared bounded SFNT metadata resolver over bundled assets
and local font paths. A final packaging step adds complete editable font faces
to DOCX font parts or ODT font-face sources and manifest entries. It runs before
atomic publication and does not change the PDF route. Missing or restricted
fonts retain their references with warnings; see [font embedding](font-embedding.md).

The separate-line dialogue template gives its block the available width, so
configured speaker alignment operates across the dialogue area. British
same-line dialogue keeps its existing grid layout.

## Document Models

Stage plays and manuscripts intentionally retain separate semantic models.

Stage-play parsers produce typed events for metadata, acts, scenes, directions, speakers, dialogue, character tables, transitions, props, intro material, and footnotes. Text emitters consume that model, while script PDF rendering prepares escaped template data from it.

Manuscripts use metadata plus prose blocks such as parts, chapters, sections, paragraphs, lists, tables, code, blockquotes, links, scene breaks, and footnotes. Source-document images are not yet part of the manuscript contract; generated ISBN barcode SVGs are a separate copyright-page facility. Markdown remains the canonical shared manuscript contract for the current implementation; Org is canonicalized through that contract before rendering.

Letters have a smaller recipient-oriented model because their sender, recipient, substitution, and signoff semantics do not map cleanly onto either stage plays or manuscripts.

## Configuration

`internal/config` owns YAML file loading and precedence for every mode:

1. British built-in preset.
2. Selected built-in style override.
3. Global `~/.config/first-folio/script.yaml`.
4. Global style-specific YAML.
5. First Markdown manuscript input's `folio:` and `render:` mappings.
6. Selected local `script.yaml`.
7. Selected local style-specific YAML.
8. CLI overrides.

Manuscript orchestration reads Markdown source configuration before loading the effective configuration. The parser shares one frontmatter decoder with this extraction path; metadata remains a separate document concern. Later input files cannot supply another configuration layer. Other modes do not pass a source layer to the shared loader.

Local discovery begins at the source directory and walks upwards towards HOME. Only the nearest `script.yaml` is selected; its style-specific sibling is loaded from the same directory. For a multi-file manuscript, the first resolved input determines that starting directory.

The loader deep-merges ordinary keys and partial font blocks, validates every effective font role and every shipped alignment field against its allowlist after layer merging, and provides typed values to each renderer. Alignment validation occurs before typed decode or output; renderer alignment parsers translate accepted values into layout expressions. Font properties inherit only through configuration layers at the same role path. The exhaustive British preset is the shared lowest-precedence base for scripts, letters, and manuscripts; US and screenplay presets contain only their differences.

Running matter is configured at `folio.page-header` and `folio.page-footer` for scripts and manuscripts. Deprecated manuscript-local blocks are normalized within each configuration layer, before merging, so same-layer manuscript overrides do not bypass higher-layer shared settings. Manuscript command diagnostics report their use on stderr. Letter rendering ignores running-matter settings because letters are supplementary correspondence, not manuscript pages.

## Rendering

DOCX is a sibling output route after script parsing, configuration and filtering.
The `docx_*.go` adapter builds native WordprocessingML with `encoding/xml`,
packages it with `archive/zip`, and publishes through a same-directory staged
rename. No Pandoc or Typst process runs for DOCX. British scene tables, font
mapping, resource bounds and approximations are described in [DOCX export](format-docx.md).
The existing Typst route retains source-aware Org emphasis independently.

ODT is a separate sibling output route. `odt_*.go` projects the shared office
layout nodes into Pandoc JSON and generates an ODF reference containing resolved
role styles and page masters. The layout nodes currently use Word names; no
DOCX package is written or parsed in this route. Pandoc runs with fixed arguments,
a bounded output pipe and a 120-second deadline. The resulting ZIP/XML is
validated before the existing atomic publication helper runs. DOCX serialization
is unchanged. See [ODT export](format-odt.md) for compatibility and limits.


Substantial Typst layouts live in real `.typ` files. Go code owns:

- typed template data;
- context-specific escaping and validated layout literals;
- event-to-layout policy;
- temporary-file lifecycle;
- direct subprocess invocation and diagnostics.

Templates own page composition. Product Go code does not contain generated-language heredocs or invoke a shell to run Typst.

Script title composition uses one shared title block for a dedicated title page or a compact heading above first-page content. `folio.title-page.enabled` selects between them; manuscript title-page settings remain independent. Script-only Boolean validation covers this switch and `skip-header`/`skip-footer`. Running-matter suppression uses the physical page counter and applies to the dedicated title page or compact page 1, without resetting numbering or removing title metadata. The renderer suppresses only the first act's explicit break in compact mode; later act breaks and natural pagination remain. See [script title configuration](config.md#compact-script-title-block) for defaults and untitled-script handling.

Manuscript body, quote and monospace stretch use geometric horizontal scaling in
the template. For stretched prose, `internal/manuscript` embeds English patterns
and exceptions, applies the pinned Go hyphenation library, and supplies a
document-specific dictionary of discretionary break points. The template
preserves shaping-sensitive boundaries and leaves final line and page breaking
to Typst. Generated Typst remains self-contained with no runtime download.
The shared loader validates `folio.manuscript.hyphenation`; code disables
automatic hyphenation. See [hyphenation settings and limits](config.md#manuscript-hyphenation).

## Build And Tests

`make build` compiles `dist/folio`; `make install` builds first and then links that binary into the configured installation directory. Version values are injected with Go linker flags. Homebrew builds the same command directly.

Automated coverage is Go-owned. Unit tests cover parsing, emission, configuration, escaping, and rendering. Integration tests execute both the in-process public application and a built binary outside the checkout working directory. PDF-sensitive suites invoke Typst directly and inspect supported outputs rather than source implementation text.

## Migration History

Before issue #10, conversion and letters used a Perl dispatcher, Perl parsers/emitters, embedded YAML::Tiny, and shell regression suites; manuscripts used a separately built Go helper. Issue #10 replaced that split with the single Go runtime while preserving the public CLI and accepted rendering behaviour.

## Document History

- 0.12 (2026-10-12): Record shared font discovery and native office font packaging.

- 0.11 (2026-10-12): Record bundled font lifecycle and configured dialogue alignment.

- 0.9 (2026-10-11): Document the native DOCX sibling writer and atomic publication.

- 0.8 (2026-10-07): Documented shared script title composition, compact-mode act-break policy and first-page running-matter suppression.

- 0.7 (2026-10-07): Shared running-matter configuration for scripts and manuscripts, with layered compatibility for manuscript-local keys and exclusion of letters.

- 0.6 (2026-09-27): Documented manuscript geometric stretch and offline hyphenation responsibilities.

- 0.5 (2026-09-25): Added the Markdown manuscript source layer between global and local configuration.


- 0.4 (2026-09-24): Clarified shared-loader ownership of resolved alignment validation.
- 0.3 (2026-08-10): Previous architecture baseline.

- 0.10 (2026-10-11): Add ODT JSON projection and generated ODF reference.
