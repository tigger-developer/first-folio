<!-- Version: 0.2 | Last updated: 2026-08-10 -->

# Format Overview

First Folio converts stage plays between three text-based formats (org-mode, Markdown, Fountain) and one output-only format (PDF). All conversions pass through a shared **event stream** - a sequence of typed semantic events representing the structure of a play. Each format has an independent parser (reader) and emitter (writer); no direct format-to-format conversion paths exist.

## Supported Formats

| Format | Read | Write | Reference |
|--------|------|-------|-----------|
| [Org-mode play](format-org.md) | Yes | Yes | [orgmode.org](https://orgmode.org) |
| [Markdown play](format-markdown.md) | Yes | Yes | [CommonMark](https://commonmark.org) |
| [Fountain](format-fountain.md) | Yes | Yes | [fountain.io](https://fountain.io) |
| [Markdown manuscript](format-manuscript-markdown.md) | Yes | No | [CommonMark](https://commonmark.org) |
| [Org-mode manuscript](format-manuscript-org.md) | Yes | No | [orgmode.org](https://orgmode.org) |
| PDF (via Typst) | - | Yes | [typst.app](https://typst.app) |

## The Event Stream

The event stream is the intermediate representation. Every parser emits these events; every emitter consumes them. The events correspond to the semantic elements of a stage play:

| Event | Arguments | Meaning |
|-------|-----------|---------|
| `front_matter` | key, value | Metadata (title, author, etc.) |
| `act_header` | title | Start of an act |
| `scene_header` | title | Start of a scene |
| `stage_direction` | text | Narrative/action text between dialogue |
| `character` | name, direction | Character about to speak, with optional parenthetical |
| `dialogue` | line | A line of speech |
| `character_table_start` | - | Begin cast list |
| `character_table_row` | name, description | One entry in the cast list |
| `character_table_end` | - | End cast list |
| `prop_text` | text | On-stage text (signs, placards, letters read aloud) |
| `footnote` | name, text | A footnote definition |
| `transition` | text | A transition such as `BLACKOUT` or `CUT TO:` |
| `intro_header` | title | A header before the play proper |
| `intro_text` | text | Introductory prose before the play proper |

## Fidelity Matrix

Not every format can represent every event natively. The matrix below shows which events survive each format. **Lossless** means the element round-trips without degradation. **Degraded** means the content is preserved but structural metadata is lost. **Lost** means the element cannot be represented and is dropped.

| Event | Org-mode | Markdown | Fountain | PDF |
|-------|----------|----------|----------|-----|
| front_matter (title) | Lossless | Lossless | Lossless | Lossless |
| front_matter (author) | Lossless | Lossless | Lossless | Lossless |
| front_matter (other keys) | Selected keys | Lost | Selected keys | Selected keys rendered where applicable |
| act_header | Lossless | Lossless | Degraded | Lossless |
| scene_header | Lossless | Lossless | Lossless | Lossless |
| stage_direction | Lossless | Lossless | Lossless | Lossless |
| character | Lossless | Lossless | Lossless | Lossless |
| character (direction) | Lossless | Lossless | Lossless | Lossless |
| dialogue | Lossless | Lossless | Lossless | Lossless |
| character_table | Lossless | Lossless | Degraded | Lossless |
| prop_text | Lossless | Lossless | Degraded | Lossless |
| footnote | Lossless | Lossless | Degraded | Lossless |
| transition | Lossless | Lossless | Lossless | Lossless |
| intro_header / intro_text | Lossless | Lossless | Degraded | Lossless |

### Key Fidelity Gaps

Fountain is the format with the most fidelity concerns. See [format-fountain.md §Fidelity Analysis](format-fountain.md#fidelity-analysis) for full details. In summary:

- **Act headers** are emitted as a Fountain page break followed by visible centred bold text. First Folio can parse that representation back as an act, but other Fountain tools may treat it only as centred text.
- **Character tables** have no Fountain equivalent. They are rendered as plain Action text, which means a Fountain->org round-trip loses the table structure.
- **Prop text** maps to Fountain's centred text (`>TEXT<`), which loses the semantic distinction between "on-stage text" and "centred action".
- **Footnotes** map to Fountain Notes (`[[text]]`), which are not numbered and are invisible in formatted output. The name/number of the footnote is lost.

Text emitters preserve only the metadata keys they explicitly render. Org output includes title, subtitle, author, date, and version; Markdown includes title, subtitle, and author; Fountain includes title, subtitle, author, version, and date. Other parsed metadata remains available to PDF rendering but is not guaranteed to survive a text-format round trip.

### Rendering toggles

Individual event types can be suppressed via [configuration](config.md). Keys beneath `render:` control whether certain elements appear in output:

| Config key | Events suppressed when `false` |
|------------|-------------------------------|
| `render.stage-directions` | `stage_direction` |
| `render.frontmatter` | Introductory headers and text before the play proper |
| `render.footnotes` | `footnote` |
| `render.character-table` | `character_table_start`, `character_table_row`, `character_table_end` |
| `render.transitions` | `transition` |

Suppression is applied between the parser and emitter - the parser always emits all events, and the emitter always handles all events it receives. The config layer filters the event stream.

## Manuscript Path

`folio manuscript` is a prose rendering path rather than a stage-play conversion path. It accepts Markdown and org-mode manuscript contracts, rejects Fountain, and renders directly to Typst or PDF through the Go manuscript engine. It does not use the stage-play event stream because prose manuscripts have different structural elements: parts, chapters, sections, paragraphs, scene breaks, code, and manuscript metadata.
