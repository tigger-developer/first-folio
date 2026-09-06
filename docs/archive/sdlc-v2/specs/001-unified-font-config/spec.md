# Feature Specification: Unified Font Configuration

Feature branch: `master`

Created: 2026-09-04

Input: Replace every public font-key shape with one uniform breaking contract;
make one exhaustive British YAML file the runtime base for script, letter, and
manuscript; retain only differences in style overrides; and carry the maintained
regression pack forward with traceable replacements.

## Scope

- In scope: [presets/british.yaml](../../presets/british.yaml) becomes the
  **single exhaustive British base** for script, letter, and manuscript
  configuration.
- In scope: The application **loads that same file at runtime** as its
  lowest-precedence configuration source.
- In scope: A user may **inspect or copy the British base** as a complete
  configuration reference without consulting implementation code.
- In scope: US and screenplay presets become **limited overrides containing only
  differences** from the British base.
- In scope: Every supported text role uses **one `font` block shape**.
- In scope: Existing flat and prefixed font keys are **removed without
  compatibility aliases**.
- In scope: The **entire maintained regression pack remains live**, using
  traceable replacements where syntax or assertions are superseded.
- Out of scope: Source-frontmatter configuration belongs to the separate
  [Source Frontmatter Configuration](../002-source-frontmatter-config/spec.md)
  feature.
- Out of scope: Source semantics, text conversion, pagination, non-typographic
  layout, and configuration discovery **remain unchanged**.
- Out of scope: The feature adds **no named style registry**, **user-defined
  inheritance graph**, or undeclared font property.

## User Scenarios & Testing

### User Story 1 - Inspect and reuse the complete configuration (Priority: P1)

The user **inspects one British YAML file** to find every configuration option
and may **copy it as an editable starting point**.

Why this priority: One runtime-owned reference prevents the documented
configuration contract from **drifting away from application defaults**.

Independent Test: The same exhaustive British file can be **loaded directly by
the application** and **copied unchanged as an override** without changing
output.

#### Acceptance Scenarios

1. GIVEN an installation with **no user override**

   WHEN the user renders a script, letter, or manuscript

   THEN every effective default **originates from `presets/british.yaml`** and
   its documented per-field inheritance

2. GIVEN a copy of `presets/british.yaml` at the global or nearest local
   `script.yaml` location

   WHEN the user renders the same source without editing that copy

   THEN the output **matches the unconfigured British output**

3. GIVEN the British base and a selected US or screenplay preset

   WHEN the application resolves configuration

   THEN the selected preset **overrides only its declared differences**

   AND every omitted value **continues to come from the British base**

### User Story 2 - Configure every font role consistently (Priority: P1)

The user **configures any text role** with the **same complete font vocabulary**
across script, letter, and manuscript rendering.

Why this priority: A uniform contract removes role-specific gaps that prevent
valid font variants from being selected.

Independent Test: One partial font block **changes only its declared
properties** and **inherits every omitted property** from its named parent.

#### Acceptance Scenarios

1. GIVEN a manuscript monospace role whose family and size are inherited

   WHEN the user sets **`weight: bold`** and **`stretch: 125%`**

   THEN inline code and code blocks **use the bold extended variant**

   AND prose and headings **retain their own effective fonts**

2. GIVEN any font role declared in the British base

   WHEN the user supplies a supported subset of its `font` properties

   THEN each supplied property **overrides only itself**

   AND each omitted property **inherits separately from the parent named beside
   that property**

3. GIVEN a higher-precedence file that changes one font property

   WHEN configuration layers are merged

   THEN lower-precedence sibling properties **remain effective**

### User Story 3 - Preserve every publishing mode (Priority: P1)

The user retains **script conversion**, **letter PDF generation**, and
**manuscript Typst/PDF generation** while adopting the new schema.

Why this priority: The configuration migration must not remove established
publishing behaviour.

Independent Test: Each rendering mode accepts the new font contract while its
unrelated source and output behaviour **remains protected**.

#### Acceptance Scenarios

1. GIVEN a stage play in Org, Markdown, or Fountain

   WHEN it is converted among the three text formats

   THEN its existing conversion behaviour **remains unchanged**

   AND font configuration **affects only Typst or PDF rendering**

2. GIVEN an Org source containing one or more letters

   WHEN letter PDFs are generated with a letter font override

   THEN the override **affects letter text only**

   AND recipient selection, filenames, and PDF generation **remain unchanged**

3. GIVEN a Markdown or Org manuscript

   WHEN Typst or PDF output is generated with manuscript font overrides

   THEN each override **affects only its declared manuscript role**

   AND manuscript parsing, joining, and non-typographic layout **remain
   unchanged**

### User Story 4 - Preserve regression protection (Priority: P1)

The project owner **hands an approved specification and aligned regression
pack** to a build agent as the behavioural boundary for test-driven delivery.

