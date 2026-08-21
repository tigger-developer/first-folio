<!-- Version: 0.8 | Last updated: 2026-08-10 -->

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
| 4 | Global `~/.config/first-folio/script-<style>.yaml` |
| 5 | Global `~/.config/first-folio/script.yaml` |
| 6 | Selected built-in style override |
| 7 (lowest) | British built-in base preset |

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

Manuscript mode additionally accepts top-level `attribution`, `author-attribution`, `wordcount`, `contact-name`, `address`, `phone`, `email`, and `website`. These override corresponding manuscript frontmatter values and are display strings, including `wordcount`; First Folio does not impose numeric or language-specific formatting.

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
| `heading-font` | string | inherits `font` | Default heading font family |
| `heading-font-size` | string | inherits `font-size` | Default heading font size |
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

The effective CLI layout overrides are `--font`, `--font-size`, `--margin`, and `--page`; `--style` selects the preset layer. Other layout changes belong in `script.yaml`. See `folio convert --help` for the public CLI surface.

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
| `letter-spacing` | Typst length | `0em` | inherited (`0em`) |
| `justify` | bool | `true` | inherited (`true`) |
| `widow-orphan-control` | bool | `true` | inherited (`true`) |
| `paragraph-indent` | string | `10mm` | `12.7mm` |
| `paragraph-spacing` | string | `0` | `0` |

`folio.manuscript.line-spacing` accepts a baseline multiplier or an explicit Typst length. With a multiplier, `1.0` is single-spaced, `1.5` is one-and-a-half-spaced, and `2.0` is double-spaced. A length such as `2em` is passed through without adding another unit. `folio.manuscript.letter-spacing` maps to Typst font tracking and defaults to `0em`; positive and negative Typst lengths are accepted. `folio.manuscript.paragraph-spacing` is additional space between paragraphs; `0` preserves the selected line interval across paragraph boundaries without adding a separate paragraph gap. `folio.manuscript.justify` controls body-text justification.

`folio.manuscript.widow-orphan-control` defaults to `true`, preventing a single paragraph line from being stranded at the bottom or top of a page. Set it to `false` to allow paragraphs to split freely at page boundaries. This setting does not keep whole paragraphs together.

`folio.manuscript.page-header.content-padding-after` controls the clearance between the running header and the manuscript body on every running-header page. It does not affect the title page or table of contents.

#### Complete manuscript key inventory

The built-in files [british-manuscript.yaml](../presets/british-manuscript.yaml) and [us-overrides-manuscript.yaml](../presets/us-overrides-manuscript.yaml) are the canonical default values. Every accepted `folio.manuscript` key is listed below; omitted child typography inherits from the manuscript heading or body typography as described above.

| Group | Accepted keys |
|---|---|
| Core | `style`, `page`, `margin`, `gutter`, `line-spacing`, `justify`, `widow-orphan-control`, `paragraph-indent`, `paragraph-spacing` |
| Body typography | `font`, `font-size`, `font-weight`, `letter-spacing` |
| Heading typography | `heading-font`, `heading-font-size`, `heading-font-weight` |
| Monospace typography | `mono-font`, `mono-font-size`, `mono-font-weight` |
| Title typography | `title-font`, `title-font-size`, `title-font-weight` |
| Subtitle typography | `subtitle-font`, `subtitle-font-size`, `subtitle-font-weight`, `subtitle-font-style` |
| Author typography | `author-font`, `author-font-size`, `author-font-weight`, `attribution`, `author-attribution` |
| Date typography | `date-font`, `date-font-size`, `date-font-weight`, `date-format` |
| Version typography | `version-font`, `version-font-size`, `version-font-weight` |
| Word-count typography | `wordcount-font`, `wordcount-font-size`, `wordcount-font-weight` |
| Contact typography | `contact-font`, `contact-font-size`, `contact-font-weight` |

| Nested block | Accepted child keys |
|---|---|
| `page-header` | `enabled`, `font`, `font-size`, `font-weight`, `font-style`, `letter-spacing`, `format`, `alt-format`, `frontmatter-format`, `alt-frontmatter-format`, `align`, `distance-from-edge`, `content-padding-after` |
| `page-footer` | `enabled`, `font`, `font-size`, `font-weight`, `font-style`, `letter-spacing`, `format`, `alt-format`, `frontmatter-format`, `alt-frontmatter-format`, `align`, `distance-from-edge`, `content-padding-after` |
| `toc` | `enabled`, `links`, `title`, `font`, `font-size`, `font-weight`, `heading-font`, `heading-font-size`, `heading-font-weight`, `include-parts`, `include-chapters`, `include-sections`, `dot-leaders`, `page-numbers`, `page-break-before`, `blank-page-before`, `blank-page-after`, `line-spacing`, `part-gap-before`, `continuation-padding-before`, `part-bold` |
| `title-page` | `enabled`, `page-number`, `include-title`, `include-subtitle`, `include-author`, `include-date`, `include-wordcount`, `include-contact-name`, `include-address`, `include-phone`, `include-email`, `include-website`, `include-version`, `title-block-align`, `footer-align` |
| `title-page.<item>` | `align`, where `<item>` is `title`, `subtitle`, `author`, `date`, `wordcount`, `version`, or `contact` |
| `scene-break` | `marker` |
| `list`, `table`, `code-block` | `space-before`, `space-after` |
| `page-numbering` | `frontmatter-format`, `body-format`, `body-reset` |

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
| Typography | `font`, `font-size`, `heading-font-weight`, `line-spacing`, `block-spacing` |

