---
title: First Folio Markdown Script Schema
version: "0.3"
updated: "2026-09-10"
---

# Markdown Script Schema

First Folio supports the **limited Markdown script syntax defined here** for
stage plays and screenplays. The rendering style controls page layout; both
use the same source grammar. Markdown constructs outside this grammar have no
promised interpretation or conversion fidelity.

**Reference example:** [One Day](../examples/one-day.md).

## Frontmatter and title

The schema declaration is YAML frontmatter. Title, subtitle, author, version,
and date retain First Folio's existing visible forms below it:

```markdown
---
schema: https://github.com/tigger-developer/first-folio/blob/master/schema/script.md
---

# A Short Play

**A Play in One Act**

*by A. Playwright*

--- Draft 1 | 2026-09-10 ---
```

- The first level-one heading supplies the title.
- A bold line immediately below it supplies the optional subtitle.
- An italic `by` line supplies the author.
- The version/date line uses the exact `--- version | date ---` form.
- YAML `schema` identifies the format. Other YAML fields do not replace these
  visible metadata forms.

Every conversion to `.md` or `.markdown` adds this schema declaration, including
when the input has none. A source schema value is replaced with the destination
schema. A declaration is consumed as metadata when read back, never as dialogue.
Malformed YAML or a missing closing frontmatter delimiter produces an error.
Schema URLs are not fetched and do not select a different document type.

## Dramatic structure

| Source form | Meaning |
|---|---|
| `## Act I` | Act or other major play division |
| `### Scene 1` | Scene within an act |
| `*A kitchen. Morning.*` on its own line | Stage direction or screenplay action |
| `**ALEX:**` | Speaker cue |
| `**ALEX:** *(softly)*` | Speaker cue with a direction |
| Plain lines following a cue | Dialogue |
| `> BLACKOUT` | Transition |
| `***"KEEP OUT"***` | Text displayed within the play |

Speaker names support Unicode. The colon belongs inside the bold cue. Separate
cues, directions, and scene or act headings with blank lines. Dialogue may
span several lines.

A `>` line means a **dramatic transition**, not a general prose quotation.
The supported standalone direction form starts and ends with one asterisk.
These conventions do not establish support for arbitrary nested Markdown.

## Cast and introductory material

A cast table precedes the first speaker cue and has name and description
columns:

```markdown
| Character | Description |
|-----------|-------------|
| ALEX      | A visitor   |
| CÁIT      | The host    |
```

Major sections before the first act containing a speaker cue are introductory
material, for example Synopsis or Setting. They use `##` headings and plain
text. Their rendering is controlled by `render.frontmatter`.

## Footnotes

A reference uses `[^note]`; its definition starts a line with `[^note]:`.
Definitions are emitted after the play body.

```markdown
**ALEX:**
That was the ninth bell.[^bell]

[^bell]: The clock normally strikes eight times.
```

## Conversion limits

Org and Markdown have different heading and cue forms. The converter maps
the dramatic structure between them. It does not translate every inline
emphasis marker between formats, so identical inline styling is not guaranteed.

Fountain conversion has additional losses described in the
[format and fidelity reference](../docs/formats.md). This schema declaration
is emitted only for Markdown output; Org uses its own declaration and
Fountain, Typst, and PDF receive none.

## Related documentation

- [Markdown play format reference](../docs/format-markdown.md).
- [Rendering configuration](../docs/config.md).

## Revision history

- 0.3, 2026-09-10: Define only First Folio's limited script grammar; use YAML
  document-version metadata. General Markdown references and the misplaced
  Org-section discussion have been removed.
- 0.2, 2026-09-10: Record schema emission and output boundaries.
- 0.1, 2026-09-10: Initial definition, superseded by this rewrite.
