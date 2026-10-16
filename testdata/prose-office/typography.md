---
title: Labelled manuscript typography
author: Example Author
date: 2026-10-16
folio:
  page-header:
    enabled: false
  page-footer:
    enabled: false
  manuscript:
    page: a4
    margin: 20mm
    font:
      family: Iosevka Publish
      size: 12pt
      weight: regular
      style: regular
      stretch: 100%
      letter-spacing: 0pt
    heading:
      font:
        family: Iosevka Publish Extended
        size: 16pt
        weight: bold
        style: regular
        stretch: 100%
        letter-spacing: 0pt
    quoted-block:
      font:
        family: Iosevka Publish Extended
        size: 10pt
        weight: regular
        style: italic
        stretch: 100%
        letter-spacing: 0pt
    mono:
      font:
        family: Iosevka Publish
        size: 9pt
        weight: regular
        style: regular
        stretch: 100%
        letter-spacing: 0pt
    title-page:
      enabled: false
    toc:
      enabled: false
    line-spacing: 1.5
    paragraph-indent: 10mm
    paragraph-spacing: 3mm
    quote-block-indent: 6mm
    quoted-block-spacing: 4mm
    code-block-indent: 8mm
    code-block-spacing: 5mm
    justify: false
    hyphenation: false
    scene-break:
      marker: '#'
    chapter:
      page-break-before: false
      position: 0mm
      show-number: false
---

## HEADING-SAMPLE: 16pt bold Iosevka Publish Extended

OPENING-SAMPLE: 12pt Iosevka Publish, with no first-line indentation because this paragraph immediately follows its chapter heading.

BODY-SAMPLE: 12pt regular Iosevka Publish, with a 10mm first-line indent. This longer paragraph demonstrates wrapped lines returning to the body margin. Paragraph separation adds 3mm to the normal line interval.

This paragraph contains **BOLD-SAMPLE**, *ITALIC-SAMPLE*, ~~STRIKE-SAMPLE~~ and `INLINE-CODE-SAMPLE: 9pt Iosevka Publish`. A [LINK-SAMPLE](https://example.invalid/manuscript/path) retains its target. CÁIT and MAIRÉAD retain their accents.

> QUOTE-SAMPLE: 10pt italic Iosevka Publish Extended. Every line begins 6mm from the body edge; quote clearance adds 4mm above and below.
>
> QUOTE-SECOND-PARAGRAPH: the same 6mm inset, with no extra native quote indentation.

```text
CODE-SAMPLE: 9pt Iosevka Publish.
All lines start 8mm from the body edge.
Clearance adds 5mm above and below.
```

`STANDALONE-CODE-SAMPLE: this complete-line code span becomes a block.`

- LIST-FIRST: an editable bullet.
  - LIST-NESTED: a nested editable bullet.
- LIST-LAST: the second top-level bullet.

1. ORDERED-FIRST: the first numbered item.
2. ORDERED-LAST: the second numbered item.

| TABLE-HEADER-LEFT | TABLE-HEADER-RIGHT |
|-------------------|--------------------|
| TABLE-CELL-LEFT | TABLE-CELL-RIGHT |

---

NOTE-ANCHOR: this sentence links to an editable native footnote.[^labelled]

[^labelled]: NOTE-SAMPLE: native footnote with **bold note text** and *italic note text*. Expected Iosevka Publish at 85% of the 12pt body: 10.2pt; DOCX half-point rounding permits 10pt.

<!-- PRIVATE-SENTINEL must never appear in an export. -->

FINAL-SENTINEL: this final paragraph proves that the document was not truncated.
