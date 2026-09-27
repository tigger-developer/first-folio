---
title: Architecture
version: "0.6"
last-updated: 2026-09-27
---

# Architecture

First Folio is a CLI-first Go application. One `folio` executable owns stage-play conversion, cover-letter generation, and prose-manuscript rendering.

## Runtime

`cmd/folio` is the sole product entry point. It delegates process-independent command handling to `internal/app`, which returns an exit status and writes through injected streams.

The subcommands are:

- `folio convert`: parses Org, Markdown, or Fountain stage plays into typed events and emits Org, Markdown, Fountain, Typst, or PDF.
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

The root `assets.go` embeds help documents, presets, and Typst templates. The installed binary therefore does not discover a project root, depend on its working directory, or require a root environment variable.

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

## Rendering

Substantial Typst layouts live in real `.typ` files. Go code owns:

- typed template data;
- context-specific escaping and validated layout literals;
- event-to-layout policy;
- temporary-file lifecycle;
- direct subprocess invocation and diagnostics.

Templates own page composition. Product Go code does not contain generated-language heredocs or invoke a shell to run Typst.

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

- 0.6 (2026-09-27): Documented manuscript geometric stretch and offline hyphenation responsibilities.

- 0.5 (2026-09-25): Added the Markdown manuscript source layer between global and local configuration.


- 0.4 (2026-09-24): Clarified shared-loader ownership of resolved alignment validation.
- 0.3 (2026-08-10): Previous architecture baseline.
