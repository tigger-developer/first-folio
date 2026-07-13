<!-- Version: 0.1 | Last updated: 2026-07-06 -->

# Architecture

First Folio is a CLI-first document renderer with two implementation paths.

## Current Runtime

`bin/folio` is the public command. Existing stage-play conversion and cover-letter generation remain implemented in Perl:

- `folio convert` uses `lib/Folio/Parser/*`, `lib/Folio/Emitter/*`, `Folio::Config`, and Typst for PDF output.
- `folio letter` uses `Folio::CoverLetter` and the existing `script.yaml` configuration model.

Prose manuscript rendering is implemented in Go:

- `folio manuscript` is dispatched by `bin/folio`.
- The dispatcher executes `bin/folio-manuscript` when built; the Go helper resolves the project root from its executable realpath.
- During development, the dispatcher falls back to `go run ./cmd/folio-manuscript`.
- `cmd/folio-manuscript` delegates all behaviour to `internal/manuscript`.

The Perl and Go paths share user-facing configuration and presets, but manuscript parsing and rendering do not reuse the stage-play event stream. Prose manuscripts have their own document model: metadata, parts, chapters, sections, paragraphs, scene breaks, code, and footnotes.

## Go Manuscript Engine

The Go manuscript engine is organized as follows:

| Path | Responsibility |
|---|---|
| `cmd/folio-manuscript/` | CLI entry point for manuscript rendering |
| `internal/manuscript/input.go` | Explicit path and quoted-glob resolution, sorting, deduplication, and format validation |
| `internal/manuscript/config.go` | Layered YAML loading and manuscript inheritance |
| `internal/manuscript/parser.go` | Markdown and org-mode manuscript contracts |
| `internal/manuscript/serialize.go` | Canonical manuscript Markdown serialization |
| `internal/manuscript/render.go` | Typst template execution and Typst-safe markup generation |
| `templates/manuscript.typ` | File-backed Typst layout template |
| `presets/british-manuscript.yaml` | British manuscript base preset |
| `presets/us-overrides-manuscript.yaml` | Shunn-style US manuscript override preset |

The Go path uses `gopkg.in/yaml.v3` for YAML and `text/template` for Typst generation. The Typst template is a real file rather than an embedded heredoc so the layout language remains reviewable.

For issue #9, Markdown is the canonical manuscript render contract. Markdown input is parsed and serialized back to canonical Markdown before Typst rendering. Org-mode input is parsed, serialized to canonical Markdown, reparsed, and then rendered through the same Typst path. This keeps Markdown and org-mode manuscript PDFs identical for the shared v1 manuscript contract. A future architecture issue tracks whether org-mode should become the canonical manuscript representation for richer semantics.

## Configuration Layers

Configuration remains based on `script.yaml`. Manuscript rendering uses the same precedence model as existing script rendering:

1. British manuscript preset.
2. US manuscript override preset, only when `--style us` or equivalent config is selected.
3. Global `~/.config/first-folio/script.yaml`.
4. Global style-specific `script-british.yaml` or `script-us.yaml`.
5. Local `script.yaml`: beside the source for one manuscript input, or in the process working directory for multiple manuscript inputs.
6. Local style-specific `script-british.yaml` or `script-us.yaml`.

The same selected local directory supplies both the base and style-specific files. A multi-input request does not merge configuration from individual input directories.
7. CLI overrides.

Root `folio.*` values remain the shared defaults. `folio.manuscript.*` values are manuscript-specific overrides. Child settings such as `folio.manuscript.toc.font` override only their own element.

## Migration Direction

Go is the preferred target for new complex rendering features. The manuscript engine is the proving ground for a future migration because it provides:

- typed configuration structures;
- table-driven tests;
- file-backed templates;
- clearer subprocess handling;
- a smaller shell surface.

Existing Perl code should not be rewritten opportunistically. Future migration should proceed feature by feature behind the stable `folio` CLI contract, with regression tests proving parity before each public path moves.
