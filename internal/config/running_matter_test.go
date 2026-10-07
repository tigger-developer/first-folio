// ABOUTME: Verifies shared running matter and manuscript compatibility layering.
// ABOUTME: Protects public paths, safe validation, and scoped migration diagnostics.
package config

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestSharedRunningMatterPresetDefaults(t *testing.T) {
	for _, mode := range []Mode{ModeScript, ModeManuscript, ModeLetter} {
		for _, style := range []string{"british", "us"} {
			cfg, err := Load(Options{Mode: mode, Home: t.TempDir(), LocalDir: t.TempDir(), CLI: map[string]any{"style": style}})
			if err != nil {
				t.Fatal(err)
			}
			assertValue(t, cfg, "folio.page-header.enabled", false)
			assertValue(t, cfg, "folio.page-header.format", "[title] • [chapter] • [author]")
			assertValue(t, cfg, "folio.page-footer.enabled", true)
			assertValue(t, cfg, "folio.page-footer.format", "[page]")
			for _, role := range []string{"page-header", "page-footer"} {
				assertValue(t, cfg, "folio."+role+".distance-from-edge", "20mm")
				assertValue(t, cfg, "folio."+role+".content-padding-after", "10mm")
				font, err := cfg.Font("folio." + role + ".font")
				if err != nil {
					t.Fatal(err)
				}
				wantFamily := "Libertinus Sans"
				if style == "us" {
					wantFamily = "Menlo"
				}
				if font.Family != wantFamily || font.Size != "10pt" {
					t.Fatalf("%s %s font = %+v", style, role, font)
				}
			}
		}
	}
}

func TestLaterSharedRunningMatterBeatsEarlierLegacyFields(t *testing.T) {
	home, project := t.TempDir(), t.TempDir()
	writeYAML(t, filepath.Join(home, ".config", "first-folio", "script.yaml"), "folio:\n  manuscript:\n    page-header:\n      format: Global legacy\n      font:\n        size: 16pt\n")
	writeYAML(t, filepath.Join(project, "script.yaml"), "folio:\n  page-header:\n    format: Local shared\n    font:\n      size: 11pt\n")
	cfg, err := Load(Options{Mode: ModeManuscript, Home: home, LocalDir: project})
	if err != nil {
		t.Fatal(err)
	}
	assertValue(t, cfg, "folio.manuscript.page-header.format", "Local shared")
	assertValue(t, cfg, "folio.manuscript.page-header.font.size", "11pt")
}

func TestSameLayerLegacyRunningMatterOverridesOnlyItsFields(t *testing.T) {
	source, err := parseYAML("source", []byte(`folio:
  page-header:
    enabled: true
    format: Shared [unknown-token]
    font:
      family: Shared Family
      size: 13pt
  manuscript:
    page-header:
      align: right
      font:
        size: 9pt
`))
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(Options{Mode: ModeManuscript, Home: t.TempDir(), LocalDir: t.TempDir(), Source: source})
	if err != nil {
		t.Fatal(err)
	}
	assertValue(t, cfg, "folio.manuscript.page-header.enabled", true)
	assertValue(t, cfg, "folio.manuscript.page-header.format", "Shared [unknown-token]")
	assertValue(t, cfg, "folio.manuscript.page-header.font.family", "Shared Family")
	assertValue(t, cfg, "folio.manuscript.page-header.font.size", "9pt")
	assertValue(t, cfg, "folio.page-header.font.size", "13pt")
	warnings, ok := any(cfg).(interface{ Warnings() []string })
	if !ok {
		t.Fatal("Config must expose Warnings()")
	}
	if got := warnings.Warnings(); len(got) != 1 || !strings.Contains(got[0], "folio.manuscript.page-header") || !strings.Contains(got[0], "folio.page-header") {
		t.Fatalf("warnings = %v", got)
	}
}

func TestLegacyRunningMatterWarningsAreManuscriptUserOnly(t *testing.T) {
	for _, mode := range []Mode{ModeManuscript, ModeScript, ModeLetter} {
		for _, legacy := range []bool{false, true} {
			source := map[string]any{}
			if legacy {
				setPath(source, "folio.manuscript.page-footer.format", "Legacy")
			}
			cfg, err := Load(Options{Mode: mode, Home: t.TempDir(), LocalDir: t.TempDir(), Source: source})
			if err != nil {
				t.Fatal(err)
			}
			collector, ok := any(cfg).(interface{ Warnings() []string })
			if !ok {
				t.Fatal("Config must expose Warnings()")
			}
			want := 0
			if legacy && mode == ModeManuscript {
				want = 1
			}
			if got := collector.Warnings(); len(got) != want {
				t.Fatalf("%s legacy=%v warnings=%v", mode, legacy, got)
			}
			if mode != ModeManuscript {
				assertValue(t, cfg, "folio.page-footer.format", "[page]")
			}
		}
	}
}

