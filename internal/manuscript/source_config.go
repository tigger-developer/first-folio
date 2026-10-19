// ABOUTME: Extracts configuration from the first resolved Markdown manuscript input.
// ABOUTME: Keeps source metadata separate and rejects ambiguous later configuration layers.
package manuscript

import (
	"fmt"
	"os"
	"strings"
)

func readConfiguredInputs(inputs InputSet) (string, map[string]any, error) {
	if inputs.Format != "markdown" {
		text, err := ReadJoined(inputs.Paths)
		return text, nil, err
	}
	var joined strings.Builder
	var source map[string]any
	for index, path := range inputs.Paths {
		raw, err := os.ReadFile(path)
		if err != nil {
			return "", nil, fmt.Errorf("reading %s: %w", path, err)
		}
		values, _, probeErr := markdownFrontmatter(string(raw))
		if index == 0 && probeErr != nil {
			return "", nil, fmt.Errorf("%s: %w", path, probeErr)
		}
		var layer map[string]any
		// Later files are Markdown body, not metadata documents. A failed YAML
		// probe can be ordinary prose between scene breaks; normal body parsing
		// remains authoritative. Reject only a decoded configuration declaration.
		if probeErr == nil {
			layer, err = sourceConfiguration(values, path, index == 0)
			if err != nil {
				return "", nil, err
			}
		}
		if index == 0 {
			source = layer
		} else {
			joined.WriteString("\n\n")
		}
		joined.Write(raw)
		joined.WriteString("\n")
	}
	return joined.String(), source, nil
}

func sourceConfiguration(values map[string]any, path string, first bool) (map[string]any, error) {
	source := map[string]any{}
	for _, key := range []string{"folio", "render"} {
		value, exists := values[key]
		if !exists {
			continue
		}
		if !first {
			return nil, fmt.Errorf("%s: %s configuration frontmatter is only allowed in the first resolved manuscript input", path, key)
		}
		if _, ok := value.(map[string]any); !ok {
			return nil, fmt.Errorf("%s: frontmatter %s must be a mapping", path, key)
		}
		source[key] = value
	}
	return source, nil
}
