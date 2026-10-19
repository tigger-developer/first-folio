---
title: W063 internal delivery review
version: 3
last-updated: 2026-10-19
---

# 2026-10-19: Superseded delivery review

The previous delivery and its internal PASS are rejected as acceptance evidence.
The implementation used a separate Pandoc DOCX writer rather than extending the
existing native script exporter. The historical review below is retained; it
does not validate the corrective implementation. The fresh corrective delivery review appears at the end of this record.

# 2026-10-18: Delivery-code review

Review method: internal (current session; not independent)

Gate: delivery-code

Mode: SDLC

Revision: `8c49c6e`, clean working tree at review start. Production ends at
`1079f63`; the later checkpoint settles specification and validation evidence.
This review record is the only addition after the reviewed checkpoint.

External review: suspended by standing policy

INTERNAL REVIEW: PASS

No unresolved blocking finding was identified. This is an author self-review,
not an independent verdict. It permits the agreed operator-validation handoff;
it does not authorize installation, publication or ticket closure.

## Authority, scope and method

Reviewed the signed-off [specification](spec.org), accepted fidelity decisions,
[validation](validation.org), maintained fixture checklist and historical
[definition/test-code evidence](audits.yaml). The historical record was not
edited or replaced. The delivery review began after required checks finished
and the delivery-code readiness validator returned `ready: true`: ten acceptance
criteria, twelve GREEN RT groups, two GREEN OTs and twelve AMBER operator UTs.

Applied MAIN, AUDITS, CODING, TESTING, ARCHITECTURE, SECURITY, GO, SHELL,
DOCUMENTATION, GIT, ORG-SCHEMA, MARKDOWN and ORGMODE. These standards were read
in full during this invocation; unchanged readings were reused.

Reviewed affected tests before production. The complete W063 production delta
and relevant unchanged interfaces were inspected during delivery preparation;
the settled candidate and subsequent changes were reassessed at this gate.
Scope includes the manuscript and letter adapters, shared office package,
font extraction compatibility, application routing, Makefile prerequisites,
affected preservation tests, CLI help and required product documentation.
The unrelated W064 warning implementation and operator preset edits are retained
baseline context, not W063-authored changes.

| Criteria | Review conclusion and evidence |
|---|---|
| AC063.1-3: commands and content | Both suffix-selected manuscript formats and recipient-specific letter formats reach the real Pandoc writer. Structured source assertions cover ordering, inline syntax, private sections and native editable blocks. Separate consuming-directory commands establish executable packaging. |
| AC063.4: configuration | Native properties are compared with authored YAML, including independent roles, relative lengths, stretch, tracking and additive spacing. Rendered measurements cover fonts, baseline geometry, margins and letter gaps; they do not rely solely on ZIP presence or rounded PDF font-size labels. |
| AC063.5-6: generated matter and pages | Native fields, anchors, grouped title/copyright placement, barcode geometry and page policies are supplemented by explicit reader refresh, contents-reference checks and semantic-page PNG inspection. Accepted DOCX opening-band and ODT transition limits remain explicit. |
| AC063.7-8: letters and fonts | Letter composition retains recipient substitution, Org emphasis and protected literals. Font tests cover permitted full faces, missing/restricted warnings, deduplication and budgets; the bounded uninstalled-family probe supports reader use of packaged fonts. |
| AC063.9: failure handling | Source/AST/ZIP/XML validation, fixed process arguments, deadlines, diagnostic bounds, private staging, source aliases, collisions and partial publication are tested at their owning boundaries. No remote service or executable document payload is introduced. |
| AC063.10: preservation | Existing PDF/Typst and script routes remain separate. Shared font extraction retains the script policy. The full regression suite and focused script-font preservation evidence passed. |

## Findings addressed during delivery preparation

These are internal findings, with retained RED/GREEN evidence in validation;
they are not copied external-auditor findings.

| Finding | Remediation and reassessment |
|---|---|
| Copyright strings were parsed as full Markdown | `office_copyright.go` preserves the existing small formatting contract, literal unsupported syntax, line breaks and configured rule markers. `TestRT063_5_OfficeCopyrightMiniMarkdown` passes. |
| Barcode clipping and displaced bottom grouping | Natural image line height and section markers outside positioned paragraphs retain full SVG proportions and copyright placement. Native regression and renewed PNG inspection pass. |
| Chapter layout leaked into local sections | Local headings retain independent left alignment and existing 1.2em/0.4em spacing. The named section-layout regression passes. |
| Generated-matter boundary was lost before tables/lists | A minimal native paragraph carries the boundary without flattening the structured block. Paragraph-control and table/list regressions pass in both formats. |
| Refreshed DOCX page fields lost band typography | Named running-band styles and whole-band baseline positioning retain configured family, size, style and alignment after refresh. The rendered field test, existing geometry tests and three book rerenders pass. |
| ODT discarded one clearance between two tables | Resolved generated styles retain earlier edge properties. The new public-command regression first failed only ODT; it and the existing additive/table groups pass after the correction. |
| Verification needed reconciliation | The release assertion now requires the declared 20-minute aggregate timeout while retaining failed-test publication blocking. The continuous ODT test explicitly checks the shared physical page. Root help names all manuscript output formats. |

