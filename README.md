---
title: First Folio
version: "1.12"
last-updated: 2026-10-19
---

# First Folio

A command-line publishing tool for stage plays, submission letters, and prose manuscripts. It converts structured play sources between Org, Markdown, and Fountain, and renders scripts, letters, and manuscripts through Typst.

## Quickstart

```bash
# Convert an org-mode play to Markdown
folio convert play.org play.md

# Convert to PDF (requires Typst)
folio convert play.org play.pdf

# Export an editable script (no external writer required)
folio convert play.org play.docx

# Export ODT with configured typography (requires Pandoc)
folio convert play.org play.odt

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

# Editable manuscripts and letters
folio manuscript manuscript.md manuscript.docx
folio manuscript manuscript.org manuscript.odt
folio letter letters.org --format docx
folio letter letters.org --format odt
```

## Installation

```bash
make install    # symlinks folio to ~/.local/bin/
make uninstall  # removes the symlink
```

Homebrew installation uses the maintained tap:

```bash
brew install tigger-developer/tap/first-folio
```

### Publishing a release

From a committed checkout, `make release` runs the regression suite, creates the
next patch tag, pushes the branch and that tag, then updates and pushes the
Homebrew formula with the archive URL and SHA-256 checksum. Go, Typst and Pandoc
and the contributor prerequisites below must be installed for the release checks. The tap checkout must be clean, current
and configured with an authenticated, writable upstream. SSH remotes can use
the same GitHub access as the source repository; an HTTPS clone needs separate
HTTPS credentials.

`HOMEBREW_TAP_DIR` selects the tap checkout; its default is
`~/code/tigoss/homebrew-tap`. `RELEASE_VERSION` can select an explicit tag:

```bash
make release RELEASE_VERSION=v0.4.16 HOMEBREW_TAP_DIR=/path/to/homebrew-tap
```

Only `SKIP_TESTS=1` skips the release test prerequisite and prints a notice;
unset or `0` runs it. The updated local `homebrew/Formula/first-folio.rb` remains
available to review and commit after publication. If publication fails after
tagging, preserve that tag. If the tap commit already exists, retry its push
from the tap checkout. Otherwise retry the formula step:

```bash
go run ./cmd/update-homebrew --publish-tap /path/to/homebrew-tap 0.4.16
```

### Dependencies

- **Go 1.26+** (build only; Homebrew installs a compiled binary)
- **Typst** (required only for PDF output)
- **Pandoc** (required for ODT writing and the existing manuscript PDF/Typst route; every DOCX route is native)

Iosevka Publish, a customized Iosevka family, is bundled with the executable
under the [SIL Open Font Licence](fonts/iosevka-publish/LICENSE.md). US script
PDFs use it by default, with regular-weight speaker cues starting at a fixed
left indent. Manuscript inline and block code also default to this family. No separate font installation is needed for
PDF output. Font choices and speaker alignment remain YAML settings; explicit
role overrides take precedence.

DOCX and ODT exports [embed complete available font faces](docs/font-embedding.md),
including bundled Iosevka Publish, where editable embedding is permitted.
Missing or restricted fonts produce warnings and retain their font-name references.
Standalone `.typ` files still require fonts on the compiling host.

## Supported Formats

| Format | Read | Write | Schema |
|--------|------|-------|--------|
| Org-mode | Yes | Yes | [docs/format-org.md](docs/format-org.md) |
| Markdown | Yes | Yes | [docs/format-markdown.md](docs/format-markdown.md) |
| Fountain | Yes | Yes | [docs/format-fountain.md](docs/format-fountain.md) |
| PDF (via Typst) | - | Yes | - |
| DOCX script, manuscript or letter | - | Yes | [Support and limitations](docs/format-docx.md) |
| ODT script, manuscript or letter | - | Yes | [Support and limitations](docs/format-odt.md) |
| Manuscript Markdown | Yes | - | [docs/format-manuscript-markdown.md](docs/format-manuscript-markdown.md) |
| Manuscript org-mode | Yes | - | [docs/format-manuscript-org.md](docs/format-manuscript-org.md) |

