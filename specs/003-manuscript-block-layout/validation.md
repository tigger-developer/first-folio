# Validation: Manuscript Block Layout

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
