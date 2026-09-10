---
title: First Folio Source Schemas
version: "0.3"
updated: "2026-09-10"
---

# Source Schemas

These documents define First Folio's supported source syntax.

| Document type | Org | Markdown |
|---|---|---|
| Stage play or screenplay | [Script schema](script.org) | [Script schema](script.md) |
| Prose manuscript | [Manuscript schema](manuscript.org) | [Manuscript schema](manuscript.md) |

The Markdown schemas describe First Folio's limited syntax, not general
Markdown conformance. Stage plays and screenplays share a source structure;
the selected rendering style determines their layout.

[Letters](letter.org) are **Org-only**. An Org script can contain both play and
letter sections. The script declares the Org script schema, which references
the separate Org letter schema.

Org uses `#+SCHEMA: <full URL>`. Markdown uses YAML `schema: <full URL>`.
The URL points to the corresponding file under
`https://github.com/tigger-developer/first-folio/blob/master/schema/`.
Each schema links to a matching source example under `../examples/`.

Conversion adds a schema declaration only to Org and Markdown output. The
existing manuscript command writes Typst/PDF, and the letter command writes
PDF; a schema definition does not add an output mode.
