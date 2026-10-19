---
title: W063 internal delivery review
version: 12
last-updated: 2026-10-19
---

# 2026-10-19: Editor-remediation delivery-code review

Review method: internal (current session; not independent)

Gate: delivery-code

Mode: SDLC

Revision: `623af4e9b155d6803e7091624d997e177945e781`, clean worktree at review.
Reviewed the complete remediation delta from the original `0eaa6e5` handoff.
Only this verdict and its derived validation/lifecycle records follow review;
no production, executable-test or product-documentation change is part of the
review output.

External review: suspended by standing policy

INTERNAL REVIEW: PASS

No unresolved blocking findings remain in the reviewed implementation and
required automated/one-off evidence. This author self-review authorizes the
agreed editing retest handoff only. W063 remains unaccepted; no Word or
LibreOffice interactive caret, selection, clipping, save/reopen or style-editing
PASS is inferred. All twelve UTs remain AMBER for the corrected candidate.

## Readiness, criteria and review

All required checks had finished before this final review. The installed
validator reports valid schema and delivery-code readiness true, with one AT,
thirteen RTs and two OTs GREEN and twelve UTs AMBER. Reused unchanged full
readings of AUDITS, TESTING, CODING, ARCHITECTURE, SECURITY, Go, documentation,
Org/Markdown and Git standards. Reviewed affected tests first, including their
helpers and observed failures, then the complete production delta, relevant
unchanged parser/export/publication interfaces and affected product documents.

| Criterion | Review conclusion |
|---|---|
| Ordinary editable baselines | `resolveBaselines` clears character offsets from ordinary paragraph roles. Word run normalization removes position properties; ODT ordinary paragraph/text styles do not serialize them. Native list markers use the same natural baseline. Font sizing and genuine `vertAlign`/note superscripts remain distinct. Both manuscript sources and both outputs, plus letters, have public-command guards against direct/inherited offsets. |
| Native geometry | Header metrics position the header paragraph box, with a compensating ODT following gap preserving the body boundary. Native margins and authored line intervals remain strict. The old ink-edge oracle was replaced because it rewarded the defective character shifts; its first-line-box bound now uses the independently specified 24.93pt interval. Genuine notes, font sizes, list alignment and field baselines retain rendered assertions. |
| Shared formatting, separate identities | Exact serialized formatting is interned within each semantic level/role after instance layout has been applied. Copies protect retained nodes while deriving signatures. Alias rewriting is restricted to style references; bookmarks, contents targets, numbering, fields and frame names remain individual. ODT page-transition overrides move into automatic styles. Positioned heading frames also share formatting. |
| Source and scale boundaries | The classifier accepts every already-supported positive Org outline level, rather than imposing Markdown's six-level recognition boundary. A level-seven case has intended RED and final GREEN. Four versus forty chapters with parts/frames retain constant common named-style inventories, independent of name prefixes, and distinct contents targets. The fifty-chapter novels retain three relevant heading/contents definitions each. |
| Preservation and resource/security rules | Native parser reuse, the PDF YAML cascade, zero-Pandoc DOCX and configured-writer-only ODT paths are unchanged and retain full-suite evidence. No dependencies, source fetching, public flags or access were added. The normalizer operates on already bounded/validated XML, then existing font/package validation and atomic publication continue. Regex/maps have no mutable cross-export state. |
| Maintainability and documentation | The normalization is a bounded collection/alias-rewrite phase, with separate renaming and reference helpers. Its main function exceeds the fifty-line tripwire; retaining one visible style pass avoids spreading shared local state across a new framework. Architecture, both format guides, independent fixture expectations and the UT handoff describe the native-baseline/shared-style contract. The expectation sheet's revision date was corrected and rechecked before this verdict. |

## Findings addressed and verification boundary

The observed baseline and per-instance style defects are addressed under W063,
not separate tickets. Preparation also caught ODT positioned-part styles,
accepted Org levels beyond six, the obsolete note compensation and two invalid
representation-coupled test expectations. Their RED, setup-failure and GREEN
records remain distinguishable in `validation.org`; none was erased or treated
as proof of successful editing.

