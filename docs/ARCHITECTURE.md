---
title: Architecture
version: "0.29"
last-updated: 2026-11-07
---

# Architecture

First Folio is a CLI-first Go application. One `folio` executable owns stage-play conversion, cover-letter generation, and prose-manuscript rendering.

## Runtime

`cmd/folio` is the sole product entry point. It delegates process-independent command handling to `internal/app`, which returns an exit status and writes through injected streams.

The subcommands are:

- `folio convert`: parses Org, Markdown, or Fountain stage plays into typed events and emits Org, Markdown, Fountain, Typst, PDF, DOCX, or ODT.
- `folio letter`: parses Org `:letter:` sections and renders recipient-specific PDF, DOCX or ODT files.
- `folio manuscript`: parses Markdown or Org prose manuscripts and renders Typst, PDF, DOCX or ODT.

Typst and Pandoc remain explicit external tools where their documented features require them. No command invokes Perl or a project-owned shell script.

## Packages

| Path | Responsibility |
|---|---|
| `cmd/folio/` | Process entry point for the single executable |
| `internal/app/` | Public CLI dispatch, conversion orchestration, script rendering, and process-level integration tests |
| `internal/config/` | Shared embedded-preset and YAML configuration loading |
| `internal/play/` | Typed stage-play model plus Org, Markdown, and Fountain parsers/emitters |
| `internal/source/` | Shared Org/Markdown source recognition, literal/inline parsing and bounded structured prose |
| `internal/letter/` | Cover-letter model, Org parser, Typst renderer, and compiler |
| `internal/manuscript/` | Manuscript input, parser, canonicalization, rendering, and compilation |
| `internal/office/` | Shared native DOCX and configured ODT writers, bounded manuscript interchange, layout, packages and fonts |
| `templates/` | File-backed Typst layouts for scripts, letters, and manuscripts |
| `presets/` | British base and explicit style override YAML |
| `cmd/update-homebrew/` | Checked release-formula update and publication tooling |

The root `assets.go` embeds help documents, presets, Typst templates, and licensed font assets. The installed binary therefore does not discover a project root, depend on its working directory, or require a root environment variable.

`internal/typst.Compile` supplies the bundled Iosevka Publish fonts to script,
manuscript and letter PDF compilation. Each invocation extracts the embedded
faces into a private temporary directory, passes it through Typst's
`--font-path` alongside existing `TYPST_FONT_PATHS`, and removes it on success
or failure. It contains no style-specific behaviour. All font selection stays
in layered YAML. Exported Typst files require fonts on the compiling host.
Office exports use a shared bounded SFNT metadata resolver over bundled assets
and local font paths. A final packaging step adds complete editable font faces
to DOCX font parts or ODT font-face sources and manifest entries. It runs before
atomic publication and does not change the PDF route. Missing or restricted
fonts retain their references with warnings; see [font embedding](font-embedding.md).

The separate-line dialogue template gives its block the available width, so
configured speaker alignment operates across the dialogue area. British
same-line dialogue keeps its existing grid layout.

## Document Models

Stage plays and manuscripts intentionally retain separate semantic models.

Stage-play parsers produce typed events for metadata, acts, scenes, direct scene-introduction paragraphs, directions, speakers, dialogue, character tables, transitions, props, global intro material, and footnotes. Text emitters preserve scene introductions as semantic source content. PDF, DOCX and ODT apply `render.frontmatter` to them at the rendered-output boundary, while script PDF rendering prepares escaped template data from the event stream.

Manuscripts use metadata plus prose blocks such as parts, chapters, sections, paragraphs, lists, tables, code, blockquotes, links, scene breaks, and footnotes. Source-document images are not yet part of the manuscript contract; generated ISBN barcode SVGs are a separate copyright-page facility. The PDF/Typst route retains Org canonicalization through Markdown. Office output parses both source forms natively through `internal/source`, retaining the same documented manuscript semantics.

Letters have a smaller recipient-oriented model because their sender, recipient, substitution, and signoff semantics do not map cleanly onto either stage plays or manuscripts.

## Configuration

`internal/config` owns YAML file loading and precedence for every mode:

1. British built-in preset.
2. Selected built-in style override.
3. Global `~/.config/first-folio/script.yaml`.
4. Global style-specific YAML.
5. Selected local `script.yaml`.
6. Selected local style-specific YAML.
7. Supported source-document `folio:` and `render:` configuration.
8. CLI overrides.

