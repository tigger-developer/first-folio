# Feature Specification: Source Frontmatter Configuration

Feature branch: `master`

Created: 2026-09-05

Input: Allow the first Markdown manuscript source to declare `folio:` and
`render:` configuration in YAML frontmatter, above every file-based
configuration layer but below explicit command-line options.

## Scope

- In scope: The first resolved Markdown manuscript input may supply **`folio:`
  and `render:` configuration** in its existing YAML frontmatter.
- In scope: Source configuration **overrides every file-based layer** while
  explicit command-line options remain highest.
- In scope: Source configuration uses the **same public schema, validation, and
  merge semantics** as external YAML configuration.
- In scope: Existing manuscript metadata may **coexist with configuration** in
  the same frontmatter block.
- Out of scope: Stage-play Markdown, Org manuscripts, and Org letters **gain no
  source-embedded configuration**.
- Out of scope: YAML-shaped `folio:` and `render:` lines in stage-play Markdown
  **remain ordinary play source**, without configuration effect or new
  source-configuration diagnostics.
- Out of scope: The feature **does not write or synchronize** external
  `script.yaml` files.
- Out of scope: Source configuration **does not alter manuscript content
  semantics** or appear in generated Typst/PDF content.

## User Scenarios & Testing

### User Story 1 - Keep manuscript layout with its source (Priority: P1)

The user **stores manuscript-specific configuration beside the manuscript** so
the intended rendering travels with the source.

Why this priority: Source-owned configuration removes dependence on a matching
local configuration file when the manuscript is moved or shared.

Independent Test: One Markdown manuscript with embedded configuration renders
with that configuration **without a local or global file change**.

#### Acceptance Scenarios

1. GIVEN a Markdown manuscript whose YAML frontmatter contains a `folio:` block

   WHEN the manuscript is rendered to Typst or PDF

   THEN its source configuration **affects the generated manuscript**

   AND the `folio:` block **does not appear as manuscript content**

2. GIVEN frontmatter containing both manuscript metadata and configuration

   WHEN the manuscript is rendered

   THEN the metadata **retains its existing meaning**

   AND `folio:` and `render:` values **apply as configuration**

3. GIVEN a valid external configuration property

   WHEN the same property is placed under its documented frontmatter path

   THEN it has the **same value semantics and rendered effect**

### User Story 2 - Resolve overrides predictably (Priority: P1)

The user can **predict the effective manuscript configuration** when defaults,
files, frontmatter, and command-line options specify the same property.

Why this priority: A portable source override is unsafe if its authority is
ambiguous.

Independent Test: Conflicting values resolve according to the complete declared
precedence chain.

#### Acceptance Scenarios

1. GIVEN different values for one property in source frontmatter and the nearest
   local `script.yaml`

   WHEN the manuscript is rendered without a matching command-line option

   THEN the **source-frontmatter value is effective**

2. GIVEN different values for one property in source frontmatter and an explicit
   command-line option

   WHEN the manuscript is rendered

   THEN the **command-line value is effective**

3. GIVEN `folio.style: us` in source frontmatter and no command-line style

   WHEN the manuscript is rendered

   THEN the US built-in differences are **selected beneath the source layer**

   AND other source-frontmatter properties **override those differences**

### User Story 3 - Bound multi-file configuration safely (Priority: P1)

The user receives **one deterministic configuration** for a manuscript assembled
from multiple Markdown files.

Why this priority: Per-chapter configuration would make one generated document
depend on input order in obscure ways.

Independent Test: The first resolved input controls source configuration, and a
later configuration block fails visibly.

#### Acceptance Scenarios

1. GIVEN a multi-file Markdown manuscript whose first resolved input contains
   configuration frontmatter and whose process working directory contains the
   local configuration files

   WHEN the inputs are rendered together

   THEN that configuration **applies to the complete generated manuscript**

   AND the lower-precedence local files **resolve from the process working
   directory**

2. GIVEN a later resolved input containing `folio:` or `render:` frontmatter

   WHEN the inputs are rendered together

   THEN the command **identifies that file as an invalid configuration source**

   AND **returns a non-zero status before producing output**

3. GIVEN malformed or invalid source configuration

   WHEN the manuscript command loads the source

   THEN the command **writes a path-specific diagnostic to stderr**

   AND **returns a non-zero status before producing output**

