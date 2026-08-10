# First Folio

A command-line publishing tool for stage plays, submission letters, and prose manuscripts. It converts structured play sources between Org, Markdown, and Fountain, and renders scripts, letters, and manuscripts through Typst.

## Quickstart

```bash
# Convert an org-mode play to Markdown
folio convert play.org play.md

# Convert to PDF (requires Typst)
folio convert play.org play.pdf

# Convert to Fountain
folio convert play.org play.fountain

# Output to stdout
folio convert play.org --to md

# Convert between any supported formats
folio convert play.fountain play.md
folio convert play.md play.org

# Generate letters from :letter: tagged sections
folio letter play.org              # all recipients
folio letter play.org --to Abbey   # specific recipient

# Render one or more prose manuscript chapters to PDF
folio manuscript 'chapters/ch??.md' submission.pdf
folio manuscript examples/dummy-manuscript.org manuscript.typ
```

## Installation

```bash
make install    # symlinks folio to ~/.local/bin/
make uninstall  # removes the symlink
```

### Dependencies

- **Go 1.26+** (build only; Homebrew installs a compiled binary)
- **Typst** (required only for PDF output)
- **Pandoc** (required for rich manuscript Markdown/org parsing and conversion)

## Supported Formats

| Format | Read | Write | Schema |
|--------|------|-------|--------|
| Org-mode | Yes | Yes | [docs/format-org.md](docs/format-org.md) |
| Markdown | Yes | Yes | [docs/format-markdown.md](docs/format-markdown.md) |
| Fountain | Yes | Yes | [docs/format-fountain.md](docs/format-fountain.md) |
| PDF (via Typst) | - | Yes | - |
| Manuscript Markdown | Yes | - | [docs/format-manuscript-markdown.md](docs/format-manuscript-markdown.md) |
| Manuscript org-mode | Yes | - | [docs/format-manuscript-org.md](docs/format-manuscript-org.md) |

Org-mode uses heading levels to encode play structure. Markdown uses headers, bold, and italic conventions. Fountain follows the [Fountain spec](https://fountain.io/syntax). See the schema docs for full element mappings, and [docs/formats.md](docs/formats.md) for the event stream and fidelity matrix.

**Intro sections** (Synopsis, Setting, Scene List, etc.) are automatically distinguished from the play proper. Any headers and prose before the first character dialogue are treated as intro material and can be toggled on/off via `render.frontmatter` in config.

## Canonical Examples

Each format has a complete reference example demonstrating all supported features:

- [about-time.org](about-time.org) --- org-mode (master authoring format, includes cover letter structure)
- [one-day.md](one-day.md) --- Markdown
- [one-day.fountain](one-day.fountain) --- Fountain

## Configuration

First Folio reads configuration from `script.yaml` files. It never creates or modifies config files.

```yaml
# ~/.config/first-folio/script.yaml or the nearest project ancestor

date: "2026-04-26"
version: "Draft v3"

folio:
  font: EB Garamond
  font-size: 11pt
  page: a4
```

Configuration is merged by key. In descending precedence: CLI overrides, the nearest local style-specific file, the nearest local `script.yaml`, the global style-specific file, global `script.yaml`, the selected built-in style override, and the British built-in base. Local discovery starts at the source directory and walks towards HOME; the nearest `script.yaml` wins. Documented top-level metadata and `render` keys may be shared with Yapper. The `folio:` block belongs exclusively to First Folio; a top-level `yapper:` block belongs exclusively to Yapper and is ignored by First Folio.

See [docs/config.md](docs/config.md) for the configuration reference and [examples/script.yaml.example](examples/script.yaml.example) for an annotated example.

## Project Structure

| Path | Purpose |
|------|---------|
| `cmd/folio/` | Single Go CLI entry point |
| `internal/app/` | Public command dispatch and integration |
| `internal/config/` | Shared layered YAML configuration |
| `internal/play/` | Stage-play model, parsers, and text emitters |
| `internal/letter/` | Cover-letter parser and renderer |
| `internal/manuscript/` | Go manuscript parsing, config, and Typst rendering |
| `templates/` | File-backed Typst templates |
| `examples/` | Annotated config example |
| `docs/` | Format schemas, config reference, vision |

## Running Tests

```bash
make test   # Go unit, integration, Typst, and PDF regression suite
make lint   # Go static analysis
```

## Documentation

- [Vision](docs/vision.md) - project goals, supported formats, and direction of travel
- [Architecture](ARCHITECTURE.md) - Go runtime, document models, configuration, and rendering boundaries
- [Configuration](docs/config.md) - config schema, precedence, shared keys, migration
- [Formats](docs/formats.md) - format overview, event stream, and fidelity matrix
  - [Org-mode](docs/format-org.md) - org-mode play format schema
  - [Markdown](docs/format-markdown.md) - Markdown play format schema
  - [Fountain](docs/format-fountain.md) - Fountain format schema and fidelity analysis
  - [Markdown manuscript](docs/format-manuscript-markdown.md) - prose manuscript Markdown contract
  - [Org manuscript](docs/format-manuscript-org.md) - prose manuscript org-mode contract

## Licence

MIT - Copyright Tadhg O'Brien

## Acknowledgements

- Historical dependency: [YAML::Tiny](https://metacpan.org/pod/YAML::Tiny) v1.76 was embedded before the Go migration and was removed with the Perl runtime in issue #10.
