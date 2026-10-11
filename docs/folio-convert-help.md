Usage: folio convert <source> [target] [options]

Convert a play from one format to another.
Format is deduced from file extensions (.org, .md, .fountain, .ftn, .typ, .pdf, .docx).

When no target file is given, output goes to stdout.
Use --to to specify the output format for stdout mode.

Options:
  --to FORMAT              Output format (org, md, fountain, pdf, docx)
  --style STYLE            Layout style: british (default), us, or screenplay
  --force                  Force binary output to terminal

Layout options for PDF/Typst and DOCX output:
  --font FONT              Body font family
  --font-size SIZE         Body font size
  --margin SIZE            Page margins
  --page SIZE              Page size (a4, letter, etc.)
  -h, --help               Show this help message
  --version                Show application version

DOCX is export-only and needs no external writer. British scenes use continuous
tables with full-width stage-direction rows. Title positioning and running matter
are approximate; British vertical spacing is compacted. ODT is not supported.
DOCX currently treats single-asterisk spans as italic, including Org input;
the Org-aware bold rule applies to PDF/Typst. See docs/format-docx.md for details.

Config: ~/.config/first-folio/script.yaml plus the nearest script.yaml found by
walking from the source directory towards HOME. Style-specific sibling files
(for example script-us.yaml) are applied after their corresponding base file.

Character names from Org/Markdown cast tables are automatically capitalized in
stage directions, never dialogue. Name cells accept aliases: ALISON (NURSE ALISON)
or MARGARET (MAGS, MARG.). Hiding the cast table does not disable lookup.
PDF/Typst show primary cast names only; Org/Markdown preserve alias declarations.
Fountain input has no structured cast lookup.

Running headers and footers: folio.page-header and folio.page-footer in script.yaml.
Each supports enabled, font, format, alignment, distance-from-edge and
content-padding-after. Shared placeholders: [title], [author], [page], [total-pages].
Script title pages retain their separate layout; running matter is skipped by default.

Page-limited submissions: set folio.title-page.enabled: false to put title,
subtitle, author, date and version at the top of the first content page instead
of a separate title page. The first act follows without a forced page break;
later act breaks and running headers/footers retain their configured behaviour.
folio.title-page.skip-header and skip-footer default to true: suppress running
matter on the separate title page, or page 1 when using a compact title. Set
either to false to allow its enabled shared header/footer on that first page.
These switches hide running text, not title date/version metadata or reserved
running-page margins. Without a source title, enabled: true creates no title
page and does not suppress first-page running matter. All three switches must
be YAML booleans. Manuscript title-page settings are independent.