The full regression passes with application 439.946s and manuscript 129.803s.
Later affected RT063.3/4/13 checks pass in 45.696s, the tightened rendered group
passes in 69.625s and final whole-inventory/deep-Org cases pass in 2.316s. This is
a full run plus explicitly identified affected follow-ups, not a claim that one
full invocation contained every later test refinement. Lint and affected race
checks pass. New normalization code has 90.9% statement and 86.0% instrumented-line
coverage; whole-office coverage from the selected application cases is 63.3%.
No new dependency graph was introduced; retained vulnerability evidence applies.

All four novel reader cases pass in 108.536s, retaining complete content,
references, counters/parity and measured font/margins across 654 pages per file.
The seven-profile comparison, affected composed-book refresh, unique uninstalled
font and consuming-directory CLI checks pass. Actual selected page images were
inspected, including positioned part frames and first-body pages. No clipping or
overlap was observed in those regions. These checks remain print/reader evidence;
the operator's reported editing-view defect requires the explicit fresh UT retest.

The live symlink/checksum remain at their recorded baseline. The feature branch
and new editable files may be returned for operator validation. No merge,
installation, release, ticket closure or independent assurance is claimed.

# 2026-10-19: Editor-remediation test-code review

Review method: internal (current session; not independent)

Gate: test-code

Mode: SDLC

Revision: `0eaa6e5` plus `prose_office_editing_test.go` and the recorded
operator-authorized W063 correction/RED evidence. Production is unchanged.

External review: suspended by standing policy

INTERNAL REVIEW: PASS

Reviewed the two actual public-command regressions, their package/style helpers,
relevant current baseline/layout and generated-matter code, and the operator's
reported editing failure against existing AC063.3/4/5/7/11 and UT063.1. Reused
unchanged AUDITS, TESTING, CODING, ARCHITECTURE, SECURITY and Go standards already
read in full. This is a routine correction within unaccepted W063, not a new
ticket or a change to approved product scope.

The baseline test covers Markdown and Org manuscripts plus composed letters in
both formats, inspecting actual ordinary paragraphs and their inherited styles.
It fails on nonzero character-position offsets; genuine note references remain
outside that ordinary-text assertion. Existing font, note and geometry tests
remain preservation checks. The scaling test compares four with forty chapters
in all four manuscript combinations and independently requires one distinct
live contents reference per heading and retained body markers. It rejects style
counts proportional to instances, without prescribing a production algorithm.
Existing differing-role/formatting and page-policy tests protect against merging
incompatible formatting or losing instance-specific layout.

Both tests compiled and failed for the intended defects in `editing-red.log`
(3.144s): body offsets and nine versus eighty-one heading/contents styles.
No execution scaffolding or production implementation was needed. Test-code
readiness passes in `editing-test-readiness.yaml`; other AT/RT records retain
previous candidate evidence. No blocking test finding remains. This PASS approves
the affected tests only. Fresh production/reader verification and a delivery-code
review are required; human editor behaviour cannot be inferred from package XML.
The earlier delivery-code PASS below is superseded for the correction candidate.

The final style-inventory assertion compares all common named style definitions
at four and forty chapters, in addition to counting the known heading/contents
families. This prevents a prefix rename from hiding per-instance growth. Its
initial GREEN is justified preservation after the already-observed RED/fix, not
a fresh implementation claim: both growth cases and accepted deep Org cases
pass together in 2.316s (=editing-style-inventory-final.log=). This final
affected-test refinement is internally PASS for test-code review.


Final source-boundary reassessment before remediation: the accepted Org parser
allows positive heading levels beyond Markdown's six. A new public-command
level-seven preservation test reproduces twenty-one named styles for twenty
headings in both outputs (=editing-deep-org-red.log=, 0.962s). Its content
assertions pass while style reuse fails, isolating an over-narrow classifier
rather than introducing new source syntax. The current-session affected-test
review is PASS for this additional case. The intended correction admits every
already accepted positive outline level to the same role-based style reuse;
existing Markdown levels and instance-specific layout remain unchanged.


Final affected-test review also preserves the genuine note superscript while
rejecting inheritance of legacy body offsets: the updated stable-anchor unit
test fails at 73 percent before correction and passes at its ordinary 33 percent
semantic offset after correction. The entire office package passes in
=editing-office-final.log= (0.858s). A public footnote test previously demanded
one explicit zero-position element; it now accepts the native absent/default
zero too, while retaining the strict superscript assertion and rejecting every
nonzero position. Its focused follow-up passes in 0.846s. These are corrections
to representation-coupled assertions, not waivers of editing or font behaviour.