Org-mode uses heading levels to encode play structure. Markdown uses headers, bold, and italic conventions. Fountain follows the [Fountain spec](https://fountain.io/syntax). See the schema docs for full element mappings, and [docs/formats.md](docs/formats.md) for the event stream and fidelity matrix.

**Intro sections** (Synopsis, Setting, Scene List, etc.) are automatically distinguished from the play proper. Any headers and prose before the first character dialogue are treated as intro material and can be toggled on/off via `render.frontmatter` in config.

## Canonical Examples

DOCX and ODT export editable script tables, paragraphs and footnotes with
documented layout compromises. DOCX uses its native writer; ODT uses Pandoc
with generated styles from the resolved configuration. See
[DOCX export](docs/format-docx.md) and [ODT export](docs/format-odt.md).

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
  font:
    family: EB Garamond
    size: 11pt
  page: a4
```

Configuration is merged by key. Every typography role uses the same six-property `font` block; partial blocks inherit from the same role in lower layers. Script introduction paragraphs also inherit omitted properties from `folio.font`; override them independently with `folio.positioning.frontmatter.paragraph.font`. In descending precedence: CLI overrides, the nearest local style-specific file, the nearest local `script.yaml`, the first Markdown manuscript input's `folio:`/`render:` frontmatter, the global style-specific file, global `script.yaml`, the selected built-in style overlay, and the shared British built-in base. Local discovery starts at the source directory and walks towards HOME; the nearest `script.yaml` wins. Documented top-level metadata and `render` keys may be shared with Yapper. The `folio:` block belongs exclusively to First Folio; a top-level `yapper:` block belongs exclusively to Yapper and is ignored by First Folio.

See [docs/config.md](docs/config.md) for the configuration reference and [examples/script.yaml.example](examples/script.yaml.example) for an annotated example.

Running headers and footers for scripts and manuscripts use `folio.page-header` and `folio.page-footer`, including configurable typography, format strings, edge distance and body clearance. Letters do not inherit running matter. Former manuscript-local header/footer blocks remain supported with deprecation warnings in manuscript mode; see [shared running matter](docs/config.md#shared-running-matter-foliopage-header-and-foliopage-footer).

For page-limited script submissions, set `folio.title-page.enabled: false` to place the title-page items above the script on its first content page instead of using a separate title page. `folio.title-page.skip-header` and `skip-footer` default to `true`, hiding running matter on that first page while retaining it on later pages. Set either to `false` to allow its enabled shared role on page 1. Manuscript title-page settings remain independent. See [compact script titles and first-page running matter](docs/config.md#compact-script-title-block).

## Project Structure

| Path | Purpose |
|------|---------|
| `cmd/folio/` | Single Go CLI entry point |
| `internal/app/` | Public command dispatch and integration |
| `internal/config/` | Shared layered YAML configuration |
| `internal/play/` | Stage-play model, parsers, and text emitters |
| `internal/source/` | Shared native source recognition and structured office prose parsing |
| `internal/letter/` | Cover-letter parser and renderer |
| `internal/manuscript/` | Go manuscript parsing, config, and Typst rendering |
| `internal/office/` | Shared native DOCX and configured ODT writers, mode adapters and font packaging |
| `templates/` | File-backed Typst templates |
| `examples/` | Annotated config example |
| `docs/` | Format schemas, config reference, vision |

## Running Tests

### Contributor prerequisites

The regression suite requires Go, Typst, Pandoc, LibreOffice (`soffice`) and
Poppler's `pdftotext`, `pdftohtml` and `pdftoppm` on `PATH`. `make test` checks
these dependencies and fails with the missing executable's name. Normal
`make release` runs this suite too. LibreOffice and Poppler are test tools;
ordinary DOCX/ODT conversion does not invoke them. The suite allows 20 minutes
per Go package; each office subprocess retains its 120-second deadline.

Office regressions refresh native fields in isolated LibreOffice profiles,
export PDFs and measure font, spacing, numbering and contents evidence. They
do not replace editing and save/reopen validation in the intended application.
The labelled fixtures and YAML checklist live in
[`testdata/prose-office`](testdata/prose-office/EXPECTATIONS.md).

The ticket-scoped full-novel acceptance test runs separately from `make test`:

```bash
go test -tags folio_acceptance ./internal/app -run '^TestAT063_1_' -count=1 -timeout=150m -v
```

It generates equivalent Markdown and Org novels with 10,000 paragraphs and
50 chapters, converts each to DOCX and ODT, refreshes native fields in an isolated
LibreOffice profile and checks the complete rendered content. Set
`FOLIO_W063_ACCEPTANCE_DIR` to an OS-allocated temporary directory to retain the
labelled documents, PDFs, measurements and selected page images for user tests.
Word and LibreOffice editing, save and reopen remain human validation.

```bash
make test   # Go unit, integration, Typst, and PDF regression suite
make lint   # Go static analysis
```

## Documentation

- [Source schemas](schema/README.md) - script, screenplay, manuscript, and letter definitions with source examples
- [Vision](docs/vision.md) - project goals, supported formats, and direction of travel
- [Architecture](docs/ARCHITECTURE.md) - Go runtime, document models, configuration, and rendering boundaries
- [Configuration](docs/config.md) - config schema, precedence, shared keys, migration
- [Formats](docs/formats.md) - format overview, event stream, and fidelity matrix
  - [Org-mode](docs/format-org.md) - org-mode play format schema
  - [Markdown](docs/format-markdown.md) - Markdown play format schema
  - [Fountain](docs/format-fountain.md) - Fountain format schema and fidelity analysis
  - [Markdown manuscript](docs/format-manuscript-markdown.md) - prose manuscript Markdown contract
  - [Org manuscript](docs/format-manuscript-org.md) - prose manuscript org-mode contract

## Documentation History

- 1.12 (2026-10-19): Document native office source parsing and separate novel-scale acceptance.

- 2026-10-19: Extend the shared native DOCX and configured ODT writers to manuscript and letter adapters.

- 1.10 (2026-10-18): State the rendered suite's aggregate package deadline.

- 1.9 (2026-10-16): Add manuscript/letter office commands and rendered regression prerequisites.

- 1.8 (2026-10-12): Documented independent script frontmatter paragraph fonts and body-font fallback.

- 1.7 (2026-10-12): Document font embedding in editable DOCX and ODT exports.

- 1.6 (2026-10-12): Document bundled Iosevka Publish fonts and US script defaults.

- 1.5 (2026-10-11): Documented gated Homebrew publication, configurable tap checkout and partial-release recovery.

- 1.2 (2026-10-07): Documented compact script first-page title blocks and running-matter suppression for page-limited submissions.

- 1.1 (2026-10-07): Documented shared script/manuscript running matter and manuscript-local migration.

## Licence

MIT - Copyright Tadhg O'Brien

## Acknowledgements

- Historical dependency: [YAML::Tiny](https://metacpan.org/pod/YAML::Tiny) v1.76 was embedded before the Go migration and was removed with the Perl runtime in issue #10.