Script orchestration projects Markdown YAML or dotted Org frontmatter into the
source layer before parsing and rendering. Manuscript orchestration reads the
first Markdown input's YAML source configuration through its shared frontmatter
decoder; metadata remains a separate document concern, and later input files
cannot supply another configuration layer. Org manuscripts, Fountain scripts
and letters do not currently define source-document configuration syntax. Once
loaded, the same effective configuration is supplied to every output route for
that mode, so PDF, Typst, DOCX, ODT and semantic text conversion do not acquire
format-specific precedence.

Local discovery begins at the source directory and walks upwards towards HOME. Only the nearest `script.yaml` is selected; its style-specific sibling is loaded from the same directory. For a multi-file manuscript, the first resolved input determines that starting directory.

The loader deep-merges ordinary keys and partial font blocks, validates every effective font role and every shipped alignment field against its allowlist after layer merging, and provides typed values to each renderer. Alignment validation occurs before typed decode or output; renderer alignment parsers translate accepted values into layout expressions. Font properties inherit through configuration layers at the same role path. Script frontmatter paragraphs additionally fill omitted properties from a copy of the final resolved `folio.font`; this fallback runs after layer merging and cannot mutate the body font. PDF uses a scoped text wrapper, while DOCX and ODT use a dedicated paragraph role. The exhaustive British preset is the shared lowest-precedence base for scripts, letters, and manuscripts; US and screenplay presets contain only their differences.

Running matter is configured at `folio.page-header` and `folio.page-footer` for scripts and manuscripts. Deprecated manuscript-local blocks are normalized within each configuration layer, before merging, so same-layer manuscript overrides do not bypass higher-layer shared settings. Manuscript command diagnostics report their use on stderr. Letter rendering ignores running-matter settings because letters are supplementary correspondence, not manuscript pages.

Monospace typography is configured once at `folio.mono.font` for script, letter and manuscript inline code and supported code blocks. The deprecated `folio.manuscript.mono.font` path is normalized into that shared role within each layer, with explicit top-level properties winning overlaps. Public commands report legacy use on stderr. The effective role then follows the ordinary PDF YAML cascade and is projected into Typst, DOCX and ODT renderers.

For manuscript mode, the shared loader records each external YAML file successfully loaded as a layer. The existing warning channel passes one resolved-path diagnostic per file, in cascade order, to both Typst/PDF and office command routes for stderr emission. Built-in presets, source frontmatter, CLI values, and other command modes do not produce load diagnostics. Warning collection does not change the merged configuration.

## Rendering

DOCX is a sibling output route after script parsing, configuration and filtering.
The `docx_*.go` adapter builds native WordprocessingML with `encoding/xml`,
packages it with `archive/zip`, and publishes through a same-directory staged
rename. No Pandoc or Typst process runs at any stage of script, manuscript or letter DOCX, including source parsing and dry-run. British scene tables, font
mapping, resource bounds and approximations are described in [DOCX export](format-docx.md).
The conversion entry point passes the source format into the shared office inline renderer, so Org single-asterisk bold also applies to DOCX and ODT, including footnotes. The existing Typst route retains its independent source-aware renderer.

After native DOCX generation, the shared attributed-footnote transformer places each Word comment range start before the reference paragraph's first content and replaces the original footnote-reference run with the range end and comment reference. This preserves complete sentence or line highlighting across multiple Word runs. Multiple ranges in one paragraph are nested in reference order before bounded package validation and atomic publication.

ODT is a separate sibling output route. `odt_*.go` projects the shared office
layout nodes into Pandoc JSON and generates an ODF reference containing resolved
role styles and page masters. The layout nodes currently use Word names; no
DOCX package is written or parsed in this route. Pandoc runs with fixed arguments,
a bounded output pipe and a 120-second deadline. The resulting ZIP/XML is
validated before the existing atomic publication helper runs. DOCX serialization
is unchanged. See [ODT export](format-odt.md) for compatibility and limits.