4. GIVEN a Markdown stage play beginning with YAML-shaped `folio:` or `render:`
   lines

   WHEN the stage play is converted or rendered

   THEN those lines **do not alter effective configuration**

   AND they **remain ordinary play source without a source-configuration
   diagnostic**

### Edge Cases

- YAML frontmatter containing metadata but **no `folio:` or `render:` block**
  retains its existing behaviour.
- An empty `folio:` or `render:` mapping **changes no effective value**.
- A scalar where either namespace requires a mapping **fails validation**.
- Unknown or invalid properties inside `folio:` or `render:` **fail under the
  same rules** as external configuration.
- A source-selected style **does not outrank an explicit CLI style**.
- A source property omitted from a partial mapping **inherits from lower
  precedence layers**.
- Configuration frontmatter in any input after the first **fails rather than
  being silently ignored**.

## Requirements

### Functional Requirements

- FR-001 - **Accept source configuration**: The first resolved Markdown
  manuscript input MAY declare `folio:` and `render:` mappings in its leading
  YAML frontmatter.
- FR-002 - **Use the complete public schema**: Source-frontmatter configuration
  MUST accept the same properties, types, values, and nested paths as external
  First Folio YAML configuration. Top-level frontmatter keys outside `folio:`
  and `render:` MUST retain the existing AC2.3 and AC10.5 tolerance and MUST NOT
  become configuration merely because either namespace is present.
- FR-003 - **Preserve metadata**: Existing manuscript metadata in YAML
  frontmatter MUST retain its current meanings when configuration namespaces are
  present. This feature MUST apply only `folio:` and `render:` as source
  configuration; external configuration metadata MUST continue to override
  source metadata under AC2.5 - configuration metadata precedence.
- FR-004 - **Apply explicit precedence**: Effective manuscript configuration
  MUST resolve from highest to lowest as command-line options, first-input source
  frontmatter, nearest local style-specific YAML, nearest local `script.yaml`,
  global style-specific YAML, global `script.yaml`, selected built-in style
  override, and the British base. For one resolved input, local files MUST resolve
  from the input directory; for multiple resolved inputs, local files MUST
  resolve from the process working directory.
- FR-005 - **Merge source mappings by property**: Source frontmatter MUST replace
  only properties it declares and MUST preserve lower-precedence sibling
  properties.
- FR-006 - **Select styles from source**: A valid source-frontmatter style MUST
  select the matching global and local style-specific sibling files and the
  built-in style override, identically to a CLI-selected style. An explicit
  command-line style MUST supersede a source-selected style.
- FR-007 - **Apply one configuration to multi-file output**: Configuration from
  the first resolved input MUST govern the complete combined manuscript.
- FR-008 - **Reject later configuration sources**: A `folio:` or `render:`
  frontmatter mapping in any later resolved input MUST terminate processing,
  identify that input, and return a non-zero status before output.
- FR-009 - **Reject invalid source configuration**: Malformed YAML, unknown
  properties inside `folio:` or `render:`, invalid configuration values, and
  wrong configuration shapes MUST terminate processing. Malformed YAML MUST
  identify the source file. Any parseable unknown property, invalid value, or
  wrong shape MUST identify both the source file and its full configuration
  path. Every failure MUST return a non-zero status before output.
- FR-010 - **Prevent partial output on failure**: A source-configuration failure
  MUST produce no target Typst, PDF, or related sidecar artefact.
- FR-011 - **Preserve non-manuscript Markdown behaviour**: In a Markdown stage
  play, YAML-shaped `folio:` or `render:` lines MUST remain ordinary play source,
  MUST NOT alter effective configuration, and MUST NOT trigger the manuscript
  source-configuration diagnostics introduced by this feature. Org manuscripts
  and letter sources MUST retain their existing configuration behaviour.
- FR-012 - **Keep source configuration out of content**: `folio:` and `render:`
  mappings used as configuration MUST NOT appear in manuscript blocks or
  generated output.
- FR-013 - **Preserve regression protection**: Existing metadata, manuscript
  parsing, multi-input ordering, rendering, external configuration, and
  command-line tests MUST remain active or have traceable active replacements.

### Key Entities

- **Source configuration**: The `folio:` and `render:` mappings in the first
  Markdown manuscript's leading YAML frontmatter.
- **Metadata**: Existing descriptive frontmatter values such as title, author,
  date, version, and word count.
- **First resolved input**: The first file after existing path and glob
  resolution; it owns source configuration for the combined manuscript.
