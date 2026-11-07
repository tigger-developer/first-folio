package config

import (
	"path/filepath"
	"testing"
)

func TestRT078_1SourceConfigurationOverridesLocalBaseAndStyleLayers(t *testing.T) {
	for _, tc := range []struct {
		name, path string
		mode       Mode
		source     map[string]any
	}{
		{
			name: "script", mode: ModeScript, path: "folio.margin",
			source: map[string]any{"folio": map[string]any{"style": "us", "margin": "20mm"}},
		},
		{
			name: "manuscript", mode: ModeManuscript, path: "folio.manuscript.margin",
			source: map[string]any{"folio": map[string]any{"manuscript": map[string]any{"style": "us", "margin": "20mm"}}},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			writeYAML(t, filepath.Join(dir, "script.yaml"), "folio:\n  style: british\n  margin: 10mm\n  manuscript:\n    style: british\n    margin: 10mm\n")
			writeYAML(t, filepath.Join(dir, "script-us.yaml"), "folio:\n  margin: 15mm\n  manuscript:\n    margin: 15mm\n")
			cfg, err := Load(Options{Mode: tc.mode, Home: t.TempDir(), LocalDir: dir, Source: tc.source})
			if err != nil {
				t.Fatal(err)
			}
			if got := cfg.String(tc.path, ""); got != "20mm" {
				t.Fatalf("%s = %q, want source 20mm", tc.path, got)
			}
			if got := cfg.String("folio.style", ""); got != "us" {
				t.Fatalf("folio.style = %q, want source us", got)
			}
		})
	}
}

func TestRT078_1CLIOverridesSourceConfiguration(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, filepath.Join(dir, "script.yaml"), "folio:\n  style: us\n  margin: 10mm\n")
	cfg, err := Load(Options{
		Mode: ModeScript, Home: t.TempDir(), LocalDir: dir,
		Source: map[string]any{"folio": map[string]any{"style": "us", "margin": "20mm"}},
		CLI:    map[string]any{"style": "british", "margin": "25mm"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := cfg.String("folio.style", ""); got != "british" {
		t.Fatalf("folio.style = %q, want CLI british", got)
	}
	if got := cfg.String("folio.margin", ""); got != "25mm" {
		t.Fatalf("folio.margin = %q, want CLI 25mm", got)
	}
}
