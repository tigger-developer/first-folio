# Source Document Schemas

First Folio distinguishes scripts, prose manuscripts, and letters. Stage plays
and screenplays share the script structure; their rendering styles differ.

| Document | Org schema | Markdown schema |
|---|---|---|
| Stage play or screenplay | [script.org](script.org) | [script.md](script.md) |
| Prose manuscript | [manuscript.org](manuscript.org) | [manuscript.md](manuscript.md) |
| Letter sections | [letter.org](letter.org) | Not defined |

Org documents declare `#+SCHEMA: <full URL>` in keyword frontmatter. Markdown
documents declare `schema: <full URL>` in YAML frontmatter. The URL points to
the corresponding file under
`https://github.com/tigger-developer/first-folio/blob/master/schema/`.
Each schema links back to an appropriate example through `../examples/`.

An Org script may include both the play and `:letter:` sections. It declares
the script schema, which references the letter schema. Play conversion must
exclude letter sections; `folio letter` generates only recipient letters.

These documents define the requested schema declarations and conversion rules.
Automatic schema emission and play-conversion exclusion still require the
implementation work recorded in [W041 - Document schemas and schema frontmatter](../docs/work.org#w-041)
and [W040 - Exclude cover-letter sections from document rendering](../docs/work.org#w-040).
