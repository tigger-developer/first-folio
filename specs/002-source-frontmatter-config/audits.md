# Audit Record: Source Frontmatter Configuration

## Specification audit

Audit: `audit-spec`

Candidate:

- `spec.md`: `7a9f2764ee9e34ce3d5ba6df676d8969054753d12a40cc8fc5eb471099b8ce05`
- Related configuration contract `presets/british.yaml`:
  `18cb3781fe5da95ed8009084f345aa59379894812272bc5fbbc7e031c7b5ff2a`

### Attempt 1

- Candidate `spec.md`:
  `7b3844eba60b905cbb25548ec790a8f85b5428de205c0a8c0a06429c2355578b`
- Harness: Hermes
- Provider: `nous`
- Model: `z-ai/glm-5.3-flash`
- Verdict: **FAIL**
- Blocking finding: FR-004 - apply explicit precedence omitted AC13.1-AC13.3 -
  manuscript local-directory resolution and did not state how the source layer
  composes with single-input and multi-input file discovery.
- Blocking finding: FR-006 - select styles from source did not state whether a
  source-selected style selects global and local style-specific sibling files.
- Advisory: FR-009 - reject invalid source configuration qualified diagnostic
  paths with an unbounded "where available".
- Advisory: SC-002 - resolve every precedence conflict did not enumerate the
  same layers as FR-004 - apply explicit precedence.
- Advisory: The dependency on Draft Feature 001 - Unified Font Configuration
  was not explicitly conditional on its approval.
- Remediation: AC13 directory rules and their implementation conflict are
  explicit; source styles select sibling files; diagnostics, success coverage,
  and the approval dependency are bounded.

### Attempt 2

- Candidate `spec.md`:
  `222cce47e140bc177493655525ee8c4f38816f9408633b42a7daa99d29d4740f`
- Harness: Hermes
- Provider: Not reported
- Model: Not reported
- Result: **Harness failure, not an audit verdict**
- Evidence: `sdlc-audit: validating audit report: FAIL verdict requires a
  blocking finding and no condition`
- Context: Remediated specification and the same bounded authority set as
  Attempt 1.

### Attempt 3

- Candidate `spec.md`:
  `222cce47e140bc177493655525ee8c4f38816f9408633b42a7daa99d29d4740f`
- Harness: Hermes
- Provider: `nous`
- Model: `z-ai/glm-5.3-flash`
- Verdict: **FAIL**
- Blocking finding: FR-011 - confine the feature to Markdown manuscripts did not
  define the observable treatment of YAML-shaped configuration lines in a
  Markdown stage play.
- Advisory: FR-009 - reject invalid source configuration used an undefined
  diagnostic-path qualification.
- Advisory: The multi-file scenario used "invocation directory" while FR-004 -
  apply explicit precedence used "process working directory".
- Advisory: FR-003 - preserve metadata omitted AC2.5 - configuration metadata
  precedence.
- Advisory: FR-013 - preserve regression protection omitted Feature 001's
  reconciliation lineage for the five source-directory lookup tests.
- Advisory: SC-005 - fail without misleading output used an undefined sampled
  population.
- Remediation: Non-manuscript Markdown handling, diagnostic classes, working-
  directory terminology, metadata precedence, regression lineage, and the
  complete failure population are now explicit.

### Attempt 4

- Candidate `spec.md`:
  `6b65c18cac7445dfdcea7f14abf91f4957467ce985f5e904524da3eda5120145`
- Harness: Hermes
- Provider: `nous`
- Model: `z-ai/glm-5.3-flash`
- Verdict: **PROVISIONAL**
- Condition: FR-009 - reject invalid source configuration duplicated the
  malformed-YAML source-file identification rule.
- Advisory: FR-002 and FR-003 did not explicitly preserve tolerance of
  unrelated top-level frontmatter keys when configuration namespaces coexist.
- Advisory: SC-003 used an unenumerated maintained-metadata test population.
- Remediation: FR-009 now states each diagnostic guarantee once; FR-002
  preserves AC2.3 and AC10.5 top-level tolerance; SC-003 identifies the
  maintained Markdown frontmatter metadata cases in
  `internal/manuscript/manuscript_test.go`.
- Condition verification: Candidate
  `7a9f2764ee9e34ce3d5ba6df676d8969054753d12a40cc8fc5eb471099b8ce05`
  contains no duplicated malformed-YAML clause and retains both required
  diagnostic classes and the non-zero, pre-output failure boundary.

## Current gate

- Effective verdict: **PASS** after satisfaction of Attempt 4's condition
- Audit attempts: **4**
- Consequence: The specification is eligible for planning, subject to approval
  of Feature 001 or another approved specification supplying the canonical
  schema and validation contract.
