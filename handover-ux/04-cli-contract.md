# 04 · CLI contract

Everything the app spawns as `folio ...`. Verified against `folio 0.4.10` on 2026-07-20 by Claude.

**Before implementation, re-verify by running `folio --version` and each `folio <cmd> --help` on the developer machine.** The user's `folio` may be newer.

Authoritative help sources (both matched live output at time of writing):

- `/Users/tigger/code/tigoss/first-folio/docs/folio-help.md`
- `/Users/tigger/code/tigoss/first-folio/docs/folio-convert-help.md`
- `/Users/tigger/code/tigoss/first-folio/docs/folio-letter-help.md`
- `/Users/tigger/code/tigoss/first-folio/docs/folio-manuscript-help.md`

## Top-level

```
folio <command> [options]
```

- `folio --version` → `folio 0.4.10\n` (single line).
- `folio -h` / `folio --help` → usage summary listing `convert`, `letter`, `manuscript`.
- No global `--verbose` / `--quiet` / `--json` flag today. Do not assume any.

## `folio convert`

```
folio convert <source> [target] [options]
```

Sources: `.org`, `.md`, `.fountain`, `.ftn`. Targets: `.org`, `.md`, `.fountain`, `.ftn`, `.pdf`.

Flags the app maps 1:1 from the Convert screen:

| UI control | Flag | Values / notes |
|---|---|---|
| Style (segmented) | `--style` | `british` (default), `us`, `screenplay` |
| Stdout format (when Preview-as-text) | `--to` | `org`, `md`, `fountain` (not `pdf` in Preview-as-text mode) |
| Font | `--font FONT` | free-form string |
| Font size | `--font-size SIZE` | Typst length, e.g. `12pt` |
| Margin | `--margin SIZE` | Typst length |
| Page size | `--page SIZE` | `a4`, `letter`, others |
| Dialogue indent | `--indent SIZE` | Typst length |
| Dialogue spacing | `--dialogue-spacing SIZE` | Typst length |
| Direction spacing | `--direction-spacing SIZE` | Typst length |
| Italic directions | `--direction-italic` / `--no-direction-italic` | mutually exclusive |
| Centre directions | `--direction-centre` / `--no-direction-centre` | mutually exclusive |
| (never expose) | `--force` | forces binary output to terminal — irrelevant to the GUI |

PDF options are silently ignored for non-PDF targets. The app enforces this by only enabling the PDF disclosure group when the target extension is `.pdf`.

**Output invocation shapes:**

- Save-as-file: `folio convert /abs/source.org /abs/target.pdf [--style ...] [pdf flags]`
- Preview-as-text: `folio convert /abs/source.org --to md [--style ...]` — output goes to stdout, no target file.

## `folio letter`

```
folio letter <source.org> [options]
```

Only `.org` sources. Never `.md` or `.fountain`.

| UI control | Flag | Values / notes |
|---|---|---|
| Selected recipients | `--to RECIPIENT` | **regex substring** per help text. If the user selects a subset, pass an anchored alternation like `^(Alice|Bob|Carol)$`. Escape each name — recipient names may contain regex metacharacters. If all recipients ticked, omit `--to` entirely. |
| Output directory | `--dir DIR` | absolute path preferred |
| Filename prefix | `--prefix PREFIX` | plain string, default `letter` |

**Recipient discovery:** the app needs a list of recipients from an Org file to populate the checkbox table.

- If decision R4 in `02-decisions.md` is confirmed and shipped: use `folio letter <source.org> --list`. **This flag does not exist yet.** Do not stub the app around a flag that has not been added to the CLI.
- If R4 is rejected: parse `**** Name :tag:` H4 headings under `:letter:` sections in Swift via `OrgLetterScanner`. Deliberately shallow; see `03-architecture.md`.

## `folio manuscript`

```
folio manuscript <input>... <target> [options]
```

Inputs: one or more `.md` files, or one or more `.org` files. **Do not mix formats in one invocation** — the CLI exits non-zero with a diagnostic. Fountain is not a valid manuscript input.