Why this priority: The regression pack protects established behaviour from
silent loss during the migration.

Independent Test: Every pre-change maintained test **remains active** or has an
**active traceable replacement** protecting the same behaviour.

#### Acceptance Scenarios

1. GIVEN a maintained test whose fixture uses a superseded font key

   WHEN the regression pack is aligned with this specification

   THEN the test **remains active with an equivalent new-schema fixture**

2. GIVEN a maintained assertion that specifically requires superseded syntax

   WHEN the regression pack is aligned with this specification

   THEN the old test **remains as superseded provenance**

   AND an active replacement **preserves its behavioural intent**

3. GIVEN an invalid font configuration

   WHEN any public command loads it

   THEN the command **writes a path-specific diagnostic to stderr**

   AND **returns a non-zero status before producing output**

### Edge Cases

- A font block with **one explicit property inherits the other five**
  independently.
- A higher-precedence property **does not erase inherited sibling properties**.
- An **unknown property inside a font block**, **legacy font key**, **scalar font
  value**, **empty value**, or **invalid font value** causes a hard failure.
- A copied British base with no edits **does not become a second default
  authority**; it is an ordinary higher-precedence override.
- A style override that repeats an unchanged British value **violates the
  minimal-override contract**.
- Default upright text **remains upright** without relying on an omitted output
  argument.

## Requirements

### Functional Requirements

- FR-001 - **Provide one canonical British base**: `presets/british.yaml` MUST
  be the **sole exhaustive built-in base** for script, letter, and manuscript
  configuration.
- FR-002 - **Use the canonical base at runtime**: Every rendering mode MUST load
  `presets/british.yaml` as its **lowest-precedence configuration source**; no
  duplicate runtime default may independently define the same contract.
- FR-003 - **Expose the complete public contract**: The British base MUST show
  **every accepted public configuration property** in its correct hierarchy.
  Explicit defaults MUST be assigned, while inherited properties MUST remain
  visible as commented options naming their source.
- FR-004 - **Keep the canonical base copyable**: Copying the British base
  unchanged to a supported global or local `script.yaml` location MUST produce
  the **same effective British configuration**.
- FR-005 - **Keep style presets minimal**: The US and screenplay built-in
  presets MUST contain **only properties whose effective values differ** from
  the British base and MUST deep-merge over that base.
- FR-006 - **Use one font block shape**: Every configurable text role MUST use
  a mapping named `font` with the ordered properties `family`, `size`, `weight`,
  `stretch`, `style`, and `letter-spacing`.
- FR-007 - **Inherit font properties independently**: An omitted font property
  MUST inherit **that property only** from the parent identified beside it in
  the British base. Resolution MUST continue transitively until it reaches an
  explicitly assigned value.
- FR-008 - **Merge font properties across layers**: A higher-precedence layer
  MUST replace only its declared properties and **preserve lower-precedence
  siblings**.
- FR-009 - **Support every role consistently**: Every font role MUST accept all
  six font properties, including roles previously limited to family, size, or
  weight.
- FR-010 - **Confine overrides to their roles**: A role-specific font override
  MUST affect **only that role's rendered text** across script, letter, and
  manuscript output.
- FR-011 - **Resolve explicit effective defaults**: An unconfigured render MUST
  resolve every font property through the British base and its transitive
  inheritance while preserving approved British **rendered output behaviour**.
  Generated Typst MAY change only where the new contract emits explicit font
  values that previously relied on omission.
- FR-012 - **Reject invalid font configuration**: Unknown properties inside a
  `font` block, invalid font values, scalar font values, empty or null font
  properties, and superseded font keys MUST **terminate before output**, write a
  diagnostic identifying the offending path, and return a **non-zero status**.
- FR-013 - **Preserve configuration loading**: Source-directory discovery,
  HOME-boundary stopping, nearest-file selection, style-sibling selection, deep
  merging, and existing layer precedence MUST remain unchanged.
- FR-014 - **Make a clean breaking migration**: After delivery, the application
  MUST accept **only the new font block contract**, without compatibility aliases
  or dual decoding.
- FR-015 - **Preserve documented font flags**: `folio convert --font` and
  `folio convert --font-size` MUST retain their observable meanings and highest
  precedence while targeting `folio.font.family` and `folio.font.size`.
- FR-016 - **Preserve mode contracts**: The migration MUST leave Org, Markdown,
  and Fountain script conversion; script Typst/PDF rendering; Org letter PDF
  generation; and Markdown/Org manuscript Typst/PDF generation unchanged except
  for the specified configuration schema.
- FR-017 - **Keep the regression pack live**: Every pre-change maintained test
  MUST remain active or have an active traceable replacement. No behavioural
  check may be deleted, disabled, or weakened solely because its syntax changes.
- FR-018 - **Preserve replacement lineage**: A replaced test MUST remain
  identifiable as superseded provenance. Its active replacement MUST name the
  same originating requirement or this specification's replacement requirement.

