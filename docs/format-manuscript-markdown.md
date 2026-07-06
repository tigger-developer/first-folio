<!-- Version: 0.1 | Last updated: 2026-07-06 -->

# Markdown Manuscript Format

Markdown manuscript input is a prose contract, separate from the Markdown stage-play contract.

## Element Schema

| Markdown syntax | Manuscript meaning |
|---|---|
| `# Title` | Manuscript title |
| `**Subtitle**` after the title | Subtitle |
| `*by Author*` | Author name |
| `--- Draft | Date ---` | Version and date |
| `## PART ONE` | Part divider page |
| `### Chapter 1` | Chapter start page |
| `#### Section` and deeper | Local section heading |
| Plain paragraphs | Body prose |
| `***` on its own line | Scene break |
| Backticks and fenced code blocks | Monospace text |
| `[^name]` and `[^name]: text` | Footnote reference and definition |
| HTML comments | Private notes, excluded |
| Heading ending `<!-- noexport -->` | Private section excluded until the next same-or-higher heading |

Fountain is not accepted by manuscript mode.

## Example

```markdown
# About Time

**A Novel**

*by Tadhg Paul*

--- Draft 4 | July 2026 ---

## PART ONE

### Chapter 1

The rain had been falling since Tuesday.

***

By noon, the hands had moved backwards twice.

### Notes <!-- noexport -->

This planning note is excluded.
```
