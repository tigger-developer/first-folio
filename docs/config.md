<!-- Version: 0.1 | Last updated: 2026-04-26 -->

# Configuration

First Folio reads configuration from YAML files named `script.yaml`. It never creates, modifies, or writes to config files - they are maintained by the user or by other tools.

The config file is designed to be shared with [yapper](https://github.com/tadg-paul/yapper) (a TTS rendering tool for play scripts). Each tool reads shared top-level keys and its own namespace, silently ignoring everything else.

See [examples/script.yaml](../examples/script.yaml) for a complete annotated example.

## File locations

| Location | Purpose |
|----------|---------|
| `~/.config/first-folio/script.yaml` | Global user defaults |
| `<source-file-directory>/script.yaml` | Per-project overrides |

Both files are read when they exist. Per-project values override global values. CLI flags override everything.

## Precedence - layered merge

All config sources are read and merged. Each layer overrides individual keys from the layers below - not the entire config. This allows global defaults (e.g. font, page size) to coexist with per-project overrides (e.g. title, author).

| Priority | Source |
|----------|--------|
| 1 (highest) | CLI flags |
| 2 | Local `script.yaml` (source file directory) |
| 3 | Global `~/.config/first-folio/script.yaml` |
| 4 (lowest) | Built-in defaults |

**Example:** Global config sets `folio.font: "EB Garamond"` and `folio.page: a4`. A local config sets only `folio.font: "Georgia"`. The merged result uses Georgia for the font and a4 for the page - the local config overrides one key without erasing the rest.

## Schema

### Shared metadata

These keys are read by both First Folio and yapper. When present, they override any corresponding values found in the source document (e.g. `#+TITLE` in org-mode).

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `title` | string | (from source) | Play title |
| `subtitle` | string | (from source) | Play subtitle |
| `author` | string | (from source) | Author name |
| `draft-date` | string | (none) | Draft date, displayed on the PDF title page |

### Shared rendering options

Control which elements appear in output. Read by both First Folio and yapper.

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `render-stage-directions` | bool | `true` | Include stage directions in output |
| `render-intro` | bool | `true` | Include title page / opening block |
| `render-footnotes` | bool | `true` | Include footnotes in output |
| `render-character-table` | bool | `true` | Include cast list in output |

### First Folio PDF settings (`folio:`)

All First Folio-specific settings live under the `folio:` key. These control PDF rendering via Typst and are silently ignored by yapper.

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `font` | string | `New Computer Modern` | Body font family |
| `font-size` | string | `12pt` | Body font size |
| `heading-font` | string | inherits `font` | Heading font family |
| `heading-font-size` | string | inherits `font-size` | Heading font size |
| `margin` | string | `25mm` | Page margins |
| `page` | string | `a4` | Page size (a4, us-letter, etc.) |
| `indent` | string | `5em` | Dialogue indent depth |
| `dialogue-spacing` | string | `1.6em` | Vertical space before dialogue blocks |
| `direction-spacing` | string | `1.6em` | Vertical space before stage directions |
| `direction-italic` | bool | `true` | Italicize stage directions |
| `direction-center` | bool | `false` | Centre stage directions |
| `default-format` | string | `pdf` | Default output format when no target file or `--to` given |

All `folio:` keys correspond to CLI flags of the same name. CLI flags override config values.

### Manuscript settings (`folio.manuscript:`)

Manuscript settings inherit from root `folio:` values unless a manuscript or child override is present. The inheritance model is:

1. Child override, such as `folio.manuscript.toc.font`
2. Manuscript override, such as `folio.manuscript.font`
3. Root default, such as `folio.font`
4. Active preset default

For heading fonts, `folio.manuscript.heading-font` inherits from `folio.heading-font`, which inherits from `folio.font`.

Common manuscript keys:

| Key | Type | British default | US override |
|---|---|---|---|
| `page` | string | `a4` | inherited |
| `margin` | string | `20mm` | `25mm` |
| `font` | string | `Libertinus Serif` | `Liberation Mono` |
| `heading-font` | string | `Libertinus Sans` | `Liberation Mono` |
| `mono-font` | string | `Libertinus Mono` | `Menlo` |
| `line-spacing` | number | `1.5` | `2` |
| `paragraph-indent` | string | `10mm` | `12.7mm` |
| `paragraph-spacing` | string | `0` | `0` |

`folio.manuscript.toc.enabled` defaults to `true`. Set it to `false` to suppress the generated table of contents.

`folio.manuscript.toc.line-spacing` controls table-of-contents item line spacing. The British default is `1.15em`.

US manuscript style is selected with `folio.manuscript.style: us` or `folio.style: us`, or with `folio manuscript --style us ...`. The US override is layered on top of the British manuscript preset and does not change the page size to `us-letter`; page size changes require explicit user config.

Manuscript metadata supports `title`, `subtitle`, `author`, `attribution`, `date`, `version`, `wordcount`, `contact-name`, `address`, `phone`, `email`, and `website`. `wordcount` is display text, not a numeric field; values such as `about 90,000 words`, `approx 100k words`, and `20.000 mots` render as entered.

`folio.manuscript.date-format` controls title-page date rendering for ISO frontmatter dates using Go date layouts. British defaults to `2 January 2006`; US overrides default to `January 2, 2006`.

`folio.manuscript.toc.part-gap-before` controls extra vertical space before part entries in the table of contents. The default is `0.5em`.

`folio.manuscript.toc.part-bold` controls whether part entries are bold in the table of contents. The default is `true`.

### Yapper-specific keys

The following keys are examples of yapper configuration. First Folio silently ignores them. See [yapper documentation](https://github.com/tadg-paul/yapper) for the full reference.

- `auto-assign-voices`, `character-voices`, `narrator-voice`, `intro-voice`
- `dialogue-speed`, `stage-direction-speed`, `gap-after-dialogue`, `gap-after-stage-direction`, `gap-after-scene`
- `speech-substitution`, `threads`

## YAML subset

Config files use a restricted YAML subset. The parser handles:

- Scalar values: `key: value`, `key: "quoted"`, `key: 'single quoted'`
- One-level-deep maps: a key followed by indented `key: value` lines
- Comments: `# comment` (full-line or inline)
- Booleans: `true`/`false`/`yes`/`no`/`on`/`off`

Not supported: multi-line strings, anchors/aliases, flow style (`{}`/`[]`), multi-document (`---`), nested maps beyond one level, sequences.

Malformed YAML produces a descriptive error with the file path and line number.

## Migration from ~/.config/org-script/

The old flat key=value config at `~/.config/org-script/config` is no longer read. To migrate:

1. Create `~/.config/first-folio/script.yaml`
2. Move settings into the `folio:` namespace:

**Old format (`~/.config/org-script/config`):**
```
font = EB Garamond
font-size = 11pt
margin = 25mm
page = a4
indent = 5em
```

**New format (`~/.config/first-folio/script.yaml`):**
```yaml
folio:
  font: EB Garamond
  font-size: 11pt
  margin: 25mm
  page: a4
  indent: 5em
```
