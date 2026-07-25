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
  --version [VERSION]         Show command version, or override manuscript version when VALUE is supplied
  --wordcount WORDS           Override manuscript word count
  --contact-name NAME         Override title-page contact name
  --dry-run                   Validate inputs and print the render plan
  -h, --help                  Show this help message

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
