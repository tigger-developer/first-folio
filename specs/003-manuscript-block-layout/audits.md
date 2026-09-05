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

## Current gate

- Effective verdict: **PASS**
- Consequence: Final verification may run against the audited implementation,
  including the focused indent amendment.
