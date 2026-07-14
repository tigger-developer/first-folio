<!-- Version: 0.1 | Last updated: 2026-04-26 -->

# First Folio - Vision

## Purpose

First Folio is a format converter for stage plays. It reads plays written in a structured source format and produces output in any supported target format, preserving the semantic structure of the work: acts, scenes, stage directions, character names, dialogue, character tables, and front matter.

## Name

The name references the 1623 First Folio of Shakespeare's plays - the first collected edition that preserved works which might otherwise have been lost. This tool serves a similar role: taking a play in one format and faithfully rendering it in another.

## Problem

Playwrights and dramaturgs work across multiple tools and workflows. A play may be drafted in Emacs org-mode, submitted in Fountain format, typeset as PDF for rehearsal, or published as Markdown. No single tool handles all of these conversions while preserving the semantic structure of the text. Existing converters (pandoc, etc.) treat plays as generic documents, losing the distinction between stage directions, character names, dialogue, and other dramatic elements.

## Goals

1. **Format-agnostic internal representation.** The parser emits a stream of typed semantic events (act, scene, stage direction, character, dialogue, etc.). Output backends consume these events and produce format-specific output. Adding a new format means writing a new parser or a new emitter - not modifying existing code.

2. **Lossless round-tripping where possible.** Converting from format A to format B and back should preserve the semantic content of the play. Formatting details (whitespace, indentation) may change, but no acts, scenes, directions, characters, or dialogue lines should be lost.

3. **Faithful formatting.** Each output format follows its own conventions. Markdown uses headers and bold/italic. PDF (via Typst) uses British stage play layout with proper indentation. Fountain follows the Fountain markup specification. The tool does not impose one format's conventions on another.

4. **CLI-first, scriptable.** All operations are available as command-line tools that read from files or stdin and write to files or stdout. Batch processing, piping, and scripting are first-class use cases.

5. **Minimal dependencies.** The installed application is one Go binary. PDF output requires Typst. Rich manuscript Markdown/org parsing and conversion may depend on Pandoc where using a standard document AST avoids custom parser complexity.

6. **Project configuration.** A project may keep First Folio and Yapper settings in one `script.yaml`. Only documented top-level metadata and `render` keys are shared. The `folio:` and `yapper:` namespaces belong exclusively to their respective applications. Per-project config files override global defaults. See [docs/config.md](config.md).

## Supported Formats

| Format | Read | Write | Notes |
|--------|------|-------|-------|
| Org-mode play | Yes | No | Structured org with heading-level semantics |
| Markdown | No | Yes | Clean idiomatic Markdown |
| PDF | No | Yes | Via Typst, British stage play layout |
| Fountain | Planned | Planned | Industry-standard screenplay/stage play format |

The direction of travel is toward full read/write support for all text-based formats (org, Markdown, Fountain). PDF remains write-only as it is a final-output format.

## Future: a gentle interface

First Folio is CLI-first, but many playwrights are not comfortable with terminals, markup syntax, or YAML configuration. A lightweight graphical interface could bridge this gap without compromising the power of the underlying tools.

The vision is not a full IDE or editor — org-mode and Emacs already serve that role beautifully. Instead, a simple companion app that:

- Opens an org or Fountain file and shows a live-rendered preview
- Provides a "Convert to..." menu (PDF, Markdown, Fountain) with one click
- Exposes style selection (British / American / Screenplay) as a dropdown
- Generates cover letters from the embedded `:letter:` section with a recipient picker
- Wraps the CLI — the app calls `folio convert` and `folio letter` under the hood

The interface should feel approachable to a writer who has never used a terminal. The underlying format remains plain text — the app is a window onto it, not a replacement for it.

## Non-goals

- **Word processor formats.** DOCX, ODT, and similar formats are out of scope. Use pandoc to convert Markdown output if needed.
- **Full screenplay tooling.** First Folio supports screenplay formatting via `--style=screenplay` but is optimised for stage plays. Dedicated screenplay software (Final Draft, Highland) serves that market.
- **Content editing.** First Folio converts between formats and renders to PDF. It does not provide editing, linting, or structural validation of play content.