The last ODT change adds one resolved-style assignment in a per-export plan.
It introduces no shared state or new font/layout policy. Its affected regression
groups were rerun; unaffected full-suite and visual-profile evidence remains
current under TESTING's affected-check rule.

## Settled checks

- `make test` passed at `02a89f7`: application 483.243s, manuscript 116.086s,
  office 2.630s and release gate 3.599s. All other packages passed.
  Log: `.agent/w063-final-make-test.log`.
- The later `1079f63` inter-table correction passed its new regression and
  existing additive/table groups in 9.620s. Log:
  `.agent/w063-inter-table-green.log`. The full-suite run did not include this
  later assignment; the evidence boundary is deliberate and explicit.
- Current `make lint`, office race check, `gofmt -l` over changed Go files and
  `git diff --check` passed. The current race result reused the valid cache;
  the preceding uncached office race run took 4.930s.
- Combined integration coverage at `736e7b6` covers 2,565/2,853 statements
  (89.9%) and 3,015/3,519 instrumented source-line ranges (85.7%) in the new
  office mechanics and prose adapters. Duplicate blocks were merged by source
  location. This is not whole-repository coverage. The later style assignment
  is exercised by its regression. Coverage is supporting evidence, not an oracle.
- The retained vulnerability scan reports no vulnerabilities. `go.mod` and
  `go.sum` are unchanged since its candidate; no new dependency was added.
- OT063.1 inspected seven profiles, including repaired title/copyright,
  section and field cases. The final comparison retains 17,303,496 bytes.
  OT063.2 passed four conversions and four help invocations outside the checkout,
  producing six office files. Exact candidates, bounds and locations are in
  [validation](validation.org).
- Required documentation describes commands, dependencies, field refresh,
  publication/resource limits and accepted fidelity limitations. Authored
  documentation was processed with `sanitize`. No harness call or modification
  of `audits.yaml` occurred in this review.

## Operator boundary

All twelve named UTs remain AMBER. Editing and Microsoft Word behaviour require
operator validation; LibreOffice rendering does not establish Word fidelity.
The accepted DOCX opening-band parity, ODT frontmatter/body transition and weight
limitations remain documented. Reader-dependent pagination and Extended-face
selection are not promised to match PDF output.

Use the [fixture checklist](../../testdata/prose-office/EXPECTATIONS.md) and
[validation inventory](validation.org) for the operator checks. The work item
remains ACTIVE. No installation, release, implementation push or closure is
included in this handoff.

# 2026-10-19: Corrected definition and test-code review

Review method: internal (current session; not independent)

Gates: definition; test-code

Mode: SDLC

Revision: `d1ed1b9` plus the corrected specification, restored public-boundary
fixtures/tests, revised dependency tests and declaration-only resource scaffolding.
External review: suspended by standing policy.

The architecture correction requires shared native DOCX generation and the
existing configured ODT writer. Manuscript source parsing may still use Pandoc;
letter DOCX may not. Existing layout, safety and fixture requirements remain.
The correction is approved on 2026-10-19; historical delivery results are superseded.

Reviewed the corrected design against the project architecture, separate semantic
models, native script writers and scoped rollback. Reviewed public command tests,
ZIP/XML property oracles, configuration matrices, rendered measurements, publication
fault injection and resource-boundary tests against the retained acceptance criteria.
Applied coding, Go, architecture, testing, security, documentation and internal-review
standards. No external reviewer or harness was invoked.

Findings and remediation:

- The old dependency matrix incorrectly required Pandoc for letter DOCX. Corrected
  it and added named absent-Pandoc and manuscript-reader-only cases. The latter
  executes the real source reader while rejecting DOCX writer/reference calls.
- Historical GREEN results could imply acceptance of the replacement. Current RT
  states now record observed RED, with a separately qualified preservation subset;
  OTs and UTs remain pending.
- Low-level safety coverage was removed with the rejected implementation. Restored
  the reviewed tests and declaration-only placeholders; no production caller uses
  those placeholders. They fail on missing validation/publication behaviour.

Checks: public prose regressions executed and failed at absent output selection;
resource tests failed on no-op boundaries; the selected existing application and
manuscript preservation tests passed. See `validation.org` for commands and scope.
The shared-exporter requirement also needs implementation inspection; dependency
checks alone cannot establish code reuse. Full regressions and fresh rendered
comparison remain delivery prerequisites.