Affected-test reassessment before final remediation: expanded the scaling fixture
to include part headings and their positioned ODT frames. Actual intended RED in
=editing-parts-red-actual.log= shows ODT named heading/frame/contents styles grow
from seven to twenty-five; DOCX remains five. The first attempted expansion did
not insert its part source and failed an inconsistent fixture expectation; retain
=editing-parts-red.log= as a setup error, not product RED. The corrected test
retains every part/chapter contents target and rejects exposed named page
overrides. The expanded package is internally PASS for this test-code purpose.
Geometry-oracle reassessment: exact native margins and line intervals remain
asserted independently. The old cap-edge-equals-margin assertions rewarded raised
characters. Replaced only those ink-edge assertions with first-line-box bounds;
normal body positioning is now an independent mandatory invariant. The actual
ODT header-baseline regression remains strict and requires native paragraph
placement correction. Human editing-view validation remains outstanding.


# 2026-10-19: Shared native parser delivery-code review (historical candidate)

Review method: internal (current session; not independent)

Gate: delivery-code

Mode: SDLC

Revision: `ec5113f514e3d3757a6e6b3cdca3f043225a796f`, clean worktree at review start.
The production/test candidate is `723734d`; subsequent checkpoints record
operator authorization and settled verification. This entry and the linked
handoff records are review outputs, with no production/test changes.

External review: suspended by standing policy

INTERNAL REVIEW: PASS

No unresolved blocking findings remain. This author self-review permits the
agreed user-validation handoff; it does not establish independent assurance,
Word editing behaviour, merge, installation, release or closure. The pending
checkpoint below is historical and superseded by this result.

## Readiness and criteria

Before beginning this gate, all required checks had finished, AT063.1, all
thirteen RTs and both OTs were GREEN, and the installed validator returned
`schema.valid: true` and `delivery-code.ready: true` in
`.agent/logs/delivery-readiness-current.yaml`. All twelve UTs remain AMBER.
The operator's "run as many as you like" authorization supersedes the local
AT execution-count cap; it does not erase the earlier recorded tracking error.

Applied AUDITS, CODING, TESTING, ARCHITECTURE, SECURITY and the applicable Go,
Shell, documentation, Org, Markdown and Git standards. Reused unchanged
standards already read in full. Reviewed affected tests first, then the complete
shared-parser production delta, relevant unchanged source/configuration,
writer and publication interfaces, and affected product documentation.

| Criterion | Review conclusion and evidence |
|---|---|
| Parser reuse and mode ownership, AC063.3/7/10/11 | Convert's `play.Parse` consumes shared heading/metadata/note recognition and table cells. Script serialization, manuscript prose and composed letters consume the same bounded inline implementation. Rich prose extends that recognition; play events, manuscript semantics and letter recipients retain separate adapters. Fountain and PDF/Typst remain separate. This conclusion includes actual call-path inspection, not only passing tests. |
| End-to-end DOCX independence, AC063.1/12 | Manuscript `runOffice` and native `parseOffice`, letter composition and the existing DOCX serializer start no Pandoc process. Public absent/trapping-Pandoc cases cover Markdown, Org, letters, scripts and dry-run; the novel DOCX regression rejects any call. Native generation and JSON validation require no executable. |
| ODT extension, AC063.13 | Native content and generated reference reach the existing `writeODT` with fixed JSON-to-ODT arguments. Recording/rejecting wrappers permit that writer only. Native layout, fonts, validation and publication follow; there is no source-reader or DOCX intermediate. |
| Same PDF YAML cascade, AC063.4 | Manuscript PDF and office reuse `readConfiguredMarkdown`, `loadConfigWithSource` and `config.Load`; letters pass the existing resolved `ModeLetter` config. Independent family, size and margin expectations cover global/source/local/style and supported CLI winners, with existing inheritance and later-source rejection checks. |
| Native features and fonts, AC063.4/5/6/8 | Existing role, generated matter, spacing, page/field and font regressions pass. Seven-profile comparison and actual selected PNG inspection cover the required visual effects, including corrected ODT note positioning. Accepted ODT boundary, DOCX opening-band and weight mappings remain explicit. Uninstalled-family evidence uses retained files without installation. |
| Novel-scale adequacy | The independent recipe has 10,000 varied-length Unicode body paragraphs and fifty chapters. All four corrected cases pass complete ordered package and refreshed-reader inventories, notes, contents references, body counters, physical header parity and start/middle/end measurements. Retained outputs each contain 654 rendered pages. Selected contents, middle and final-page PNGs were inspected; this is not inspection of every page or human editing evidence. |
| Failure, limits and security, AC063.9 | Source, aggregate letter, semantic, output, ZIP/XML and font limits remain. Public preservation/failure cases and native exact-limit/overflow checks cover their owning boundaries. Fixed subprocess arguments, private staging and per-file atomic publication remain; hyperlinks are not fetched. No dependency change, executable source content or network service was introduced. |
| Maintainability and documentation | Shared grammar stays below mode adapters; it does not introduce a universal semantic model or third-party parser. The long bounded grammar dispatch is the recorded fifty-line tripwire tradeoff, with lists/tables separated. Product help, dependency/format/configuration documents and architecture describe the implemented routes and limitations. |