func TestSharedRunningMatterRejectsMalformedFields(t *testing.T) {
	for _, test := range []struct {
		suffix string
		value  any
	}{
		{"", "not a mapping"}, {"", nil}, {"enabled", "yes"}, {"enabled", 1},
		{"align", "diagonal"}, {"font.size", "0pt"}, {"distance-from-edge", "1pt; panic()"},
		{"distance-from-edge", "-1mm"}, {"content-padding-after", 1}, {"format", []any{"bad"}},
	} {
		for _, role := range []string{"page-header", "page-footer"} {
			path := "folio." + role
			if test.suffix != "" {
				path += "." + test.suffix
			}
			source := map[string]any{}
			setPath(source, path, test.value)
			_, err := Load(Options{Mode: ModeScript, Home: t.TempDir(), LocalDir: t.TempDir(), Source: source})
			if err == nil || !strings.Contains(err.Error(), path) {
				t.Errorf("%s=%#v error=%v", path, test.value, err)
			}
		}
	}
}

func TestRunningMatterPreservesUnknownKeysAndNullableFrontmatter(t *testing.T) {
	source := map[string]any{}
	setPath(source, "folio.page-header.extension", map[string]any{"future": true})
	setPath(source, "folio.page-header.frontmatter-format", nil)
	setPath(source, "folio.manuscript.page-footer.alt-frontmatter-format", nil)
	cfg, err := Load(Options{Mode: ModeManuscript, Home: t.TempDir(), LocalDir: t.TempDir(), Source: source})
	if err != nil {
		t.Fatal(err)
	}
	assertValue(t, cfg, "folio.page-header.extension.future", true)
	assertValue(t, cfg, "folio.manuscript.page-header.frontmatter-format", nil)
}

func TestRunningMatterWarningsDeduplicateAcrossLayers(t *testing.T) {
	home, project := t.TempDir(), t.TempDir()
	legacy := "folio:\n  manuscript:\n    page-header:\n      enabled: true\n"
	writeYAML(t, filepath.Join(home, ".config", "first-folio", "script.yaml"), legacy)
	writeYAML(t, filepath.Join(project, "script.yaml"), legacy)
	cfg, err := Load(Options{Mode: ModeManuscript, Home: home, LocalDir: project})
	if err != nil {
		t.Fatal(err)
	}
	if got := cfg.Warnings(); len(got) != 1 {
		t.Fatalf("warnings = %v", got)
	}
}

func TestRunningMatterNormalizesEveryPrecedenceLayer(t *testing.T) {
	home, project := t.TempDir(), t.TempDir()
	globalDir := filepath.Join(home, ".config", "first-folio")
	writeYAML(t, filepath.Join(globalDir, "script.yaml"), "folio:\n  manuscript:\n    page-footer:\n      format: Global legacy\n      font:\n        family: Global Family\n")
	writeYAML(t, filepath.Join(globalDir, "script-us.yaml"), "folio:\n  page-footer:\n    format: Global style shared\n")
	source := map[string]any{}
	setPath(source, "folio.manuscript.page-footer.format", "Source legacy")
	writeYAML(t, filepath.Join(project, "script.yaml"), "folio:\n  page-footer:\n    format: Local shared\n    font:\n      size: 11pt\n")
	writeYAML(t, filepath.Join(project, "script-us.yaml"), "folio:\n  manuscript:\n    page-footer:\n      format: Local style legacy\n")
	cfg, err := Load(Options{Mode: ModeManuscript, Home: home, LocalDir: project, Source: source, CLI: map[string]any{
		"style": "us", "page-footer.format": "CLI shared", "manuscript.page-footer.align": "right", "arbitrary-new-key": true,
	}})
	if err != nil {
		t.Fatal(err)
	}
	assertValue(t, cfg, "folio.manuscript.page-footer.format", "CLI shared")
	assertValue(t, cfg, "folio.manuscript.page-footer.align", "right")
	assertValue(t, cfg, "folio.manuscript.page-footer.font.family", "Global Family")
	assertValue(t, cfg, "folio.manuscript.page-footer.font.size", "11pt")
	assertValue(t, Config{data: source}, "folio.manuscript.page-footer.format", "Source legacy")
	if _, ok := cfg.Get("folio.arbitrary-new-key"); ok {
		t.Fatal("CLI created an unsupported key")
	}
}

func TestMalformedLegacyRunningMatterIsManuscriptOnly(t *testing.T) {
	for _, mode := range []Mode{ModeManuscript, ModeScript, ModeLetter} {
		source := map[string]any{}
		setPath(source, "folio.manuscript.page-header", []any{"invalid"})
		_, err := Load(Options{Mode: mode, Home: t.TempDir(), LocalDir: t.TempDir(), Source: source})
		if mode == ModeManuscript {
			if err == nil || !strings.Contains(err.Error(), "folio.manuscript.page-header") {
				t.Fatalf("error=%v", err)
			}
		} else if err != nil {
			t.Fatalf("%s legacy block should not affect other modes: %v", mode, err)
		}
	}
}

func TestRunningMatterRejectsOverflowLengths(t *testing.T) {
	source := map[string]any{}
	setPath(source, "folio.page-header.distance-from-edge", strings.Repeat("9", 400)+"mm")
	_, err := Load(Options{Mode: ModeScript, Home: t.TempDir(), LocalDir: t.TempDir(), Source: source})
	if err == nil || !strings.Contains(err.Error(), "folio.page-header.distance-from-edge") {
		t.Fatalf("error=%v", err)
	}
}