### Key Entities

- **British base**: The single exhaustive runtime configuration from which every
  built-in style and rendering mode begins.
- **Style override**: A limited YAML layer containing only effective differences
  from the British base.
- **Font block**: The uniform six-property value resolving typography for one
  text role.
- **Parent role**: The documented source from which one omitted font property
  inherits.
- **Effective configuration**: The result after base selection, style overlays,
  file layers, command-line overrides, and per-property inheritance.

## Success Criteria

### Measurable Outcomes

- SC-001 - **Maintain one exhaustive base**: **100% of accepted public
  configuration properties** appear in `presets/british.yaml`, assigned or
  visibly commented for inheritance.
- SC-002 - **Use one runtime authority**: **100% of unconfigured script, letter,
  and manuscript renders** derive their defaults from the same British base.
- SC-003 - **Keep overlays limited**: **100% of US and screenplay preset
  properties differ effectively** from the British base.
- SC-004 - **Use one font vocabulary**: **100% of configurable font roles** use
  the same six public properties.
- SC-005 - **Retain publishing behaviour**: Org, Markdown, and Fountain script
  conversion; letter PDFs; and Markdown/Org manuscript Typst/PDF output retain
  all protected non-schema behaviour.
- SC-006 - **Retain regression protection**: **100% of pre-change maintained
  tests remain active or have active traceable replacements**, with none removed
  or skipped solely for the migration.
- SC-007 - **Make failures visible**: **100% of sampled legacy, unknown, and
  malformed configuration paths** return non-zero with an identifying diagnostic
  and produce no output artefact.

## Assumptions

- The sole current user has **accepted a breaking configuration migration**.
- The six font properties in the British base **bound this feature**; additional
  text or OpenType controls require another specification.
- A commented property **inherits from its named parent**; an uncommented
  property owns its explicit default.
- Inheritance **continues through commented parent properties** until an
  assigned value is reached, identically in the built-in base and an unedited
  copy.
- Existing built-in British, US, and screenplay rendering semantics **remain
  authoritative** unless this specification explicitly changes their schema.
- Source-frontmatter configuration is **specified separately** and does not
  alter this feature's existing file-layer precedence.
- Automated tests are **aligned before production code changes** so they
  establish the failing RED boundary.

## Existing Baseline

- Requirement authority consulted: `docs/ACs.org` requirements **AC1.1-AC1.7 -
  script conversion and PDF configuration**, **AC2.1-AC2.3 - YAML configuration
  and precedence**, **AC3.2-AC3.6 - British and US preset selection**,
  **AC8.1-AC8.7 - font weight and stretch**, **AC9.1-AC9.12 - manuscript input,
  rendering, and inheritance**, and **AC10.1-AC10.5 - unified runtime and loader**.
- Design authority consulted: `ARCHITECTURE.md`, `docs/config.md`, the current
  preset files, and the active configuration and rendering code establish the
  **split runtime baseline**.
- Historical context consulted: Archived tickets **1 - multi-format script
  conversion**, **2 - unified YAML configuration**, **3 - style presets**,
  **8 - font weight and stretch**, **9 - manuscript mode**, **10 - unified Go
  runtime**, **15 - manuscript configuration extensions**, **19 - running-matter
  font style**, **20 - source-directory lookup**, and **34 - manuscript letter
  spacing** establish lineage and rationale.
- Regression evidence consulted: Maintained tests under `internal/config`,
  `internal/app`, `internal/letter`, `internal/manuscript`, and `internal/play`
  establish the **implemented configuration, mode, and output boundaries**.
- Preserves: **British-first styling**, **US and screenplay overlays**, existing
  file discovery and precedence, all rendering modes, and the maintained
  regression pack.
- Changes: One exhaustive British runtime file **replaces the separate script
  and manuscript base files**; one US override **replaces the separate US
  script and manuscript overrides**; every font key adopts the nested contract.
- Supersedes: This specification supersedes the **preset filenames and split-file
  structure** named by AC9.9 - US manuscript override, and the conflicting font
  key paths and omission rules in AC8.1-AC8.7, AC9.5-AC9.7, AC9.12, AC15.2, and
  AC34.1. Their rendering intent, value semantics, and lineage remain
  authoritative.
- Unaffected: **AC2.3 - unrelated top-level key tolerance** and **AC10.5 -
  companion-tool key tolerance** remain authoritative; unknown-key rejection is
  confined to font blocks and superseded font keys.
- Unaffected: Source-format semantics, content conversion, non-typographic
  layout, and external Typst, Pandoc, and Fountain ownership boundaries do not
  change.
- Reconciled evidence: On approval, the five header/footer style tests and five
  source-directory discovery tests **gain current targets** in FR-010 - confine
  overrides to roles, FR-011 - resolve effective defaults, FR-013 - preserve
  loading, and FR-018 - preserve replacement lineage.
