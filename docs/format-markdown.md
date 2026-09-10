---
version: "0.5"
updated: "2026-09-10"
---

# Markdown Play Format

First Folio supports the limited Markdown play syntax described here and in the [Markdown script schema](../schema/script.md). Support is limited to the specified structural conventions; it does not imply general Markdown conformance.

**Intro sections:** `##` headings before the first character dialogue (e.g. Synopsis, Setting, Scene List) are automatically detected as intro material. These render identically to act headers but can be toggled with `render.frontmatter` in [config](config.md).

## Element Schema

### Front Matter

The document begins with YAML frontmatter containing the destination `schema` URL. The document title is a level-1 ATX heading. The subtitle (if present) appears as a bold line below the title. The author appears as an italic string prefixed with "by". An optional version/date line uses the exact delimiter form shown below.

```markdown
---
schema: https://github.com/tigger-developer/first-folio/blob/master/schema/script.md
---

# The Importance of Being Earnest

**A Trivial Comedy for Serious People**

*by Oscar Wilde*

--- Draft 3 | 2026-08-10 ---
```

`schema` is emitted in YAML. `title`, `subtitle`, `author`, `version`, and `date` retain their visible forms. Other YAML fields do not replace those forms. The writer adds the schema even when the source has none; input schema values do not override the destination URL. Malformed YAML or unclosed frontmatter produces an error.

### Acts (H2)

Level-2 headings represent acts.

```markdown
## Act I
## Act II
## Epilogue
```

### Scenes (H3)

Level-3 headings represent scenes.

```markdown
### Scene 1
### Scene 2 — The Garden
```

### Stage Directions

Standalone italic paragraphs (a paragraph consisting entirely of `*text*`). These must be separated from surrounding elements by blank lines.

```markdown
*A morning room in Algernon's flat in Half-Moon Street.*

*JACK enters through the French windows.*
```

When parsing, a paragraph is identified as a stage direction if it begins and ends with `*` and contains no `**bold**` character-name pattern.

### Characters

A bold name followed by a colon, optionally followed by an italic parenthetical direction.

```markdown
**ALGERNON:**
**JACK:** *(earnestly)*
**LADY BRACKNELL:** *(rising)*
```

The parser detects the pattern `**NAME:**` at the start of a line. If `*(direction)*` follows on the same line, it is captured as the character's parenthetical.

### Dialogue

Plain text following a character line. Multiple lines are preserved as-is.

```markdown
**ALGERNON:**
I don't think there is much likelihood, Jack,
of you and Miss Fairfax being united.
```

A blank line after dialogue separates it from the next element.

### Character Table

A standard Markdown table with "Character" and "Description" headers (or similar).

```markdown
| Character | Description                  |
|-----------|------------------------------|
| ALGERNON  | A young man about town       |
| JACK      | His friend, also young       |
| LANE      | Algernon's manservant        |
```

### Prop Text

Bold-italic quoted text, standalone on its own line.

```markdown
***"WELCOME TO THE GARDEN PARTY"***
```

### Transitions

A blockquote-style line represents a transition. First Folio treats this as a dramatic transition, not a general Markdown blockquote.

```markdown
> BLACKOUT
> CUT TO:
```

### Footnotes

Standard Markdown footnote syntax.

```markdown
A famous verse[^verse] is quoted here.

[^verse]: From Tennyson's "In Memoriam", Canto 27.
```

## Complete Example

```markdown
---
schema: https://github.com/tigger-developer/first-folio/blob/master/schema/script.md
---

# A Short Play

**A Trivial Comedy**

*by A. Playwright*

| Character | Description        |
|-----------|--------------------|
| BOB       | An ordinary man    |
| CÁIT      | His neighbour      |

## Act I

### Scene 1

*A kitchen. Morning. Sunlight through the window.*

**BOB:**
Good morning.

*BOB crosses to the kettle.*

**CÁIT:** *(entering)*
Is the kettle on?

**BOB:** *(cheerfully)*
Just boiled.
```

## Revision History

- 0.5, 2026-09-10: Add schema frontmatter, move document-version metadata to YAML, and remove general Markdown-standard references that did not define First Folio's supported subset.
- 0.4, 2026-08-10: Previous format-reference revision.
