<!-- Version: 0.4 | Last updated: 2026-08-10 -->

# Org-mode Play Format

Org-mode is the primary authoring format for First Folio. It uses Emacs org-mode heading levels to encode the hierarchical structure of a stage play. This format supports all event-stream elements without loss.

**External reference:** [orgmode.org - Document Structure](https://orgmode.org/manual/Document-Structure.html)

**Intro sections:** H1 headings before the first character dialogue (e.g. Synopsis, Setting, Scene List) are automatically detected as intro material. These render the same as act headers but can be toggled with `render.frontmatter` in [config](config.md).

## Element Schema

### Front Matter

Org-mode keyword lines at the top of the file. Any `#+KEY: value` line is captured by the parser. Org emission preserves `TITLE`, `SUBTITLE`, `AUTHOR`, `DATE`, and `VERSION`; arbitrary keys are not guaranteed to survive a text-format round trip.

```org
#+TITLE: The Importance of Being Earnest
#+AUTHOR: Oscar Wilde
#+SUBTITLE: A Trivial Comedy for Serious People
#+TEMPLATE: play
```

Standard emitted keys are `TITLE`, `SUBTITLE`, `AUTHOR`, `DATE`, and `VERSION`.

### Acts (H1)

Level-1 headings represent acts. The heading text becomes the act title.

```org
* Act I
* Act II
* Epilogue
```

The special heading `* CHARACTERS` (case-insensitive, singular or plural) is not an act - it introduces a character table (see below).

### Scenes (H2)

Level-2 headings represent scenes within an act.

```org
** Scene 1
** Scene 2 — The Garden
```

### Stage Directions (H3)

Level-3 headings represent stage directions (also called action text or narrative). These describe setting, movement, and physical action.

```org
*** A morning room in Algernon's flat in Half-Moon Street.
*** JACK enters through the French windows.
```

Org-mode inline markup is retained in event text and interpreted by the Org-aware PDF renderer. Text-format emitters do not translate every inline marker between Org, Markdown, and Fountain, so inline emphasis is not guaranteed to retain identical semantics across a text-format round trip.

### Characters (H4)

Level-4 headings introduce a character who is about to speak. The heading contains the character name in ALL CAPS, optionally followed by a parenthetical direction.

```org
**** ALGERNON
**** JACK earnestly
**** LADY BRACKNELL (rising)
**** GWENDOLEN, with great feeling
```

All of the following direction formats are normalized to the same internal representation:

| Syntax | Name | Direction |
|--------|------|-----------|
| `**** BOB softly` | BOB | softly |
| `**** BOB (softly)` | BOB | softly |
| `**** BOB, softly` | BOB | softly |
| `**** BOB, (softly)` | BOB | softly |
| `**** BOB` | BOB | (none) |

Character names support Unicode: `CÁIT`, `MAIRÉAD`, `SÉAN`.

### Dialogue

Plain text following an H4 heading. Multiple lines are preserved. Blank lines are consumed silently.

```org
**** ALGERNON
I don't think there is much likelihood, Jack,
of you and Miss Fairfax being united.
```

The same inline-markup limitation applies to dialogue: PDF rendering understands the Org source markers, while text-format emission preserves the event text rather than normalizing all target markers.

### Character Table

A level-1 heading `* CHARACTERS` (or `* CHARACTER`) followed by an org table listing the cast.

```org
* CHARACTERS
|----------+------------------------------|
| ALGERNON | A young man about town       |
| JACK     | His friend, also young       |
| LANE     | Algernon's manservant        |
|----------+------------------------------|
```

Separator rows (`|---+---|`) are ignored. Each data row emits a `character_table_row` event with the name and description.

### Prop Text

On-stage text (signs, letters, placards) uses a level-3 heading tagged `:prop:`. The parser also accepts the historical list-item form.

```org
*** WELCOME TO THE GARDEN PARTY                              :prop:

- *"WELCOME TO THE GARDEN PARTY"*
```

Both forms emit the same `prop_text` event. First Folio emits the tagged-heading form when writing Org.

### Transitions

A transition uses a level-3 heading tagged `:transition:`. The parser also accepts a level-5 heading as a compatibility form.

```org
*** BLACKOUT                                                :transition:
```

### Footnotes

Org-mode footnote definitions. The reference `[fn:name]` appears inline in dialogue or directions; the definition appears on its own line.

```org
[fn:verse] From Tennyson's "In Memoriam", Canto 27.
```

### Noexport Sections

Any heading tagged `:noexport:` is excluded from output, along with all its children.

```org
*** Notes on staging :noexport:
These are private notes and will not appear in any output.
```

## Complete Example

```org
#+TITLE: A Short Play
#+AUTHOR: A. Playwright

* CHARACTERS
|------+------------------|
| BOB  | An ordinary man  |
| CÁIT | His neighbour    |
|------+------------------|

* Act I
** Scene 1
*** A kitchen. Morning. Sunlight through the window.
**** BOB
Good morning.

*** BOB crosses to the kettle.

**** CÁIT (entering)
Is the kettle on?

**** BOB cheerfully
Just boiled.

*** Research notes :noexport:
- Look up kettle brands for period accuracy.
```
