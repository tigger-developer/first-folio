<!-- SYNC IMPACT: 0.1.0 -> 0.1.1 | Principles: NONE | Added: NONE | Removed: Perl Standards | TODOs: Human ratification; reconcile requirement targets for font-style and configuration-lookup regression groups -->

# First Folio Constitution

<!-- SDLC-GENERATED-SCAFFOLD: editable until ratification. -->

## Engineering Standards

This project MUST comply with the following canonical standards. The standards are referenced, not copied; load only those relevant to the current operation.

- **Universal engineering behaviour:** `~/.agents/sdlc/MAIN.md`.
- **Specification and requirement quality:** `~/.agents/sdlc/ISSUES.md`.
- **Implementation and design:** `~/.agents/sdlc/CODING.md`.
- **Testing and evidence:** `~/.agents/sdlc/TESTING.md`.
- **Security and vulnerability checking:** `~/.agents/sdlc/SECURITY.md`.
- **Independent audits:** `~/.agents/sdlc/AUDITS.md`.
- **Paired development:** `~/.agents/sdlc/PAIRING.md`.
- **Documentation:** `~/.agents/sdlc/DOCUMENTATION.md`.
- **Source control:** `~/.agents/sdlc/GIT.md`.
- **Go Standards:** `~/.agents/sdlc/technologies/GO.md`.
- **Shell Standards:** `~/.agents/sdlc/technologies/SHELL.md`.

A deviation MUST name the standard, reason, risk, and approving authority. Silence is not a deviation.

The adopted SDLC revision is `99b9509c548d10cb1613b128b1c8986a69f19c9a`.

## Specification and Evidence

No implementation may begin without a defined specification. Before drafting a brownfield specification or design, the author MUST examine the current requirement and design authorities, relevant historical work records, the maintained regression test pack, and the affected implementation. The resulting artefact MUST identify what existing behaviour and decisions it preserves, changes, supersedes, or leaves unaffected. Tests MUST be used to trace actively protected behaviour to its originating requirements and compatibility constraints. Tests and code are implementation evidence; they do not approve requirements. Project documentation and the active specification MUST be updated when delivered behaviour or ownership boundaries change.

## Specification Baseline

**Project classification:** Brownfield

### Requirement Authority

`docs/ACs.org` is the sole requirement authority for requirements established
under the completed legacy ticket-led process. Approved Spec Kit feature
specifications govern requirements established or changed through Spec Kit. A
later approved Spec Kit specification may supersede a legacy requirement only
explicitly and MUST preserve its lineage.

### Migration Record and Historical Context

`docs/ticket-migration.org` is the disposition index for the completed SDLC v1
legacy-ticket migration. `docs/archive/migrated-tickets/` is its lossless
historical source. The index, archived tickets, ticket bodies, and comments
provide disposition, provenance, and rationale only; they are not current
requirement, acceptance-criterion, or design authorities. Scope classified as
defective, undelivered, or abandoned requires an approved Spec Kit
specification before implementation.

### Design Authority

`ARCHITECTURE.md` is the current architecture and design authority. Approved
Spec Kit plans govern their bounded design deltas. Archived implementation
plans are historical provenance only and are not design authority.

### Regression Evidence and Traceability

The maintained regression pack is the Go test suite under `cmd/` and
`internal/`, selected by `make test`. `docs/ACs.org` records legacy requirement
traceability; test names and adjacent descriptors record additional executable
traceability. Ten maintained tests covering page-header and page-footer
font-style behaviour and source-directory configuration lookup do not yet have
unambiguous current requirement targets. This conflict blocks ratification
until the human project owner reconciles their authority without inventing a
requirement.

Tests and code provide evidence of implemented behaviour. They do not approve requirements.

### Precedence and Supersession

Authority is concern-specific:

1. Taḋg O'Brien, the human project owner, controls ratification, amendments,
   and standards deviations.
2. After ratification, this constitution and its selected engineering standards
   govern engineering and project governance.
3. `docs/vision.md` governs durable product purpose and policy.
4. Approved requirement authorities govern observable behaviour within their
   stated scopes.
5. `ARCHITECTURE.md` and approved Spec Kit plans govern technical choices
   within approved requirements.
6. Operational, testing, migration, and user documentation govern their
   respective procedures and evidence. An external integration contract is
   authoritative only within its named ownership boundary.
7. Code and tests record implemented state and evidence; they do not approve
   requirements.

A later authority supersedes an earlier authority within the same concern only
when the supersession is explicit and preserves lineage. Historical records do
not override current authorities. Unresolved cross-concern conflicts require a
decision from the human project owner.

## Mandatory Independent Audits

For staged Spec Kit delivery, each audit MUST run in a fresh agent context that did not author the artefact and MUST emit the exact structured verdict required by its skill. PASS may retain advisories. A PROVISIONAL verdict becomes effective PASS only through the exact condition receipt defined by the audit standard. On FAIL, the author remediates and a fresh independent audit runs. The next stage MUST NOT begin until the required audit records effective PASS.

1. Specification and clarification require `audit-spec` PASS before planning.
2. Plan and design require `audit-design` PASS before test design and tasks.
3. Test design and traceability require `audit-tests` PASS before implementation.
4. Implementation requires `audit-code` PASS before completion or convergence.

Record each audit name, auditor provider and model, artefact revision, exact verdict, findings, and superseding rerun in the active feature's `audits.md`. `speckit-analyze` is a consistency check and does not replace an independent audit.

When the operator explicitly selects paired development under `~/.agents/sdlc/PAIRING.md`, its change-scoped closure and user-validation contract replaces these staged transitions for that change. Engineering standards and applicable audit requirements remain mandatory.

## Project-Specific Principles

### I. Semantic Fidelity and Authorial Integrity

First Folio MUST preserve the semantic content and authorial intent of source
material across conversion and rendering boundaries. A change that permits
silent semantic loss or transfers editorial authority from the writer to the
tool requires a constitutional amendment.

### II. Open Sources and Scriptable Core

User-owned, open plain-text sources and a stable scriptable command-line core
MUST remain the foundation of First Folio. Additional interfaces may wrap that
core, but MUST NOT replace it as the automation boundary or make a proprietary
authoring environment authoritative.

## Project Ownership and Architecture Boundaries

First Folio owns its application semantics and the contracts it publishes for
its command-line interface, configuration, supported source formats, and
generated documents. As the provider, it defines, implements, evolves, and MUST
honour those contracts.

First Folio consumes external contracts, including Typst, Pandoc, and the
supported Fountain boundary. Those contracts are authoritative only within
their owners' integration boundaries; they do not approve First Folio product
requirements or transfer First Folio's provider obligations.

## Governance

This constitution governs project specifications, plans, tasks,
implementation, and review after ratification. Before ratification, this
scaffold has no authority. Taḋg O'Brien is the human authority for initial
ratification and every later amendment. No standards deviation is approved.

Compliance review MUST report the applicable constitutional principles, every
approved standards deviation, and every unresolved constitutional conflict.
Changes to governance require human approval and MUST explain compatibility
and migration effects.

Before ratification, draft versions use pre-1.0 numbering and revisions are not
amendments. Initial ratification sets version `1.0.0`. After ratification,
MAJOR removes or incompatibly redefines governance, MINOR adds or materially
expands governance, and PATCH clarifies governance without changing meaning.

### Ratification Blockers

- Human ratification has not occurred.
- The page-header and page-footer font-style regression group and the
  source-directory configuration lookup regression group, comprising ten
  maintained tests, lack unambiguous current requirement targets.

**Version**: 0.1.1 | **Ratified**: UNRATIFIED | **Last Revised**: 2026-09-04
