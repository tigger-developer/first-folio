// ABOUTME: Defines script configuration leaves accepted from dotted Org frontmatter.
// ABOUTME: Reuses embedded preset keys and supplements optional public script fields.
package config

import "strings"

// ScriptSourcePaths returns the public leaf paths accepted from Org script
// frontmatter. The source syntax is a projection of the existing YAML tree, not
// a second configuration schema.
func ScriptSourcePaths() (map[string]struct{}, error) {
	paths := map[string]struct{}{}
	for _, name := range []string{"presets/british.yaml", "presets/us.yaml", "presets/screenplay.yaml"} {
		data, err := readEmbedded(name)
		if err != nil {
			return nil, err
		}
		collectSourcePaths(paths, "", data)
	}
	for _, path := range []string{
		"folio.cover-image.path",
		"folio.cover-image.mode",
		"folio.cover-image.placement",
		"folio.cover-image.size",
		"folio.cover-image.image.treatment",
		"folio.cover-image.image.opacity",
		"folio.cover-image.superscale",
		"folio.cover-image.dodge.padding",
		"folio.cover-image.overlay.blur",
		"folio.cover-image.overlay.wash.treatment",
		"folio.cover-image.overlay.wash.opacity",
		"folio.cover-image.overlay.padding",
		"folio.page-header.frontmatter-format",
		"folio.page-header.alt-frontmatter-format",
		"folio.page-footer.alt-format",
		"folio.page-footer.frontmatter-format",
		"folio.page-footer.alt-frontmatter-format",
	} {
		paths[path] = struct{}{}
	}
	return paths, nil
}

func collectSourcePaths(paths map[string]struct{}, prefix string, data map[string]any) {
	for key, value := range data {
		path := key
		if prefix != "" {
			path = prefix + "." + key
		}
		if child, ok := value.(map[string]any); ok {
			collectSourcePaths(paths, path, child)
			continue
		}
		if strings.HasPrefix(path, "folio.") || strings.HasPrefix(path, "render.") {
			paths[path] = struct{}{}
		}
	}
}
