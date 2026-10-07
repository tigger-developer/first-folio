// ABOUTME: Verifies shared running matter never changes supplementary letters.
// ABOUTME: Compares actual letter output with enabled and disabled shared settings.
package letter

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tigger-developer/first-folio/internal/config"
)

func TestSharedRunningMatterDoesNotChangeLetters(t *testing.T) {
	home := t.TempDir()
	dir := t.TempDir()
	baseline, err := config.Load(config.Options{Mode: config.ModeLetter, Home: home, LocalDir: dir})
	if err != nil {
		t.Fatal(err)
	}
	item := Letter{Subject: "Subject", Body: "Body", Date: "2026-10-07"}
	want, err := RenderTypst(item, baseline)
	if err != nil {
		t.Fatal(err)
	}
	for _, enabled := range []string{"true", "false"} {
		raw := "folio:\n  page-header:\n    enabled: " + enabled + "\n    format: HEADER MUST NOT APPEAR\n    distance-from-edge: 45mm\n  page-footer:\n    enabled: " + enabled + "\n    format: FOOTER MUST NOT APPEAR\n    content-padding-after: 25mm\n"
		if err := os.WriteFile(filepath.Join(dir, "script.yaml"), []byte(raw), 0o644); err != nil {
			t.Fatal(err)
		}
		cfg, err := config.Load(config.Options{Mode: config.ModeLetter, Home: home, LocalDir: dir})
		if err != nil {
			t.Fatal(err)
		}
		got, err := RenderTypst(item, cfg)
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Fatalf("shared running matter enabled=%s changed letter output", enabled)
		}
	}
}
