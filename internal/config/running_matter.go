// ABOUTME: Projects shared running matter into the manuscript renderer's compatibility paths.
// ABOUTME: Normalizes each layer before merging so deprecated overrides never outlive precedence.
package config

import (
	"fmt"
	"math"
	"strings"
)

// runningMatterLayer keeps canonical values independent of manuscript-only
// overrides. The projection intentionally retains the renderer's existing typed
// paths and template contract until the deprecated public paths are retired.
func runningMatterLayer(data map[string]any, mode Mode) map[string]any {
	layer := cloneMapping(data)
	cfg := Config{data: layer}
	for _, role := range []string{"page-header", "page-footer"} {
		shared, present := cfg.Get("folio." + role)
		legacy, legacyPresent := cfg.Get("folio.manuscript." + role)
		if present {
			setPath(layer, "folio.manuscript."+role, cloneValue(shared))
		}
		if mode == ModeManuscript && legacyPresent {
			sharedBlock, sharedOK := shared.(map[string]any)
			legacyBlock, legacyOK := legacy.(map[string]any)
			if present && sharedOK && legacyOK {
				effective := cloneMapping(sharedBlock)
				deepMerge(effective, legacyBlock)
				setPath(layer, "folio.manuscript."+role, effective)
			} else {
				setPath(layer, "folio.manuscript."+role, cloneValue(legacy))
			}
		} else if !present && legacyPresent {
			manuscript, _ := cfg.Get("folio.manuscript")
			delete(manuscript.(map[string]any), role)
		}
	}
	return layer
}

func cloneValue(value any) any {
	if mapping, ok := value.(map[string]any); ok {
		return cloneMapping(mapping)
	}
	return value
}

// Letters neither consume nor validate running matter, including legacy fields.
func isRunningMatterPath(path string) bool {
	for _, prefix := range []string{"folio.page-header", "folio.page-footer", "folio.manuscript.page-header", "folio.manuscript.page-footer"} {
		if path == prefix || strings.HasPrefix(path, prefix+".") {
			return true
		}
	}
	return false
}

func cloneMapping(data map[string]any) map[string]any {
	result := make(map[string]any, len(data))
	for key, value := range data {
		result[key] = cloneValue(value)
	}
	return result
}

// Warnings returns migration diagnostics without writing to process streams.
// Only user-supplied deprecated manuscript blocks produce diagnostics.
func (c Config) Warnings() []string {
	return append([]string(nil), c.warnings...)
}

func legacyRunningMatterWarnings(layer map[string]any) []string {
	warnings := []string{}
	for _, role := range []string{"page-header", "page-footer"} {
		path := "folio.manuscript." + role
		if _, present := (Config{data: layer}).Get(path); present {
			warnings = append(warnings, path+" is deprecated; move it to folio."+role+" (legacy fields remain manuscript-only overrides)")
		}
	}
	return warnings
}

func validateRunningMatter(data map[string]any) error {
	cfg := Config{data: data}
	for _, prefix := range []string{"folio.", "folio.manuscript."} {
		for _, role := range []string{"page-header", "page-footer"} {
			path := prefix + role
			raw, _ := cfg.Get(path)
			block, ok := raw.(map[string]any)
			if !ok {
				return fmt.Errorf("%s must be a mapping", path)
			}
			for key, value := range block {
				field := path + "." + key
				switch key {
				case "font", "align": // Validated by the shared font and alignment contracts.
				case "enabled":
					if _, ok := value.(bool); !ok {
						return fmt.Errorf("%s must be true or false", field)
					}
				case "frontmatter-format", "alt-frontmatter-format":
					if value == nil {
						continue
					}
					fallthrough
				case "format", "alt-format":
					if _, ok := value.(string); !ok {
						return fmt.Errorf("%s must be a string", field)
					}
				case "distance-from-edge", "content-padding-after":
					text, ok := value.(string)
					text = strings.TrimSpace(text)
					if !ok || !fontSizeRE.MatchString(text) || numericPart(text) < 0 || math.IsInf(numericPart(text), 0) || math.IsNaN(numericPart(text)) {
						return fmt.Errorf("%s must be a non-negative length in pt, mm, cm, in, or em", field)
					}
					block[key] = text
				default:
					// Unknown keys remain available to companion tools; only known fields are validated.
				}
			}
		}
	}
	return nil
}
