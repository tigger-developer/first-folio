---
title: Fonts in Editable Office Exports
version: "1.1"
last-updated: 2026-10-16
---

# Fonts in editable office exports

DOCX and ODT conversions embed complete font files for the families referenced
by the generated styles. Available regular, bold, italic and bold italic faces
are included, allowing text to be added without a glyph subset restriction.
Font selection remains in YAML; no embedding option is required.

This also applies to manuscript and letter DOCX/ODT output. Their font roles
remain `folio.manuscript.*.font` and `folio.letter.font`; script role settings
do not replace those mode-specific roles. An office reader may still substitute
a face despite its presence in the package. The rendered prose regression
checks the PDF's actual font names separately from embedded-file assertions.

The bundled Iosevka Publish and Iosevka Publish Extended families are available
without a host installation. Other families are discovered in `TYPST_FONT_PATHS`
and standard user/system font directories. On Linux, user directories include
`XDG_DATA_HOME/fonts` (defaulting to `~/.local/share/fonts`) and `~/.fonts`.
The bundled faces take precedence, followed by explicit paths and installed
fonts. Nothing is downloaded, built or installed during conversion.

## Supported files and permissions

Embedding supports standalone `.ttf` and `.otf` files with Unicode family names
and regular or bold weight metadata. Font collections, variable fonts and other
weights are not extracted or instantiated. The resolver checks OpenType OS/2
embedding permissions: installable and editable faces are accepted; restricted,
preview/print-only and bitmap-only faces are omitted. Full embedding honours
the no-subsetting flag.

Missing, unsupported or restricted families produce a warning on stderr while
retaining their configured font names. Conversion continues; a reading application
may then use an installed font or substitute another. Embedding permission flags
do not replace the font licence. The bundled customized Iosevka faces retain
their [SIL Open Font Licence](../fonts/iosevka-publish/LICENSE.md).

## Prose weight fidelity

Family selection and percentage stretch are independent of font weight. Current
DOCX prose styles express weight through regular/bold flags. ODT prose styles
retain thin through black as numeric weights in steps of 100; explicit generated
heading bold overrides take precedence. Intermediate numeric values, such as
643, still use the regular/bold fallback. These are accepted conversion limits; they do not imply that every configured
weight is faithfully reproduced.

Embedding still includes supported regular/bold/italic faces only. A numeric
ODT style therefore does not prove that a distinct light or semibold face was
embedded or selected by the reader. The ordinary-family rendering checks do
not establish that separate weight claim.

## Limits and portability

Each font is limited to 32 MiB, with 64 MiB of font data per document. Discovery
visits at most 10,000 directory entries and reads bounded font metadata before
loading selected files. Exports include at most 256 ZIP members and remain
within the existing 128 MiB compressed package limit and 256 MiB expanded limit.
Font files increase document size, even for short scripts.

DOCX stores obfuscated OpenType font parts with a font table and full-font save
settings. ODT stores font files with manifest entries and font-face source
references after Pandoc writes the configured document. Readers must support
embedded fonts; pagination still depends on the reading application. LibreOffice
rendering has been checked with an uninstalled test family in both formats.
Microsoft Word editing and save/reopen behaviour remain user validation.

Standalone `.typ` files still require the selected fonts on the compiling host.
PDF generation continues through the existing Typst font path.

## History

- 1.1 (2026-10-16): Document manuscript and letter office output.

- 1.0 (2026-10-12): Add full-face embedding to DOCX and ODT exports.
