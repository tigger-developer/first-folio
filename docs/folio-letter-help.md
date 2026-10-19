Usage: folio letter <source.org> [options]

Generate letters from :letter: tagged sections in an org file.
One document is generated per recipient; PDF is the default format.

Options:
  --format FORMAT   Output format: pdf, docx or odt; default pdf
  --to RECIPIENT     Generate only for matching recipients, using a regex substring
  --dir DIR          Output directory, default same as source file
  --prefix PREFIX    Output filename prefix, default "letter"
  -h, --help         Show this help message
  --version          Show application version

Placeholders:
  Any [keyword] in the letter body is replaced from two sources:

  1. Recipient tags, H4 under :to:, highest priority:
       **** The Abbey                :org:       -> [org]
       **** Dave                     :moniker:   -> [moniker]
       **** Friday                   :deadline:  -> [deadline]
     Any tag name works. Define custom tags as needed.

  2. Document front matter fallback:
       #+TITLE:    -> [title]      #+AUTHOR:   -> [author]
       #+SUBTITLE: -> [subtitle]   #+DATE:     -> [date]
       #+VERSION:  -> [version]

  Unresolved placeholders pass through as [name].

Config:
  ~/.config/first-folio/script.yaml plus the nearest script.yaml found by
  walking from the source directory towards HOME.

Editable office output:
  DOCX is parsed and generated natively. ODT uses the configured Pandoc writer
  after native parsing. Neither uses Typst. Both follow the PDF YAML cascade and use the existing
  folio.letter font, margin, spacing and subject settings. Letters do not inherit
  manuscript headers or footers. Available permitted full fonts are embedded.
  Org emphasis and =verbatim= spans remain editable; use verbatim for literal URLs.
  Duplicate output names and source/output aliases fail before publication.
  Each recipient file is validated and published separately. If a later file
  fails, the command reports the recipient files already written.
