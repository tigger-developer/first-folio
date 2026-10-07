// ABOUTME: Tests strict script title-page enablement and mode isolation.
// ABOUTME: Protects the dedicated-page default and rejects malformed switches.
package config

import (
	"fmt"
	"strings"
	"testing"
)

func TestScriptTitlePageEnabledValidation(t *testing.T) {
	for _, field := range []string{"enabled", "skip-header", "skip-footer"} {
		for _, value := range []any{nil, "false", "true", "no", 0, []any{false}, map[string]any{"enabled": false}} {
			t.Run(fmt.Sprintf("%s/%#v", field, value), func(t *testing.T) {
				_, err := Load(Options{Mode: ModeScript, Home: t.TempDir(), LocalDir: t.TempDir(), Source: map[string]any{"folio": map[string]any{"title-page": map[string]any{field: value}}}})
				if err == nil || !strings.Contains(err.Error(), "folio.title-page."+field+" must be a boolean") {
					t.Fatalf("expected precise boolean validation error, got %v", err)
				}
			})
		}
	}
}

func TestScriptTitlePageDefaultAndModeIsolation(t *testing.T) {
	for _, style := range []string{"british", "us", "screenplay"} {
		t.Run(style, func(t *testing.T) {
			cfg, err := Load(Options{Mode: ModeScript, Home: t.TempDir(), LocalDir: t.TempDir(), CLI: map[string]any{"style": style}})
			if err != nil {
				t.Fatal(err)
			}
			assertValue(t, cfg, "folio.title-page.enabled", true)
			assertValue(t, cfg, "folio.title-page.skip-header", true)
			assertValue(t, cfg, "folio.title-page.skip-footer", true)
		})
	}
	for _, mode := range []Mode{ModeScript, ModeManuscript, ModeLetter} {
		for _, field := range []string{"enabled", "skip-header", "skip-footer"} {
			for _, value := range []any{false, true} {
				cfg, err := Load(Options{Mode: mode, Home: t.TempDir(), LocalDir: t.TempDir(), Source: map[string]any{"folio": map[string]any{"title-page": map[string]any{field: value}}}})
				if err != nil {
					t.Fatal(err)
				}
				assertValue(t, cfg, "folio.title-page."+field, value)
				assertValue(t, cfg, "folio.manuscript.title-page.enabled", true)
			}
			if mode != ModeScript {
				_, err := Load(Options{Mode: mode, Home: t.TempDir(), LocalDir: t.TempDir(), Source: map[string]any{"folio": map[string]any{"title-page": map[string]any{field: "irrelevant"}}}})
				if err != nil {
					t.Fatalf("root script switch must not validate in %s: %v", mode, err)
				}
			}
		}
	}
}
