# Feature Specification: Unified Font Configuration

Feature branch: `master`

Created: 2026-09-04

Status: Draft

Input: Replace every public font-related key shape with one uniform, breaking
configuration block; publish all defaults and visible inheritance options; and
carry the maintained regression pack forward with traceable replacements where
the old syntax itself is superseded.

## Scope

- In scope: A **single `font` block shape** configures every supported text role
  across script, letter, and manuscript rendering.
- In scope: [config.yaml](config.yaml) is the **normative configuration
  contract** for the block shape, default values, roles, and inheritance.
- In scope: The existing flat and prefixed font keys are a **deliberately
  breaking change** and receive no compatibility aliases.
- In scope: The **entire maintained regression pack remains live**, with
  traceable replacements where a test's old syntax or output assertion is
  superseded.
- Out of scope: Source parsing, document semantics, pagination, non-typographic
  layout, configuration discovery, layer precedence, and command-line option
  names do not otherwise change.
- Out of scope: Named reusable style registries, user-defined inheritance
  graphs, and additional font properties beyond those declared in the
  normative configuration are not introduced.

## User Scenarios & Testing

### User Story 1 - Configure any font role consistently (Priority: P1)

The user can configure a text role with the same complete font vocabulary in
every rendering mode, including selecting a bold extended monospace face.

Why this priority: A uniform contract removes role-specific gaps that currently
prevent valid font variants from being selected.

Independent Test: A partial monospace font block changes weight and stretch
while inheriting every omitted property and leaving non-monospace text
unchanged.

#### Acceptance Scenarios

1. GIVEN a manuscript whose monospace role inherits its configured family and
   size

   WHEN the user sets **`weight: bold`** and **`stretch: 125%`** in that role's
   `font` block

   THEN inline code and code blocks use the **bold extended variant**

   AND prose and headings retain their own effective font values

2. GIVEN any configurable script, letter, or manuscript text role listed in
   the normative configuration

   WHEN the user supplies any supported subset of the role's `font` fields

   THEN each supplied field **overrides that field only**

   AND each omitted field **inherits independently** from the documented
   parent role

### User Story 2 - Inspect the complete contract (Priority: P1)

The user can inspect one configuration artefact to see every effective default
and every available inherited font field without consulting implementation
code.

Why this priority: Invisible defaults and inconsistent omissions make the
current configuration difficult to reason about safely.

Independent Test: The normative configuration exposes every supported font
role with all six fields either assigned or visibly commented with its
inheritance source.

#### Acceptance Scenarios

1. GIVEN the normative configuration artefact

   WHEN the user inspects a `font` block

   THEN all supported fields are **visible in the same order**

   AND an inherited field is **commented out with its source named**

2. GIVEN an unconfigured installation

   WHEN each rendering mode produces a document

   THEN its effective font values match the **active built-in defaults**
   declared by the normative configuration and its style overrides

### User Story 3 - Preserve regression protection (Priority: P1)

The project owner can hand the approved specification and an aligned regression
pack to a build agent as the behavioural boundary for test-driven delivery.

Why this priority: The regression pack is the strongest evidence that the
schema migration has not silently removed established behaviour.

Independent Test: Every pre-change maintained test remains active or has an
active, traceable replacement that protects the same behaviour under the new
contract.

#### Acceptance Scenarios

1. GIVEN a maintained test whose behaviour is unchanged but whose fixture uses
   a superseded font key

   WHEN the regression pack is aligned with this specification

   THEN the test remains **active with an equivalent new-schema fixture**

2. GIVEN a maintained test whose assertion specifically requires superseded
   syntax or omitted output

   WHEN the regression pack is aligned with this specification

   THEN the old test is **retained as superseded provenance**

   AND an **active replacement preserves its behavioural intent** under this
   specification

3. GIVEN configuration files at the existing global, local, and style-specific
   locations

   WHEN their font values use the new block shape

   THEN the existing **discovery, precedence, and deep-merge behaviour is
   unchanged**

### Edge Cases

- A font block containing **only one field** inherits the other five fields
  independently.
- A higher-precedence layer that changes one font field **does not erase
  sibling fields** supplied by lower-precedence layers.
- An **unknown font field**, an old flat font key, or a scalar value where a
  block is required is rejected with a diagnostic naming the offending path.
- Empty values do not silently erase inherited values; a field is either a
  valid explicit value or is omitted to inherit.
- Numeric and named weights retain their established meanings; stretch values
  retain percentage semantics, including the established plain-number
  shorthand.
- Default upright text remains upright even though the old implementation
  sometimes represented that state by omitting an emitted style argument.

## Requirements

### Functional Requirements

- FR-001 - One font block shape: Every configurable text role MUST use a
  **mapping named `font`** with the same ordered field vocabulary: `family`,
  `size`, `weight`, `stretch`, `style`, and `letter-spacing`.
- FR-002 - Normative YAML contract: [config.yaml](config.yaml) MUST be the
  **authoritative feature-level schema example**, with every effective British
  default assigned and every inheritable field present as a commented option
  naming its source.
- FR-003 - Per-field inheritance: An omitted font field MUST **inherit only
  that field** from the parent role identified in the normative configuration.
- FR-004 - Per-field layered merge: Configuration layering MUST **merge font
  mappings by field**, preserving lower-precedence sibling fields that a
  higher-precedence layer does not replace.
