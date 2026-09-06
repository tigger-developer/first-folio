# Feature Specification: Manuscript Block Layout

## Specification Summary

- **Outcome:** Manuscript blockquotes and fenced code have configurable
  clearance and whole-block indentation, and blockquotes have configurable
  typography; a code span occupying a complete source line uses the same block
  layout as fenced code; prose after a block retains normal paragraph
  indentation.
- **Before:** Fenced-code clearance used only side-specific settings;
  blockquotes had no dedicated clearance or font contract.
- **After:** Equal spacing is available through `quoted-block-spacing` and
  `code-block-spacing`; whole-block left indentation is available through
  `quote-block-indent` and `code-block-indent`; quoted blocks use the shared
  six-property `font` block; complete-line code spans become code blocks; only
  a paragraph directly opening a chapter is flush left.
- **Changes:** Side-specific code-block spacing remains available and overrides
  the equal-spacing value on its respective side; block indentation is separate
  from prose first-line indentation; structural blocks no longer suppress the
  following paragraph's configured first-line indent.
- **Unchanged:** Code embedded within a mixed-content line remains inline;
  inline and block code retain the manuscript monospace font; paragraph spacing
  and the configured indent length do not change.
- **Edge cases:** Omitted quote-font properties inherit independently;
  side-specific code spacing may override only one side; invalid values fail
  before output; omitted block indents preserve the existing layout; existing
  fences and code spans sharing a source line with prose are not reclassified.
- **Decisions:** Both spacing defaults are `0.5em`; both indent defaults are
  `0em`; the public quote style `regular` maps to Typst `normal`; complete-line
  classification is purely syntactic and never infers a programming language.
- **Evidence:** The project owner's `BYPASS-GATE-7` instruction dated 2026-09-05,
  the maintained manuscript regression pack, and [audits.md](audits.md).
- **Next step:** A code-audit PASS and current regression evidence permit the
  emergency delivery checkpoint.

***

Feature branch: `master`

Created: 2026-09-05

Status: Approved emergency change

Input: Add configurable clearance around quoted and fenced-code manuscript
blocks, separate whole-block left indents, and a reusable six-property font
block for quoted text; classify a complete-line code span as a code block
without changing mixed-content inline-code semantics; preserve normal prose
indentation after blocks while keeping only chapter-opening prose flush left.

## Scope

- In scope: Markdown blockquotes and fenced code in manuscript generation gain
  **equal spacing above and below** and **separate whole-block left indents**.
- In scope: Org manuscript blockquotes and source blocks gain the same layout
  through the existing canonical Markdown conversion.
- In scope: Quoted blocks gain the **Feature 001 font-block shape**.
- In scope: Existing side-specific code-block spacing remains supported.
- In scope: A Markdown code span that is the only non-whitespace content on its
  source line gains fenced-code block semantics. Org verbatim spans gain the
  same behaviour through the existing canonical Markdown conversion.
- In scope: Prose paragraphs after code, quote, and other structural blocks use
  the configured `paragraph-indent`; a prose paragraph directly opening a
  chapter remains flush left.
- Out of scope: Inline-code typography, code-block typography, paragraph
  spacing and first-line indentation, all other source parsing, and
  non-manuscript rendering **do not change**. The feature does not parse code or
  infer a programming language.
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

4. GIVEN a manuscript blockquote containing multiple rendered lines

   WHEN `quote-block-indent` is configured

   THEN **every line in the block is inset from the left by that value**

   AND prose first-line indentation **does not change**

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

3. GIVEN a fenced code block containing multiple lines

   WHEN `code-block-indent` is configured

   THEN **every line in the block is inset from the left by that value**

   AND its **monospace typography is unchanged**

4. GIVEN a code span that is the only non-whitespace content on its source line

   WHEN the manuscript is rendered

   THEN it uses **code-block spacing, indentation, and monospace typography**

5. GIVEN a code span preceded by `but ` or any other source content outside its
   delimiters on the same line

   WHEN the manuscript is rendered

   THEN it **remains inline code**

### User Story 3 - Preserve prose indentation after blocks (Priority: P1)

The user reads prose after displayed code or quotation with the same first-line
indent used by other body paragraphs.

Independent Test: A rendered manuscript leaves chapter-opening prose flush and
applies the configured indent to prose after code and quote blocks.

#### Acceptance Scenarios

