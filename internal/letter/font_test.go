// ABOUTME: W039 regression for the letter renderer's uniform font block.
// ABOUTME: Verifies all six properties reach the role-scoped Typst text setting.
package letter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tigger-developer/first-folio/internal/config"
)

func TestRT039_1EveryLetterFontPropertyReachesOnlyItsRole(t *testing.T) {
	properties := []struct {
		name, value, marker string
	}{
		{"family", "W039 Family", `"W039 Family"`},
		{"size", "13.5pt", "13.5pt"},
		{"weight", "643", "643"},
		{"stretch", "117.5%", "117.5%"},
		{"style", "oblique", `"oblique"`},
		{"letter-spacing", "-0.03em", "-0.03em"},
	}
	for _, property := range properties {
		t.Run(property.name, func(t *testing.T) {
			home := t.TempDir()
			dir := t.TempDir()
			raw := fmt.Sprintf("folio:\n  letter:\n    font:\n      %s: %s\n", property.name, property.value)
			if err := os.WriteFile(filepath.Join(dir, "script.yaml"), []byte(raw), 0o644); err != nil {
				t.Fatal(err)
			}
			cfg, err := config.Load(config.Options{Mode: config.ModeLetter, Home: home, LocalDir: dir})
			if err != nil {
				t.Fatal(err)
			}
			typst, err := RenderTypst(Letter{Subject: "Subject", Body: "Body"}, cfg)
			if err != nil {
				t.Fatal(err)
			}
			if count := strings.Count(typst, property.marker); count != 1 {
				t.Fatalf("letter marker %q occurs %d times, want exactly once", property.marker, count)
			}
		})
	}
}

func TestRT039_1LetterFontBlockEmitsAllProperties(t *testing.T) {
	home := t.TempDir()
	dir := t.TempDir()
	raw := `folio:
  letter:
    font:
      family: W039 Letter
      size: 13.5pt
      weight: 650
      stretch: 117.5%
      style: oblique
      letter-spacing: -0.03em
`
	if err := os.WriteFile(filepath.Join(dir, "script.yaml"), []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(config.Options{Mode: config.ModeLetter, Home: home, LocalDir: dir})
	if err != nil {
		t.Fatal(err)
	}
	typst, err := RenderTypst(Letter{Subject: "Subject", Body: "Body"}, cfg)
	if err != nil {
		t.Fatal(err)
	}
	want := `#set text(font: "W039 Letter", size: 13.5pt, weight: 650, stretch: 117.5%, style: "oblique", tracking: -0.03em)`
	if !strings.Contains(typst, want) {
		t.Fatalf("missing %q in:\n%s", want, typst)
	}
}
