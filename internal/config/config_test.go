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
  font: Global Font
  page: a5
  margin: 30mm
character-voices:
  CÁIT: voice-id
`)
	writeYAML(t, filepath.Join(home, ".config", "first-folio", "script-us.yaml"), "folio:\n  margin: 22mm\n")
	writeYAML(t, filepath.Join(project, "script.yaml"), "folio:\n  style: us\n  font: Local Font\n")
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
	assertValue(t, cfg, "folio.font", "CLI Font")
	assertValue(t, cfg, "folio.page", "a4")
	assertValue(t, cfg, "folio.margin", "22mm")
	assertValue(t, cfg, "folio.style", "us")
	assertValue(t, cfg, "character-voices.CÁIT", "voice-id")
}

func TestInheritedValue(t *testing.T) {
	cfg := Config{data: map[string]any{
		"folio": map[string]any{
			"font": "Root Font",
			"positioning": map[string]any{
				"speech": map[string]any{
					"font":    "Speech Font",
					"speaker": map[string]any{"bold": false},
				},
			},
		},
	}}
	if got := cfg.InheritedString("folio.positioning.speech.speaker", "font", ""); got != "Speech Font" {
		t.Fatalf("inherited font = %q", got)
	}
	if cfg.Bool("folio.positioning.speech.speaker.bold", true) {
		t.Fatal("explicit false should override true fallback")
	}
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