1. GIVEN a prose paragraph directly after a chapter heading

   WHEN the manuscript is rendered

   THEN that paragraph is **flush left**

2. GIVEN prose immediately after a code block or blockquote

   WHEN the manuscript is rendered

   THEN its first line uses the configured **`paragraph-indent`**

3. GIVEN a chapter whose first content is a code block or blockquote

   WHEN a prose paragraph follows that block

   THEN the paragraph uses the configured **`paragraph-indent`**

### Edge Cases

- Omitted equal-spacing properties use `0.5em`.
- Omitted block-indent properties use `0em` and preserve the existing layout.
- `code-block.space-before` and `code-block.space-after` override their
  respective sides independently.
- A partial quoted-block font overrides only its declared properties.
- Public quote style `regular` renders through Typst's equivalent `normal`
  value.
- Surrounding source-line whitespace does not prevent a complete-line code span
  from becoming a block.
- Existing fenced code is preserved without reclassification.
- A prose prefix such as `but ` keeps the span inline even when text beginning
  with `#` remains inside the code delimiters.
- A structural block at the beginning of a chapter does not grant flush-left
  treatment to a later prose paragraph.

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
  block-indent, quote-font size, weight, stretch, style, or letter-spacing
  values MUST return non-zero with the full configuration path and MUST produce
  no output artefact. Quote-font family values MUST be escaped for Typst string
  context.
- FR-010 - **Indent quoted blocks**: `folio.manuscript.quote-block-indent` MUST
  apply the configured left inset to every rendered line of a manuscript
  blockquote, MUST remain independent of prose first-line indentation, and MUST
  default to `0em`.
- FR-011 - **Indent code blocks**: `folio.manuscript.code-block-indent` MUST
  apply the configured left inset to every rendered line of a fenced-code
  block, MUST remain independent of prose first-line indentation, and MUST
  default to `0em`.
- FR-012 - **Promote complete-line code spans**: A Markdown code span that is
  the only non-whitespace content on its source line MUST render with code-block
  semantics, including configured code-block spacing and indentation. Org
  verbatim spans MUST follow the same rule after canonical Markdown conversion.
  Classification MUST depend only on source delimiters and line occupancy and
  MUST NOT inspect code content or infer a programming language.
- FR-013 - **Preserve mixed-content inline code**: A code span with any content
  outside its delimiters on the same source line MUST remain inline and MUST NOT
  receive code-block spacing or indentation. Content inside the delimiters,
  including text beginning with `#`, MUST remain code and MUST NOT affect this
  classification.
- FR-014 - **Indent prose after blocks**: Every prose paragraph other than one
  directly opening a chapter MUST use the configured
  `folio.manuscript.paragraph-indent`, including prose after code blocks and
  blockquotes.
- FR-015 - **Keep chapter-opening prose flush**: A prose paragraph directly
  opening a chapter MUST remain flush left. If a structural block opens the
  chapter, a later prose paragraph MUST use the configured paragraph indent.

### Key Entities

- **Equal block spacing**: One length applied above and below a quoted or fenced
  block.
- **Quoted-block font**: The shared six-property font mapping applied to quoted
  text.
- **Precise code spacing**: Existing independent clearance above or below a
  fenced-code block.
- **Whole-block indent**: A left inset applied uniformly to every rendered line
  of a quoted or fenced-code block.
- **Complete-line code span**: Inline-code source syntax whose delimiters and
  surrounding whitespace occupy the whole source line.
- **Chapter-opening prose**: A prose paragraph directly following a chapter
  heading, before any structural block.

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
- SC-006 - **Apply whole-block indents**: Each configured block-indent value is
  emitted as the left inset of its named block type without changing paragraph
  indentation.
- SC-007 - **Classify source lines exactly**: Complete-line code spans receive
  block layout in Markdown and Org manuscripts while code spans sharing a line
  with prose remain inline.
- SC-008 - **Retain paragraph rhythm**: PDF layout evidence shows the first
  chapter paragraph at the body margin and every tested paragraph after code or
  quotation at the body margin plus the configured paragraph indent.

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
- Changes: Blockquotes gain dedicated spacing, indentation, and font
  configuration; fenced code gains equal-spacing and whole-block-indentation
  settings; complete-line code spans gain fenced-code layout semantics; blocks
  no longer suppress the following prose indent.
- Supersedes: No established requirement is removed or weakened.
