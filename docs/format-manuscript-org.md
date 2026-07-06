<!-- Version: 0.1 | Last updated: 2026-07-06 -->

# Org-mode Manuscript Format

Org-mode manuscript input uses org front matter and headings for prose manuscript structure. It is separate from the org-mode stage-play contract.

## Element Schema

| Org syntax | Manuscript meaning |
|---|---|
| `#+TITLE: About Time` | Manuscript title |
| `#+SUBTITLE: A Novel` | Subtitle |
| `#+AUTHOR: Tadhg Paul` | Author name |
| `#+DATE: July 2026` | Manuscript date |
| `#+VERSION: Draft 4` | Draft/version marker |
| `#+WORDCOUNT: 80000` | Approximate word count |
| `#+ADDRESS: ...` | Postal address |
| `#+EMAIL: ...` | Email address |
| `#+WEBSITE: ...` | Website |
| `* PART ONE` | Part divider page |
| `** Chapter 1` | Chapter start page |
| `*** Section` and deeper | Local section heading |
| Plain paragraphs | Body prose |
| `-----` on its own line | Scene break |
| `~code~`, `=verbatim=`, and source blocks | Monospace text |
| `[fn:name]` and `[fn:name] Text` | Footnote reference and definition |
| Heading tagged `:noexport:` | Private section excluded with children |

Fountain is not accepted by manuscript mode.

## Example

```org
#+TITLE: About Time
#+SUBTITLE: A Novel
#+AUTHOR: Tadhg Paul
#+DATE: July 2026
#+VERSION: Draft 4
#+WORDCOUNT: 80000

* PART ONE
** Chapter 1
The rain had been falling since Tuesday.

-----

By noon, the hands had moved backwards twice.

*** Notes :noexport:
This planning note is excluded.
```
