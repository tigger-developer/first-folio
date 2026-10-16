# Prose office fixture expectations

This checklist is being assembled during W063 delivery. **Pending rows are uncovered**, not evidence of support. Values come from authored fixture YAML, independently of office output.

| YAML path | Fixture/profile | Expected value | Named user test |
|---|---|---|---|
| `folio.letter.font.family` | letter | letter: `"Iosevka Publish"` | UT063.11 - Letter typography |
| `folio.letter.font.letter-spacing` | letter | letter: `"0pt"` | UT063.11 - Letter typography |
| `folio.letter.font.size` | letter | letter: `"11pt"` | UT063.11 - Letter typography |
| `folio.letter.font.stretch` | letter | letter: `"100%"` | UT063.11 - Letter typography |
| `folio.letter.font.style` | letter | letter: `"regular"` | UT063.11 - Letter typography |
| `folio.letter.font.weight` | letter | letter: `"regular"` | UT063.11 - Letter typography |
| `folio.letter.margin-bottom` | letter | letter: `"25mm"` | UT063.11 - Letter typography |
| `folio.letter.margin-left` | letter | letter: `"30mm"` | UT063.11 - Letter typography |
| `folio.letter.margin-right` | letter | letter: `"25mm"` | UT063.11 - Letter typography |
| `folio.letter.margin-top` | letter | letter: `"25mm"` | UT063.11 - Letter typography |
| `folio.letter.page` | letter | letter: `"a4"` | UT063.11 - Letter typography |
| `folio.letter.space-after-date` | letter | letter: `"5mm"` | UT063.11 - Letter typography |
| `folio.letter.space-after-recipient` | letter | letter: `"6mm"` | UT063.11 - Letter typography |
| `folio.letter.space-after-sender` | letter | letter: `"8mm"` | UT063.11 - Letter typography |
| `folio.letter.space-after-subject` | letter | letter: `"4mm"` | UT063.11 - Letter typography |
| `folio.letter.space-before-closing` | letter | letter: `"7mm"` | UT063.11 - Letter typography |
| `folio.letter.space-before-signoff` | letter | letter: `"9mm"` | UT063.11 - Letter typography |
| `folio.manuscript.chapter.align` | Pending | Pending contrasting/inherited profile | UT063.7 - Headings and numbering |
| `folio.manuscript.chapter.blank-page-after` | Pending | Pending contrasting/inherited profile | UT063.7 - Headings and numbering |
| `folio.manuscript.chapter.blank-page-before` | Pending | Pending contrasting/inherited profile | UT063.7 - Headings and numbering |
| `folio.manuscript.chapter.case-transform` | Pending | Pending contrasting/inherited profile | UT063.7 - Headings and numbering |
| `folio.manuscript.chapter.explicit-numbering` | Pending | Pending contrasting/inherited profile | UT063.7 - Headings and numbering |
| `folio.manuscript.chapter.name-case` | Pending | Pending contrasting/inherited profile | UT063.7 - Headings and numbering |
| `folio.manuscript.chapter.number-format` | Pending | Pending contrasting/inherited profile | UT063.7 - Headings and numbering |
| `folio.manuscript.chapter.number-reset` | Pending | Pending contrasting/inherited profile | UT063.7 - Headings and numbering |
| `folio.manuscript.chapter.page-break-before` | typography, book | typography: `false`; book: `false` | UT063.7 - Headings and numbering |
| `folio.manuscript.chapter.position` | typography, book | typography: `"0mm"`; book: `"0mm"` | UT063.7 - Headings and numbering |
| `folio.manuscript.chapter.prefix` | Pending | Pending contrasting/inherited profile | UT063.7 - Headings and numbering |
| `folio.manuscript.chapter.separator` | Pending | Pending contrasting/inherited profile | UT063.7 - Headings and numbering |
| `folio.manuscript.chapter.show-name` | Pending | Pending contrasting/inherited profile | UT063.7 - Headings and numbering |
| `folio.manuscript.chapter.show-number` | typography, book | typography: `false`; book: `false` | UT063.7 - Headings and numbering |
| `folio.manuscript.chapter.skip-footer` | book | book: `false` | UT063.7 - Headings and numbering |
| `folio.manuscript.chapter.skip-header` | book | book: `false` | UT063.7 - Headings and numbering |
| `folio.manuscript.chapter.space-after` | Pending | Pending contrasting/inherited profile | UT063.7 - Headings and numbering |
| `folio.manuscript.code-block` | Pending | Pending contrasting/inherited profile | UT063.4 - Block layout |
| `folio.manuscript.code-block-indent` | typography | typography: `"8mm"` | UT063.4 - Block layout |
| `folio.manuscript.code-block-spacing` | typography | typography: `"5mm"` | UT063.4 - Block layout |
| `folio.manuscript.copyright.align` | Pending | Pending contrasting/inherited profile | UT063.9 - Copyright |
| `folio.manuscript.copyright.blank-page-after` | Pending | Pending contrasting/inherited profile | UT063.9 - Copyright |
| `folio.manuscript.copyright.blank-page-before` | Pending | Pending contrasting/inherited profile | UT063.9 - Copyright |
| `folio.manuscript.copyright.block-spacing` | Pending | Pending contrasting/inherited profile | UT063.9 - Copyright |
| `folio.manuscript.copyright.body` | Pending | Pending contrasting/inherited profile | UT063.9 - Copyright |
| `folio.manuscript.copyright.credits` | Pending | Pending contrasting/inherited profile | UT063.9 - Copyright |
| `folio.manuscript.copyright.enabled` | Pending | Pending contrasting/inherited profile | UT063.9 - Copyright |
| `folio.manuscript.copyright.font.family` | Pending | Pending contrasting/inherited profile | UT063.9 - Copyright |
| `folio.manuscript.copyright.font.letter-spacing` | Pending | Pending contrasting/inherited profile | UT063.9 - Copyright |
| `folio.manuscript.copyright.font.size` | Pending | Pending contrasting/inherited profile | UT063.9 - Copyright |
| `folio.manuscript.copyright.font.stretch` | Pending | Pending contrasting/inherited profile | UT063.9 - Copyright |
| `folio.manuscript.copyright.font.style` | Pending | Pending contrasting/inherited profile | UT063.9 - Copyright |
| `folio.manuscript.copyright.font.weight` | Pending | Pending contrasting/inherited profile | UT063.9 - Copyright |
| `folio.manuscript.copyright.isbn` | Pending | Pending contrasting/inherited profile | UT063.9 - Copyright |
| `folio.manuscript.copyright.isbn-barcode` | Pending | Pending contrasting/inherited profile | UT063.9 - Copyright |
| `folio.manuscript.copyright.isbn-label` | Pending | Pending contrasting/inherited profile | UT063.9 - Copyright |
| `folio.manuscript.copyright.label.font.family` | Pending | Pending contrasting/inherited profile | UT063.9 - Copyright |
| `folio.manuscript.copyright.label.font.letter-spacing` | Pending | Pending contrasting/inherited profile | UT063.9 - Copyright |
| `folio.manuscript.copyright.label.font.size` | Pending | Pending contrasting/inherited profile | UT063.9 - Copyright |
| `folio.manuscript.copyright.label.font.stretch` | Pending | Pending contrasting/inherited profile | UT063.9 - Copyright |
| `folio.manuscript.copyright.label.font.style` | Pending | Pending contrasting/inherited profile | UT063.9 - Copyright |
| `folio.manuscript.copyright.label.font.weight` | Pending | Pending contrasting/inherited profile | UT063.9 - Copyright |
| `folio.manuscript.copyright.line-spacing` | Pending | Pending contrasting/inherited profile | UT063.9 - Copyright |
| `folio.manuscript.copyright.position` | Pending | Pending contrasting/inherited profile | UT063.9 - Copyright |
| `folio.manuscript.copyright.publication` | Pending | Pending contrasting/inherited profile | UT063.9 - Copyright |
| `folio.manuscript.copyright.publisher` | Pending | Pending contrasting/inherited profile | UT063.9 - Copyright |
| `folio.manuscript.copyright.publisher-preposition` | Pending | Pending contrasting/inherited profile | UT063.9 - Copyright |
| `folio.manuscript.copyright.separator` | Pending | Pending contrasting/inherited profile | UT063.9 - Copyright |
| `folio.manuscript.copyright.separator-space-after` | Pending | Pending contrasting/inherited profile | UT063.9 - Copyright |
| `folio.manuscript.copyright.separator-space-before` | Pending | Pending contrasting/inherited profile | UT063.9 - Copyright |
| `folio.manuscript.copyright.skip-footer` | Pending | Pending contrasting/inherited profile | UT063.9 - Copyright |
| `folio.manuscript.copyright.skip-header` | Pending | Pending contrasting/inherited profile | UT063.9 - Copyright |
| `folio.manuscript.date-format` | Pending | Pending contrasting/inherited profile | UT063.4 - Block layout |
| `folio.manuscript.font.family` | typography, book | typography: `"Iosevka Publish"`; book: `"Iosevka Publish"` | UT063.3 - Font roles |
| `folio.manuscript.font.letter-spacing` | typography, book | typography: `"0pt"`; book: `"0pt"` | UT063.3 - Font roles |
| `folio.manuscript.font.size` | typography, book | typography: `"12pt"`; book: `"12pt"` | UT063.3 - Font roles |
| `folio.manuscript.font.stretch` | typography, book | typography: `"100%"`; book: `"100%"` | UT063.3 - Font roles |
| `folio.manuscript.font.style` | typography, book | typography: `"regular"`; book: `"regular"` | UT063.3 - Font roles |
| `folio.manuscript.font.weight` | typography, book | typography: `"regular"`; book: `"regular"` | UT063.3 - Font roles |
| `folio.manuscript.gutter` | book | book: `"5mm"` | UT063.5 - Page geometry |
| `folio.manuscript.heading.font.family` | typography | typography: `"Iosevka Publish Extended"` | UT063.3 - Font roles |
| `folio.manuscript.heading.font.letter-spacing` | typography | typography: `"0pt"` | UT063.3 - Font roles |
| `folio.manuscript.heading.font.size` | typography | typography: `"16pt"` | UT063.3 - Font roles |
| `folio.manuscript.heading.font.stretch` | typography | typography: `"100%"` | UT063.3 - Font roles |
| `folio.manuscript.heading.font.style` | typography | typography: `"regular"` | UT063.3 - Font roles |
| `folio.manuscript.heading.font.weight` | typography | typography: `"bold"` | UT063.3 - Font roles |
| `folio.manuscript.hyphenation` | typography | typography: `false` | UT063.4 - Block layout |
| `folio.manuscript.justify` | typography | typography: `false` | UT063.4 - Block layout |
| `folio.manuscript.line-spacing` | typography, book | typography: `1.5`; book: `"1"` | UT063.4 - Block layout |
| `folio.manuscript.list.space-after` | Pending | Pending contrasting/inherited profile | UT063.4 - Block layout |
| `folio.manuscript.list.space-before` | Pending | Pending contrasting/inherited profile | UT063.4 - Block layout |
| `folio.manuscript.margin` | typography, book | typography: `"20mm"`; book: `"20mm"` | UT063.5 - Page geometry |
| `folio.manuscript.mono.font.family` | typography | typography: `"Iosevka Publish"` | UT063.3 - Font roles |
| `folio.manuscript.mono.font.letter-spacing` | typography | typography: `"0pt"` | UT063.3 - Font roles |
| `folio.manuscript.mono.font.size` | typography | typography: `"9pt"` | UT063.3 - Font roles |
| `folio.manuscript.mono.font.stretch` | typography | typography: `"100%"` | UT063.3 - Font roles |
| `folio.manuscript.mono.font.style` | typography | typography: `"regular"` | UT063.3 - Font roles |
| `folio.manuscript.mono.font.weight` | typography | typography: `"regular"` | UT063.3 - Font roles |
| `folio.manuscript.page` | typography, book | typography: `"a4"`; book: `"a4"` | UT063.5 - Page geometry |
| `folio.manuscript.page-numbering.body-format` | book | book: `"1"` | UT063.7 - Headings and numbering |
| `folio.manuscript.page-numbering.body-reset` | book | book: `"first-part-or-chapter"` | UT063.7 - Headings and numbering |
| `folio.manuscript.page-numbering.frontmatter-format` | book | book: `"i"` | UT063.7 - Headings and numbering |
| `folio.manuscript.paragraph-indent` | typography | typography: `"10mm"` | UT063.4 - Block layout |
| `folio.manuscript.paragraph-spacing` | typography | typography: `"3mm"` | UT063.4 - Block layout |
| `folio.manuscript.part.align` | Pending | Pending contrasting/inherited profile | UT063.7 - Headings and numbering |
| `folio.manuscript.part.blank-page-after` | Pending | Pending contrasting/inherited profile | UT063.7 - Headings and numbering |
| `folio.manuscript.part.blank-page-before` | book | book: `"enforce-right"` | UT063.7 - Headings and numbering |
| `folio.manuscript.part.case-transform` | Pending | Pending contrasting/inherited profile | UT063.7 - Headings and numbering |
| `folio.manuscript.part.explicit-numbering` | Pending | Pending contrasting/inherited profile | UT063.7 - Headings and numbering |
| `folio.manuscript.part.name-case` | Pending | Pending contrasting/inherited profile | UT063.7 - Headings and numbering |
| `folio.manuscript.part.number-format` | Pending | Pending contrasting/inherited profile | UT063.7 - Headings and numbering |
| `folio.manuscript.part.number-reset` | Pending | Pending contrasting/inherited profile | UT063.7 - Headings and numbering |
| `folio.manuscript.part.page-break-before` | book | book: `true` | UT063.7 - Headings and numbering |
| `folio.manuscript.part.prefix` | Pending | Pending contrasting/inherited profile | UT063.7 - Headings and numbering |
| `folio.manuscript.part.separator` | Pending | Pending contrasting/inherited profile | UT063.7 - Headings and numbering |
| `folio.manuscript.part.show-name` | Pending | Pending contrasting/inherited profile | UT063.7 - Headings and numbering |
| `folio.manuscript.part.show-number` | book | book: `false` | UT063.7 - Headings and numbering |
| `folio.manuscript.part.skip-footer` | book | book: `true` | UT063.7 - Headings and numbering |
| `folio.manuscript.part.skip-header` | book | book: `true` | UT063.7 - Headings and numbering |
| `folio.manuscript.part.vertical-align` | Pending | Pending contrasting/inherited profile | UT063.7 - Headings and numbering |
| `folio.manuscript.quote-block-indent` | typography | typography: `"6mm"` | UT063.4 - Block layout |
| `folio.manuscript.quoted-block-spacing` | typography | typography: `"4mm"` | UT063.4 - Block layout |
| `folio.manuscript.quoted-block.font.family` | typography | typography: `"Iosevka Publish Extended"` | UT063.3 - Font roles |
| `folio.manuscript.quoted-block.font.letter-spacing` | typography | typography: `"0pt"` | UT063.3 - Font roles |
| `folio.manuscript.quoted-block.font.size` | typography | typography: `"10pt"` | UT063.3 - Font roles |
| `folio.manuscript.quoted-block.font.stretch` | typography | typography: `"100%"` | UT063.3 - Font roles |
| `folio.manuscript.quoted-block.font.style` | typography | typography: `"italic"` | UT063.3 - Font roles |
| `folio.manuscript.quoted-block.font.weight` | typography | typography: `"regular"` | UT063.3 - Font roles |
| `folio.manuscript.scene-break.marker` | typography | typography: `"#"` | UT063.4 - Block layout |
| `folio.manuscript.style` | Pending | Pending contrasting/inherited profile | UT063.4 - Block layout |
| `folio.manuscript.table.space-after` | Pending | Pending contrasting/inherited profile | UT063.4 - Block layout |
| `folio.manuscript.table.space-before` | Pending | Pending contrasting/inherited profile | UT063.4 - Block layout |
| `folio.manuscript.title-page.author.align` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.title-page.author.attribution` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.title-page.author.font.family` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.title-page.author.font.letter-spacing` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.title-page.author.font.size` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.title-page.author.font.stretch` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.title-page.author.font.style` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.title-page.author.font.weight` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.title-page.contact.align` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.title-page.contact.font.family` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.title-page.contact.font.letter-spacing` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.title-page.contact.font.size` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.title-page.contact.font.stretch` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.title-page.contact.font.style` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.title-page.contact.font.weight` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.title-page.date.align` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.title-page.date.font.family` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.title-page.date.font.letter-spacing` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.title-page.date.font.size` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.title-page.date.font.stretch` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.title-page.date.font.style` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.title-page.date.font.weight` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.title-page.enabled` | typography, book | typography: `false`; book: `true` | UT063.6 - Title page |
| `folio.manuscript.title-page.footer-align` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.title-page.include-address` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.title-page.include-author` | book | book: `true` | UT063.6 - Title page |
| `folio.manuscript.title-page.include-contact-name` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.title-page.include-date` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.title-page.include-email` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.title-page.include-phone` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.title-page.include-subtitle` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.title-page.include-title` | book | book: `true` | UT063.6 - Title page |
| `folio.manuscript.title-page.include-version` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.title-page.include-website` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.title-page.include-wordcount` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.title-page.page-number` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.title-page.subtitle.align` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.title-page.subtitle.font.family` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.title-page.subtitle.font.letter-spacing` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.title-page.subtitle.font.size` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.title-page.subtitle.font.stretch` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.title-page.subtitle.font.style` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.title-page.subtitle.font.weight` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.title-page.title-block-align` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.title-page.title.align` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.title-page.title.font.family` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.title-page.title.font.letter-spacing` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.title-page.title.font.size` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.title-page.title.font.stretch` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.title-page.title.font.style` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.title-page.title.font.weight` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.title-page.version.align` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.title-page.version.font.family` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.title-page.version.font.letter-spacing` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.title-page.version.font.size` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.title-page.version.font.stretch` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.title-page.version.font.style` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.title-page.version.font.weight` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.title-page.wordcount.align` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.title-page.wordcount.font.family` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.title-page.wordcount.font.letter-spacing` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.title-page.wordcount.font.size` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.title-page.wordcount.font.stretch` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.title-page.wordcount.font.style` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.title-page.wordcount.font.weight` | Pending | Pending contrasting/inherited profile | UT063.6 - Title page |
| `folio.manuscript.toc.blank-page-after` | Pending | Pending contrasting/inherited profile | UT063.8 - Contents |
| `folio.manuscript.toc.blank-page-before` | Pending | Pending contrasting/inherited profile | UT063.8 - Contents |
| `folio.manuscript.toc.continuation-padding-before` | book | book: `"10mm"` | UT063.8 - Contents |
| `folio.manuscript.toc.dot-leaders` | book | book: `true` | UT063.8 - Contents |
| `folio.manuscript.toc.enabled` | typography, book | typography: `false`; book: `true` | UT063.8 - Contents |
| `folio.manuscript.toc.font.family` | book | book: `"Iosevka Publish"` | UT063.8 - Contents |
| `folio.manuscript.toc.font.letter-spacing` | book | book: `"0pt"` | UT063.8 - Contents |
| `folio.manuscript.toc.font.size` | book | book: `"12pt"` | UT063.8 - Contents |
| `folio.manuscript.toc.font.stretch` | book | book: `"100%"` | UT063.8 - Contents |
| `folio.manuscript.toc.font.style` | book | book: `"regular"` | UT063.8 - Contents |
| `folio.manuscript.toc.font.weight` | book | book: `"regular"` | UT063.8 - Contents |
| `folio.manuscript.toc.heading.font.family` | Pending | Pending contrasting/inherited profile | UT063.8 - Contents |
| `folio.manuscript.toc.heading.font.letter-spacing` | Pending | Pending contrasting/inherited profile | UT063.8 - Contents |
| `folio.manuscript.toc.heading.font.size` | Pending | Pending contrasting/inherited profile | UT063.8 - Contents |
| `folio.manuscript.toc.heading.font.stretch` | Pending | Pending contrasting/inherited profile | UT063.8 - Contents |
| `folio.manuscript.toc.heading.font.style` | Pending | Pending contrasting/inherited profile | UT063.8 - Contents |
| `folio.manuscript.toc.heading.font.weight` | Pending | Pending contrasting/inherited profile | UT063.8 - Contents |
| `folio.manuscript.toc.include-chapters` | book | book: `true` | UT063.8 - Contents |
| `folio.manuscript.toc.include-parts` | book | book: `true` | UT063.8 - Contents |
| `folio.manuscript.toc.include-sections` | book | book: `true` | UT063.8 - Contents |
| `folio.manuscript.toc.line-spacing` | book | book: `"1.5"` | UT063.8 - Contents |
| `folio.manuscript.toc.links` | book | book: `true` | UT063.8 - Contents |
| `folio.manuscript.toc.page-break-before` | Pending | Pending contrasting/inherited profile | UT063.8 - Contents |
| `folio.manuscript.toc.page-numbers` | book | book: `true` | UT063.8 - Contents |
| `folio.manuscript.toc.part-bold` | Pending | Pending contrasting/inherited profile | UT063.8 - Contents |
| `folio.manuscript.toc.part-gap-before` | Pending | Pending contrasting/inherited profile | UT063.8 - Contents |
| `folio.manuscript.toc.title` | book | book: `"CONTENTS-SAMPLE"` | UT063.8 - Contents |
| `folio.manuscript.widow-orphan-control` | Pending | Pending contrasting/inherited profile | UT063.4 - Block layout |
| `folio.page-footer.align` | Pending | Pending contrasting/inherited profile | UT063.5 - Page geometry |
| `folio.page-footer.content-padding-after` | book | book: `"5mm"` | UT063.5 - Page geometry |
| `folio.page-footer.distance-from-edge` | book | book: `"20mm"` | UT063.5 - Page geometry |
| `folio.page-footer.enabled` | typography, book, letter | typography: `false`; book: `true`; letter: `true` | UT063.5 - Page geometry |
| `folio.page-footer.font.family` | book | book: `"Iosevka Publish"` | UT063.5 - Page geometry |
| `folio.page-footer.font.letter-spacing` | book | book: `"0pt"` | UT063.5 - Page geometry |
| `folio.page-footer.font.size` | book | book: `"10pt"` | UT063.5 - Page geometry |
| `folio.page-footer.font.stretch` | book | book: `"100%"` | UT063.5 - Page geometry |
| `folio.page-footer.font.style` | book | book: `"regular"` | UT063.5 - Page geometry |
| `folio.page-footer.font.weight` | book | book: `"regular"` | UT063.5 - Page geometry |
| `folio.page-footer.format` | book, letter | book: `"BODY-FOOT [page]"`; letter: `"LETTER-MUST-NOT-HAVE-THIS-FOOTER"` | UT063.5 - Page geometry |
| `folio.page-header.align` | Pending | Pending contrasting/inherited profile | UT063.5 - Page geometry |
| `folio.page-header.content-padding-after` | book | book: `"5mm"` | UT063.5 - Page geometry |
| `folio.page-header.distance-from-edge` | book | book: `"20mm"` | UT063.5 - Page geometry |
| `folio.page-header.enabled` | typography, book, letter | typography: `false`; book: `true`; letter: `true` | UT063.5 - Page geometry |
| `folio.page-header.font.family` | book | book: `"Iosevka Publish"` | UT063.5 - Page geometry |
| `folio.page-header.font.letter-spacing` | book | book: `"0pt"` | UT063.5 - Page geometry |
| `folio.page-header.font.size` | book | book: `"10pt"` | UT063.5 - Page geometry |
| `folio.page-header.font.stretch` | book | book: `"100%"` | UT063.5 - Page geometry |
| `folio.page-header.font.style` | book | book: `"regular"` | UT063.5 - Page geometry |
| `folio.page-header.font.weight` | book | book: `"regular"` | UT063.5 - Page geometry |
| `folio.page-header.format` | book, letter | book: `"BODY-EVEN [page]"`; letter: `"LETTER-MUST-NOT-HAVE-THIS-HEADER"` | UT063.5 - Page geometry |

## Measurement rules

- PDF units are points: 72 per inch. Page dimensions have a 0.25pt tolerance and measured positions a 1pt tolerance.
- Typography body edge is 20mm from the left page edge. Ordinary first-line indent adds 10mm, quote inset adds 6mm, and code inset adds 8mm. A chapter-opening paragraph has no first-line indent.
- Header edge distance and body clearance describe layout boxes. Do not compare a glyph top with a paragraph-box top without the font metrics.
- Native font-size values must match within one serialization unit. The Poppler font-size attribute is integer-rounded even with `-noroundcoord`; it cannot alone prove a 0.25pt size tolerance. Retain native assertions and use precise glyph metrics or explicit user validation.
- Footnote family is inherited from the body. The Typst baseline and upstream footnote rule establish an 85% size; this is not a separate YAML role. Native office rounding still needs verification.

## Remaining fixture work

- Add contrasting profiles for absent, empty, disabled and inherited settings, copyright/barcodes, title placement and remaining numbering policies.
- Add generation commands, visible markers, named RT mappings and verified measurement references before user handoff.
- Reconcile metadata and deprecated running-matter aliases with the config types. This initial inventory enumerates current preset leaf keys.