- **Effective configuration**: The property-level merge of the declared
  precedence layers.

## Success Criteria

### Measurable Outcomes

- SC-001 - **Support the complete configuration vocabulary**: **100% of public
  `folio:` and `render:` properties** accepted in external YAML are accepted with
  the same semantics in first-input manuscript frontmatter.
- SC-002 - **Resolve every precedence conflict**: **100% of pairwise conflicts**
  among CLI, source frontmatter, local style-specific, local base, global
  style-specific, global base, built-in style, and British base layers resolve
  according to the declared chain.
- SC-003 - **Preserve metadata behaviour**: **100% of maintained Markdown
  frontmatter metadata cases in `internal/manuscript/manuscript_test.go`** retain
  their existing effective values when no source configuration is present.
- SC-004 - **Keep multi-file configuration deterministic**: Every multi-file
  render has **exactly one permitted source-configuration owner**.
- SC-005 - **Fail without misleading output**: Every malformed YAML block,
  unknown configuration property, invalid configuration value, wrong
  configuration shape, and later-input configuration block **returns non-zero
  with its required diagnostic and leaves no output artefact**.
- SC-006 - **Retain regression protection**: All affected pre-change maintained
  tests **remain active or have traceable active replacements**.

## Assumptions

- Explicit command-line options **remain the highest authority**.
- Only the first resolved Markdown manuscript input may own source configuration;
  later inputs retain their ordinary manuscript-content contract.
- Source frontmatter permits the complete public `folio:` and `render:`
  namespaces rather than a feature-specific subset.
- Empty namespace mappings are valid no-ops; null or scalar namespaces are
  invalid.
- The [Unified Font Configuration](../001-unified-font-config/spec.md) feature
  supplies the canonical British base and validation contract consumed by this
  feature.
- This feature's planning and delivery are **conditional on approval of Unified
  Font Configuration** or another approved specification supplying the same
  canonical schema and validation contract.

## Existing Baseline

- Requirement authority consulted: `docs/ACs.org` requirements **AC2.1-AC2.3 -
  YAML files and precedence**, **AC2.5 - configuration metadata precedence**,
  **AC9.1-AC9.4 - manuscript inputs and semantics**, **AC9.11 - manuscript
  failure behaviour**, and **AC10.5 - shared configuration loading**, together
  with **AC3.7 - style-specific file selection** and **AC13.1-AC13.3 - manuscript
  local-directory resolution and precedence**.
- Design authority consulted: `ARCHITECTURE.md`, `docs/config.md`, and
  `docs/format-manuscript-markdown.md` establish the existing **file-layer and
  metadata boundary**.
- Historical context consulted: Archived tickets **2 - unified YAML
  configuration**, **9 - manuscript mode**, **10 - unified Go runtime**, and
  **20 - source-directory configuration lookup** establish provenance and
  rationale.
- Regression evidence consulted: Maintained tests under `internal/config`,
  `internal/app`, and `internal/manuscript`, together with the active loader,
  input resolver, frontmatter parser, and manuscript command, establish the
  **implemented precedence and parsing baseline**.
- Regression lineage: The five maintained source-directory lookup tests are
  reconciled by Feature 001 requirement **FR-018 - preserve replacement
  lineage**; this feature preserves that dependency for its file-layer
  composition.
- Baseline conflict: **AC13.2 - multi-input configuration uses the working
  directory** is PENDING, while `ARCHITECTURE.md` and the active manuscript
  command currently use the first input directory. Requirement authority governs
  this feature's composition with local file layers.
- Preserves: Existing metadata meanings, CLI authority, AC13.1-AC13.3 directory
  rules, property-level deep merging, input ordering, and Markdown manuscript
  output.
- Changes: The first Markdown manuscript's frontmatter **becomes a configuration
  layer** for `folio:` and `render:` mappings.
- Changes: The affected multi-input path must **conform to AC13.2's working-
  directory rule** rather than perpetuate the conflicting implemented state.
- Supersedes: For Markdown manuscript rendering only, this specification extends
  **AC2.2 - CLI, local, global, and built-in precedence** by inserting source
  frontmatter below CLI and above every file-based layer while preserving its
  lineage.
- Unaffected: Stage-play conversion, Org manuscript input, letter generation,
  external configuration locations, and Typst, Pandoc, and Fountain ownership
  boundaries do not change.