- FR-005 - Consistent field support: Every font role MUST accept **all six font
  fields**, including roles that previously supported only family, size, or
  weight.
- FR-006 - Scoped application: A role-specific font override MUST affect
  **only that role's rendered text**, including body, heading, monospace,
  title-page item, running-matter, contents, copyright, script-element, and
  letter roles declared by the normative configuration.
- FR-007 - Explicit defaults: An unconfigured render MUST resolve every font
  field to an **explicit effective value**; the British defaults and selected
  style overrides MUST preserve the currently approved rendering semantics.
- FR-008 - Invalid configuration: Unknown fields, invalid values, scalar font
  values, and superseded flat or prefixed font keys MUST **fail rather than be
  ignored**, with a diagnostic naming the offending path and the replacement
  block where applicable.
- FR-009 - Unchanged configuration loading: The schema migration MUST preserve
  **source-directory discovery, HOME-boundary stopping, nearest-file selection,
  style-sibling selection, layer precedence, and command-line precedence**.
- FR-010 - Breaking migration: The application MUST accept **only the new font
  block contract** after delivery; legacy aliases and dual decoding are not
  supported.
- FR-011 - Command-line continuity: Existing font-related command-line options
  MUST retain their observable meanings and precedence while targeting the new
  configuration model.
- FR-012 - Regression continuity: **Every test in the maintained pre-change
  regression pack** MUST remain active or have an active traceable replacement;
  no behavioural check may be deleted, disabled, or weakened solely because
  this feature changes configuration syntax.
- FR-013 - Replacement lineage: A test replaced because its old syntax or
  emitted-source assertion is no longer valid MUST remain identifiable as
  **superseded provenance**, and its active replacement MUST identify the same
  originating requirement or this specification's replacement requirement.

### Key Entities

- Font block: The uniform six-field value used to resolve typography for one
  text role.
- Font role: A bounded class of rendered text, such as manuscript monospace,
  a page header, a script speaker, or a letter body.
- Parent role: The documented source from which an omitted font field inherits.
- Effective font: The complete six-field value after preset selection,
  configuration layering, command-line overrides, and role inheritance.

## Success Criteria

### Measurable Outcomes

- SC-001 - Uniformity: **100% of configurable font roles** use the same six
  public fields and no role-specific font-field spelling remains.
- SC-002 - Visible contract: **100% of font blocks** in the normative
  configuration show all six fields as either assigned defaults or commented
  inherited options with named sources.
- SC-003 - Predictable variants: The user can select **family, size, weight,
  stretch, style, and letter spacing independently** for every declared role.
- SC-004 - Regression retention: **100% of the pre-change maintained tests**
  remain active or have active traceable replacements, with zero tests removed
  or skipped solely for this migration.
- SC-005 - Default continuity: British, US, and screenplay outputs retain
  **all protected non-schema behaviour** when rendered without user font
  overrides.
- SC-006 - Failure visibility: **100% of sampled legacy and malformed font
  paths** fail with a diagnostic that identifies the invalid path; none are
  silently ignored.

## Assumptions

- Tadhg O'Brien is the sole current user and has explicitly accepted a
  **breaking configuration migration**.
- The six fields in [config.yaml](config.yaml) are the complete scope of this
  feature; additional text or OpenType controls require a later specification.
- Omission means inheritance. Explicit empty strings and null values are not a
  separate reset mechanism.
- Existing style selection and built-in override layering remain authoritative
  outside the font-key migration.
- Test alignment is a pre-implementation activity: changed automated tests are
  expected to establish the failing RED boundary before production code is
  changed.

## Existing Baseline

- Sources consulted: `docs/ACs.org` requirements AC8.1-AC8.7 for weight and
  stretch, AC9.5-AC9.7 for cross-mode font inheritance, AC9.12 for contents
  typography, AC15.2 for running-footer inheritance, and AC34.1 for manuscript
  tracking; `ARCHITECTURE.md`; `docs/config.md`; the current British, US, and
  screenplay presets; archived tickets 8, 9, 15, 19, 20, and 34; the shared
  configuration loader, three render paths, manuscript normalization, and
  affected maintained tests under `internal/`.
- Preserves: **Rendered-role boundaries, default document semantics,
  configuration discovery and precedence, style layering, CLI precedence, and
  the complete maintained regression pack** are retained.
- Changes: Flat keys, prefixed keys, boolean `bold` and `italic` typography
  controls, inconsistent role field sets, and silent unknown-key acceptance are
  replaced by the normative nested block contract.
- Supersedes: This specification explicitly supersedes the **key paths and
  omission rules** in AC8.1-AC8.7, AC9.5-AC9.7, AC9.12, AC15.2, and AC34.1 only
  where they conflict with [config.yaml](config.yaml). Their rendering intent,
  role scoping, value semantics, and lineage remain in force.
- Unaffected: Source formats, semantic models, non-typographic layout,
  external Typst and Pandoc ownership boundaries, and output-format contracts
  do not change.
- Reconciled regression evidence: The five maintained header/footer style tests
  and five maintained source-directory discovery tests gain explicit current
  targets in FR-006 - scoped application, FR-007 - explicit defaults, FR-009 -
  unchanged configuration loading, and FR-013 - replacement lineage when this
  specification is approved.
