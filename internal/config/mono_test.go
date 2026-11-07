// ABOUTME: Verifies the shared monospace font migration and compatibility projection.
// ABOUTME: Protects layer precedence, legacy warnings, and all-mode availability.
package config

import (
	"strings"
	"testing"
)

func TestRT068_1SharedMonospaceFontIsAvailableInEveryMode(t *testing.T) {
	for _, mode := range []Mode{ModeScript, ModeLetter, ModeManuscript} {
		cfg, err := Load(Options{Mode: mode, Home: t.TempDir(), LocalDir: t.TempDir(), Source: map[string]any{
			"folio": map[string]any{"mono": map[string]any{"font": map[string]any{"family": "Shared Mono", "size": "13pt"}}},
		}})
		if err != nil {
			t.Fatalf("%s: %v", mode, err)
		}
		font, err := cfg.Font("folio.mono.font")
		if err != nil || font.Family != "Shared Mono" || font.Size != "13pt" {
			t.Fatalf("%s shared font = %+v, %v", mode, font, err)
		}
		if mode == ModeManuscript {
			legacy, legacyErr := cfg.Font("folio.manuscript.mono.font")
			if legacyErr != nil || legacy != font {
				t.Fatalf("manuscript projection = %+v, %v; shared = %+v", legacy, legacyErr, font)
			}
		}
	}
}

func TestRT068_2LegacyMonospaceFontWarnsAndRemainsEffective(t *testing.T) {
	cfg, err := Load(Options{Mode: ModeManuscript, Home: t.TempDir(), LocalDir: t.TempDir(), Source: map[string]any{
		"folio": map[string]any{"manuscript": map[string]any{"mono": map[string]any{"font": map[string]any{"family": "Legacy Mono", "stretch": "125%"}}}},
	}})
	if err != nil {
		t.Fatal(err)
	}
	font, err := cfg.Font("folio.mono.font")
	if err != nil || font.Family != "Legacy Mono" || font.Stretch != "125%" {
		t.Fatalf("shared compatibility font = %+v, %v", font, err)
	}
	warnings := strings.Join(cfg.Warnings(), "\n")
	if !strings.Contains(warnings, "folio.manuscript.mono.font is deprecated") || !strings.Contains(warnings, "folio.mono.font") {
		t.Fatalf("warnings = %q", warnings)
	}
}

func TestRT068_3SharedMonospaceValuesWinWithinLayerAndSourceWinsOverLocal(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	writeYAML(t, project+"/script.yaml", "folio:\n  mono:\n    font:\n      size: 14pt\n      weight: bold\n  manuscript:\n    mono:\n      font:\n        family: Legacy Local\n        size: 8pt\n        style: italic\n")
	cfg, err := Load(Options{Mode: ModeManuscript, Home: home, LocalDir: project, Source: map[string]any{
		"folio": map[string]any{"mono": map[string]any{"font": map[string]any{"family": "Shared Source", "stretch": "80%"}}},
	}})
	if err != nil {
		t.Fatal(err)
	}
	font, err := cfg.Font("folio.mono.font")
	if err != nil {
		t.Fatal(err)
	}
	if font.Family != "Shared Source" || font.Size != "14pt" || font.Weight != "bold" || font.Style != "italic" || font.Stretch != "80%" {
		t.Fatalf("effective shared mono = %+v", font)
	}
}

func TestMonospaceLayerSafelyHandlesNonMappingManuscriptData(t *testing.T) {
	for _, tc := range []struct {
		mode    Mode
		wantErr bool
	}{{ModeScript, true}, {ModeLetter, true}} {
		_, err := Load(Options{Mode: tc.mode, Home: t.TempDir(), LocalDir: t.TempDir(), Source: map[string]any{
			"folio": map[string]any{"manuscript": "not applicable to this mode"},
		}})
		if (err != nil) != tc.wantErr {
			t.Fatalf("%s error = %v; want error %t", tc.mode, err, tc.wantErr)
		}
	}
}
