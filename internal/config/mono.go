// ABOUTME: Projects the shared monospace role into the manuscript compatibility path.
// ABOUTME: Preserves legacy input while making the top-level role authoritative per layer.
package config

func monospaceLayer(data map[string]any, mode Mode) map[string]any {
	layer := cloneMapping(data)
	cfg := Config{data: layer}
	shared, sharedPresent := cfg.Get("folio.mono")
	legacy, legacyPresent := cfg.Get("folio.manuscript.mono")

	var effective any
	switch {
	case sharedPresent && legacyPresent:
		sharedBlock, sharedOK := shared.(map[string]any)
		legacyBlock, legacyOK := legacy.(map[string]any)
		if sharedOK && legacyOK {
			effectiveBlock := cloneMapping(legacyBlock)
			deepMerge(effectiveBlock, sharedBlock)
			effective = effectiveBlock
		} else {
			effective = cloneValue(shared)
		}
	case sharedPresent:
		effective = cloneValue(shared)
	case legacyPresent:
		effective = cloneValue(legacy)
	}
	if effective != nil {
		setPath(layer, "folio.mono", cloneValue(effective))
		setPath(layer, "folio.manuscript.mono", cloneValue(effective))
	} else if mode != ModeManuscript {
		if manuscript, ok := cfg.Get("folio.manuscript"); ok {
			delete(manuscript.(map[string]any), "mono")
		}
	}
	return layer
}

func legacyMonospaceWarnings(layer map[string]any) []string {
	if _, present := (Config{data: layer}).Get("folio.manuscript.mono"); !present {
		return nil
	}
	return []string{"folio.manuscript.mono.font is deprecated; move it to folio.mono.font"}
}