Manuscript and letter office output use separate mode adapters into
`internal/office.Prose`. Manuscripts extend the convert parser's native source
recognition into structured prose; letters compose the existing recipient model
and use the same inline parser. Convert's script events retain mode-specific
classification. The shared source node shape also supplies JSON to ODT, but
constructing or validating it requires no Pandoc process. Contiguous text stays
in one span rather than expanding into per-word nodes.
Prose layout keeps normal characters on their native baseline; paragraph boxes
own leading and the running-header edge offset. Temporary instance identities
carry bookmarks, fields and page transitions through layout. Final serialization
reuses exact formatting definitions within each semantic level/role, rewrites
style references and keeps ODT page overrides in automatic styles. Sharing
formatting never merges the instance's content, anchors or numbering.

All three modes use the existing native WordprocessingML serializer and DOCX
packager. Manuscript and letter adapters add structured prose content, while
script events retain their existing layout. ODT uses the same bounded Pandoc
writer and native reference-package generator as script export, with
mode-specific styles and page masters. Native layout adds fields, editable
frames and generated images. It validates packages before full-font embedding
and again before publication. This does not route scripts through the prose
model or round-trip DOCX into ODT.

Bounds apply to aggregate input, semantic depth/counts, subprocess time and
diagnostics, JSON, ZIP entries and expanded data. Manuscript output publishes
before its optional barcode sidecar. Letters preflight all target names and
publish individually, reporting prior successful paths on later failure.
ODT frontmatter/body number-format changes, requested resets and changed
running-band page-field structure use a new physical page. Displayed numbering
and configured page-side policy remain separate. DOCX special first-page bands
retain the documented odd-variant limitation.


Substantial Typst layouts live in real `.typ` files. Go code owns:

- typed template data;
- context-specific escaping and validated layout literals;
- event-to-layout policy;
- temporary-file lifecycle;
- direct subprocess invocation and diagnostics.

Templates own page composition. Product Go code does not contain generated-language heredocs or invoke a shell to run Typst.

Script title composition uses one shared title block for a dedicated title page or a compact heading above first-page content. `folio.title-page.enabled` selects between them; manuscript title-page settings remain independent. Script-only Boolean validation covers this switch and `skip-header`/`skip-footer`. Running-matter suppression uses the physical page counter and applies to the dedicated title page or compact page 1, without resetting numbering or removing title metadata. The renderer suppresses only the first act's explicit break in compact mode; later act breaks and natural pagination remain. See [script title configuration](config.md#compact-script-title-block) for defaults and untitled-script handling.

The renderer-neutral `folio.cover-image` model is validated by the shared
configuration loader, while W073 adopts it only in the dedicated script PDF
route. `internal/app` resolves assets from the source document, validates PNG
or JPEG headers and bounded source dimensions, optionally invokes Superscale
with literal arguments, and prepares normalized sharp PNGs and blurred JPEG
backdrops.
The script Typst template owns background cover-crop, inset geometry, localized
overlay regions and dodge composition. Dodge placement first measures and
positions each live-text group in the region bounded by the actual image edge,
then applies configured image clearance as a one-sided minimum adjustment. The
clearance does not shrink or recalculate that region. The adjustment may enter
the title-page margin and fails only at the physical page edge. Title metadata
remains live text. Script author placement defaults to the movable title group;
the optional bottom-centre position instead joins date and version in a fixed
three-slot metadata row. Dodge displacement never moves that row, and an image
clearance collision fails before atomic publication. One validated script-title
position resolver preserves named compatibility aliases. Ordinary and overlay
layouts use a literal percentage as preceding space; dodge uses the same value
as the centre anchor of each movable group before applying image clearance.
For inset backdrops, the Go preparation boundary independently derives a
page-aspect JPEG at fixed quality 85 from the original bounded source, before any optional
Superscale selection for the sharp inset. It applies the shared treatment,
centre-crops, downsamples to a 512-pixel longest-axis ceiling without upscaling
and applies the configured physical blur. Solid backdrops remain validated
colour data. Typed template data supplies the selected asset or colour, a
page-wide wash and its text treatment; the Typst template paints that full
bleed before composing the existing sharp inset and live text. Localized
overlay treatment remains the more specific text and wash rule.
Preparation and layout complete before a PDF is atomically renamed into place;
warnings fall back to the original image, while invalid configuration, assets
or measured dodge regions leave the target unchanged. Standalone Typst, DOCX,
ODT, compact script titles, manuscripts and letters reject configured covers
until their own adoption tickets are delivered.

