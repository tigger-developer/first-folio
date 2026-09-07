// ABOUTME: Characterizes shared YAML layering and inherited configuration access.
// ABOUTME: Covers global, local, style-specific, CLI, malformed, and companion keys.
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadScriptPrecedence(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	writeYAML(t, filepath.Join(home, ".config", "first-folio", "script.yaml"), `folio:
  font:
    family: Global Font
  page: a5
  margin: 30mm
yapper:
  character-voices:
    CÁIT: voice-id
`)
	writeYAML(t, filepath.Join(home, ".config", "first-folio", "script-us.yaml"), "folio:\n  margin: 22mm\n")
	writeYAML(t, filepath.Join(project, "script.yaml"), "folio:\n  style: us\n  font:\n    family: Local Font\n")
	writeYAML(t, filepath.Join(project, "script-us.yaml"), "folio:\n  page: a4\n")

	cfg, err := Load(Options{
		Mode:     ModeScript,
		Home:     home,
		LocalDir: project,
		CLI:      map[string]any{"font": "CLI Font"},
	})
	if err != nil {
		t.Fatal(err)
	}
	assertValue(t, cfg, "folio.font.family", "CLI Font")
	assertValue(t, cfg, "folio.page", "a4")
	assertValue(t, cfg, "folio.margin", "22mm")
	assertValue(t, cfg, "folio.style", "us")
	assertValue(t, cfg, "yapper.character-voices.CÁIT", "voice-id")
}

func TestInheritedValue(t *testing.T) {
	base := map[string]any{
		"folio": map[string]any{
			"font": map[string]any{"family": "Root Font", "size": "12pt"},
			"positioning": map[string]any{
				"speech": map[string]any{
					"speaker": map[string]any{"font": map[string]any{"family": "Speaker Font"}},
				},
			},
		},
	}
	overlay := map[string]any{"folio": map[string]any{"font": map[string]any{"size": "11pt"}}}
	deepMerge(base, overlay)
	cfg := Config{data: base}
	assertValue(t, cfg, "folio.font.family", "Root Font")
	assertValue(t, cfg, "folio.font.size", "11pt")
	assertValue(t, cfg, "folio.positioning.speech.speaker.font.family", "Speaker Font")
	if _, ok := cfg.Get("folio.positioning.speech.speaker.font.size"); ok {
		t.Fatal("speaker font must not inherit size from folio.font")
	}
}

func TestLoadScreenplayPresetAndStyleSpecificConfig(t *testing.T) {
	project := t.TempDir()
	writeYAML(t, filepath.Join(project, "script.yaml"), "folio:\n  style: screenplay\n")
	writeYAML(t, filepath.Join(project, "script-screenplay.yaml"), "folio:\n  margin: 18mm\n")

	cfg, err := Load(Options{Mode: ModeScript, Home: t.TempDir(), LocalDir: project})
	if err != nil {
		t.Fatal(err)
	}
	assertValue(t, cfg, "folio.style", "screenplay")
	assertValue(t, cfg, "folio.font.family", "Courier Prime")
	assertValue(t, cfg, "folio.title-page.title.font.size", "12pt")
	assertValue(t, cfg, "folio.positioning.speech.speaker.align", "center")
	assertValue(t, cfg, "folio.positioning.stage-direction.font.style", "regular")
	assertValue(t, cfg, "folio.margin", "18mm")
}

func TestMalformedYAMLNamesFile(t *testing.T) {
	project := t.TempDir()
	path := filepath.Join(project, "script.yaml")
	writeYAML(t, path, "folio:\n  font: [unterminated\n")
	_, err := Load(Options{Mode: ModeScript, Home: t.TempDir(), LocalDir: project})
	if err == nil {
		t.Fatal("malformed YAML should fail")
	}
}

func writeYAML(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func assertValue(t *testing.T, cfg Config, path string, want any) {
	t.Helper()
	got, ok := cfg.Get(path)
	if !ok || got != want {
		t.Errorf("%s = %#v, %t; want %#v", path, got, ok, want)
	}
}
