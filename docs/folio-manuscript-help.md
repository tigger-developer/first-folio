Usage: folio manuscript <input>... <target> [options]

Render Markdown or org-mode prose manuscript chapters to .typ or .pdf.

Options:
  --style british|us          Manuscript preset, default british
  --title TITLE               Override manuscript title
  --subtitle SUBTITLE         Override manuscript subtitle
  --author AUTHOR             Override author name
  --attribution TEXT          Prefix author name, for example by
  --author-attribution TEXT   Compatibility alias for --attribution
  --date DATE                 Override manuscript date
  --version                   Show application version
  --wordcount WORDS           Override manuscript word count
  --contact-name NAME         Override title-page contact name
  --dry-run                   Validate inputs and print the render plan
  -h, --help                  Show this help message

Set manuscript version metadata in source frontmatter or script.yaml.

Config:
  ~/.config/first-folio/script.yaml plus the nearest script.yaml found by
  walking from the first resolved input's directory towards HOME.

Source configuration:
  The first Markdown input may provide folio: and render: frontmatter mappings.
  Properties merge recursively: presets < global files < source < local files
  < explicit CLI options. A local font size preserves the source font family.
  Later input files cannot supply another configuration layer.

Hyphenation:
  Set folio.manuscript.hyphenation to auto (follow justify), true (allow),
  or false (disable). Applies to body and quotes, including stretched text;
  code stays unhyphenated. Stretched prose uses ASCII English break points,
  preserving ligatures and kerning; inserted hyphens keep native width.
  See the configuration reference for the full limits.

External ISBN barcode SVG:
  Configure the copyright block in the local script.yaml:

    folio:
      manuscript:
        copyright:
          enabled: true
          isbn: "978-0-000000-00-2"
          isbn-barcode: file

  Then render the manuscript:

    folio manuscript manuscript.md manuscript.pdf

  This writes manuscript.barcode.svg beside manuscript.pdf.
  Use render-and-file instead of file to embed the barcode in the
  manuscript and write the external SVG.
