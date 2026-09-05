# Feature Specification: Manuscript Block Layout

## Specification Summary

- **Outcome:** Manuscript blockquotes and fenced code have configurable
  clearance from surrounding paragraphs, and blockquotes have configurable
  typography.
- **Before:** Fenced-code clearance used only side-specific settings;
  blockquotes had no dedicated clearance or font contract.
- **After:** Equal spacing is available through `quoted-block-spacing` and
  `code-block-spacing`; quoted blocks use the shared six-property `font` block.
- **Changes:** Side-specific code-block spacing remains available and overrides
  the equal-spacing value on its respective side.
- **Unchanged:** Inline code and fenced code retain the manuscript monospace
  font; paragraph spacing and source semantics do not change.
- **Edge cases:** Omitted quote-font properties inherit independently;
  side-specific code spacing may override only one side; invalid values fail
  before output.
- **Decisions:** Both spacing defaults are `0.5em`; the public quote style
  `regular` maps to Typst `normal`.
- **Evidence:** The project owner's `BYPASS-GATE-7` instruction dated 2026-09-05,
  the maintained manuscript regression pack, and [audits.md](audits.md).
- **Next step:** A code-audit PASS and current regression evidence permit the
  emergency delivery checkpoint.

***

Feature branch: `master`

Created: 2026-09-05

Status: Approved emergency change

Input: Add configurable clearance around quoted and fenced-code manuscript
blocks, plus a reusable six-property font block for quoted text, without
changing code-block monospace typography.

## Scope

- In scope: Markdown blockquotes and fenced code in manuscript generation gain
  **equal spacing above and below**.
- In scope: Org manuscript blockquotes and source blocks gain the same layout
  through the existing canonical Markdown conversion.
- In scope: Quoted blocks gain the **Feature 001 font-block shape**.
- In scope: Existing side-specific code-block spacing remains supported.
- Out of scope: Inline-code typography, code-block typography, paragraph
  spacing, source parsing, and non-manuscript rendering **do not change**.
- Out of scope: This emergency change does not activate the full
  [Unified Font Configuration](../001-unified-font-config/spec.md) migration.

## User Scenarios & Testing

### User Story 1 - Separate quoted text from prose (Priority: P1)

The user **sets one quoted-block spacing value** so quoted text has deliberate
clearance from the paragraphs around it.

Why this priority: The current paragraph rhythm leaves quoted passages too close
to neighbouring prose.

Independent Test: A configured Markdown blockquote renders with the requested
space above and below.

#### Acceptance Scenarios

1. GIVEN a manuscript blockquote between two prose paragraphs

   WHEN `quoted-block-spacing` is configured

   THEN the blockquote has **that clearance above and below**

2. GIVEN a quoted-block `font` mapping containing only `style: italic`

   WHEN the manuscript is rendered

   THEN quoted text is **italic**

   AND the other five font properties **inherit independently**

3. GIVEN an invalid quoted-block spacing or font value

   WHEN manuscript generation starts

   THEN it **identifies the full configuration path and returns non-zero**

   AND it **writes no output artefact**

### User Story 2 - Separate fenced code from prose (Priority: P1)

The user **sets one code-block spacing value** without altering the established
monospace font.

Why this priority: Fenced code currently sits too close to neighbouring prose.

Independent Test: A configured fenced block uses equal clearance while
retaining the manuscript monospace font.

#### Acceptance Scenarios

1. GIVEN fenced code between prose paragraphs

   WHEN `code-block-spacing` is configured

   THEN the code block has **that clearance above and below**

   AND its **monospace typography is unchanged**

2. GIVEN `code-block-spacing` and a side-specific `code-block.space-before`

   WHEN the manuscript is rendered

   THEN `space-before` **overrides the upper clearance only**

   AND the lower clearance **retains `code-block-spacing`**

### Edge Cases

- Omitted equal-spacing properties use `0.5em`.
- `code-block.space-before` and `code-block.space-after` override their
  respective sides independently.