Script, letter and manuscript Typst templates apply every configured font stretch
as geometric horizontal scaling from a normal-width face. Explicit role wrappers
preserve baselines and word-level reflow, and role state prevents an enclosing
body role from scaling nested monospace a second time. DOCX and ODT retain their
native editable text-scaling properties. For stretched manuscript prose,
`internal/manuscript` embeds English patterns
and exceptions, applies the pinned Go hyphenation library, and supplies a
document-specific dictionary of discretionary break points. The template
preserves shaping-sensitive boundaries and leaves final line and page breaking
to Typst. Generated Typst remains self-contained with no runtime download.
The shared loader validates `folio.manuscript.hyphenation`; code disables
automatic hyphenation. See [hyphenation settings and limits](config.md#manuscript-hyphenation).

## Build And Tests

`make build` compiles `dist/folio`; `make install` builds first and then links that binary into the configured installation directory. Version values are injected with Go linker flags. Homebrew builds the same command directly.

Automated coverage is Go-owned. Unit tests cover parsing, emission, configuration, escaping, and rendering. Integration tests execute both the in-process public application and a built binary outside the checkout working directory. PDF-sensitive suites invoke Typst directly and inspect supported outputs rather than source implementation text.

## Migration History

Before issue #10, conversion and letters used a Perl dispatcher, Perl parsers/emitters, embedded YAML::Tiny, and shell regression suites; manuscripts used a separately built Go helper. Issue #10 replaced that split with the single Go runtime while preserving the public CLI and accepted rendering behaviour.

## Document History

- 0.31 (2026-11-08): Clarify the group-centre anchor used by script dodge percentages.

- 0.30 (2026-11-07): Record the fixed-quality JPEG boundary for generated blurred image backdrops.

- 0.29 (2026-11-07): Add bounded original-source image and solid-colour backdrops to the inset script PDF composition path.

- 0.28 (2026-11-07): Place supported source-document configuration above local YAML and feed one resolved cascade to every output route.

- 0.27 (2026-11-07): Honour literal script title-position percentages across PDF title-page layouts.

- 0.26 (2026-11-07): Add an anchored three-slot script title metadata row with optional bottom-centre author.

- 0.25 (2026-11-07): Use the physical page edge as the boundary for post-placement dodge clearance.

- 0.24 (2026-11-07): Add post-placement dodge clearance without changing relative-position regions.

- 0.23 (2026-11-07): Add the renderer-neutral cover-image plan and its dedicated script PDF adoption boundary.

- 0.22 (2026-10-24): Add scene-introduction semantics and target-scoped rendered frontmatter filtering.
- 0.21 (2026-10-24): Share monospace typography across modes and define geometric stretch as the application-wide PDF contract.

- 0.20 (2026-10-19): Separate editable natural baselines and shared formatting from per-instance layout.

- 0.19 (2026-10-19): Share native source recognition across convert and office prose, retaining distinct mode semantics and PDF configuration cascades.

- 2026-10-19: Extend the shared native DOCX and configured ODT writers to manuscript and letter adapters.

- 0.17 (2026-10-18): Record accepted native office pagination boundaries.

- 0.16 (2026-10-16): Describe prose office adapters and publication boundaries.

- 0.15 (2026-10-16): Record manuscript YAML load diagnostics and their stderr ownership.
- 0.14 (2026-10-12): Carry source format into shared DOCX/ODT emphasis handling.

- 0.13 (2026-10-12): Documented independent script frontmatter paragraph fonts and body-font fallback.

- 0.12 (2026-10-12): Record shared font discovery and native office font packaging.

- 0.11 (2026-10-12): Record bundled font lifecycle and configured dialogue alignment.

- 0.9 (2026-10-11): Document the native DOCX sibling writer and atomic publication.

- 0.8 (2026-10-07): Documented shared script title composition, compact-mode act-break policy and first-page running-matter suppression.

- 0.7 (2026-10-07): Shared running-matter configuration for scripts and manuscripts, with layered compatibility for manuscript-local keys and exclusion of letters.

- 0.6 (2026-09-27): Documented manuscript geometric stretch and offline hyphenation responsibilities.

- 0.5 (2026-09-25): Added the Markdown manuscript source layer between global and local configuration.


- 0.4 (2026-09-24): Clarified shared-loader ownership of resolved alignment validation.
- 0.3 (2026-08-10): Previous architecture baseline.

- 0.10 (2026-10-11): Add ODT JSON projection and generated ODF reference.
