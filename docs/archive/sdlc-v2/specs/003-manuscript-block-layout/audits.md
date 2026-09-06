# Audit Record: Manuscript Block Layout

## Code audit

### Attempt 1

- Candidate diff:
  `852879ea278bdda29e5c4d10b7fc8235d29919f46ec510e80e0422c752895ab1`
- Harness: Hermes
- Provider: `openai-codex`
- Model: `gpt-5.6-luna`
- Verdict: **FAIL**
- Blocking finding: User-configured quoted-block font values were interpolated
  into Typst without complete validation and output-context escaping.
- Remediation: The configuration boundary now validates spacing, size, weight,
  stretch, style, and letter spacing; the renderer escapes font family names,
  emits only validated raw values, and maps named or numeric weights safely.
  Failure-path tests require a path-specific error, non-zero result, and no
  output artefact.

### Attempt 2

- Candidate diff:
  `3c58dfc89ad60c58b944f1345bf702bd23fdf7c8c6f6cfc699b531e36e75ee0b`
- Harness: Hermes
- Provider: `openai-codex`
- Model: `gpt-5.6-luna`
- Verdict: **FAIL**
- Blocking finding: String-only quote-font decoding could not distinguish an
  omitted property, which must inherit, from an explicitly empty or null
  property, which must fail.
- Remediation: The shared `FontConfig` now records YAML property presence,
  inherits only omitted values, and preserves explicit empty or null values for
  validation. Tests cover an omitted font block, a style-only override, every
  explicit empty font property, a null property, and empty spacing controls.

### Attempt 3

- Candidate diff:
  `72ace5559e985319fac209a922a4716bbd25a841c362380fc5c2f3b3f1d9c64b`
- Harness: Hermes
- Provider: `openai-codex`
- Model: `gpt-5.6-luna`
- Verdict: **FAIL**
- Blocking finding: `strconv.ParseFloat` accepts non-finite `NaN` and `Inf`
  stretch values, allowing them to reach generated Typst.
- Remediation: Stretch validation now rejects `NaN`, positive infinity, and
  negative infinity explicitly. Regression cases cover `NaN` and `Inf`.

### Attempt 4

- Candidate diff:
  `30d23d8c386eb8d98d0aad9c09578ec2b21c588fd43be9bb1c3ea15c4e781036`
- Harness: Hermes
- Provider: `openai-codex`
- Model: `gpt-5.6-luna`
- Verdict: **FAIL**
- Blocking finding: Pre-normalization spacing validation depended on the
  embedded preset already supplying both documented defaults.
- Remediation: Presence-aware spacing values now distinguish omission from an
  explicit empty or null value. Omitted values default before validation;
  explicit invalid values remain available for path-specific rejection. Tests
  cover operation without preset values, mapping-shaped values, numeric font
  weight, and plain-number stretch.

### Attempt 5

- Candidate diff:
  `f0f41bb1d3cd4977a96c8a6be14fd65bb36131aabf502586fb9cf0e5ea28d316`
- Harness: Hermes
- Provider: `openai-codex`
- Model: `gpt-5.6-luna`
- Verdict: **PASS**

### Post-PASS delta audit 1

- Candidate diff:
  `2367d61ae2db27c9c7be786bcdc9da3680207cdc8dcabdb3ed38fd6da05afcec`
- Scope: The template accessor required by the presence-aware spacing values
  was found unstaged during final integrity checks and audited separately.
- Harness: Hermes
- Provider: `openai-codex`
- Model: `gpt-5.6-luna`
- Verdict: **PASS**

### Indent amendment audit 1

- Candidate diff:
  `6dd5a52047dd399e0a8b95a3d7803711c93e7a9541e1687be5552daaf7c715f8`
- Scope: Add separate whole-block left indents for manuscript blockquotes and
  fenced code within the existing emergency feature.
- Harness: Hermes
- Provider: `openai-codex`
- Model: `gpt-5.6-luna`
- Verdict: **PASS**

## Complete-line code-span amendment

### Test audit harness incident 1

- Scope: New complete-line code-span regression coverage.
- Result: The harness rejected its own output because a FAIL verdict contained
  no blocking finding.
- Consequence: No audit verdict was recorded from this invocation.

### Test audit attempt 1

- Verdict: **FAIL**
- Blocking findings: The selected artefact omitted existing spacing, font,
  invalid-value, indentation, and compatibility tests from the audit context;
  alternate complete-line source forms were covered only at the helper
  boundary; normal-lifecycle PENDING records and planning rationale were absent.
- Remediation: The audit was rerun with the relevant maintained test files and
  emergency authority. Whitespace, backtick-delimiter, existing-fence,
  tilde-fence, unclosed-span, Markdown, and Org cases were expanded.

### Test audit attempt 2

- Verdict: **FAIL**
- Blocking findings: Lower-side precise spacing, each independent quote-font
  inheritance path, explicit `regular` and `oblique` style mapping, deterministic
  rendered-layout evidence, application-dispatch coverage, and the normal
  PENDING lifecycle record were requested.