The last positional argument is the target (`.typ` or `.pdf`); every positional before it is an input.

| UI control | Flag |
|---|---|
| Style (segmented, British/US only) | `--style british\|us` |
| Metadata: Title | `--title TITLE` |
| Metadata: Subtitle | `--subtitle SUBTITLE` |
| Metadata: Author | `--author AUTHOR` |
| Metadata: Attribution (e.g. "by") | `--attribution TEXT` (alias: `--author-attribution`) |
| Metadata: Date | `--date DATE` |
| Metadata: Version | `--version VERSION` — see note |
| Metadata: Word count | `--wordcount WORDS` |
| Metadata: Contact name | `--contact-name NAME` |
| Dry-run checkbox | `--dry-run` |

**Note on `--version`:** `folio manuscript --help` documents this as `--version [VERSION]` — bare `--version` prints the tool version. To override the manuscript version, pass a value (`--version "Draft v3"`). The app should always pass a value or omit the flag entirely; never send bare `--version`.

**File globbing:** the CLI accepts globs on the command line when the shell expands them. The app does not spawn a shell — `Process` executes `folio` directly. If the user pastes a glob into the "Add glob..." sheet, the app must expand it in Swift (using `FileManager` enumeration or `Glob`) and append the resolved matches to the input list. Do not pass unexpanded globs.

## Output patterns

Verified with `folio 0.4.10` on 2026-07-20:

### `folio convert` — silent on success

No stdout, no "wrote X" line. Non-zero exit + stderr message on failure. The app knows the target path from its own argv; use that to build the Reveal-in-Finder and Open-in-Preview affordances after a zero exit.

### `folio letter` — one line per PDF, plus a summary

Example live output:

```
Generated: /tmp/folio-probe/probe-the-abbey.pdf
Generated: /tmp/folio-probe/probe-the-mill.pdf
Generated: /tmp/folio-probe/probe-dodgy-derek-secret-alley-greystones.pdf
Generated: /tmp/folio-probe/probe-milly-o-murphy-1234-sixties-towerblock-london-se1.pdf
4 cover letter(s) generated.
```

`OutputParser` regex: `^Generated: (.+)$` per PDF, and `^(\d+) cover letter\(s\) generated\.$` for the summary. Emit a Reveal-in-Finder row for each captured path.

### `folio manuscript --dry-run` — YAML-ish plan

Example live output:

```
format: org
output: /tmp/folio-probe/ms.typ
style: british
page: a4
margin: 20mm
inputs:
  - /Users/tigger/code/tigoss/first-folio/examples/dummy-manuscript.org
```

Show this raw in the output log. Do not try to render it as a form — it is a human-readable summary, not a stable machine format.

### `folio manuscript` (real run) — silent on success

Same principle as `convert`: use the app's known target path.

## Errors

On non-zero exit the CLI writes a diagnostic to stderr and does not create a misleading success artefact (see `docs/ACs.org` AC9.11, AC9.5). The app surfaces this as an alert (last stderr line as title, full stderr in a disclosure) and offers no auto-open.

Do not swallow stderr. Do not attempt to re-word CLI diagnostics — relay them verbatim.

## Config the app reads (not writes)

Path: `~/.config/first-folio/script.yaml`.

Schema authoritative source: `/Users/tigger/code/tigoss/first-folio/docs/config.md`. For MVP the app only cares about a handful of keys for prefill:

- `folio.style` — default for Convert's style segmented control (`british` / `us` / `screenplay`).
- `folio.font`, `folio.font-size`, `folio.margin`, `folio.page` — defaults for the Convert PDF disclosure.
- `folio.manuscript.style` (if present) — default for Manuscript's style segmented control.

Missing keys are normal. Return an empty struct, let the CLI's built-in defaults win.

**Do not** parse `yapper:` blocks — that namespace belongs to another tool and the app must leave it untouched.

## Absolute paths

Always pass absolute paths to `folio`. `Process` does not inherit a useful cwd for the user's document, and relative paths would resolve against the app's launch directory. Resolve every user-picked file to its absolute path via `URL.standardizedFileURL.path` before appending to argv.