## Findings, remediation and checks

Preparation findings and their RED/GREEN evidence are retained in the historical
checkpoint below: format diagnostics, empty table cells, ODT note positioning,
letter-wide budgets, remote Org images and the novel recipe. Reassessed those
corrections against the settled candidate. No additional blocking finding was
identified at this gate.

Full regression passes in `full-suite-final.log`; subsequent affected parser,
letter, cascade, novel and rendered-note checks pass in their named logs in
`validation.org`. The full run preceded later refinements, so it is not claimed
as a single run of every final test. Current shared-source coverage is 86.4% of
statements and 81.3% of instrumented lines. Lint, affected race checks and the
unchanged dependency graph's vulnerability check pass. Corrected DOCX reader
acceptance passes in 43.193s and ODT in 39.795s. Their retained artefacts total
42,814,234 bytes and include twenty-four selected PNGs.

The live symlink and SHA-256 were rechecked unchanged. The feature branch is the
handoff candidate. Only the twelve named human UTs remain; none is inferred
GREEN from automated LibreOffice rendering. Editable files and commands are
linked from `validation.org` and the maintained fixture expectations.

# 2026-10-19: Historical shared-parser verification checkpoint, gate pending at that time

Current candidate: `723734d`, with documentation and evidence updates following.

The current delivery-code gate has **no verdict**. The installed validator
reports valid schema, thirteen RTs and two OTs GREEN, twelve UTs AMBER, and
AT063.1 AMBER. The corrected Unicode/varied-length fixture passes both DOCX
reader cases; its two ODT reader cases await the operator's verification-budget
decision. Historical delivery PASS records below do not apply to this candidate.
Do not start or claim the delivery-code gate before that prerequisite is met.

## Verification and review preparation

Method: internal, current session; not independent. External review remains
suspended. Reused the standards already read in full for this delivery: MAIN,
AUDITS, TESTING, CODING, ARCHITECTURE, SECURITY, DOCUMENTATION, GIT, GO and
Org/Markdown documentation rules. Inspected the production delta, written
tests, affected unchanged configuration/publication boundaries and product
documentation. These preparation observations are not an audit PASS.

The implementation uses these actual paths:

- `convert` calls `play.Parse`. Its Org and Markdown adapters now consume
  `source.Recognize` for headings, metadata and notes, and the shared table
  recognition. Script classification remains in `play`; Fountain is preserved.
  The shared office script serializer calls `source.ScriptInline`, which uses
  the same bounded `parser.inline` implementation as prose and letters.
- Manuscript office routing calls `runOffice`, then `parseOffice`,
  `source.Prose`, `Recognize` and `parser.inline`. Mode-owned metadata, private
  sections, heading composition, generated matter and layout remain in
  `internal/manuscript`. DOCX uses the existing native XML/package machinery.
  Constructing and validating the internal JSON-shaped nodes starts no process.
- Letter routing retains its typed recipient parser and substitution. Each
  `letter.Office` composition creates one `source.LetterParser`; its `Parse`
  method shares `parser.inline` and carries budgets across all recipient blocks.
  It maps literals to the existing letter role without rediscovering YAML.
