# Feature Specification: Unified Font Configuration

Feature branch: `master`

Created: 2026-09-04

Status: Draft

Input: Replace every public font-related key shape with one uniform, breaking
configuration block; publish all defaults and visible inheritance options; and
carry the maintained regression pack forward with traceable replacements where
the old syntax itself is superseded.

## Scope

- In scope: Every supported text role uses **one `font` block shape** across
  script, letter, and manuscript rendering.
- In scope: [config.yaml](config.yaml) **defines** the block shape, defaults,
  roles, and inheritance as the **normative configuration contract**.
- In scope: Existing flat and prefixed font keys are **removed without
  compatibility aliases**.
- In scope: The **entire maintained regression pack remains live**, using
  traceable replacements where old syntax or assertions are superseded.
- Out of scope: Source parsing, document semantics, pagination, and
  non-typographic layout **remain unchanged**.
- Out of scope: Configuration discovery, layer precedence, and command-line
  option names **remain unchanged**.
- Out of scope: The feature adds **no named style registry**, **user-defined
  inheritance graph**, or undeclared font property.

## User Scenarios & Testing

### User Story 1 - Configure any font role consistently (Priority: P1)

The user **configures any text role** with the **same complete font vocabulary**
in every rendering mode, including a bold extended monospace face.

Why this priority: A **uniform contract removes role-specific gaps** that
prevent valid font variants from being selected.

Independent Test: A **partial monospace font block changes weight and stretch**,
**inherits omitted properties**, and **leaves other text unchanged**.

#### Acceptance Scenarios

1. GIVEN a manuscript with a **partially inherited monospace font**

   WHEN the user sets **`weight: bold`** and **`stretch: 125%`** in that role's
   `font` block

   THEN **inline code and code blocks use** the **bold extended variant**

   AND **prose and headings retain** their own effective font values

2. GIVEN any configurable script, letter, or manuscript text role listed in
   the normative configuration

   WHEN the user **supplies any supported subset** of the role's `font` fields

   THEN each supplied field **overrides only itself**

   AND each omitted field **inherits independently** from its documented parent

### User Story 2 - Inspect the complete contract (Priority: P1)

The user **inspects one configuration artefact** to see **every default** and
**every inherited font option** without consulting implementation code.

Why this priority: **Visible defaults and inheritance** make configuration
decisions reviewable before rendering.

Independent Test: Every supported font role **shows all six fields** as an
assigned default or a commented option naming its inheritance source.

#### Acceptance Scenarios

1. GIVEN the normative configuration artefact

   WHEN the user **inspects a `font` block**

   THEN **all supported fields appear** in the same order

   AND **each inherited field is commented out** with its source named

2. GIVEN an unconfigured installation

   WHEN each rendering mode **produces a document**

   THEN its effective font values **match the active built-in defaults** and
   selected style overrides

### User Story 3 - Preserve regression protection (Priority: P1)

The project owner **hands an approved specification and aligned regression
pack** to a build agent as the behavioural boundary for test-driven delivery.

Why this priority: The **regression pack protects established behaviour** from
silent loss during the schema migration.

Independent Test: Every pre-change maintained test **remains active** or has an
**active traceable replacement** protecting the same behaviour.

#### Acceptance Scenarios

1. GIVEN a maintained test whose behaviour is unchanged but whose fixture uses
   a superseded font key

   WHEN the regression pack is **aligned with this specification**

   THEN the test **remains active** with an equivalent new-schema fixture

2. GIVEN a maintained test whose assertion specifically requires superseded
   syntax or omitted output

   WHEN the regression pack is **aligned with this specification**

   THEN the old test **remains as superseded provenance**

   AND an active replacement **preserves its behavioural intent**

3. GIVEN configuration files at the existing global, local, and style-specific
   locations

   WHEN their font values **use the new block shape**

   THEN **discovery, precedence, and deep merging remain unchanged**

### Edge Cases

- A font block with **one explicit field inherits the other five**
  independently.
- A higher-precedence field **does not erase inherited sibling fields**.
- An **unknown field**, **legacy key**, or **scalar font value is rejected**
  with a diagnostic naming its path.
- An **empty or null field is rejected** rather than erasing an inherited
  value.
- Named and numeric weights **retain their established meanings**.
- Stretch values **retain percentage semantics** and the established
  plain-number shorthand.
- Default upright text **remains upright** without requiring the former omitted
  output argument.

## Requirements

### Functional Requirements

- FR-001 - **Use one font block shape**: Every configurable text role MUST use
  a **mapping named `font`** with the same ordered fields: `family`,
  `size`, `weight`, `stretch`, `style`, and `letter-spacing`.
- FR-002 - **Publish the normative YAML contract**:
  [config.yaml](config.yaml) MUST **assign every effective British default** and
  show every inheritable field as a **commented option naming its source**.
- FR-003 - **Inherit each field independently**: An omitted font field MUST
  **inherit that field only** from the parent named in the normative
  configuration.
- FR-004 - **Merge font fields across layers**: Configuration layering MUST
  **preserve lower-precedence sibling fields** that a higher layer does not
  replace.
- FR-005 - **Support every field consistently**: Every font role MUST accept
  **all six font fields**, including roles previously limited to family, size,
  or weight.
