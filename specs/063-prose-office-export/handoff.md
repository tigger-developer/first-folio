---
title: W063 operator validation handoff
version: 1
last-updated: 2026-10-19
---

# W063 operator validation

The feature branch is `feature/W063-prose-office-export`. Internal delivery-code
review is PASS against `ec5113f`, by current-session author self-review without
independence. All required automated and one-off tests are GREEN. The twelve
user tests below remain AMBER. No merge, installation, release or closure is
part of this handoff; the live binary remains unchanged.

## Editable files

The corrected novel outputs each contain 10,000 body paragraphs, fifty chapters
and about 166,850 source words. Automated reader refresh produced 654 pages per
file. These results establish conversion/rendering, not human editing usability.
Source, profile and output hashes are in [validation](validation.org).

| Source | DOCX for Word | ODT for LibreOffice |
|---|---|---|
| Markdown | [Novel DOCX](/var/folders/wn/zk3zl2050rd7qkfgkfzn6rwr0000gp/T/folio-w063-diverse-novels.gmhc35S551/md-docx-3786990111/novel.docx) | [Novel ODT](/var/folders/wn/zk3zl2050rd7qkfgkfzn6rwr0000gp/T/folio-w063-diverse-novels.gmhc35S551/md-odt-3877756069/novel.odt) |
| Org | [Novel DOCX](/var/folders/wn/zk3zl2050rd7qkfgkfzn6rwr0000gp/T/folio-w063-diverse-novels.gmhc35S551/org-docx-293426489/novel.docx) | [Novel ODT](/var/folders/wn/zk3zl2050rd7qkfgkfzn6rwr0000gp/T/folio-w063-diverse-novels.gmhc35S551/org-odt-654992727/novel.odt) |

[Small-fixture comparison and editable profiles](/var/folders/wn/zk3zl2050rd7qkfgkfzn6rwr0000gp/T/folio-w063-profiles.a4nas7OY59/comparison.html)
cover typography, book composition and letters. Retained files are in the OS
temporary directory; their absolute paths and reproduction commands are recorded
in validation. The isolated CLI is `.agent/w063-native-bin/folio` in the feature
worktree, not the live primary-checkout executable.

## Pending user tests

Follow the exact procedures in [the signed-off specification](spec.org) and the
[independent YAML expectations](../../testdata/prose-office/EXPECTATIONS.md).
Use Word for DOCX and LibreOffice for ODT. Save edits as new files, retaining the
originals and their hashes. Record editor versions, observations, timings and
saved/reopened file paths in validation as we perform the checks together.

| Test | Human validation still required |
|---|---|
| UT063.1 | Edit, refresh, save, close and reopen all four novels, retaining beginning/middle/end content, table and footnote edits, and contents links/references. |
| UT063.2 | Edit, save and reopen both letter formats with recipient details and fonts retained. |
| UT063.3 | Manuscript font roles and inline specimens. |
| UT063.4 | Paragraph and block spacing. |
| UT063.5 | Paper geometry and running matter. |
| UT063.6 | Title-page fields and independent placement. |
| UT063.7 | Part/chapter presentation and pagination. |
| UT063.8 | Contents layout and live references after edits. |
| UT063.9 | Copyright composition and barcode modes. |
| UT063.10 | Metadata, style and configuration precedence, including first-source YAML. |
| UT063.11 | Letter typography, margins and block spacing. |
| UT063.12 | Letter substitutions, literals and excluded running matter. |

The accepted DOCX opening-band, ODT boundary and weight-mapping limitations
remain those named in the specification. LibreOffice's automated rendering does
not verify Word's behaviour. The [internal review](internal-review.md) states
the assurance boundary and retained remediation evidence.