- ODT passes the native content through `Prose.Export` and the shared
  `writeODT`: JSON to ODT, native generated reference, bounded configured writer,
  native layout, font packaging, validation and existing atomic publication.
  No source-reader, default-reference or intermediate DOCX call is needed.
- PDF and office manuscripts retain `readConfiguredMarkdown` and
  `loadConfigWithSource` into `config.Load`; letters use the command's existing
  resolved `ModeLetter` configuration for both renderers. No office cascade was
  introduced. PDF/Typst parsing and its tool requirements remain unchanged.

The grammar dispatch functions exceed the coding standard's fifty-line
tripwire. They implement one bounded source scan, with explicit node budgets,
literal protection and format/mode distinctions; list and table parsing are
separate helpers. Keeping that dispatch visible avoids a speculative parser
framework or a universal document model. This is a recorded maintainability
tradeoff to assess at the eventual gate, not evidence that tests alone establish
architecture conformance.

## Findings and remediation evidence

| Finding | Correction and evidence |
|---|---|
| New shared-parser error lost ODT diagnostic context | Preserve native DOCX context at the script adapter; the existing ODT error adapter translates it. The previously failing input-limit regression passes in 5.046s. |
| Table splitting dropped an empty trailing cell | Strip one structural pipe at each edge, preserving empty cells; native RED/GREEN and existing play tests cover the boundary. |
| Rendered ODT note anchor appeared below the body baseline | Native ODF anchor styling includes the resolved body shift; a repeated-style test prevents duplicate definitions. Both formats pass and final typography PNGs were inspected. The retained assertion reuses existing renders. |
| Letter budgets reset for each composed paragraph | Own one shared-parser state per recipient. Byte, block and inline exact-limit/overflow tests have intended RED-to-GREEN evidence, plus public letter follow-ups. |
| Unlabelled remote Org images could become ordinary links | Reject image sources with URL query/fragment suffixes while preserving explicitly labelled hyperlinks. Native RED/GREEN and affected public parser cases pass. No target is fetched. |
| Novel fixture lacked required Unicode and varied paragraph lengths | Independent inventory test reproduced both omissions. Corrected structural matrix and two DOCX reader cases pass; ODT reader evidence is pending. |
| AT conversion allowance exceeded | Twenty attempts are recorded, including both four-case RED runs. The agent initially omitted one RED run from its count. Further AT conversions are stopped pending the requested allowance of twenty-two; this does not retrospectively authorize the overrun. |

Full regression passed in `full-suite-final.log` (application 565.617s).
The W063 group run passed in 410.706s; later affected parser, letter, cascade,
fixture and rendered-font checks pass separately. Current shared-source
coverage is 86.4% of statements and 81.3% of instrumented source lines.
Lint, source/office/letter race checks and the unchanged dependency graph's
vulnerability check pass. `validation.org` records the exact affected-check
boundary, artefact paths and preserved failed/setup observations.

The feature worktree is isolated and the live binary hash is unchanged.
No installation, release, merge, operator UT or issue closure is claimed.

# 2026-10-19: Native parser correction definition review

Review method: internal (current session; not independent)

Gate: definition

Revision: `aa5def7` plus this specification, ledger and validation amendment.
Production and executable tests are unchanged.

External review: suspended by standing policy

INTERNAL REVIEW: PASS

This result concerns the revised definition only. Operator sign-off is pending.
The earlier delivery PASS below does not establish parser reuse or end-to-end
DOCX independence from Pandoc. Historical evidence remains preserved.

## Reviewed scope and findings

Reviewed the specification, its test definitions and solution design against the
operator's explicit requirement to extend the `convert` parser. Applied the
project profile and named architecture/requirements authorities, MAIN, ISSUES,
ARCHITECTURE, TESTING, CODING, SECURITY, AUDITS and the applicable Go and
Org/documentation standards, reusing prior full readings from this session.
Inspected `internal/app/convert.go`, `internal/play/parser.go`, manuscript office
routing and the typed letter adapter to distinguish current behaviour from the
required design. Historical definition feedback and accepted layout limitations
remain in force except where expressly superseded by this correction.

