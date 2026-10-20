---
title: W063 operator validation handoff
version: 3
last-updated: 2026-10-20
---

# W063 operator validation

## Accepted closure

Accepted for use and closed on 2026-10-20, with residual heading clipping and a
slight FolioBody caret/text offset retained as known limitations. Their causes
remain unverified. The implementation is integrated into master. All twelve
formal user tests remain AMBER; their remaining procedures are waived for this
closure, not recorded as passed. See [the closure evidence](validation.org).

Version 3 records acceptance and closure. The earlier retest instructions and
candidate-specific claims below are preserved as historical evidence.

## Historical editor-remediation candidate

W063 remains unaccepted. Ordinary body characters now use the natural baseline;
formatting styles are shared by semantic level/role and exact properties while
anchors, fields and page boundaries stay distinct. Automated baseline/style
regressions, full regression and the four-case novel reader acceptance pass.
The current internal delivery-code review is recorded separately in
[internal-review.md](internal-review.md); no independent review is claimed.
The earlier handoff and its files below are historical evidence for the reported
editor defects, not the current retest candidate.

| Source | Fresh DOCX | Fresh ODT |
|---|---|---|
| Markdown | [Novel for retest](/var/folders/wn/zk3zl2050rd7qkfgkfzn6rwr0000gp/T/folio-w063-editor-novels-final.r5TPd2jF0z/md-docx-3761308049/novel.docx) | [Novel for retest](/var/folders/wn/zk3zl2050rd7qkfgkfzn6rwr0000gp/T/folio-w063-editor-novels-final.r5TPd2jF0z/md-odt-614836670/novel.odt) |
| Org | [Novel for retest](/var/folders/wn/zk3zl2050rd7qkfgkfzn6rwr0000gp/T/folio-w063-editor-novels-final.r5TPd2jF0z/org-docx-40061999/novel.docx) | [Novel for retest](/var/folders/wn/zk3zl2050rd7qkfgkfzn6rwr0000gp/T/folio-w063-editor-novels-final.r5TPd2jF0z/org-odt-732247428/novel.odt) |

[Fresh small-fixture comparison and editable profiles](/var/folders/wn/zk3zl2050rd7qkfgkfzn6rwr0000gp/T/folio-w063-editor-profiles.MdVvnfyZWJ/comparison.html).
All four novels render to 654 pages and have three heading/contents definitions
in the relevant named-style families. The richer part/chapter regression uses a
constant five relevant styles in DOCX and six in ODT at both tested sizes.

Start the human retest with ordinary body text on physical page 5, then repeat
in the middle and final chapter: place the caret, type accented text, select
across lines and scroll. Caret, selection and visible text must coincide, with
no editing-view clipping. Check that genuine footnote anchors remain raised.
Confirm that editing a shared chapter style affects equivalent chapter headings
and that their distinct contents links and page references still work after
refresh. Save to a new file and reopen under the existing UT procedure.

All twelve named user tests below remain AMBER; the failed observation on the
old candidate is retained in validation. Headless printing does not establish
interactive editor usability. The live executable remains unchanged.

## Historical handoff before editor findings

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
