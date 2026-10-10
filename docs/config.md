---
title: Configuration
version: "0.25"
last-updated: 2026-10-08
---

# Configuration

First Folio reads configuration from YAML files named `script.yaml`. It never creates, modifies, or writes to config files - they are maintained by the user or by other tools.

First Folio owns the `folio:` namespace. A project may also contain a top-level `yapper:` block, which belongs exclusively to Yapper and is ignored by First Folio. Only the documented top-level metadata and `render` keys form a shared contract between the applications.

See [examples/script.yaml.example](../examples/script.yaml.example) for an annotated example.

## File locations

| Location | Purpose |
|----------|---------|
| `~/.config/first-folio/script.yaml` | Global user defaults |
| Nearest `script.yaml` at or above the source directory | Per-project overrides |

Local discovery starts in the source directory and walks upwards towards HOME. The nearest `script.yaml` wins; multiple local files are not merged. For a multi-file manuscript, the first resolved input supplies the starting directory. A style-specific sibling such as `script-us.yaml` or `script-screenplay.yaml` is loaded from the same directory as its base file.

## Precedence - layered merge

All config sources are read and merged. Each layer overrides individual keys from the layers below - not the entire config. This allows global defaults (e.g. font, page size) to coexist with per-project overrides (e.g. title, author).

| Priority | Source |
|----------|--------|
| 1 (highest) | CLI flags |
| 2 | Nearest local `script-<style>.yaml` |
| 3 | Nearest local `script.yaml` |
| 4 | First Markdown manuscript input's `folio:` and `render:` frontmatter |
| 5 | Global `~/.config/first-folio/script-<style>.yaml` |
| 6 | Global `~/.config/first-folio/script.yaml` |
| 7 | Selected built-in style override |
| 8 (lowest) | British built-in base preset |

**Example:** Global config sets `folio.font.family: "EB Garamond"` and `folio.page: a4`. A local config sets only `folio.font.size: 11pt`. The merged root font uses EB Garamond at 11pt and the page remains a4. Font properties merge only at the same role path.

## Manuscript source configuration

The first resolved Markdown manuscript input may contain `folio:` and `render:` mappings alongside its metadata. These mappings merge recursively **after global files and before local files**. A local layout changes only the properties it specifies; it does not replace an entire font or manuscript object.

For example, source frontmatter can set `folio.manuscript.font.family: DejaVu Sans Mono`, while local `script.yaml` sets `folio.manuscript.font.size: 11.5pt`. Both values apply, and other font properties inherit from lower layers at that same role path. If both source and local files set the size, the local value wins. This supports common manuscript settings with separate A4 and 6x9 layout files.

The same merge and effective-value validation rules apply as for external YAML. The `folio` and `render` namespace values themselves must be mappings; null or scalar namespaces are errors. Source values overridden by local properties are validated as part of the final merged configuration. Unknown-key handling is unchanged from external configuration; this is not a new strict-schema validator. `render` retains its existing mode-specific meaning and gains no new manuscript rendering switches.

A source `folio.style` or `folio.manuscript.style` participates in style selection below local configuration and explicit `--style`. The selected style determines the built-in overlay and style-specific files. Source metadata such as title and author retains its existing handling, separate from these configuration mappings.

For multiple input files, resolution sorts paths and uses the first. A later input with a decoded `folio:` or `render:` frontmatter mapping is rejected with its path before output is written. Other later-file content remains ordinary Markdown body; leading scene breaks are not treated as malformed configuration. Configuration in Org manuscripts, stage-play sources and letters is unchanged. Source configuration is not emitted as body content. File discovery still starts from the first input's directory.

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

Manuscript mode additionally accepts top-level `attribution`, `author-attribution`, `wordcount`, `contact-name`, `address`, `phone`, `email`, and `website`. These override corresponding manuscript frontmatter values and are display strings, including `wordcount`; First Folio does not impose numeric or language-specific formatting.

### Shared rendering options

Control which elements appear in output. Read by both First Folio and yapper.

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `render.stage-directions` | bool | `true` | Include stage directions |
| `render.frontmatter` | bool | `true` | Include introductory sections before the play proper |
| `render.footnotes` | bool | `true` | Include footnotes |
| `render.character-table` | bool | `true` | Include the cast list; hiding it does not disable character-name lookup |
| `render.transitions` | bool | `true` | Include transitions |

