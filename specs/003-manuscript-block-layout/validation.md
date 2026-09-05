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
  layout while code embedded in prose and code followed by trailing content
  remain inline.
- Procedure: Render the four source forms supplied by the project owner through
  the public manuscript command, inspect the generated Typst, compile to PDF,
  extract Poppler bounding boxes and layout-retained text, and inspect the
  rendered content page at 120 DPI.
- Status: PASS
- Tested revision: Audited candidate diff
  `393973073fe956c5f4c0181dc24d7b18ba31551fe6355d9d93da462dcfdaf5fb`
- Environment: macOS arm64, Go 1.26.3, Typst 0.14.2, Poppler 26.04.0
- Tester: Claudius
- Observed: The existing fence and the complete-line code span rendered with
  the configured code-block inset. Inline code within prose and the code span
  followed by a source comment remained at the prose margin. Text content was
  retained in all four cases. Bounding boxes placed both block-code lines at
  `xMin=85.492910`; the inline-code paragraph and the trailing-content line both
  began at `xMin=56.692913`.
- Post-audit repeat: PASS with the same three-page A4 output, 18,575-byte file
  size, extracted content, bounding-box positions, and rendered appearance.