- Remediation: Retained regressions now cover both precise-spacing sides, every
  inherited font property, all public quote styles, successful application
  dispatch, and application-level invalid-indent failure. Poppler bounding-box
  evidence now records exact block and prose positions.

### Test audit attempt 3

- Verdict: **FAIL**
- Remaining blocking findings: The auditor required retroactive
  pre-implementation PENDING entries, public-executable tests for all existing
  emergency requirements, broader preservation comparisons, expanded Org
  permutations, one-off classification rationale, and operational-path
  rationale.
- Disposition: `BYPASS-GATE-7` explicitly skips test audits under the emergency
  standard. The failed report is retained as advisory evidence. No
  pre-implementation record has been fabricated retrospectively, and demands
  beyond the bounded complete-line classification were not made delivery gates.

### Code audit 1

- Candidate diff:
  `393973073fe956c5f4c0181dc24d7b18ba31551fe6355d9d93da462dcfdaf5fb`
- Scope: Promote complete-line Markdown code spans, and Org verbatim spans after
  canonical Markdown conversion, to existing code-block layout while preserving
  mixed-content inline code and existing fences.
- Harness: Hermes
- Provider: `openai-codex`
- Model: `gpt-5.6-luna`
- Verdict: **PASS**

### Boundary clarification audit 1

- Candidate diff:
  `46bb31a075dd6f563877aeb347bd623ca43555e52932c95a488ae86cccc11c73`
- Scope: Correct the negative example so leading prose determines inline
  classification while `#because...` remains inside the code delimiters.
- Harness: Hermes
- Provider: `openai-codex`
- Model: `gpt-5.6-luna`
- Verdict: **PASS**
- Advisory: Add the equivalent Org mixed-content case containing `#`.
- Remediation: The Org regression was added through the canonical conversion
  boundary.

### Boundary clarification audit invocation incident

- Result: A malformed command included stray positional arguments and was
  rejected before the audit harness ran.
- Consequence: No audit verdict was recorded from this invocation.

### Boundary clarification audit 2

- Candidate diff:
  `e66ce92b71c5d57d610ee651dce507b2960af160f7b92d7cd3650e07cc7a9db7`
- Verdict: **FAIL**
- Blocking finding: The preprocessor excluded tab indentation and more than
  three leading spaces even though the specification ignores all surrounding
  source-line whitespace.
- Remediation: The arbitrary indentation exclusion was removed. Markdown
  integration and helper regressions now cover four-space and tab-indented
  complete-line spans.

### Boundary clarification audit 3

- Candidate diff:
  `e33af353c2b01082cb016c6b67b0aefdc7f6c4a900c6ce126312d533e5102afc`
- Harness: Hermes
- Provider: `openai-codex`
- Model: `gpt-5.6-luna`
- Verdict: **PASS**

## Post-block paragraph indentation amendment

### Code audit attempt 1

- Candidate diff:
  `d823fd71ab945ece85a86cdb66fae3d4969ba315aa4674ea52d85fa1ef84a9c3`
- Harness: Hermes
- Provider: `openai-codex`
- Model: `gpt-5.6-luna`
- Verdict: **FAIL**
- Blocking finding: A chapter-opening footnote did not consume the opening
  position, so later prose could receive the flush-left override.
- Remediation: The renderer consumed that position for footnotes and a focused
  regression reproduced the failure before passing with the correction.

### Code audit attempt 2

- Candidate diff:
  `8790bb7ed7108447371a39bdb8537a80d3cb395163a1a40ffe269280968c29a6`
- Harness: Hermes
- Provider: `openai-codex`
- Model: `gpt-5.6-luna`
- Verdict: **FAIL**
- Blocking finding: Generated Typst was classified through an incomplete
  string-prefix heuristic, allowing an unlisted structural block to be treated
  as chapter-opening prose.
- Remediation: The prefix heuristic was removed. Pandoc's source block model
  now identifies only a direct `Para` or `Plain` after a chapter heading as
  chapter-opening prose. A raw Typst grid regression reproduces the former
  failure.

### Code audit attempt 3

- Candidate diff:
  `c9f6c848a9f0872ee64562bfa9f9e62afc80780239a03e7b2fcb8857fb88e599`
- Harness: Hermes
- Provider: `openai-codex`
- Model: `gpt-5.6-luna`
- Verdict: **PASS**
- Advisory: Positively verify configured indentation after code, quote, and
  structural chapter openings through generated Typst or PDF coordinates.
- Disposition: Post-audit one-off PDF validation recorded the chapter-opening,
  post-code, quote, and post-quote coordinates in `validation.md`.

## Current gate

- Effective verdict: **PASS**
- Consequence: Final verification may run against the audited implementation,
  including the corrected prose-prefix boundary, Org parity, surrounding
  whitespace handling, and paragraph indentation after structural blocks. The
  separate test audit remains a disclosed, non-gating FAIL under the explicit
  emergency bypass.
