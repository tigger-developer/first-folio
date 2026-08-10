<!-- Version: 0.3 | Last updated: 2026-08-10 -->

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
5. Selected local `script.yaml`.
6. Selected local style-specific YAML.
7. CLI overrides.

Local discovery begins at the source directory and walks upwards towards HOME. Only the nearest `script.yaml` is selected; its style-specific sibling is loaded from the same directory. For a multi-file manuscript, the first resolved input determines that starting directory.

The loader provides dotted and inherited access for scripts and letters, and typed decoding for manuscript layout. Root `folio.*` values remain shared defaults; child mode values override only their own elements.

## Rendering

Substantial Typst layouts live in real `.typ` files. Go code owns:

- typed template data;
- context-specific escaping and validated layout literals;
- event-to-layout policy;
- temporary-file lifecycle;
- direct subprocess invocation and diagnostics.

Templates own page composition. Product Go code does not contain generated-language heredocs or invoke a shell to run Typst.

## Build And Tests

`make build` compiles `dist/folio`; `make install` builds first and then links that binary into the configured installation directory. Version values are injected with Go linker flags. Homebrew builds the same command directly.

Automated coverage is Go-owned. Unit tests cover parsing, emission, configuration, escaping, and rendering. Integration tests execute both the in-process public application and a built binary outside the checkout working directory. PDF-sensitive suites invoke Typst directly and inspect supported outputs rather than source implementation text.

## Migration History

Before issue #10, conversion and letters used a Perl dispatcher, Perl parsers/emitters, embedded YAML::Tiny, and shell regression suites; manuscripts used a separately built Go helper. Issue #10 replaced that split with the single Go runtime while preserving the public CLI and accepted rendering behaviour.