| Finding | Amendment and reassessment |
|---|---|
| Manuscript DOCX explicitly allowed Pandoc readers | AC063.12 prohibits every invocation, including metadata, canonicalization, dry-run and probes; RT063.1 now requires absent-Pandoc and executable-trap cases. |
| Exporter reuse did not require parser reuse | AC063.11 names native convert parsing; design extends its source recognition with distinct mode adapters. RT063.13 covers shared parsing and mode semantics; implementation review must trace actual callers and reject nominal wrapper reuse. |
| ODT source parsing and output generation were conflated | AC063.13 and the five-step ODT design separate native parsing, JSON projection, generated reference, configured writer, native postprocessing, fonts and publication. Recording-wrapper tests reject source-reader and DOCX intermediate calls. |
| Tiny cases missed novel-scale token expansion | Named synthetic 120,000-word/10,000-paragraph DOCX test retains native span/count limits and checks ordered content with Pandoc unavailable. It does not claim the real manuscript has passed. |
| Old approvals and GREEN states could imply current acceptance | Renewed sign-off is PENDING, ledger is REVIEW, changed dependency/CLI tests and new parser group are AMBER; earlier results are retained as history. |
| Synopsis and handoff retained an obsolete external-review requirement | Current review references now follow the standing internal-only policy; harness-owned evidence remains untouched. |

The existing YAML fidelity criteria, twelve named user tests, font policy,
publication safeguards and accepted page/weight limitations remain. PDF/Typst
parsing and dependencies are preserved. Shared parsing for ODT is explicit in
this proposed revision; Pandoc remains its configured output writer.

## Checks and evidence boundary

Reviewed the full specification for remaining current reader-only allowances,
AC-to-test links, distinct mode semantics and agreement between Synopsis and
design. Historical reader-only statements are labelled superseded. Added AC and
RT identifiers were checked against current files and Git history before use.
Authored text was processed through `sanitize`; whitespace and Org schema checks
passed for this amendment: schema valid, 13 ACs, 13 RTs, two OTs and twelve UTs. Test-code readiness correctly remains false because revised RT063.1 and new RT063.13 are AMBER, not executed evidence. Executable test results are not required for
this definition review and none are claimed. Current source still invokes
Pandoc when parsing manuscripts; implementation and fresh test-code/delivery
reviews remain outstanding. There is no independent audit or operator approval.

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

# 2026-10-19: Novel-scale specification and test-definition review

Review method: internal (current session; not independent)

Gate: definition

Revision: `82ea97be60ac472f879b3d525c484e6cfdbadee5` plus the specification and
validation amendments recorded here. Production and executable tests are unchanged.

Reviewed specification SHA-256:
`906e2cc7ba03a2cc159ef72d2e3b2bbaba2261eded83a183f2f52e3bc7a47846`.

Reviewed validation SHA-256:
`2e8bd9ad64480915f4138c60cc4ab3e2e5e048966cc82dae4b10b2f700e1a775`.

External review: suspended by standing policy

INTERNAL REVIEW: PASS

This is a **definition-only result under the explicitly required AT category**.
It is not an automated schema pass, implementation admission, test result or
operator sign-off. The installed validator rejects AT; that limitation remains
visible below and must be resolved before relying on automated delivery readiness.

## Criteria and evidence

Applied MAIN, ISSUES, TESTING, AUDITS, ARCHITECTURE, CODING, SECURITY, GO,
DOCUMENTATION, ORGMODE, ORG-SCHEMA, MARKDOWN and GIT, read in full in this retained
session. Reviewed the complete specification, including its preserved history,
and the affected validation records. Consulted the project profile, README,
product vision, architecture, W063 work-ledger entry, manuscript source contract,
letter schema, command help and font policy. The ledger and implementation were
examined in the relevant sections, not treated as an exhaustive code audit.
No processed-feedback file was found in the work-item directory.

Inspected the convert parser call, manuscript office parsing, letter adapter,
shared DOCX/ODT export and AST bounds, dependency tests and representative
generated-matter/rendered tests. These establish the current baseline and its
limitations, not successful delivery of the proposed parser correction. Historical
external evidence remains linked through [audits.yaml](audits.yaml); it was not
changed or presented as a current verdict.

## Findings, amendments and reassessment

