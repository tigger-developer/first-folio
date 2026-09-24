// ABOUTME: Validates resolved alignment fields before any renderer consumes them.
// ABOUTME: Keeps explicit grouped placement distinct from invalid empty configuration.
package config

import (
	"fmt"
	"strings"
)

// Each family lists the public fields sharing one alignment vocabulary. Defaults
// belong in the presets; missing/null values are errors after layer merging.
var alignmentFamilies = []struct {
	paths  []string
	values []string
}{
	{[]string{
		"folio.title-page.title.align", "folio.title-page.subtitle.align", "folio.title-page.author.align",
		"folio.positioning.speech.align", "folio.positioning.speech.speaker.align",
		"folio.positioning.speech.speech-instruction.align", "folio.positioning.speech.dialogue.align",
		"folio.positioning.stage-direction.align", "folio.positioning.transition.align",
		"folio.positioning.frontmatter.header.align", "folio.positioning.act-header.align", "folio.positioning.scene-header.align",
		"folio.manuscript.part.align", "folio.manuscript.chapter.align", "folio.manuscript.copyright.align",
	}, []string{"left", "center", "right"}},
	{[]string{"folio.manuscript.part.vertical-align"}, []string{"top", "center", "bottom", "middle", "horizon"}},
	{[]string{"folio.manuscript.page-header.align", "folio.manuscript.page-footer.align"},
		[]string{"left", "center", "right", "left-left", "left-center", "left-right", "center-left", "center-center", "center-right", "right-left", "right-center", "right-right"}},
	{[]string{"folio.manuscript.title-page.title-block-align", "folio.manuscript.title-page.footer-align", "folio.manuscript.title-page.contact.align"}, titlePlacementValues},
	{[]string{
		"folio.manuscript.title-page.title.align", "folio.manuscript.title-page.subtitle.align",
		"folio.manuscript.title-page.author.align", "folio.manuscript.title-page.date.align",
		"folio.manuscript.title-page.wordcount.align", "folio.manuscript.title-page.version.align",
	}, append(append([]string{}, titlePlacementValues...), "group")},
}

var titlePlacementValues = []string{
	"left", "center", "right", "top-left", "top-center", "top-right",
	"center-left", "center-center", "center-right", "bottom-left", "bottom-center", "bottom-right",
}

func validateAlignments(data map[string]any) error {
	cfg := Config{data: data}
	for _, family := range alignmentFamilies {
		for _, path := range family.paths {
			raw, _ := cfg.Get(path)
			value, isString := raw.(string)
			value = strings.TrimSpace(value)
			if path == "folio.manuscript.part.vertical-align" {
				value = strings.ToLower(value) // Preserve the existing vertical keyword aliases.
			}
			valid := false
			for _, allowed := range family.values {
				if isString && value == allowed {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("%s has invalid alignment %#v; expected %s", path, raw, strings.Join(family.values, ", "))
			}
			setPath(data, path, value)
		}
	}
	return nil
}
