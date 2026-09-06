# Feature Specification: [FEATURE NAME]

Feature branch: `[FEATURE BRANCH]`

Created: [DATE]

Status: Draft

Input: [OPERATOR REQUEST]

<!--
Follow the ABC presentation contract in ~/.agents/sdlc/ISSUES.md. Use ordinary
Markdown headings and scan-friendly bullets. Bold the semantic spine of each
statement: the smallest words or phrases carrying its distinctive state,
action, qualifier, quantity, boundary, or outcome. Do not bold headings, generic
labels, identifiers, modal verbs, GIVEN/WHEN/THEN/AND signposts, or context
already supplied nearby. The bold fragments must form an accurate compressed
summary when read in order. Remove every instruction comment and unused optional
section from the completed specification.

The complete fictional presentation example is at:
~/.agents/sdlc/presets/sdlc-standards/examples/spec-example.md
Use it only to understand structure and presentation. Never copy its
requirements, terminology, or behaviour into a project specification.
-->

## Scope

- In scope: [Observable behaviour introduced or changed.]
- Out of scope: [Adjacent behaviour deliberately left unchanged.]

## User Scenarios & Testing

<!--
Keep each user story to one or two sentences establishing actor, purpose, and
value. Do not narrate the feature or repeat its requirements here. Acceptance
scenarios carry the concrete behavioural examples.
-->

### User Story 1 - [BRIEF OUTCOME] (Priority: P1)

[State the actor, purpose, and value briefly.]

Why this priority: [Explain the value in one sentence.]

Independent Test: [State one independently observable outcome, not a test procedure.]

#### Acceptance Scenarios

1. GIVEN [initial state]

   WHEN [action]

   THEN [observable outcome]

   AND [additional outcome, when needed]

<!-- Add only the stories and scenarios required by the feature. -->

### Edge Cases

- [Material boundary, alternate path, or failure case.]

## Requirements

### Functional Requirements

<!--
State the authoritative generalized rules demonstrated by the scenarios. Give
every identifier a descriptor. Each requirement must be observable, falsifiable,
bounded, and free of implementation or test detail unless the mechanism is part
of the public contract.
-->

- FR-001 - [DESCRIPTIVE REQUIREMENT TITLE]: [State one behavioural rule.]

### Key Entities

<!-- Include only when the feature involves material data or domain entities. -->

- [ENTITY]: [Meaning and relationships without implementation detail.]

## Success Criteria

### Measurable Outcomes

<!--
Measure overall feature success. Do not restate each functional requirement or
invent arbitrary targets merely to populate this section.
-->

- SC-001 - [DESCRIPTIVE OUTCOME]: [State a measurable, technology-neutral result.]

## Assumptions

- [Material assumption, dependency, or none.]

## Existing Baseline

<!--
Required for brownfield work. Remove this section for greenfield work. Record a
concise delta relationship; do not copy the existing system specification.
-->

- Sources consulted: [Current requirements, design, relevant work history,
  regression evidence, and affected implementation.]
- Preserves: [Existing behaviour and decisions retained.]
- Changes: [Existing behaviour changed by this specification.]
- Supersedes: [Existing requirements explicitly replaced, or none.]
- Unaffected: [Important compatibility and ownership boundaries.]
