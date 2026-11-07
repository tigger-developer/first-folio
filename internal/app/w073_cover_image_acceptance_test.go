//go:build folio_acceptance

// ABOUTME: Exercises W073 cover-image behaviour through the built public Folio executable.
// ABOUTME: Verifies rendered geometry, atomic rejection, and visible Superscale fallback.
package app

import (
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAT073_1RenderedPDFGeometryFollowsTheImageContract(t *testing.T) {
	requireW073Tools(t)
	binary := officeBuild(t)
	t.Run("source-relative full-page background", func(t *testing.T) {
		root, home := t.TempDir(), t.TempDir()
		sourceDir, workDir := filepath.Join(root, "source"), filepath.Join(root, "work")
		if err := os.MkdirAll(sourceDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(workDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Join(home, ".config", "first-folio"), 0o755); err != nil {
			t.Fatal(err)
		}
		source, target := filepath.Join(sourceDir, "play.org"), filepath.Join(root, "background.pdf")
		writeAppFile(t, source, "* ACT ONE\n**** CÁIT\nBody marker.\n")
		writeW073PNG(t, filepath.Join(sourceDir, "cover.png"), 300, 500, color.RGBA{R: 230, G: 20, B: 150, A: 255})
		writeAppFile(t, filepath.Join(home, ".config", "first-folio", "script.yaml"), "folio:\n  cover-image:\n    path: cover.png\n    mode: background\n    placement: overlay\n")
		_, diagnostic, err := officeRun(t, binary, workDir, officeEnv(t, home, nil), "convert", source, target)
		if err != nil {
			t.Fatalf("background conversion: %v: %s", err, diagnostic)
		}
		page := rasterW073Page(t, target, 1)
		for _, point := range []imagePointW073{{2, 2}, {page.Bounds().Max.X - 3, 2}, {2, page.Bounds().Max.Y - 3}, {page.Bounds().Max.X - 3, page.Bounds().Max.Y - 3}} {
			if !isW073Magenta(page.At(point.x, point.y)) {
				t.Fatalf("background does not cover corner %v", point)
			}
		}
		if !w073PDFHasPage(t, target, 2) {
			t.Fatal("image-only background did not create a title page")
		}
	})

	t.Run("longest-axis inset", func(t *testing.T) {
		root, home := t.TempDir(), t.TempDir()
		source, target := filepath.Join(root, "play.org"), filepath.Join(root, "inset.pdf")
		writeAppFile(t, source, "* ACT ONE\n**** CÁIT\nBody marker.\n")
		writeW073PNG(t, filepath.Join(root, "wide.png"), 1000, 250, color.RGBA{R: 230, G: 20, B: 150, A: 255})
		writeAppFile(t, filepath.Join(root, "script.yaml"), "folio:\n  cover-image:\n    path: wide.png\n    mode: inset\n    placement: dodge\n    size: 50mm\n")
		_, diagnostic, err := officeRun(t, binary, t.TempDir(), officeEnv(t, home, nil), "convert", source, target)
		if err != nil {
			t.Fatalf("inset conversion: %v: %s", err, diagnostic)
		}
		page := rasterW073Page(t, target, 1)
		box, ok := w073ColourBounds(page)
		if !ok || box.Dx() < 137 || box.Dx() > 147 || box.Dy() < 31 || box.Dy() > 41 {
			t.Fatalf("inset bounds=%v found=%t", box, ok)
		}
	})
}

func TestAT073_2InvalidAndUnsupportedCoverInputFailsAtomically(t *testing.T) {
	requireW073Tools(t)
	binary := officeBuild(t)
	for _, tc := range []struct {
		name, targetExt, extra string
		prepare                func(*testing.T, string)
	}{
		{"unknown key", "pdf", "    typo: true\n", nil},
		{"compact title", "pdf", "  title-page:\n    enabled: false\n", nil},
		{"standalone Typst", "typ", "", nil},
		{"unsupported DOCX", "docx", "", nil},
		{"missing image", "pdf", "", func(t *testing.T, dir string) {
			if err := os.Remove(filepath.Join(dir, "cover.png")); err != nil {
				t.Fatal(err)
			}
		}},
		{"source over byte limit", "pdf", "", func(t *testing.T, dir string) {
			if err := os.Truncate(filepath.Join(dir, "cover.png"), coverSourceByteLimit+1); err != nil {
				t.Fatal(err)
			}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir, home := t.TempDir(), t.TempDir()
			source, target := filepath.Join(dir, "play.org"), filepath.Join(dir, "target."+tc.targetExt)
			writeAppFile(t, source, "#+TITLE: Atomic\n* ACT ONE\n**** CÁIT\nBody.\n")
			writeW073PNG(t, filepath.Join(dir, "cover.png"), 100, 50, color.RGBA{R: 230, G: 20, B: 150, A: 255})
			if tc.prepare != nil {
				tc.prepare(t, dir)
			}
			writeAppFile(t, filepath.Join(dir, "script.yaml"), "folio:\n  cover-image:\n    path: cover.png\n    mode: inset\n    placement: overlay\n    size: 50mm\n"+tc.extra)
			const sentinel = "existing target sentinel"
			writeAppFile(t, target, sentinel)
			_, diagnostic, err := officeRun(t, binary, t.TempDir(), officeEnv(t, home, nil), "convert", source, target)
			if err == nil || (!strings.Contains(diagnostic, "folio.cover-image") && !strings.Contains(diagnostic, "cover")) {
				t.Fatalf("expected cover diagnostic, err=%v diagnostic=%s", err, diagnostic)
			}
			if got := readAppFile(t, target); got != sentinel {
				t.Fatalf("target changed: %q", got)
			}
		})
	}
}

func TestAT073_3PublicSuperscaleFallbackPreservesPDFOutput(t *testing.T) {
	requireW073Tools(t)
	binary := officeBuild(t)
	for _, tc := range []struct {
		name, mode string
		fake       bool
	}{
		{"missing", "", false}, {"nonzero", "nonzero", true}, {"unusable output", "unusable", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir, home, tempRoot := t.TempDir(), t.TempDir(), t.TempDir()
			toolDir := controlledW073Tools(t, tc.fake)
			marker := filepath.Join(dir, "superscale.marker")
			source, target := filepath.Join(dir, "play.org"), filepath.Join(dir, "play.pdf")
			writeAppFile(t, source, "* ACT ONE\n**** CÁIT\nBody.\n")
			writeW073PNG(t, filepath.Join(dir, "cover.png"), 400, 100, color.RGBA{R: 230, G: 20, B: 150, A: 255})
			writeAppFile(t, filepath.Join(dir, "script.yaml"), "folio:\n  cover-image:\n    path: cover.png\n    mode: inset\n    placement: overlay\n    size: 50mm\n    superscale: default\n")
			env := officeEnv(t, home, map[string]string{"PATH": toolDir, "TMPDIR": tempRoot, "FOLIO_TEST_SUPERSCALE_MODE": tc.mode, "FOLIO_TEST_SUPERSCALE_MARKER": marker})
			_, diagnostic, err := officeRun(t, binary, t.TempDir(), env, "convert", source, target)
			if err != nil {
				t.Fatalf("fallback failed: %v: %s", err, diagnostic)
			}
			if strings.Count(diagnostic, "folio.cover-image.superscale") != 1 {
				t.Fatalf("warning=%q", diagnostic)
			}
			page := rasterW073Page(t, target, 1)
			box, ok := w073ColourBounds(page)
			if !ok || box.Dx() < 137 || box.Dx() > 147 {
				t.Fatalf("fallback geometry=%v found=%t", box, ok)
			}
			assertNoW073TemporaryOutput(t, tempRoot)
		})
	}
}

type imagePointW073 struct{ x, y int }

func isW073Magenta(value color.Color) bool {
	r, g, b, _ := value.RGBA()
	return int(r>>8) >= w073MagentaThreshold && int(g>>8) <= 80 && int(b>>8) >= 100
}