INTERNAL REVIEW: PASS

This result permits corrective implementation within the approved architecture.
It is not delivery acceptance, independent review, or operator closure.

# 2026-10-19: Shared-exporter delivery-code review

Review method: internal (current session; not independent)

Gate: delivery-code

Mode: SDLC

Revision: `91ad009` plus the restored malformed-AST regression and current specification/validation records. The restoration changes tests only; production ends at `91ad009`.

External review: suspended by standing policy

INTERNAL REVIEW: PASS

No unresolved blocking finding remains in this internal review. This author self-review permits operator validation, not installation, publication or closure.

## Scope and criteria

Reviewed the corrected shared-exporter architecture, production delta, affected tests, command boundaries and product documentation against the signed-off W063 scope. Applied MAIN, CODING, TESTING, ARCHITECTURE, SECURITY, AUDITS, Go and documentation/Org/Markdown standards. The earlier definition and test-code reviews remain the admission record; this is fresh delivery review of their replacement implementation.

The script DOCX path and prose DOCX path both use the extracted native paragraph/XML/package primitives. Prose preserves its own semantic adapter and resolved role/page model. Both ODT paths use the same bounded writer and native configured-reference packaging. Manuscript source parsing still calls Pandoc; letter DOCX does not. Dependency tests execute the real manuscript reader while rejecting writer/reference calls, rather than treating a source search as architectural proof.

Inspected preserved interfaces and ownership: source/configuration precedence, script target guards, letter batch preflight and partial publication, manuscript document/sidecar order, bounded process cancellation, ZIP/XML validation, local font embedding and failure diagnostics. Existing PDF/Typst routing and unrelated preset/configuration changes remain outside the corrective implementation.

## Findings and remediation

- Native rich-content review exposed missing outline and nested-list levels, footnote hyperlink relationships, code/note role names, barcode proportions and document-level automatic hyphenation. Focused regressions reproduced the defects and passed after remediation.
- The visual note-reference finding was not resolved by run positioning or a dedicated reference style. The final implementation keeps reference and surrounding text on the native paragraph baseline. The retained rendered regression measures the raised reference; the final typography PNG was inspected. Property-only intermediate failures remain recorded.
- Source semantics review found double quotation substituted for single quotation, decimal labels substituted for alphabetic labels, and missing cell alignment. `TestRT063_3_DOCXSourceSemantics` reproduced all three. The final affected suite passes. Empty notes also retain an editable paragraph and owned reference without slicing an empty child list.
- The final historical-test comparison found `TestRT063_9_ASTRejectsMalformedAdapterNodes` omitted during selective restoration. Restored it unchanged. The AST group passes in `.agent/w063-native-audit-ast.log` (0.378s). No other prior test function was removed by the correction; dependency assertions were updated to the authorized native architecture.
- Historical delivery acceptance is explicitly superseded. Current RT/OT records now point to fresh replacement evidence; old failures and results remain available as history.

## Evidence and limits

- Delivery readiness is valid and ready in `.agent/w063-native-delivery-readiness.yaml`: twelve RT groups and both OTs GREEN, twelve operator UTs AMBER.
- Full `make test` passed on the extraction candidate. All W063 groups subsequently passed, followed by the final affected RT063.3/4/11 run at the current production revision. Validation records identify the candidate boundaries; the full run is not misrepresented as including later edits.
- New prose adapters/mechanics have 87.4% statement and 82.7% instrumented-line coverage. The measure excludes extracted existing script/font implementations and avoids merging stale line positions for changed files. Coverage does not establish useful assertions by itself; public-path, native-property, failure-path and rendered oracles were reviewed separately.
- Current lint reports zero issues, vet passes, the office race check passes and the unchanged dependency graph has no reported vulnerabilities. Resource tests cover source/AST/ZIP/font budgets, malformed structures, process output/deadline failures and atomic publication failures.
- Seven comparison profiles were generated and inspected, with targeted DOCX rerenders after remediation. The uninstalled-font probe passes for both formats. The outside-checkout CLI produces six valid editable files. Exact commands, bounds and temporary evidence paths are in `validation.org`.
- Product documentation explains source-reader versus writer dependencies, commands, field refresh, resource/publication behaviour and accepted fidelity limits. Authored additions were processed with `sanitize`.

The accepted DOCX special-opening parity, ODT frontmatter/body transition and weight limitations remain. Reader pagination and font selection are not guaranteed identical to PDF. LibreOffice rendering and package structure do not establish Microsoft Word fidelity, editing, or save/reopen behaviour. The twelve named user tests remain operator-owned. No external audit, harness-record mutation, installation, release, push or work-item closure occurred.
