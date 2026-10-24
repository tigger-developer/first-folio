// ABOUTME: Defines the typed stage-play document shared by parsers and emitters.
// ABOUTME: Keeps format syntax separate from semantic play events.
package play

type EventKind string

const (
	EventFrontMatter         EventKind = "front_matter"
	EventActHeader           EventKind = "act_header"
	EventSceneHeader         EventKind = "scene_header"
	EventStageDirection      EventKind = "stage_direction"
	EventCharacter           EventKind = "character"
	EventDialogue            EventKind = "dialogue"
	EventCharacterTableStart EventKind = "character_table_start"
	EventCharacterTableRow   EventKind = "character_table_row"
	EventCharacterTableEnd   EventKind = "character_table_end"
	EventPropText            EventKind = "prop_text"
	EventFootnote            EventKind = "footnote"
	EventTransition          EventKind = "transition"
	EventIntroHeader         EventKind = "intro_header"
	EventIntroText           EventKind = "intro_text"
	EventSceneIntro          EventKind = "scene_intro"
)

type Event struct {
	Kind      EventKind
	Key       string
	Name      string
	Text      string
	Direction string
}

type Document struct {
	Metadata map[string]string
	Events   []Event
}

type Output struct {
	Text     string
	Warnings []string
}