Character names declared in the cast table are automatically capitalized in stage
directions and screenplay action, never in dialogue. No configuration switch is
needed. Name cells may declare aliases as `ALISON (NURSE ALISON)` or
`MARGARET (MAGS, MARG.)`; primary names and aliases are equal lookup matches.
PDF/Typst cast lists show only the primary name; Org and Markdown conversions
preserve alias declarations. See the [Org cast-table schema](../schema/script.org)
and [Markdown alias syntax](../schema/script.md#character-name-aliases).

### First Folio PDF settings (`folio:`)

All First Folio-specific settings live under the `folio:` key.

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `font` | font block | British root font | Body typography |
| `heading.font` | font block | British heading font | Shared heading typography |
| `margin` | string | `25mm` | Page margins |
| `page` | string | `a4` | Page size |
| `default-format` | string | `pdf` | Default output format when no target file or `--to` given |
| `style` | string | `british` | Script style: `british`, `us`, or `screenplay` |

Script layout is configured beneath `folio.title-page` and `folio.positioning`. The canonical preset documents every supported child key. Important paths include:

| Path | Purpose |
|---|---|
| `folio.title-page.enabled` | Separate script title page (`true`); compact first-page title block (`false`) |
| `folio.title-page.skip-header`, `folio.title-page.skip-footer` | Suppress corresponding running matter on the title/first page (both default `true`) |
| `folio.title-page.{title,subtitle,author,date,version}` | Title-page alignment, typography, spacing, and footer position |
| `folio.positioning.speech.space-before` | Space before a speech block |
| `folio.positioning.speech.speaker` | Speaker alignment, weight, case, prefix, and suffix |
| `folio.positioning.speech.speech-instruction` | Parenthetical placement, alignment, delimiters, and emphasis |
| `folio.positioning.speech.dialogue` | Same-line/new-line placement and wrapping indent |
| `folio.positioning.stage-direction` | Direction spacing, alignment, emphasis, case, and indentation |
| `folio.positioning.transition` | Transition spacing, alignment, and case |
| `folio.positioning.{frontmatter,act-header,scene-header}` | Header typography, spacing, alignment, case, and page breaks |

The effective CLI layout overrides are `--font`, `--font-size`, `--margin`, and `--page`; `--style` selects the preset layer. Other layout changes belong in `script.yaml`. See `folio convert --help` for the public CLI surface.

#### Compact script title block

`folio.title-page.enabled` defaults to `true`, preserving the separate title page when a script has a title. Set it to `false` for page-limited submissions:

```yaml
folio:
  title-page:
    enabled: false
```

This moves the available title, subtitle, author, date and version to a block at the top of the first content page rather than removing the metadata. Title-role fonts, subtitle/author spacing and author prefixes remain configurable. The dedicated title-page vertical position is not applied. Date/version appear below the title block rather than in the page footer, leaving running matter independent.

The script follows on the same page, without a forced title-page break or a break before the first act. Later acts retain `folio.positioning.act-header.page-break-before`; content may still paginate naturally. Introductory material remains controlled by `render.frontmatter` and `render.character-table`. Script page numbering begins at physical page 1; the first-page running header and footer are hidden by default through the suppression switches below.

This switch applies to `folio convert` script PDF/Typst output, including British, US and screenplay styles. It does not change text-format conversion or letters. Manuscripts retain their independent `folio.manuscript.title-page.enabled` setting; disabling a manuscript title page does not create this compact script block.

#### Script title/first-page running matter

`folio.title-page.skip-header` and `folio.title-page.skip-footer` both default to `true`. They suppress only the corresponding running header or footer on the first physical page:

- With a separate title page enabled, that title page has no running matter by default.
- With the separate title page disabled, suppression instead applies to page 1 containing the compact title and script.
- Subsequent pages use the shared `folio.page-header` and `folio.page-footer` settings normally.

If `enabled: true` but the source has no title, no separate title page is generated and these switches do not suppress the untitled script's first-page running matter. With `enabled: false`, they apply to physical page 1 even when title metadata are absent.

Set either switch to `false` to allow that running-matter role on the first page. The shared role must also be enabled; a suppression switch never enables a globally disabled header/footer. Hiding the footer does not reset numbering: a displayed page number on the next page remains 2. Date/version title metadata are not running matter and remain visible.

Skipping a role hides its text, not its reserved running-page margin. A dedicated title page retains its existing title-page margins and vertical position. `folio.title-page.page-number` remains the independent legacy dedicated-title-page numbering setting; `skip-footer` controls the shared running footer, not that setting. `enabled`, `skip-header` and `skip-footer` accept YAML booleans only in script mode; quoted strings, null and other types are rejected before output.

```yaml
folio:
  title-page:
    enabled: false
    skip-header: true
    skip-footer: true
```

#### Uniform font blocks

Every configurable typography role uses the same mapping:

```yaml
font:
  family: Libertinus Serif
  size: 12pt
  weight: regular
  stretch: 100%
  style: regular
  letter-spacing: 0em
```

The six properties are:

| Property | Accepted values |
|---|---|
| `family` | Non-empty string |
| `size` | Positive decimal with `pt`, `mm`, `cm`, `in`, or `em` |
| `weight` | `thin`, `extralight`, `light`, `regular`, `medium`, `semibold`, `bold`, `extrabold`, `black`, or an integer from 100 through 900 |
| `stretch` | Positive finite decimal, optionally followed by `%`; a plain number means percent |
| `style` | `regular`, `italic`, or `oblique` |
| `letter-spacing` | Signed decimal with `pt`, `mm`, `cm`, `in`, or `em`; zero is valid |

For roles other than manuscript body, quoted blocks and monospace, `stretch` selects an available width within the chosen font family. If that width is unavailable, Typst selects the nearest available face; it does not geometrically compress or expand the text. `letter-spacing` separately adjusts the space between characters.

For **manuscript body and quoted blocks**, `font.stretch` and `quoted-block.font.stretch` scale glyphs horizontally from the normal-width face: 160% means 1.6 times the width, with unchanged glyph height. Bold and italic retain their styles. Each role has its own ratio; code inside prose or a quote still uses `mono.font.stretch`. Word spaces scale too, and justification remains active. The `hyphenation` setting controls automatic word splitting at normal and stretched widths. At 100%, prose keeps native Typst line breaking.

For **manuscript monospace**, `mono.font.stretch` instead specifies proportional horizontal scaling from the normal-width face: 150% is 1.5 times the width, 200% twice, and 50% half. Glyph height stays unchanged. Words and inter-word spacing scale, so paragraphs can still wrap between code words and code blocks keep their column alignment and their page breaks. This corrects the earlier interpretation of monospace stretch as font-face selection only.

Paragraph inline monospace and fenced code blocks receive the same treatment: they are one font role, so a given `mono.font` configuration renders identically wherever monospace appears. Scaling applied to inline code from version 0.12 was extended to code blocks in version 0.13.

Manuscript monospace text uses the complete `folio.manuscript.mono.font` block. Ordinary manuscript headings use `folio.manuscript.heading.font`, including its configured size and weight without implicit heading-level scaling or bold. Table-of-contents entries, running headers and footers retain their own font roles. Existing source emphasis and title casing still apply.

Code within headings and their table-of-contents entries is outside the manuscript font-rendering repair's acceptance scope. The earlier documentation included those cases; that inclusion was withdrawn on 2026-09-12. This scope change does not disable existing syntax.

Higher-precedence files may set only the properties they change. Omitted properties come from the lower layer at the same role path. A role never inherits font properties from its parent or from another role. The old scalar, prefixed, `bold`, and `italic` font keys are rejected with their full paths.

The exhaustive [British base](../presets/british.yaml) is both the lowest-precedence runtime configuration and the public inventory. [US](../presets/us.yaml) and [screenplay](../presets/screenplay.yaml) contain only differences from it.

### Letter settings (`folio.letter:`)

Letters use one layout rather than British/US variants. Supported keys are `font` using the uniform six-property block, `page`, `margin-top`, `margin-bottom`, `margin-left`, `margin-right`, `space-before-closing`, `space-before-signoff`, `space-after-sender`, `space-after-recipient`, `space-after-date`, and `space-after-subject`.

Letters are supplementary correspondence, not manuscript pages. They never inherit `folio.page-header` or `folio.page-footer`; those settings are ignored without warnings in letter mode.

Running-matter values are not validated in letter mode, so an irrelevant malformed header or footer setting does not prevent letter generation. Letter typography and layout retain their own validation.

### Shared running matter (`folio.page-header:` and `folio.page-footer:`)

Running headers and footers apply to stage-play and screenplay conversion and to prose manuscripts. They are distinct from title-page text and act, scene, part, and chapter headings.

Both blocks accept `enabled`, `font`, `format`, `alt-format`, `frontmatter-format`, `alt-frontmatter-format`, `align`, `distance-from-edge`, and `content-padding-after`. The British base defines a complete font block for each role; partial overrides merge with that same role, never with body or heading typography.

The header defaults to disabled; the footer defaults to enabled. To display a script heading, set `folio.page-header.enabled: true` and choose a script-friendly format such as `"[title] • [author]"`. The inherited header format retains `[chapter]` for manuscript compatibility; that token contributes no text in scripts, but surrounding separators remain literal.

`distance-from-edge: 20mm` and `content-padding-after: 10mm` retain the manuscript layout geometry: the configured edge distance and body clearance determine the corresponding page margin. Disabled running matter restores the ordinary margin on its side. Lengths may use `pt`, `mm`, `cm`, `in`, or `em`.

Script running matter now uses these shared dimensions rather than the former automatic footer placement. The enabled default footer therefore reserves a 30mm bottom margin; scripts previously used the ordinary page margin on that side. Pagination can change. Script `[page]` retains physical page numbering, including the title page in the count, without the manuscript body-number reset.

The former `folio.manuscript.page-header` and `folio.manuscript.page-footer` blocks remain supported as deprecated manuscript-only overrides. Manuscript invocations report a migration warning on stderr. Within each configuration layer, their specified fields override shared fields, including partial font blocks. Normal layer precedence still applies: a local shared value overrides a global legacy value. Scripts and letters ignore these legacy manuscript blocks. Move them directly under `folio` to migrate.

For example, the following shared configuration enables running headings in both script and manuscript output. It replaces the same blocks formerly nested beneath `folio.manuscript`:

```yaml
folio:
  page-header:
    enabled: true
    format: "[title] • [author]"
    distance-from-edge: 20mm
    content-padding-after: 10mm
  page-footer:
    enabled: true
    format: "[page]/[total-pages]"
    distance-from-edge: 20mm
    content-padding-after: 10mm
```

Remove the old manuscript-local blocks after moving their fields. Leaving them present intentionally retains manuscript-specific overrides and deprecation warnings. A manuscript-only chapter heading can continue to use `[chapter]` in that override; the shared example uses only placeholders supported by both renderers.

### Manuscript settings (`folio.manuscript:`)

Manuscript settings use the same configuration-file precedence as scripts and letters. Font inheritance is strictly same-path layering. Running matter uses `folio.page-header.font` and `folio.page-footer.font`, not manuscript body or heading fonts.

Common manuscript keys:

| Key | Type | British default | US override |
|---|---|---|---|
| `page` | string | `a4` | inherited |
| `margin` | string | `20mm` | `25mm` |
| `font` | font block | Libertinus Serif, 12pt | Menlo, 10pt |
| `heading.font` | font block | Libertinus Sans, 14pt | Menlo, 10pt |
| `mono.font` | font block | Libertinus Mono, 10pt | Iosevka Custom, 9pt bold |
| `line-spacing` | number | `1.5` | `2` |
| `justify` | bool | `true` | inherited (`true`) |
| `widow-orphan-control` | bool | `true` | inherited (`true`) |
| `paragraph-indent` | string | `10mm` | `12.7mm` |
| `paragraph-spacing` | string | `0` | `0` |

`folio.manuscript.line-spacing` accepts only a **positive finite numeric multiplier**: `1.0` is single-spaced, `1.5` is one-and-a-half-spaced, and `2.0` is double-spaced. Quoted numeric values are also accepted. The multiplier applies to the selected body font's normal baseline interval at its configured size, using native Typst text metrics and its normal `0.65em` interline gap. It does not multiply the point size alone. Values below one compress that interval without being silently clamped; they can cause text to overlap.

**Migration:** lengths such as `1em`, `5mm` and `12pt`, zero, negative values and non-finite numbers are rejected before output is written. Replace a body line-spacing length with the intended multiplier, such as `1.0` or `1.5`; removing the unit is not an exact conversion from the former extra-gap behaviour. Existing numeric configurations also gain the font-relative interval, which can change pagination. The British base uses `1.5` and the US overlay uses `2`.

`folio.manuscript.paragraph-spacing` remains an **additional non-negative length** between paragraphs, such as `0.5em` or `5mm`. `0` or `0pt` preserves the selected line interval across paragraph boundaries without an extra gap. Contents and copyright spacing retain their separate settings. `folio.manuscript.font.letter-spacing` controls body tracking; header and footer tracking use their own font blocks. `folio.manuscript.justify` controls body-text justification.

`folio.manuscript.widow-orphan-control` defaults to `true`, preventing a single paragraph line from being stranded at the bottom or top of a page. Set it to `false` to allow paragraphs to split freely at page boundaries. This setting does not keep whole paragraphs together.

`folio.page-header.content-padding-after` controls the clearance between the running header and the body on every running-header page. Manuscript title pages and tables of contents retain their separate layout.

#### Complete manuscript key inventory

The built-in [British base](../presets/british.yaml) is the canonical default and exhaustive inventory. The [US overlay](../presets/us.yaml) contains only differences. Every font entry below means the uniform six-property block.

| Group | Accepted keys |
|---|---|
| Core | `style`, `page`, `margin`, `gutter`, `line-spacing`, `justify`, `hyphenation`, `widow-orphan-control`, `paragraph-indent`, `paragraph-spacing` |
| Body typography | `font` |
| Heading typography | `heading.font` |
| Monospace typography | `mono.font` |
| Title-page typography | `title-page.<item>.font`, where `<item>` is `title`, `subtitle`, `author`, `date`, `wordcount`, `version`, or `contact` |

| Nested block | Accepted child keys |
|---|---|
| `toc` | `enabled`, `links`, `title`, `font`, `heading.font`, `include-parts`, `include-chapters`, `include-sections`, `dot-leaders`, `page-numbers`, `page-break-before`, `blank-page-before`, `blank-page-after`, `line-spacing`, `part-gap-before`, `continuation-padding-before`, `part-bold` |
| `title-page` | `enabled`, `page-number`, `include-title`, `include-subtitle`, `include-author`, `include-date`, `include-wordcount`, `include-contact-name`, `include-address`, `include-phone`, `include-email`, `include-website`, `include-version`, `title-block-align`, `footer-align` |
| `title-page.<item>` | `align`, `font`, where `<item>` is `title`, `subtitle`, `author`, `date`, `wordcount`, `version`, or `contact` |
| `scene-break` | `marker` |
| `quoted-block-spacing`, `code-block-spacing` | Equal clearance above and below quoted and fenced-code blocks |
| `quote-block-indent`, `code-block-indent` | Left inset applied to every line of quoted and fenced-code blocks; defaults to `0em` |
| `quoted-block.font` | Uniform font block; omitted properties inherit only from the same path in a lower configuration layer |
| `list`, `table`, `code-block` | `space-before`, `space-after`; code-block values override `code-block-spacing` on their respective side |
| `page-numbering` | `frontmatter-format`, `body-format`, `body-reset` |

`code-block` controls layout only and accepts no font block. Fenced-code typography comes from `folio.manuscript.mono.font`, the same role as inline code. Quoted blocks differ: they carry their own `quoted-block.font`.

Quoted-block clearance also adds to normal body paragraph separation. `quote-block-indent` is the entire left inset from the body text edge: `6mm` means exactly 6mm, with no native quote inset or prose first-line indent added. Every paragraph in the quote uses that same inset. Quote clearance and indentation in `em` resolve against the body font size, independently of `quoted-block.font`.

Code-block clearance is additional to normal body paragraph separation. This applies to fenced blocks and complete-line code spans promoted to blocks, including consecutive blocks. `code-block-spacing` sets both sides; `code-block.space-before` and `.space-after` override their respective sides. These vertical `em` clearances resolve against the body font size. Code indentation retains its existing basis. Inline code within prose remains part of its paragraph.

The `part` and `chapter` blocks share this shape:

| Key | Purpose |
|---|---|
| `page-break-before`, `blank-page-before`, `blank-page-after` | Page and parity control |
| `skip-header`, `skip-footer` | Suppress running matter on the heading page |
| `vertical-align`, `position`, `align` | Heading placement; `vertical-align` is principally for parts and `position` for chapters |
| `case-transform`, `name-case` | Case of the complete heading or semantic name |
| `space-after` | Clearance after the heading |
| `prefix`, `separator`, `suffix` | Compose the displayed heading |
| `number-format`, `show-number`, `show-name` | Number/name presentation |
| `explicit-numbering` | Use `derived` source order or `source` number text |
| `number-reset` | `never` or, for chapters, `per-part` |

The `copyright` block accepts:

| Group | Accepted keys |
|---|---|
| Page control | `enabled`, `position`, `skip-header`, `skip-footer`, `blank-page-before`, `blank-page-after`, `align` |
| Content | `credits`, `body`, `separator`, `separator-space-before`, `separator-space-after`, `publication`, `publisher`, `publisher-preposition`, `isbn`, `isbn-label`, `isbn-barcode` |
| Typography | `font`, `label.font`, `line-spacing`, `block-spacing` |

### Page-header format placeholders

`folio.page-header.format` and `folio.page-footer.format` accept the following placeholders, substituted at render time. `[author]`, `[title]`, `[page]`, and `[total-pages]` apply to both scripts and manuscripts. Known part/chapter placeholders resolve to empty text in scripts; they do not select an act or scene.

- `[author]` -- the document author
- `[title]` -- the document title
- `[page]` -- the current page number
- `[total-pages]` -- the final physical page count in Arabic numerals, including frontmatter and intentional blank pages
- `[part]` -- the current part's **semantic name** (issue #18: whatever remains after `Part N:` prefix stripping; e.g. `Unbelieved` for a source heading `# PART ONE: UNBELIEVED`)
- `[part-number]` -- the current part's number, formatted per `part.number-format` (`1`, `I`, `i`)
- `[part-prefix]` -- the configured `part.prefix` string
- `[part-full]` -- the fully rendered part heading (`prefix + number + separator + name + suffix`)
- `[chapter]` -- current chapter's semantic name (analogous to `[part]`)
- `[chapter-number]` -- current chapter's formatted number
- `[chapter-prefix]` -- configured `chapter.prefix` string
- `[chapter-full]` -- fully rendered chapter heading

Unknown bracket tokens (e.g. `[unknown]`) are rendered as literal text.

The British and US presets both default to `format: "[title] • [chapter] • [author]"` for the header and `format: "[page]"` for the footer.

Use both page placeholders to render values such as `4/100`:

```yaml
folio:
  page-footer:
    format: "[page]/[total-pages]"
```

`[total-pages]` always reports the final physical page count. Unlike `[page]`, it is not affected by `page-numbering.body-reset` or frontmatter/body numbering formats.

#### `alt-format` for facing-page layouts

`page-header.alt-format` and `page-footer.alt-format` (issue #18 AC18.6) are optional companion format strings. When set, `format` renders on left (verso, even) pages and `alt-format` renders on right (recto, odd) pages. Common use: put the page number on the *outer* edge of the book on both pages by pairing the two format strings mirror-image.

```yaml
folio:
  page-header:
    format:     "[page] • [title] • [author]"     # verso (left) -- [page] on outer
    alt-format: "[author] • [title] • [page]"     # recto (right) -- [page] on outer
```

When `alt-format` is unset, `format` renders on every page (unchanged from AC15.1).

### Page-footer block

`folio.page-footer` mirrors the non-font fields of `folio.page-header`. Each has its own complete `font` block in the British base. A partial user block merges only with the lower layer at that exact header or footer path. Default: enabled with a centred `[page]` number, `distance-from-edge` and `content-padding-after` matching `page-header`. Set `folio.page-footer.enabled: false` to omit the running footer, including automatic script page numbering. Script title-page date/version text remains separate.

### Frontmatter-format (issue #24)

`page-header` and `page-footer` each accept `frontmatter-format` and `alt-frontmatter-format` that apply on running-matter frontmatter pages. Manuscripts use their existing frontmatter/body boundary before the first part or chapter, and retain title-page, copyright, contents and heading-page suppression rules. Scripts use introductory material such as synopsis and cast pages as frontmatter; dramatic content begins the body. Script title pages skip running matter by default; setting the corresponding `folio.title-page.skip-header` or `skip-footer` to `false` allows the enabled role and its frontmatter format there. Body pages use the normal `format` / `alt-format` pair.

- **Unset** (key absent from YAML) -> frontmatter pages use `format` / `alt-format` (backwards-compatible, no change).
- **Set to non-empty string** -> that string renders on frontmatter pages.
- **Set to empty string `""`** -> frontmatter pages render blank (no header or footer text).
- **`alt-frontmatter-format` set alongside `frontmatter-format`** -> verso frontmatter uses `frontmatter-format`, recto frontmatter uses `alt-frontmatter-format` (same verso/recto pairing as `format` / `alt-format`).

For manuscripts, pages before the first part or chapter block are frontmatter; from that heading onward they are body. For scripts, the first act, scene, speech, stage direction, prop text or transition begins the body. A page containing that first dramatic event uses body running formats, even if introductory material occurs earlier on the same page.

Example --- suppress the running header on frontmatter but keep body headers:

```yaml
folio:
  page-header:
    format: "[title] • [author]"
    frontmatter-format: ""       # blank on frontmatter
  page-footer:
    format: "[page]"
    frontmatter-format: "[page]" # keep page numbers on frontmatter too
```

### Book-layout page-pair alignment

`page-header.align` and `page-footer.align` accept:

- a compass keyword: `left`, `center`, `right` -- applied uniformly to every page
- a compound page-pair alias: `left-right`, `right-left`, `left-center`, `right-center`, `center-left`, `center-right` -- **first token = LEFT (verso, even) page, second token = RIGHT (recto, odd) page**, matching the reader's view of an open book. `left-right` therefore places left-alignment on verso pages and right-alignment on recto pages, which is the classical outer-edge running-head convention.

Default: `align: left-right` for the header (outer-edge, both sides), `align: center` for the footer.

### Custom page dimensions

`folio.manuscript.page` accepts either a named Typst preset (`a4`, `us-letter`, `uk-book-b`, ...) or a custom `WxHmm` dimension. Imperial custom dimensions remain accepted for backward compatibility but public configuration should use metric values.

```yaml
folio:
  manuscript:
    page: 140x216mm    # trade paperback
    # or
    page: 200x300mm    # custom hardback
```

Both dimensions must share the same unit. Values that match neither shape (e.g. `200mm`, mixed-unit dimensions, or `bogus`) are rejected at config load with a diagnostic naming the offending value.

### Binding gutter

`folio.manuscript.gutter` (default `0mm`) is a Typst length that is added to the inside (binding-side) margin on odd and even pages. Under the hood the running-page margin switches to Typst's `inside`/`outside` idiom, which mirrors sides automatically per page parity:

```yaml
folio:
  manuscript:
    gutter: 15mm
```

A `0mm` gutter leaves the running-page margin configuration byte-identical to the pre-gutter behaviour.

### Blank pages before or after headings

`folio.manuscript.part.blank-page-before`, `part.blank-page-after`, `chapter.blank-page-before`, `chapter.blank-page-after`, `toc.blank-page-before`, and `toc.blank-page-after` accept:

- `false` (default) -- no blank page.
- `true` -- insert one unconditional unnumbered blank page adjacent to the heading.
- `enforce-right` -- ensure the next section starts on a right-hand (recto/odd) page; a blank page is inserted only if needed to reach that parity. Uses Typst's `pagebreak(to: "odd")`.
- `enforce-left` -- ensure the next section starts on a left-hand (verso/even) page. Uses Typst's `pagebreak(to: "even")`.

Independent of `page-break-before`; combining `page-break-before: true` with `blank-page-before: true` produces one blank page and one heading page (no doubling). Combining with `enforce-right` / `enforce-left` inserts the parity blank if and only if the natural next page has the wrong parity.

### Page numbering (issue #16)

`folio.manuscript.page-numbering` controls the number style on frontmatter and body pages, and whether the display counter restarts at the frontmatter/body boundary.

```yaml
folio:
  manuscript:
    page-numbering:
      frontmatter-format: "i"          # "1" (default), "I", "i"
      body-format: "1"                  # "1" (default), "I", "i"
      body-reset: first-part-or-chapter # default; also "never"
```

- **`frontmatter-format`** --- style of the page number when `[page]` is used in `page-header.frontmatter-format` or `page-footer.frontmatter-format`. Accepts `"1"` (arabic, default), `"I"` (Roman upper), `"i"` (Roman lower).
- **`body-format`** --- style of the page number on body pages (from the first part or chapter onward). Same accepted values.
- **`body-reset: first-part-or-chapter`** (default) --- the display counter restarts at 1 at the first body block.
- **`body-reset: never`** --- the display counter continues through frontmatter and body without a restart.

### Chapter number reset (issue #16)

`chapter.number-reset` controls whether the chapter counter restarts per part. Continuous numbering is the default.

```yaml
folio:
  manuscript:
    chapter:
      number-reset: never               # default: continuous across all parts
      # other: "per-part" -- restart at 1 for each part
```

### Table-of-contents links (issue #32)

`folio.manuscript.toc.links` controls clickable internal links from visible TOC entries to their headings. It defaults to `true`. Set it to `false` for print-production systems such as KDP that reject PDF link annotations; the visible TOC and PDF document outline remain present.

```yaml
folio:
  manuscript:
    toc:
      links: true                       # default
```

### Copyright page (issue #21)

The `folio.manuscript.copyright` block renders a frontmatter copyright page (verso, page ii by convention) between the title page and the TOC. Disabled by default. Every field is optional.

```yaml
folio:
  manuscript:
    copyright:
      enabled: true
      credits:                          # free-text lines rendered as paragraphs
        - "Copyright © 2026 Author Name."
        - "Front cover image: © 1988 Photographer Name."
        - "Back cover illustration: © 2026 Illustrator Name."
      body:                             # legal boilerplate paragraphs
        - "The moral rights of the authors have been asserted."
        - "**All rights reserved.** No part of this publication may be..."
      separator: "———"
      publication:
        - "First published in Ireland in 2026"
      publisher: "Example Publisher"
      isbn: "978-0-000000-00-2"
      isbn-barcode: none                # none | render | file | render-and-file
      line-spacing: 1.4                 # multiplier (default 1.4; override to inherit body)
```

**Rendering order** (fixed):

1. `credits` lines (rendered as centred paragraphs; write the exact text including `©`, year, name)
2. `body` paragraphs (markdown-mini: `**bold**`, `*italic*`, `--` en-dash, `---` em-dash; a body entry that is only `---` / `***` / `___` renders as a scene-break line)
3. Separator glyph (centred, between top and bottom sections)
4. `#v(1fr)` --- pushes the sections below to the bottom of the page
5. Publication lines
6. `<preposition> <publisher>` (publisher bold)
7. `<isbn-label>: <isbn>` (label bold)
8. Barcode (when `isbn-barcode: render` or `render-and-file`)

Missing blocks silently collapse. The bottom section (publication onwards) is always page-bottom-aligned; the top section flows from the top.

**Defaults**:

- `credits` unset -> single default line: `Copyright © YEAR Author Name.` (year from `folio.date`, name from `folio.author`)
- `body` unset -> British preset ships Irish/UK moral-rights + all-rights-reserved + NLI/BL legal-deposit; US preset ships all-rights-reserved + Library of Congress CIP text
- `folio.date` unset -> defaults to today at config-load time so year derivation always resolves
- `skip-header: true` (default) -> no running header on copyright page
- `skip-footer: false` (default) -> page number renders in footer
- `blank-page-before: enforce-left` (default) -> lands on verso (page ii)
- `position: after-title` (default) -> between title page and TOC
- `line-spacing: 1.4` (default) -> generous publisher-typical spacing; override to `1.0` (single) or inherit body via explicit value

**ISBN barcode**:

- `none` --- no barcode (default)
- `render` --- embed EAN-13 SVG on the copyright page below the ISBN text
- `file` --- write `<output-basename>.barcode.svg` alongside the output PDF; do not embed
- `render-and-file` --- both

Invalid ISBNs (wrong length, non-numeric, or wrong EAN-13 check digit for a 13-digit input) are rejected at config-load time with a diagnostic naming the offending value.

To generate an external SVG for cover artwork, configure the local `script.yaml`:

```yaml
folio:
  manuscript:
    copyright:
      enabled: true
      isbn: "978-0-000000-00-2"
      isbn-barcode: file
```

Then render the manuscript:

```bash
folio manuscript manuscript.md manuscript.pdf
```

The command writes `manuscript.barcode.svg` beside `manuscript.pdf`. Set `isbn-barcode: render-and-file` to embed the barcode in the manuscript and write the external SVG.

### Semantic authoring of parts and chapters (issue #18)

Parts and chapters can be authored with just the semantic name. The parser derives numbers from source order; manuscript rendering then applies `chapter.number-reset`, whose default `never` produces one continuous chapter sequence across parts. The rendered heading is composed from configurable prefix, number, separator, name, and suffix.

**Source (author-facing):**

```markdown
# Unbelieved

## Character

The hedges were higher than he remembered.
```

**Config (presentation):**

```yaml
folio:
  manuscript:
    part:
      prefix: "PART "               # "PART "
      number-format: "1"            # "1" arabic (default), "I" roman-upper, "i" roman-lower
      separator: ": "               # ": "
      suffix: ""                    # trailing suffix (rare)
      show-name: true               # default true
      show-number: true             # default false; set true to include the number
      name-case: "as-written"       # "as-written" (default), "upper", "lower", "title"
      case-transform: "as-written"  # applies to the composed heading as a whole
      explicit-numbering: "derived" # "derived" (default) or "source"
    chapter:
      # same shape as part
      prefix: "Chapter "
      show-number: true
      number-reset: never            # default; use per-part to restart each part
      number-format: "1"             # chapter only: 1, I, i
      # number-format: "I.1"         # part.chapter; each segment accepts 1, I, i
```

Rendered outcomes for the source above with the config above:

- Part body heading: `PART 1: Unbelieved`
- Chapter body heading: `Chapter 1: Character`

**Backward compatibility:** existing manuscripts that write `# PART ONE: UNBELIEVED` or `## Chapter 12: The Watch` continue to render sensibly. The parser detects the `Part <token>` / `Chapter <token>` prefix pattern and strips it, capturing the source number in `SourceNumber` (used only when `explicit-numbering: source` is set) and the remainder as the semantic name. If the source heading is plain (no prefix, e.g. `# Unbelieved`), it's used verbatim as the semantic name.

**Numbering (`explicit-numbering`):**

- `derived` (default): part and chapter numbers come from source order (safe if you renumber chapters by moving files around).
- `source`: use the number literal from the source heading. Useful when a manuscript deliberately skips numbers (Chapter 7 is called "Chapter 7" for in-fiction reasons).

**Chapter number formats:**

- One segment formats only the chapter counter: `1`, `I`, or `i`.
- Two segments format `part.chapter`: `I.1`, `1.1`, `1.I`, and the other combinations of `1`, `I`, and `i`.
- `number-reset: never` keeps the chapter segment continuous across parts; `per-part` restarts that segment at 1.
- Unsupported patterns are rejected instead of silently falling back to Arabic.

**Name case vs case-transform (AC18.5):**

- `case-transform` applies to the *composed* heading (`prefix + number + separator + name + suffix`) as a whole. Set to `upper` to render `PART 1: UNBELIEVED`.
- `name-case` applies only to the name segment. Values: `""` (as-written, default), `"upper"`, `"lower"`, `"title"`. Set `name-case: "title"` to auto-capitalize a source like `# the watch` into `Part 1: The Watch`.

### Skipping running header or footer on part / chapter pages

`folio.manuscript.part.skip-header`, `part.skip-footer`, `chapter.skip-header`, and `chapter.skip-footer` (all default `false`) suppress the corresponding running header or footer on any page that renders the corresponding heading. Combined with a heading that has `page-break-before: true`, this cleanly hides the header/footer on the dedicated part or chapter page; combined with a heading that shares a page with a chapter, this hides the header/footer for that shared page.

### Title-page item alignment

`folio.manuscript.title-page.<item>.align` accepts `left`, `center`, `right`, or a compound `V-H` value. V is `top`, `center` or `bottom`; H is `left`, `center` or `right`. Bare compass values use vertical centre. These values position an item independently. Supported items are `title`, `subtitle`, `author`, `date`, `wordcount`, `version` and `contact`.

The six items other than `contact` also accept **`group`**, their British base default:

- Title, subtitle and author remain in the title group, retaining their order and spacing. `title-block-align` controls that group's placement.
- Version, word count and date form the bottom row, in that order. Version is left-aligned and date is right-aligned; `footer-align` controls word-count alignment in the middle column. British uses the page footer; US uses the bottom of the title-page content area.
- US explicitly sets `include-date: false` and `include-version: false`. Enabling either includes it in the row, or at its configured independent placement.
- Contact defaults to `top-left` and does not accept `group`.

The earlier description of `footer-align` as a US-only grid setting was incorrect. It controls the middle column of the British and US title-page footer groups.

**Migration:** replace the former `align: ""` on those six items with `align: group`, or omit the override to inherit the base. Replacing it with `center` instead changes the item to independent placement and can overlap other items.

All shipped alignment fields are validated **after configuration layers merge**, including fields for disabled elements and other output modes. Empty strings, whitespace-only values, null, non-string values and unknown keywords are errors naming the field, rejected value and accepted values. An omitted override inherits the lower layer. Explicit invalid values never request a fallback.

Script alignments and manuscript part/chapter/copyright horizontal alignments accept `left`, `center` or `right`. Manuscript `part.vertical-align` accepts `top`, `center` or `bottom`, retaining `middle` and `horizon` as centre aliases. Title-page `title-block-align` and `footer-align` accept the compass and V-H forms above, but not `group`. Running header/footer alignment retains its [page-pair syntax](#book-layout-page-pair-alignment). Surrounding whitespace is trimmed; vertical part keywords also retain case-insensitive handling.

Empty or null optional metadata such as `publisher` and `isbn`, and empty prefix/suffix strings, remain valid. Commented YAML examples are not active configuration.

`folio.manuscript.toc.enabled` defaults to `true`. Set it to `false` to suppress the generated table of contents.

`folio.manuscript.toc.line-spacing` uses the same positive finite multiplier definition as body `line-spacing`, calculated from the TOC's own font and size. It applies to wrapped lines within an entry and to adjacent entries: `1.0` is native single spacing, `1.5` is one-and-a-half spacing, and `2.0` is double spacing. Quoted numbers are accepted. British and US use a default of `1.15`; TOC spacing remains independent of the body setting. Values below one compress the interval and can overlap.

**TOC migration:** replace length values such as `1.15em` or `2em` with the intended numeric multiplier. Removing the unit changes the spacing definition and can change pagination. Lengths, empty strings, zero, negative and non-finite values fail before output, even when the TOC is disabled. `toc.part-gap-before` remains a separate length setting for space before part entries.

`folio.manuscript.toc.continuation-padding-before` reserves space above entries on every table-of-contents page. The Contents heading occupies that band on page one, and continuation pages leave it blank, keeping entry lists vertically aligned. The British default is `15mm`, inherited by the US preset.

US manuscript style is selected with `folio.manuscript.style: us` or `folio.style: us`, or with `folio manuscript --style us ...`. The unified US overlay is layered on top of the shared British base and does not change the page size to `us-letter`; page size changes require explicit user config.

Manuscript metadata supports `title`, `subtitle`, `author`, `attribution`, `date`, `version`, `wordcount`, `contact-name`, `address`, `phone`, `email`, and `website`. `wordcount` is display text, not a numeric field; values such as `about 90,000 words`, `approx 100k words`, and `20.000 mots` render as entered.

`folio.manuscript.date-format` controls title-page date rendering for ISO frontmatter dates using Go date layouts. British defaults to `2 January 2006`; US overrides default to `January 2, 2006`.

`folio.manuscript.toc.part-gap-before` controls extra vertical space before part entries in the table of contents. The default is `0.5em`.

`folio.manuscript.toc.part-bold` controls whether part entries are bold in the table of contents. The default is `true`.

### Yapper namespace (`yapper:`)

Anything beneath a top-level `yapper:` block is exclusively Yapper configuration and is ignored by First Folio. First Folio does not define or document Yapper's child keys; see the [Yapper documentation](https://github.com/tigger-developer/yapper) for that schema.

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
  font:
    family: EB Garamond
    size: 11pt
  margin: 25mm
  page: a4
  positioning:
    speech:
      dialogue:
        wrap-indent: 5em
```

## Migration from the retired font schema

The 0.9 font configuration is intentionally breaking. Move each old font value into the corresponding role's `font` block. Replace role-local `bold` and `italic` booleans with `weight` and `style`.

```yaml
# Retired
folio:
  font: EB Garamond
  font-size: 11pt
  positioning:
    stage-direction:
      italic: true

# Current
folio:
  font:
    family: EB Garamond
    size: 11pt
  positioning:
    stage-direction:
      font:
        style: italic
```

The current configuration may contain a partial font block because its remaining properties come from the same role in the British base. Retired keys are rejected rather than translated silently.

### Manuscript hyphenation

`folio.manuscript.hyphenation` controls automatic hyphenation in body paragraphs and quoted blocks:

- `false`: keep words whole.
- `true`: allow hyphenation with either justified or ragged text.
- `auto`: allow hyphenation when `justify: true`. This is the British default, inherited by US.

Quoted `"true"` and `"false"` are also accepted. Empty, null, numeric and other values produce a configuration error. Source frontmatter and local YAML use the normal deep-merge precedence. Inline code, standalone code paragraphs and fenced code remain unhyphenated.

```yaml
folio:
  manuscript:
    hyphenation: true
    font:
      stretch: 130%
```

At non-100% stretch, First Folio supplies English discretionary break points from bundled patterns. No download is needed. Words containing non-ASCII letters and tokens longer than 128 letters remain whole. Break points that would alter a ligature or kerned pair are omitted, preserving unbroken word widths. Typst chooses which permitted breaks to use; enabling hyphenation does not force every line to end in a hyphen. The inserted discretionary hyphen uses its native glyph width, while the source letters retain the configured geometric stretch. These restrictions apply to stretched prose; normal-width prose uses Typst's native hyphenation.

## Changelog

- 0.25 (2026-10-08): Document automatic cast-name capitalisation, aliases and hidden-table lookup.
- 0.24 (2026-10-07): Added compact script first-page titles for page-limited submissions and default-on title/first-page running-matter suppression; documented untitled scripts, mode isolation, validation and numbering.

- 0.23 (2026-10-07): Promoted running headers and footers to shared Folio settings for scripts and manuscripts; retained layered manuscript-local compatibility with deprecation diagnostics and excluded letters.

- 0.22 (2026-09-26): Added manuscript hyphenation controls for body and quoted text, including geometric stretch, with code remaining unhyphenated.


- 0.21 (2026-09-26): Stretched chapter openings preserve subsequent paragraph breaks, first-line indentation and body stretch.

- 0.20 (2026-09-26): Body and quoted-block stretch scale glyph widths proportionally; documented independent role ratios and word-level line breaking.

- 0.19 (2026-09-26): Quoted blocks use exact configured indentation and additive paragraph clearance, with body-font em units.

- 0.18 (2026-09-25): Restored paragraph separation around code blocks; vertical clearances are additional and use the body font size for em units.

- 0.17 (2026-09-25): TOC line spacing uses font-relative multipliers consistently with body text; length-valued TOC spacing requires migration.

- 0.16 (2026-09-25): Markdown manuscript configuration frontmatter now merges below local layout files and above global files, using recursive same-property precedence and existing validation.


- 0.15 (2026-09-24): Alignment defaults use explicit group placement; resolved invalid alignments fail before output. US title-page date/version omission is explicit and inclusion flags are honoured. Historical empty alignment values require migration.

- 0.13 (2026-09-17): Extended proportional monospace stretch to fenced code blocks, so one `mono.font` configuration renders identically wherever monospace appears, and recorded that `code-block` accepts no font block.
- 0.12 (2026-09-12): Corrected inline monospace stretch to proportional horizontal scaling, preserving glyph height, paragraph wrapping and neighbouring text.
- 0.11 (2026-09-12): Recorded the exclusion of heading code and its TOC appearances from the font-rendering repair; paragraph code and ordinary heading typography remain in scope.
- 0.14 (2026-09-24): Body line spacing now accepts only positive finite multipliers of the selected font's normal interval; paragraph spacing retains additive lengths. This supersedes length acceptance and font-size-based numeric intervals.
- 0.10 (2026-09-12): Clarified font-width selection and corrected application of configured monospace and manuscript heading fonts, without changing configuration keys or defaults.
- 0.9 (2026-09-07): Unified every typography role under a six-property font block and replaced split runtime presets with one British base plus limited US and screenplay overlays.
- 0.8 (2026-08-10): Documented nearest-ancestor and style-sibling discovery, the complete manuscript key inventory, justification, and metric custom-page examples.
- 0.7 (2026-08-09): Restored linked manuscript TOCs by default, documented the annotation-free override, and added continuous and part-qualified chapter numbering.
- 0.6 (2026-08-07): Added configurable reserved space above continued table-of-contents entries.
- 0.5 (2026-08-07): Clarified manuscript body and table-of-contents line-spacing behaviour.
- 0.4 (2026-08-07): Added the `[total-pages]` manuscript header/footer placeholder.
- 0.3 (2026-07-26): Added the external ISBN barcode SVG workflow.