### Page-header format placeholders

`folio.manuscript.page-header.format` and `folio.manuscript.page-footer.format` accept the following placeholders, substituted at render time:

- `[author]` -- the manuscript author
- `[title]` -- the manuscript title
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
  manuscript:
    page-footer:
      format: "[page]/[total-pages]"
```

`[total-pages]` always reports the final physical page count. Unlike `[page]`, it is not affected by `page-numbering.body-reset` or frontmatter/body numbering formats.

#### `alt-format` for facing-page layouts

`page-header.alt-format` and `page-footer.alt-format` (issue #18 AC18.6) are optional companion format strings. When set, `format` renders on left (verso, even) pages and `alt-format` renders on right (recto, odd) pages. Common use: put the page number on the *outer* edge of the book on both pages by pairing the two format strings mirror-image.

```yaml
folio:
  manuscript:
    page-header:
      format:     "[page] • [chapter] • [author]"     # verso (left) -- [page] on outer
      alt-format: "[author] • [chapter] • [page]"     # recto (right) -- [page] on outer
```

When `alt-format` is unset, `format` renders on every page (unchanged from AC15.1).

### Page-footer block

`folio.manuscript.page-footer` mirrors the fields of `folio.manuscript.page-header`. Typography fields (`font`, `font-size`, `font-weight`, `font-style`) inherit from `page-header` when unset. `letter-spacing` is independent: both header and footer inherit `folio.manuscript.letter-spacing` unless individually overridden. Default: enabled with a centred `[page]` number, `distance-from-edge` and `content-padding-after` matching `page-header`. Set `page-footer.enabled: false` to omit the running footer.

Both `page-header` and `page-footer` accept `font-style` alongside `font`, `font-size`, and `font-weight`. Accepted values are `regular` (default), `italic`, and `oblique`. When unset, no `style:` argument is emitted, preserving the default upright rendering. Their `letter-spacing` values map to Typst tracking and accept values such as `0.05em` or `-0.01em`.

### Frontmatter-format (issue #24)

`page-header` and `page-footer` each accept `frontmatter-format` and `alt-frontmatter-format` that apply on frontmatter pages (title, copyright, TOC, and any page before the first part or chapter). Body pages use the normal `format` / `alt-format` pair.

- **Unset** (key absent from YAML) -> frontmatter pages use `format` / `alt-format` (backwards-compatible, no change).
- **Set to non-empty string** -> that string renders on frontmatter pages.
- **Set to empty string `""`** -> frontmatter pages render blank (no header or footer text).
- **`alt-frontmatter-format` set alongside `frontmatter-format`** -> verso frontmatter uses `frontmatter-format`, recto frontmatter uses `alt-frontmatter-format` (same verso/recto pairing as `format` / `alt-format`).

The frontmatter/body boundary is defined as: any page before the first part or chapter block is frontmatter; from the first part or chapter onward is body. Matches standard publishing convention.

Example --- suppress the running header on frontmatter but keep body headers:

```yaml
folio:
  manuscript:
    page-header:
      format: "[title] • [chapter] • [author]"
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

`folio.manuscript.title-page.<item>.align` accepts either a compass keyword (`left`, `center`, `right`) or a compound `V-H` value where V is in `{top, center, bottom}` and H is in `{left, center, right}` (for example `top-left`, `bottom-center`). Items placed with a per-item align hug the manuscript margin at the named corner. Supported items are `title`, `subtitle`, `author`, `date`, `wordcount`, `version`, and `contact`.

Legacy `folio.manuscript.title-page.title-block-align` continues to control the title/subtitle/author group when no per-item align is set; `footer-align` continues to control the US grid footer (version/word-count/date row) on the title page.

Unknown alignment values (e.g. `middle-middle`, `bottom-diagonal`) are rejected at config load with a diagnostic naming the offending value.

`folio.manuscript.toc.enabled` defaults to `true`. Set it to `false` to suppress the generated table of contents.

`folio.manuscript.toc.line-spacing` controls the baseline interval between table-of-contents items, including one-line entries. The British default is `1.15em`; for example, `2em` renders item baselines approximately two font-heights apart.

`folio.manuscript.toc.continuation-padding-before` reserves space above entries on every table-of-contents page. The Contents heading occupies that band on page one, and continuation pages leave it blank, keeping entry lists vertically aligned. The British default is `15mm`, inherited by the US preset.

US manuscript style is selected with `folio.manuscript.style: us` or `folio.style: us`, or with `folio manuscript --style us ...`. The US override is layered on top of the British manuscript preset and does not change the page size to `us-letter`; page size changes require explicit user config.

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
  font: EB Garamond
  font-size: 11pt
  margin: 25mm
  page: a4
  positioning:
    speech:
      dialogue:
        wrap-indent: 5em
```

## Changelog

- 0.8 (2026-08-10): Documented nearest-ancestor and style-sibling discovery, the complete manuscript key inventory, justification, and metric custom-page examples.
- 0.7 (2026-08-09): Restored linked manuscript TOCs by default, documented the annotation-free override, and added continuous and part-qualified chapter numbering.
- 0.6 (2026-08-07): Added configurable reserved space above continued table-of-contents entries.
- 0.5 (2026-08-07): Clarified manuscript body and table-of-contents line-spacing behaviour.
- 0.4 (2026-08-07): Added the `[total-pages]` manuscript header/footer placeholder.
- 0.3 (2026-07-26): Added the external ISBN barcode SVG workflow.
