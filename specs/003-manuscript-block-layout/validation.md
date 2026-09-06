# Validation: Manuscript Block Layout

Emergency lifecycle note: The project owner authorized implementation with
`BYPASS-GATE-7` on 2026-09-05. The normal pre-implementation `PENDING` ledger
therefore does not exist, and this record does not fabricate one retrospectively.
Each result below records the actual post-implementation execution and tested
revision.

## OOT-001 - configured block layout compiles

- Category: One-off test
- Requirement: FR-001 - quoted-block clearance, FR-002 - code-block clearance,
  and FR-004 - quote typography
- Expected: A Markdown manuscript containing a blockquote and fenced code
  compiles to PDF with non-default equal spacing and all six quoted-font
  properties, and the PDF text retains both blocks.
- Procedure: Generate a temporary Markdown source and local `script.yaml`, run
  the repository CLI to PDF, inspect PDF metadata, and extract its text.
- Status: PASS
- Tested revision: Emergency candidate
  `f0f41bb1d3cd4977a96c8a6be14fd65bb36131aabf502586fb9cf0e5ea28d316`
- Environment: macOS arm64, Go 1.26.3, Typst 0.14.2, Poppler 26.04.0
- Tester: Claudius
- Observed: The CLI returned zero and produced a one-page A4 PDF of 23,977
  bytes. Extracted text retained the prose before and after the quote, the
  quoted passage, and the fenced-code content.
- Post-audit repeat: PASS after the focused template-delta audit, with the same
  page size, byte count, and extracted content.

## OOT-002 - configured whole-block indents render

- Category: One-off test
- Requirement: FR-010 - quoted-block indentation and FR-011 - code-block
  indentation
- Expected: A Markdown manuscript containing a wrapping blockquote and a
  two-line fenced-code block compiles to PDF with separate non-default left
  indents applied uniformly within each block.
- Procedure: Generate a temporary Markdown source and local `script.yaml`, set
  `quote-block-indent: 2em` and `code-block-indent: 3em`, run the repository CLI
  to PDF, extract Poppler bounding boxes and layout-retained text, and inspect
  the rendered content page at 120 DPI.
- Status: PASS
- Tested revision: Candidate diff
  `6dd5a52047dd399e0a8b95a3d7803711c93e7a9541e1687be5552daaf7c715f8`
- Environment: macOS arm64, Go 1.26.3, Typst 0.14.2, Poppler 26.04.0
- Tester: Claudius
- Observed: The CLI returned zero and produced a three-page A4 PDF of 19,532
  bytes. Both wrapped quote lines shared the quote inset, both code lines shared
  the code inset, and the surrounding prose retained its body margin. Bounding
  boxes placed both quote lines at `xMin=92.692920`, both code lines at
  `xMin=85.492910`, and adjacent prose at `xMin=56.692913`.

## OOT-003 - complete-line code-span classification renders

- Category: One-off test
- Requirement: FR-012 - promote complete-line code spans and FR-013 - preserve
  mixed-content inline code
- Expected: A complete-line Markdown code span uses the configured code-block
  layout while code embedded in prose remains inline. In the supplied negative
  example, `but ` is prose before the code span and `#because...` remains inside
  that span.
- Procedure: Render the four source forms supplied by the project owner through
  the public manuscript command, inspect the generated Typst, compile to PDF,
  extract Poppler bounding boxes and layout-retained text, and inspect the
  rendered content page at 120 DPI.
- Status: PASS
- Tested revision: Audited candidate diff
  `e33af353c2b01082cb016c6b67b0aefdc7f6c4a900c6ce126312d533e5102afc`
  based on production revision `e71795a`
- Environment: macOS arm64, Go 1.26.3, Typst 0.14.2, Poppler 26.04.0
- Tester: Claudius
- Observed: The existing fence and the complete-line code span rendered with
  the configured code-block inset. Inline code within prose, including the span
  preceded by `but ` with `#because...` inside its delimiters, remained at the
  prose margin. Text content was retained in all four cases. Bounding boxes
  placed both block-code lines at `xMin=85.492910`; both prose lines began at
  `xMin=56.692913`, while `#because` remained within the inline monospace span.
- Superseded fixture: The earlier 18,575-byte run placed `#because...` outside
  the closing backtick and therefore did not evidence the project owner's
  intended negative example.
- Corrected post-audit repeat: PASS with a three-page A4 PDF of 18,580 bytes.
  Extracted text, bounding boxes, and the rendered page confirm that leading
  `but ` determines prose classification while all delimited content remains
  inline code.

## OOT-004 - paragraph indentation survives structural blocks

- Category: One-off test
- Requirement: FR-014 - indent prose after blocks and FR-015 - keep
  chapter-opening prose flush
- Expected: In the project owner's supplied passage, the chapter-opening prose
  begins at the body margin while prose after every code block begins at the
  body margin plus `paragraph-indent`. A separate blockquote fixture produces
  the same post-block indent.
- Procedure: Render the supplied Markdown passage and a focused blockquote
  fixture through the public manuscript command, extract Poppler bounding boxes,
  and inspect the rendered content page at 120 DPI.
- Status: PASS
- Tested revision: Audited candidate diff
  `c9f6c848a9f0872ee64562bfa9f9e62afc80780239a03e7b2fcb8857fb88e599`
- Environment: macOS arm64, Go 1.26.3, Typst 0.14.2, Poppler 26.04.0
- Tester: Claudius
- Observed: The supplied passage produced a four-page A4 PDF of 32,290 bytes.
  The chapter-opening prose began at `xMin=56.692913`; prose after complete-line
  code blocks and subsequent ordinary prose began at `xMin=85.039370`. The
  blockquote fixture placed its opening prose at `xMin=56.692913`, quoted text
  at `xMin=68.692920`, and post-quote prose at `xMin=85.039370`. Inspection of
  the supplied passage at 120 DPI confirmed the same visible indentation.