- FR-006 - **Confine overrides to their roles**: A role-specific font override
  MUST affect **only that role's text**. The normative configuration declares
  the body, heading, monospace, title-page, running-matter, contents, copyright,
  script-element, and letter roles.
- FR-007 - **Resolve explicit effective defaults**: An unconfigured render MUST
  resolve **every font field to a value** while preserving approved British and
  selected-style rendering semantics.
- FR-008 - **Reject invalid font configuration**: Unknown fields, invalid
  values, scalar font values, and superseded keys MUST **fail with a diagnostic
  naming the path** and, for legacy keys, the replacement block.
- FR-009 - **Preserve configuration loading**: The migration MUST preserve
  **source-directory discovery**, **HOME-boundary stopping**, **nearest-file and
  style-sibling selection**, and **configuration and command-line precedence**.
- FR-010 - **Make a clean breaking migration**: The application MUST accept
  **only the new font block contract** after delivery, without legacy aliases
  or dual decoding.
- FR-011 - **Preserve command-line behaviour**: Existing font-related options
  MUST retain their **meanings and precedence** while targeting the new model.
- FR-012 - **Keep the regression pack live**: Every pre-change maintained test
  MUST **remain active or have an active traceable replacement**. No behavioural
  check may be deleted, disabled, or weakened solely because its configuration
  syntax changes.
- FR-013 - **Preserve replacement lineage**: A replaced test MUST remain
  identifiable as **superseded provenance**. Its active replacement MUST name
  the same originating requirement or this specification's replacement
  requirement.

### Key Entities

- **Font block**: The uniform six-field value used to resolve one text role.
- **Font role**: A bounded class of rendered text, such as manuscript
  monospace, a page header, a script speaker, or a letter body.
- **Parent role**: The documented source from which an omitted field inherits.
- **Effective font**: The complete six-field value after preset selection,
  configuration layering, command-line overrides, and role inheritance.

## Success Criteria

### Measurable Outcomes

- SC-001 - **Use one public vocabulary**: **100% of configurable font roles**
  use the same six fields, with no role-specific spelling.
- SC-002 - **Expose the whole contract**: **100% of normative font blocks** show
  all six fields as assigned defaults or commented inherited options naming
  their sources.
- SC-003 - **Select variants predictably**: The user can set **family, size,
  weight, stretch, style, and letter spacing independently** for every role.
- SC-004 - **Retain regression protection**: **100% of pre-change maintained
  tests remain active or have active traceable replacements**, with none removed
  or skipped solely for the migration.
- SC-005 - **Preserve default output behaviour**: British, US, and screenplay
  outputs retain **all protected non-schema behaviour** without user font
  overrides.
- SC-006 - **Make failures visible**: **100% of sampled legacy and malformed
  paths fail with an identifying diagnostic**; none are silently ignored.

## Assumptions

- The sole current user has **accepted a breaking configuration migration**.
- The six fields in [config.yaml](config.yaml) **bound this feature**;
  additional text or OpenType controls require another specification.
- **Omission means inheritance**; empty strings and null values do not reset a
  field.
- Existing style selection and built-in override layering **remain
  authoritative** outside the font-key migration.
- Automated tests are **aligned before production code changes** so they
  establish the failing RED boundary.

## Existing Baseline

- Requirement authority consulted: `docs/ACs.org` requirements **AC8.1-AC8.7 -
  font weight and stretch**, **AC9.5-AC9.7 - cross-mode font inheritance**,
  **AC9.12 - contents typography**, **AC15.2 - running-footer inheritance**, and
  **AC34.1 - manuscript tracking**.
- Design authority consulted: `ARCHITECTURE.md`, `docs/config.md`, and the
  current British, US, and screenplay presets establish the **configuration and
  rendering baseline**.
- Historical context consulted: Archived tickets **8 - font weight and
  stretch**, **9 - manuscript PDF mode**, **15 - manuscript configuration
  extensions**, **19 - running-matter font style**, **20 - source-directory
  configuration lookup**, and **34 - manuscript letter spacing** establish
  **lineage and rationale**.
- Implementation evidence consulted: The shared loader, three render paths,
  manuscript normalization, and affected maintained tests establish the
  **implemented baseline and regression boundary**.
- Preserves: **Role boundaries, document semantics, configuration discovery and
  precedence, style layering, CLI precedence, and the maintained regression
  pack remain unchanged**.
- Changes: The normative nested contract **replaces flat and prefixed keys**,
  boolean typography controls, inconsistent role fields, and silent unknown-key
  acceptance.
- Supersedes: This specification **supersedes conflicting key paths and omission
  rules only** in AC8.1-AC8.7, AC9.5-AC9.7, AC9.12, AC15.2, and AC34.1. Their
  rendering intent, scoping, value semantics, and lineage remain authoritative.
- Unaffected: Source formats, semantic models, non-typographic layout, external
  ownership boundaries, and output-format contracts **do not change**.
- Reconciled evidence: On approval, the five header/footer style tests and five
  source-directory discovery tests **gain current targets** in FR-006 - confine
  overrides to their roles, FR-007 - resolve explicit effective defaults,
  FR-009 - preserve configuration loading, and FR-013 - preserve replacement
  lineage.