| Finding | Amendment and reassessment |
|---|---|
| The named scale case specified DOCX only | RT063.13 - Shared parser and mode semantics now explicitly covers Markdown/Org to DOCX/ODT, with real configured ODT writing and trapped DOCX dependencies. |
| Novel-scale support appeared only in a test definition | AC063.3 - Manuscript semantic content now requires the novel-scale content and usability outcome, traced to RT, AT and UT evidence. |
| Package inspection plus small rendered specimens could miss unusable long documents | AT063.1 - Novel-scale rendered manuscripts requires full reader rendering, complete ordered content, refreshed contents references and distributed geometry checks for all four cases. UT063.1 - Editing manuscripts adds human inspection, edits, refresh, save and reopen of those same outputs. |
| Repeatable ticket-scoped rendering must not be labelled OT | The new rendering check is AT, outside the retained regression pack, with intended RED/GREEN, explicit local bounds and fourfold conversion/render execution headroom. Existing OT evidence remains preserved. |
| Old successes could be mistaken for the expanded outcome | A dated validation note separates historical small-fixture evidence from unexecuted scale requirements. The changed RT and UT and the new AT remain AMBER. |

The revised definitions specify independent expectations, exact matrix rows,
complete-content assertions, field refresh, known layout exceptions, error
outcomes and candidate-specific evidence. Reader rendering is separate from Word
behaviour and human editability. The scale fixture remains synthetic and bounded;
no private manuscript, external service or new test framework is required.

The solution design retains native convert-parser reuse below distinct mode
adapters, native DOCX throughout, and the existing configured ODT writer after
native parsing. Parser reuse requires call-path review as well as behavioural
tests. The amendment adds verification, not another exporter or semantic model.
It preserves source/configuration contracts, output safety and the existing
PDF/Typst boundary. No additional product-design blocker was identified.

## Checks and limitations

- `sanitize` processed the amended specification and validation text without
  changes. `git diff --check` passed.
- The installed `sdlc-validate --readiness-check=test-code` returned exit 2 with
  exactly two `INVALID_ID` diagnostics: the AT heading in each Org file. Its
  documented ID allowlist is AC/RT/OT/UT. The explicit instruction to include AT
  governs the definition; relabelling it OT or removing it would conceal the gap.
- Manual record comparison includes **13 ACs and 28 tests: 13 RTs, one AT,
  two OTs and twelve UTs**. The new AT has matching definition/result identities,
  a unique custom anchor, reciprocal AC links and an AMBER reason. The validator's
  27-test inventory excludes AT and is not a complete readiness assessment.
- This internal definition review accepts the explicit AT extension. It does
  **not** waive the implementation readiness checks, authorize editing the
  external SDLC, or report schema validity. Tooling support for AT remains an
  implementation-admission limitation. Revised RT/AT cases are also unexecuted.
- No product tests, document conversions, reader checks or user tests were run
  for this documentation amendment. Test implementation and execution belong to
  delivery after sign-off. The earlier internal definition result remains
  historical for its own candidate; this entry assesses the amended candidate.

# 2026-10-19: Corrected AT validator reconciliation

Review method: internal (current session; not independent)

Gate: definition

Revision: `4bc7f73` (manuscript-scale test definitions) plus the specification,
validation and review-record reconciliation in this entry.

Specification SHA-256:
`fee98f9912e516c4c9e1efa868caf63a1a07e4699a116e82162905e3404fd040`.

Validation SHA-256:
`7ac88e7db2cab508d1870c44511e1acb5a05003817dcb83dffd0843336312c54`.

External review: suspended by standing policy

INTERNAL REVIEW: PASS

Read the corrected TESTING and ORG-SCHEMA standards and validator help in full.
The four-category contract now supports the existing definition directly.
Reviewed the documentation delta against the preceding definition review:
requirements, architecture and test expectations are unchanged. The handoff now
explicitly names the AT/RT package as its first delivery step. Earlier tooling
failures remain historical evidence and are superseded by this check.

The validator reports **schema valid**, 13 acceptance criteria and all 28 tests:
one AT, thirteen RTs, two OTs and twelve UTs. Its test-code check returns exit 1
only for the three unexecuted automated records: RT063.1 - Output selection and
dependencies, RT063.13 - Shared parser and mode semantics, and AT063.1 - Novel-scale
rendered manuscripts. It reports no schema error or missing category.

