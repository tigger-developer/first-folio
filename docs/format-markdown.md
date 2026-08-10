<!-- Version: 0.4 | Last updated: 2026-08-10 -->

# Markdown Play Format

First Folio uses a convention-based Markdown format to represent stage plays. The format uses standard Markdown elements (headers, bold, italic, tables) with specific structural conventions that allow round-trip parsing.

**External references:**
- [CommonMark Specification](https://commonmark.org)
- [GitHub Flavoured Markdown](https://github.github.com/gfm/) (tables)

**Intro sections:** `##` headings before the first character dialogue (e.g. Synopsis, Setting, Scene List) are automatically detected as intro material. These render identically to act headers but can be toggled with `render.frontmatter` in [config](config.md).

## Element Schema

### Front Matter

The document title is a level-1 ATX heading. The subtitle (if present) appears as a bold line below the title. The author appears as an italic string prefixed with "by". An optional version/date line uses the exact delimiter form shown below.

```markdown
# The Importance of Being Earnest

**A Trivial Comedy for Serious People**

*by Oscar Wilde*

--- Draft 3 | 2026-08-10 ---
```

`title`, `subtitle`, `author`, `version`, and `date` are represented. Other front matter keys are not included in Markdown output.

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
# A Short Play

*A Trivial Comedy*

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
