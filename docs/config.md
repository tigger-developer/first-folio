<!-- Version: 0.2 | Last updated: 2026-07-14 -->

# Configuration

First Folio reads configuration from YAML files named `script.yaml`. It never creates, modifies, or writes to config files - they are maintained by the user or by other tools.

First Folio owns the `folio:` namespace. A project may also contain a top-level `yapper:` block, which belongs exclusively to Yapper and is ignored by First Folio. Only the documented top-level metadata and `render` keys form a shared contract between the applications.

See [examples/script.yaml.example](../examples/script.yaml.example) for an annotated example.

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
| `date` | string | (from source) | Date displayed on the title page |
| `version` | string | (from source) | Draft or version displayed on the title page |

### Shared rendering options

Control which elements appear in output. Read by both First Folio and yapper.

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `render.stage-directions` | bool | `true` | Include stage directions |
| `render.frontmatter` | bool | `true` | Include introductory sections before the play proper |
| `render.footnotes` | bool | `true` | Include footnotes |
| `render.character-table` | bool | `true` | Include the cast list |
| `render.transitions` | bool | `true` | Include transitions |

### First Folio PDF settings (`folio:`)

All First Folio-specific settings live under the `folio:` key.

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `font` | string | `Libertinus Serif` | Body font family |
| `font-size` | string | `12pt` | Body font size |
| `font-weight` | string | font default | Optional Typst font weight |
| `font-stretch` | string | font default | Optional Typst font stretch |
| `margin` | string | `25mm` | Page margins |
| `page` | string | `a4` | Page size |
| `default-format` | string | `pdf` | Default output format when no target file or `--to` given |
| `style` | string | `british` | Script style: `british`, `us`, or `screenplay` |

Script layout is configured beneath `folio.title-page` and `folio.positioning`. The canonical preset documents every supported child key. Important paths include:

| Path | Purpose |
|---|---|
| `folio.title-page.{title,subtitle,author,date,version}` | Title-page alignment, typography, spacing, and footer position |
| `folio.positioning.speech.space-before` | Space before a speech block |
| `folio.positioning.speech.speaker` | Speaker alignment, weight, case, prefix, and suffix |
| `folio.positioning.speech.speech-instruction` | Parenthetical placement, alignment, delimiters, and emphasis |
| `folio.positioning.speech.dialogue` | Same-line/new-line placement and wrapping indent |
| `folio.positioning.stage-direction` | Direction spacing, alignment, emphasis, case, and indentation |
| `folio.positioning.transition` | Transition spacing, alignment, and case |
| `folio.positioning.{frontmatter,act-header,scene-header}` | Header typography, spacing, alignment, case, and page breaks |

CLI layout options override their documented configuration equivalents. See `folio convert --help` for the public CLI surface.

### Letter settings (`folio.letter:`)

Letters use one layout rather than British/US variants. Supported keys are `font`, `font-size`, `font-weight`, `font-stretch`, `page`, `margin-top`, `margin-bottom`, `margin-left`, `margin-right`, `space-before-closing`, `space-before-signoff`, `space-after-sender`, `space-after-recipient`, `space-after-date`, and `space-after-subject`.

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
| `font` | string | `Libertinus Serif` | `Menlo` |
| `heading-font` | string | `Libertinus Sans` | `Menlo` |
| `mono-font` | string | `Libertinus Mono` | `Iosevka Custom` |
| `line-spacing` | number | `1.5` | `2` |
| `paragraph-indent` | string | `10mm` | `12.7mm` |
| `paragraph-spacing` | string | `0` | `0` |

`folio.manuscript.line-spacing` is a baseline multiplier: `1.0` is single-spaced, `1.5` is one-and-a-half-spaced, and `2.0` is double-spaced. `folio.manuscript.paragraph-spacing` is additional space between paragraphs; `0` preserves the selected line interval across paragraph boundaries without adding a separate paragraph gap.

`folio.manuscript.page-header.content-padding-after` controls the clearance between the running header and the manuscript body on every running-header page. It does not affect the title page or table of contents.

`folio.manuscript.toc.enabled` defaults to `true`. Set it to `false` to suppress the generated table of contents.

`folio.manuscript.toc.line-spacing` controls table-of-contents item line spacing. The British default is `1.15em`.

US manuscript style is selected with `folio.manuscript.style: us` or `folio.style: us`, or with `folio manuscript --style us ...`. The US override is layered on top of the British manuscript preset and does not change the page size to `us-letter`; page size changes require explicit user config.

Manuscript metadata supports `title`, `subtitle`, `author`, `attribution`, `date`, `version`, `wordcount`, `contact-name`, `address`, `phone`, `email`, and `website`. `wordcount` is display text, not a numeric field; values such as `about 90,000 words`, `approx 100k words`, and `20.000 mots` render as entered.

`folio.manuscript.date-format` controls title-page date rendering for ISO frontmatter dates using Go date layouts. British defaults to `2 January 2006`; US overrides default to `January 2, 2006`.

`folio.manuscript.toc.part-gap-before` controls extra vertical space before part entries in the table of contents. The default is `0.5em`.

`folio.manuscript.toc.part-bold` controls whether part entries are bold in the table of contents. The default is `true`.

### Yapper namespace (`yapper:`)

Anything beneath a top-level `yapper:` block is exclusively Yapper configuration and is ignored by First Folio. First Folio does not define or document Yapper's child keys; see the [Yapper documentation](https://github.com/tadg-paul/yapper) for that schema.

## YAML

Config files are parsed with `gopkg.in/yaml.v3` and support standard YAML mappings and scalar values. Common project configuration uses:

- Scalar values: `key: value`, `key: "quoted"`, `key: 'single quoted'`
- Nested maps: a key followed by indented `key: value` lines
- Comments: `# comment` (full-line or inline)
- Booleans: `true`/`false`/`yes`/`no`/`on`/`off`

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
  positioning:
    speech:
      dialogue:
        wrap-indent: 5em
```
