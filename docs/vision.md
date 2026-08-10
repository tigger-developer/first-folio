<!-- Version: 0.2 | Last updated: 2026-08-10 -->

# First Folio - Vision

## Purpose

First Folio is a command-line publishing tool for stage plays, submission letters, and prose manuscripts. It converts structured play sources without discarding their dramatic semantics and renders scripts, letters, and manuscripts as production-quality Typst or PDF documents.

## Name

The name references the 1623 First Folio of Shakespeare's plays - the first collected edition that preserved works which might otherwise have been lost. This tool serves a similar role: taking a play in one format and faithfully rendering it in another.

## Problem

Writers work across multiple tools and workflows. A play may be drafted in Emacs Org mode, submitted in Fountain, typeset as PDF for rehearsal, or published as Markdown. A prose manuscript needs repeatable title pages, contents, running matter, chapter layout, and regional submission conventions without repeated word-processor repair. Generic converters do not preserve dramatic semantics, while word processors make deterministic project-wide typography and rendering unnecessarily fragile.

## Goals

1. **Format-agnostic internal representation.** The parser emits a stream of typed semantic events (act, scene, stage direction, character, dialogue, etc.). Output backends consume these events and produce format-specific output. Adding a new format means writing a new parser or a new emitter - not modifying existing code.

2. **Lossless round-tripping where possible.** Converting from format A to format B and back should preserve the semantic content of the play. Formatting details (whitespace, indentation) may change, but no acts, scenes, directions, characters, or dialogue lines should be lost.

3. **Faithful formatting.** Each output format follows its own conventions. Play PDFs support British stage-play, US stage-play, and US screenplay presets. Manuscript PDFs support British and US manuscript presets. Fountain follows the supported subset documented by First Folio rather than pretending unsupported Fountain features round-trip.

4. **CLI-first, scriptable.** All operations are available as command-line tools that read from files or stdin and write to files or stdout. Batch processing, piping, and scripting are first-class use cases.

5. **Minimal dependencies.** The installed application is one Go binary. PDF output requires Typst. Rich manuscript Markdown/org parsing and conversion may depend on Pandoc where using a standard document AST avoids custom parser complexity.

6. **Project configuration.** A project may keep First Folio and Yapper settings in one `script.yaml`. Only documented top-level metadata and `render` keys are shared. The `folio:` and `yapper:` namespaces belong exclusively to their respective applications. The nearest project config and its style-specific sibling override global settings and built-in presets by key. See [docs/config.md](config.md).

## Supported Formats

| Path | Read | Write | Notes |
|------|------|-------|-------|
| Org-mode play | Yes | Yes | Structured Org with heading-level dramatic semantics |
| Markdown play | Yes | Yes | Convention-based Markdown play contract |
| Fountain play | Yes | Yes | Documented supported Fountain subset |
| Typst/PDF play | No | Yes | British, US stage-play, or US screenplay layout |
| Org-mode manuscript | Yes | No | Canonicalized through the Markdown manuscript contract |
| Markdown manuscript | Yes | No | Canonical manuscript input contract |
| Typst/PDF manuscript | No | Yes | British or US manuscript layout |
| Org-mode letter sections | Yes | No | Recipient-specific letters rendered to PDF |

PDF and Typst are final-output formats. Manuscript text-format emission and source-document images remain outside the current contract.

## Future: a gentle interface

First Folio is CLI-first, but many playwrights are not comfortable with terminals, markup syntax, or YAML configuration. A lightweight graphical interface could bridge this gap without compromising the power of the underlying tools.

The vision is not a full IDE or editor — org-mode and Emacs already serve that role beautifully. Instead, a simple companion app that:

- Opens an Org, Markdown, or Fountain file and shows a live-rendered preview
- Provides a "Convert to..." menu (PDF, Markdown, Fountain) with one click
- Exposes style selection (British / American / Screenplay) as a dropdown
- Generates cover letters from the embedded `:letter:` section with a recipient picker
- Wraps the CLI — the app calls `folio convert` and `folio letter` under the hood

The interface should feel approachable to a writer who has never used a terminal. The underlying format remains plain text — the app is a window onto it, not a replacement for it.

## Non-goals

- **Word processor formats.** DOCX, ODT, and similar formats are out of scope. Use pandoc to convert Markdown output if needed.
- **Full screenplay tooling.** First Folio supports screenplay formatting via `--style=screenplay` but is optimised for stage plays. Dedicated screenplay software (Final Draft, Highland) serves that market.
- **Content editing.** First Folio converts between formats and renders to PDF. It does not provide editing, linting, or structural validation of play content.