- A partial quoted-block font overrides only its declared properties.
- Public quote style `regular` renders through Typst's equivalent `normal`
  value.

## Requirements

### Functional Requirements

- FR-001 - **Configure quoted-block clearance**: `folio.manuscript.quoted-block-spacing`
  MUST set equal space above and below manuscript blockquotes and MUST default to
  `0.5em`.
- FR-002 - **Configure code-block clearance**: `folio.manuscript.code-block-spacing`
  MUST set equal space above and below fenced-code blocks and MUST default to
  `0.5em`.
- FR-003 - **Preserve precise code spacing**: Existing
  `folio.manuscript.code-block.space-before` and `.space-after` values MUST
  override `code-block-spacing` on their respective sides.
- FR-004 - **Configure quote typography**: `folio.manuscript.quoted-block.font`
  MUST accept `family`, `size`, `weight`, `stretch`, `style`, and
  `letter-spacing`.
- FR-005 - **Inherit quote font properties**: Each omitted quoted-block font
  property MUST inherit independently from the equivalent manuscript font
  property, using `100%` stretch where the current manuscript configuration has
  no explicit stretch property.
- FR-006 - **Map regular style**: Public quote style `regular` MUST render as
  Typst `normal`; `italic` and `oblique` MUST retain their public meanings.
- FR-007 - **Preserve code typography**: Fenced and inline code MUST retain the
  existing manuscript monospace font configuration.
- FR-008 - **Preserve manuscript semantics**: The change MUST NOT alter
  blockquote or code content, paragraph spacing, headings, pagination controls,
  or non-manuscript output.
- FR-009 - **Reject invalid values before output**: Invalid equal-spacing,
  quote-font size, weight, stretch, style, or letter-spacing values MUST return
  non-zero with the full configuration path and MUST produce no output
  artefact. Quote-font family values MUST be escaped for Typst string context.

### Key Entities

- **Equal block spacing**: One length applied above and below a quoted or fenced
  block.
- **Quoted-block font**: The shared six-property font mapping applied to quoted
  text.
- **Precise code spacing**: Existing independent clearance above or below a
  fenced-code block.

## Success Criteria

### Measurable Outcomes

- SC-001 - **Apply equal spacing**: Configured quoted and fenced-code blocks
  emit the requested value on **both sides**.
- SC-002 - **Preserve precise overrides**: Each configured side-specific code
  value affects **only its named side**.
- SC-003 - **Expose one quote font vocabulary**: Quoted text accepts **all six
  Feature 001 font properties**.
- SC-004 - **Retain regression protection**: The maintained Markdown and Org
  manuscript examples containing blockquotes and fenced code continue to
  compile and retain equivalent textual content.
- SC-005 - **Fail safely**: Every invalid new spacing or quote-font value is
  rejected before output with its full configuration path.

## Assumptions

- The project owner authorized immediate delivery with `BYPASS-GATE-7` on
  2026-09-05.
- Feature 001 remains the authority for the future project-wide font migration;
  this feature adopts only its font-block shape for quoted manuscript text.
- Existing side-specific code-block spacing is more precise than the new equal
  spacing and therefore wins on the side it declares.

## Existing Baseline

- Requirement authority consulted: `docs/ACs.org` requirements **AC9.3 -
  Markdown manuscript elements**, **AC9.4 - Org manuscript elements**, and
  **AC9.5 - monospace prose and code**.
- Design authority consulted: `ARCHITECTURE.md`, `docs/config.md`, and
  `docs/format-manuscript-markdown.md` establish the manuscript configuration
  and canonical Markdown rendering boundary.
- Regression evidence consulted: Maintained tests and examples under
  `internal/manuscript` and `examples` establish blockquote, fenced-code,
  Markdown, Org, Typst, and PDF behaviour.
- Preserves: Existing source parsing, content, monospace typography,
  side-specific code spacing, and non-manuscript rendering.
- Changes: Blockquotes gain dedicated spacing and font configuration; fenced
  code gains an equal-spacing shorthand.
- Supersedes: No established requirement is removed or weakened.
