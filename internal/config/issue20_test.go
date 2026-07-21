// ABOUTME: Regression tests for issue #20 -- script.yaml walk-up lookup.
// ABOUTME: When LocalDir has no script.yaml, walk upward until one is found or
// ABOUTME: the home boundary is reached.
package config

import (
	"path/filepath"
	"testing"
)

// RT-20.1: script.yaml at project root is discovered when input is in a subdirectory.
func TestRT_20_1_WalkUpFindsScriptInParent(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	subdir := filepath.Join(project, "part0")
	writeYAML(t, filepath.Join(project, "script.yaml"), "folio:\n  font: Root Font\n")
	// Create the subdirectory so LocalDir is a real path.
	writeYAML(t, filepath.Join(subdir, "placeholder.md"), "ignored")

	cfg, err := Load(Options{Mode: ModeManuscript, Home: home, LocalDir: subdir})
	if err != nil {
		t.Fatal(err)
	}
	assertValue(t, cfg, "folio.font", "Root Font")
}

// RT-20.2: script.yaml directly in LocalDir wins over any parent script.yaml.
func TestRT_20_2_NearerScriptWinsOverAncestor(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	subdir := filepath.Join(project, "part1")
	writeYAML(t, filepath.Join(project, "script.yaml"), "folio:\n  font: Root Font\n")
	writeYAML(t, filepath.Join(subdir, "script.yaml"), "folio:\n  font: Subdir Font\n")

	cfg, err := Load(Options{Mode: ModeManuscript, Home: home, LocalDir: subdir})
	if err != nil {
		t.Fatal(err)
	}
	// Nearest wins: subdir/script.yaml overrides project/script.yaml.
	assertValue(t, cfg, "folio.font", "Subdir Font")
}

// RT-20.3: with no script.yaml anywhere on the walk path, Load still succeeds
// and uses embedded preset defaults (backwards-compatible with the no-local case).
func TestRT_20_3_NoScriptOnWalkPathSucceeds(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	subdir := filepath.Join(project, "empty-sub")
	writeYAML(t, filepath.Join(subdir, "placeholder.md"), "ignored")

	cfg, err := Load(Options{Mode: ModeManuscript, Home: home, LocalDir: subdir})
	if err != nil {
		t.Fatalf("expected success with no local script.yaml on walk path, got %v", err)
	}
	// Preset font should be present (Libertinus Serif from british-manuscript.yaml).
	assertValue(t, cfg, "folio.font", "Libertinus Serif")
}

// RT-20.4: the walk stops at the home directory boundary; a script.yaml above
// home is NOT discovered. Safety guard: we don't accidentally read from other
// projects, /etc, or another user's home.
func TestRT_20_4_WalkStopsAtHomeBoundary(t *testing.T) {
	home := t.TempDir()
	// Place a script.yaml *above* the home directory. Walk from a subdir of
	// home should NOT reach it.
	parentOfHome := filepath.Dir(home)
	writeYAML(t, filepath.Join(parentOfHome, "script.yaml"), "folio:\n  font: Should Not Be Loaded\n")
	subdir := filepath.Join(home, "project", "part0")
	writeYAML(t, filepath.Join(subdir, "placeholder.md"), "ignored")

	cfg, err := Load(Options{Mode: ModeManuscript, Home: home, LocalDir: subdir})
	if err != nil {
		t.Fatal(err)
	}
	// Preset default, NOT "Should Not Be Loaded".
	if font, _ := cfg.Get("folio.font"); font == "Should Not Be Loaded" {
		t.Fatalf("walk crossed home boundary and read forbidden script.yaml")
	}
}

// RT-20.5: style-suffixed sibling override is found next to the walk-discovered
// script.yaml, not next to LocalDir itself.
func TestRT_20_5_StyleSuffixedSiblingFoundAtWalkTarget(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	subdir := filepath.Join(project, "part0")
	writeYAML(t, filepath.Join(project, "script.yaml"), "folio:\n  style: us\n  font: Root Font\n")
	writeYAML(t, filepath.Join(project, "script-us.yaml"), "folio:\n  font: US Override Font\n")
	writeYAML(t, filepath.Join(subdir, "placeholder.md"), "ignored")

	cfg, err := Load(Options{Mode: ModeManuscript, Home: home, LocalDir: subdir})
	if err != nil {
		t.Fatal(err)
	}
	assertValue(t, cfg, "folio.font", "US Override Font")
}
