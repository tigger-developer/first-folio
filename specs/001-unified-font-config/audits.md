# Audit Record: Unified Font Configuration

## Specification audit

Audit: `audit-spec`

Candidate:

- `spec.md`: `29f7f9b2f746bfc500e0da9d63d52fd44cddb2676d1e08f740aa4f547f85961f`
- `config.yaml`: `057b4d61c0181c44ef00b779b58ea10ebf7ae969d3f7540916c7247ee83d03eb`

### Attempt 1

- Harness: Hermes
- Provider: Not reported
- Model: Not reported
- Result: **Harness failure, not an audit verdict**
- Evidence: `sdlc-audit: hermes audit exceeded 15m0s timeout`
- Context: Full relevant brownfield authority, history, implementation, preset,
  and regression evidence set.

### Attempt 2

- Harness: Hermes
- Provider: Not reported
- Model: Not reported
- Result: **Harness failure, not an audit verdict**
- Evidence: `sdlc-audit: hermes audit exceeded 15m0s timeout`
- Context: Constitution, current requirement and architecture authorities,
  public configuration contract, and the two ten-test reconciliation files.

### Attempt 3

- Harness: Hermes
- Provider: Not reported
- Model: Not reported
- Result: **Harness failure, not an audit verdict**
- Evidence: `sdlc-audit: hermes audit exceeded 15m0s timeout`
- Context: Candidate specification, normative YAML, and constitution only, to
  isolate context volume from harness availability.

## Post-feedback candidate

- `spec.md`: `041ed5d2f8afee7b844ba6cb6fae083c4794434e79ec5210307bce8cc887b873`
- `config.yaml`: `057b4d61c0181c44ef00b779b58ea10ebf7ae969d3f7540916c7247ee83d03eb`
- Revision: Presentation was aligned with the project styling contract after
  project-owner feedback on 2026-09-05. Requirements and behaviour were not
  intentionally changed.
- Audit status: The three attempts above apply to the earlier candidate digest,
  not this revised candidate. No independent verdict covers this digest.

## Canonical-base candidate

- `spec.md`: `d68119ae6b30bd81a53b284c16c51ca3ce24f9f33bf16f02b257f1bbf4a833a3`
- `presets/british.yaml`:
  `18cb3781fe5da95ed8009084f345aa59379894812272bc5fbbc7e031c7b5ff2a`
- Revision: Project-owner clarification restored one exhaustive British runtime
  base, minimal US and screenplay overlays, explicit mode coverage, named CLI
  flags, and hard-failure semantics.
- Audit status: The earlier attempts do not cover this candidate. A fresh
  independent verdict is required.

### Attempt 4

- Candidate `spec.md`:
  `4d3d9538618f147ec876c4f4b40de1848fec3c873734fcd939be10fa8017bdab`
- Candidate `presets/british.yaml`:
  `18cb3781fe5da95ed8009084f345aa59379894812272bc5fbbc7e031c7b5ff2a`
- Harness: Hermes
- Provider: `nous`
- Model: `z-ai/glm-5.3-flash`
- Verdict: **FAIL**
- Blocking finding: FR-012 - reject invalid configuration described unknown
  properties without confining the rule to font blocks, contradicting AC2.3 -
  unrelated top-level key tolerance and AC10.5 - companion-tool key tolerance.
- Advisory: FR-007 - inherit font properties independently did not state that
  inheritance resolves transitively through commented parent properties.
- Advisory: FR-011 - resolve explicit effective defaults did not distinguish
  rendered-output continuity from deliberately changed generated Typst text.
- Remediation: The invalid-property boundary is confined to font blocks and
  superseded font keys; unrelated-key tolerance is explicit; transitive
  inheritance and rendered-output equivalence are stated.

### Attempt 5

- Candidate `spec.md`:
  `d68119ae6b30bd81a53b284c16c51ca3ce24f9f33bf16f02b257f1bbf4a833a3`
- Candidate `presets/british.yaml`:
  `18cb3781fe5da95ed8009084f345aa59379894812272bc5fbbc7e031c7b5ff2a`
- Harness: Hermes
- Provider: Not reported
- Model: Not reported
- Result: **Harness failure, not an audit verdict**
- Evidence: `sdlc-audit: hermes audit exceeded 15m0s timeout`
- Context: Remediated specification, canonical British base, constitution,
  requirement and design authorities, configuration documentation, format
  documentation, and current built-in presets.

## Current gate

- Effective verdict: **NONE**
- Audit attempts: **5**
- Blocker: Attempts 1-3 and 5 returned no structured verdict. Attempt 4 returned
  FAIL; its blocking finding was remediated, but the final permitted attempt
  timed out without reviewing that correction.
- Consequence: The draft may be reviewed by the project owner, but planning
  MUST NOT begin until a current `audit-spec` PASS or effective PASS is recorded.