The earlier AT tooling limitation no longer applies. **Definition PASS does not
establish test-code readiness**: those records remain AMBER. No product tests
were run and no execution state or operator sign-off changed. Sanitization and
whitespace checks pass for this documentation reconciliation.

# 2026-10-19: Frontmatter configuration synopsis clarification

Review method: internal (current session; not independent)

Gate: definition

Revision: `2f648e9` (AT validator reconciliation) plus the synopsis clarification;
specification SHA-256 `25740ce30d60838d52716172a5c06b05b4589cc9516693f3b65dce5f8f8d2460`.

External review: suspended by standing policy

INTERNAL REVIEW: PASS

The synopsis now explicitly includes the first resolved Markdown manuscript's
`folio:` and `render:` frontmatter mappings, above global YAML files and below
local YAML files and CLI overrides. Checked against manuscript source-configuration
extraction, command help, RT063.4 - Configuration and role geometry and
UT063.10 - Metadata, style and configuration precedence. This clarifies the
existing requirement; it adds no source format, configuration layer or test
obligation. Earlier definition findings remain addressed. No implementation,
test execution or sign-off is implied.

# 2026-10-19: PDF YAML-cascade alignment review

Review method: internal (current session; not independent)

Gate: definition

Revision: `7956961` (frontmatter synopsis clarification) plus the specification
and validation amendments requiring the PDF renderer's YAML cascade.

Specification SHA-256:
`a8ba6fb00cb6d35f26abde7cbfb542b8951701d36556d3cd4a88e3635d379496`.

Validation SHA-256:
`761fd821e1b9e8e77acd92c448f3b2b7e169a50489a8b125e70c5b8ff5cd5085`.

External review: suspended by standing policy

INTERNAL REVIEW: PASS

Reviewed the changed Synopsis, AC063.4 - Resolved typography and block geometry,
RT063.4 - Configuration and role geometry, shared-cascade design and both office
workflow entry steps against the existing configuration authority and inspected
PDF/office callers. The named cascade and first-Markdown-source boundary agree
across those sections. The design reuses manuscript source extraction and the
existing shared loader; letters use the command's already-resolved ModeLetter
configuration. No office-specific precedence or additional configuration source
is introduced. The expanded behavioural comparison has independent expectations
and preserves historical evidence under its AMBER execution record.

The schema check accepts all 13 criteria and 28 tests. Test-code readiness remains
false for unexecuted RT063.1 - Output selection and dependencies, RT063.4 -
Configuration and role geometry, RT063.13 - Shared parser and mode semantics,
and AT063.1 - Novel-scale rendered manuscripts. This is expected before the
test-writing stage and does not block definition sign-off. Sanitization and
whitespace checks pass. No product tests were executed; operator sign-off remains
pending, and implementation has not started in this workflow.

# 2026-10-19: Shared native parser test-code review

Review method: internal (current session; not independent)

Gate: test-code

Mode: SDLC

Revision: `e29559d` (delivery admission) plus the parser/dependency, novel fixture,
cascade-parity and tagged acceptance tests; production code is unchanged.

External review: suspended by standing policy

INTERNAL REVIEW: PASS

Applied the signed-off specification and AUDITS, TESTING, CODING, ARCHITECTURE,
SECURITY, GO and Org evidence standards. Reviewed the changed test package and
existing helpers in this session. The validator reports valid schema and
test-code readiness for all 28 specified records. The new dependency/scale RTs
and AT have intended observed RED; cross-format cascade and mode semantics
are initially GREEN preservation checks (4.976s and 2.567s respectively).

The traps operate only in child executable paths. DOCX asserts no Pandoc call,
ODT delegates permitted writer calls to the real tool, and the independent novel
generator supplies all 10,000 expected paragraphs, fifty chapters and distributed
structured specimens. Native package checks and the tagged reader acceptance
test use public commands; the latter stays outside make test. Full reader
assertions are written but not claimed exercised by a pre-render RED failure.

Review corrected fixture setup errors (copyright body type, retired font key,
Org cast heading and optional native bold flags) rather than counting them as
product RED. The AT font check now rejects arbitrary fallback families. The
complete package has no target-production scaffolding. Existing requirement
groups retain their baseline tests and history; new affected results will be
rerun after implementation. Parser sharing still requires implementation call-
path review, not matching output alone. No independent audit is claimed.
