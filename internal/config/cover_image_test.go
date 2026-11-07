// ABOUTME: Defines W073 cover-image configuration validation and strict-key regression cases.
// ABOUTME: Protects the new block without making established configuration namespaces strict.
package config

import (
	"strings"
	"testing"
)

func TestRT073_1CoverImageConfigurationKeysAreValidated(t *testing.T) {
	valid := map[string]any{"folio": map[string]any{"cover-image": map[string]any{
		"path": "cover.png", "mode": "inset", "placement": "overlay", "size": "80mm", "superscale": "4x",
		"image":   map[string]any{"treatment": "dark", "opacity": "75%"},
		"overlay": map[string]any{"blur": "2mm", "padding": "1em", "wash": map[string]any{"treatment": "light", "opacity": "40%"}},
	}}}
	if _, err := Load(Options{Mode: ModeScript, Home: t.TempDir(), LocalDir: t.TempDir(), Source: valid}); err != nil {
		t.Fatalf("valid cover-image configuration: %v", err)
	}

	invalid := []struct {
		name, path string
		cover      map[string]any
	}{
		{"unknown direct key", "folio.cover-image.typo", map[string]any{"path": "cover.png", "mode": "background", "typo": true}},
		{"unknown image key", "folio.cover-image.image.typo", map[string]any{"path": "cover.png", "mode": "background", "image": map[string]any{"typo": true}}},
		{"unknown overlay key", "folio.cover-image.overlay.typo", map[string]any{"path": "cover.png", "mode": "background", "overlay": map[string]any{"typo": true}}},
		{"unknown wash key", "folio.cover-image.overlay.wash.typo", map[string]any{"path": "cover.png", "mode": "background", "overlay": map[string]any{"wash": map[string]any{"typo": true}}}},
		{"missing mode", "folio.cover-image.mode", map[string]any{"path": "cover.png"}},
		{"invalid mode", "folio.cover-image.mode", map[string]any{"path": "cover.png", "mode": "stretch"}},
		{"background size", "folio.cover-image.size", map[string]any{"path": "cover.png", "mode": "background", "size": "80mm"}},
		{"inset missing size", "folio.cover-image.size", map[string]any{"path": "cover.png", "mode": "inset"}},
		{"background dodge", "folio.cover-image.placement", map[string]any{"path": "cover.png", "mode": "background", "placement": "dodge"}},
		{"invalid opacity", "folio.cover-image.image.opacity", map[string]any{"path": "cover.png", "mode": "background", "image": map[string]any{"opacity": "101%"}}},
		{"invalid blur unit", "folio.cover-image.overlay.blur", map[string]any{"path": "cover.png", "mode": "background", "overlay": map[string]any{"blur": "1em"}}},
		{"invalid padding unit", "folio.cover-image.overlay.padding", map[string]any{"path": "cover.png", "mode": "background", "overlay": map[string]any{"padding": "1pt"}}},
		{"invalid superscale", "folio.cover-image.superscale", map[string]any{"path": "cover.png", "mode": "background", "superscale": "16x"}},
	}
	for _, tc := range invalid {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Load(Options{Mode: ModeScript, Home: t.TempDir(), LocalDir: t.TempDir(), Source: map[string]any{"folio": map[string]any{"cover-image": tc.cover}}})
			if err == nil || !strings.Contains(err.Error(), tc.path) {
				t.Fatalf("expected %s validation error, got %v", tc.path, err)
			}
		})
	}

	if _, err := Load(Options{Mode: ModeScript, Home: t.TempDir(), LocalDir: t.TempDir(), Source: map[string]any{"folio": map[string]any{"unrelated-future-key": true}}}); err != nil {
		t.Fatalf("cover-image strictness leaked outside its block: %v", err)
	}
}
