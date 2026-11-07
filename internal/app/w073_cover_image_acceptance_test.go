//go:build folio_acceptance

// ABOUTME: Exercises W073 cover-image behaviour through the built public Folio executable.
// ABOUTME: Verifies rendered geometry, atomic rejection, and visible Superscale fallback.
package app

import (
	"encoding/binary"
	"hash/crc32"
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
		writeW073CropMarkerPNG(t, filepath.Join(sourceDir, "cover.png"), 300, 500)
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
		if !strings.Contains(w073PDFText(t, target), "Body marker") {
			t.Fatal("body page content changed after cover insertion")
		}
	})

	for _, tc := range []struct {
		name, file            string
		width, height         int
		size                  string
		wantWidth, wantHeight int
	}{
		{"landscape longest-axis inset", "wide.png", 1000, 250, "50mm", 142, 35},
		{"portrait longest-axis inset", "portrait.png", 200, 400, "80mm", 113, 227},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root, home := t.TempDir(), t.TempDir()
			source, target := filepath.Join(root, "play.org"), filepath.Join(root, "inset.pdf")
			writeAppFile(t, source, "* ACT ONE\n**** CÁIT\nBody marker.\n")
			writeW073PNG(t, filepath.Join(root, tc.file), tc.width, tc.height, color.RGBA{R: 230, G: 20, B: 150, A: 255})
			writeAppFile(t, filepath.Join(root, "script.yaml"), "folio:\n  cover-image:\n    path: "+tc.file+"\n    mode: inset\n    placement: dodge\n    size: "+tc.size+"\n")
			_, diagnostic, err := officeRun(t, binary, t.TempDir(), officeEnv(t, home, nil), "convert", source, target)
			if err != nil {
				t.Fatalf("inset conversion: %v: %s", err, diagnostic)
			}
			page := rasterW073Page(t, target, 1)
			box, ok := w073ColourBounds(page)
			if !ok || absW073(box.Dx()-tc.wantWidth) > 2 || absW073(box.Dy()-tc.wantHeight) > 2 {
				t.Fatalf("inset bounds=%v found=%t want approximately %dx%d pixels", box, ok, tc.wantWidth, tc.wantHeight)
			}
		})
	}

	t.Run("overlay treatment remains localized and text remains live", func(t *testing.T) {
		root, home := t.TempDir(), t.TempDir()
		source, target := filepath.Join(root, "play.org"), filepath.Join(root, "overlay.pdf")
		writeAppFile(t, source, "#+TITLE: OVERLAYMARKER\n#+SUBTITLE: SUBMARKER\n#+AUTHOR: AUTHORMARKER\n#+DATE: DATEMARKER\n#+VERSION: VERSIONMARKER\n* ACT ONE\n**** CÁIT\nBODYMARKER.\n")
		writeW073PNG(t, filepath.Join(root, "cover.png"), 600, 900, color.RGBA{R: 230, G: 20, B: 150, A: 255})
		writeAppFile(t, filepath.Join(root, "script.yaml"), "folio:\n  cover-image:\n    path: cover.png\n    mode: background\n    placement: overlay\n    image:\n      treatment: dark\n      opacity: 70%\n    overlay:\n      blur: 2mm\n      padding: 5mm\n      wash:\n        treatment: light\n        opacity: 80%\n")
		_, diagnostic, err := officeRun(t, binary, t.TempDir(), officeEnv(t, home, nil), "convert", source, target)
		if err != nil {
			t.Fatalf("overlay conversion: %v: %s", err, diagnostic)
		}
		text := w073PDFText(t, target)
		for _, marker := range []string{"OVERLAYMARKER", "SUBMARKER", "AUTHORMARKER", "DATEMARKER", "VERSIONMARKER", "BODYMARKER"} {
			if !strings.Contains(text, marker) {
				t.Fatalf("live PDF text lacks %s: %q", marker, text)
			}
		}
		page := rasterW073Page(t, target, 1)
		white, pixels := 0, page.Bounds().Dx()*page.Bounds().Dy()
		for y := page.Bounds().Min.Y; y < page.Bounds().Max.Y; y++ {
			for x := page.Bounds().Min.X; x < page.Bounds().Max.X; x++ {
				r, g, b, _ := page.At(x, y).RGBA()
				if r>>8 > 220 && g>>8 > 180 && b>>8 > 200 {
					white++
				}
			}
		}
		if white < 500 || white > pixels/3 {
			t.Fatalf("localized light wash pixels=%d of %d", white, pixels)
		}
	})

	for _, position := range []string{"center", "third"} {
		t.Run("dodge "+position+" positions groups around image", func(t *testing.T) {
			root, home := t.TempDir(), t.TempDir()
			source, target := filepath.Join(root, "play.org"), filepath.Join(root, "dodge.pdf")
			writeAppFile(t, source, "#+TITLE: UPPERMARKER\n#+AUTHOR: LOWERMARKER\n* ACT ONE\n**** CÁIT\nBody.\n")
			writeW073PNG(t, filepath.Join(root, "cover.png"), 400, 200, color.RGBA{R: 230, G: 20, B: 150, A: 255})
			writeAppFile(t, filepath.Join(root, "script.yaml"), "folio:\n  title-page:\n    title:\n      position: "+position+"\n  cover-image:\n    path: cover.png\n    mode: inset\n    placement: dodge\n    size: 100mm\n")
			_, diagnostic, err := officeRun(t, binary, t.TempDir(), officeEnv(t, home, nil), "convert", source, target)
			if err != nil {
				t.Fatalf("dodge conversion: %v: %s", err, diagnostic)
			}
			page := rasterW073Page(t, target, 1)
			imageBox, ok := w073ColourBounds(page)
			if !ok {
				t.Fatal("dodge image marker absent")
			}
			upper := w073WordBounds(t, target, "UPPERMARKER")
			lower := w073WordBounds(t, target, "LOWERMARKER")
			margin := 25 * 72 / 25.4
			factor := 0.5
			if position == "third" {
				factor = 1.0 / 3.0
			}
			wantUpper := margin + (float64(imageBox.Min.Y)-margin)*factor
			wantLower := float64(imageBox.Max.Y) + (float64(page.Bounds().Max.Y)-margin-float64(imageBox.Max.Y))*factor
			if absoluteFloatW073(upper.centerY()-wantUpper) > 18 || absoluteFloatW073(lower.centerY()-wantLower) > 18 {
				t.Fatalf("position=%s image=%v upper=%#v wantY=%.1f lower=%#v wantY=%.1f", position, imageBox, upper, wantUpper, lower, wantLower)
			}
			if upper.yMax >= float64(imageBox.Min.Y) || lower.yMin <= float64(imageBox.Max.Y) {
				t.Fatalf("dodge overlap image=%v upper=%#v lower=%#v", imageBox, upper, lower)
			}
		})
	}
}

func TestAT073_2InvalidAndUnsupportedCoverInputFailsAtomically(t *testing.T) {
	requireW073Tools(t)
	binary := officeBuild(t)
	for _, tc := range []struct {
		name, targetExt, extra, completeConfig string
		prepare                                func(*testing.T, string)
	}{
		{"unknown key", "pdf", "    typo: true\n", "", nil},
		{"compact title", "pdf", "  title-page:\n    enabled: false\n", "", nil},
		{"standalone Typst", "typ", "", "", nil},
		{"unsupported DOCX", "docx", "", "", nil},
		{"unsupported ODT", "odt", "", "", nil},
		{"unsupported image type", "pdf", "", "", func(t *testing.T, dir string) {
			writeAppFile(t, filepath.Join(dir, "cover.png"), "not an image")
		}},
		{"missing image", "pdf", "", "", func(t *testing.T, dir string) {
			if err := os.Remove(filepath.Join(dir, "cover.png")); err != nil {
				t.Fatal(err)
			}
		}},
		{"source over byte limit", "pdf", "", "", func(t *testing.T, dir string) {
			if err := os.Truncate(filepath.Join(dir, "cover.png"), coverSourceByteLimit+1); err != nil {
				t.Fatal(err)
			}
		}},
		{"source over axis limit", "pdf", "", "", func(t *testing.T, dir string) {
			writeW073PNGHeader(t, filepath.Join(dir, "cover.png"), coverSourceAxisLimit+1, 1)
		}},
		{"source over pixel limit", "pdf", "", "", func(t *testing.T, dir string) {
			writeW073PNGHeader(t, filepath.Join(dir, "cover.png"), 10_001, 10_000)
		}},
		{"invalid background dodge combination", "pdf", "", "folio:\n  cover-image:\n    path: cover.png\n    mode: background\n    placement: dodge\n", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir, home := t.TempDir(), t.TempDir()
			source, target := filepath.Join(dir, "play.org"), filepath.Join(dir, "target."+tc.targetExt)
			writeAppFile(t, source, "#+TITLE: Atomic\n* ACT ONE\n**** CÁIT\nBody.\n")
			writeW073PNG(t, filepath.Join(dir, "cover.png"), 100, 50, color.RGBA{R: 230, G: 20, B: 150, A: 255})
			if tc.prepare != nil {
				tc.prepare(t, dir)
			}
			configuration := tc.completeConfig
			if configuration == "" {
				configuration = "folio:\n  cover-image:\n    path: cover.png\n    mode: inset\n    placement: overlay\n    size: 50mm\n" + tc.extra
			}
			writeAppFile(t, filepath.Join(dir, "script.yaml"), configuration)
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

	t.Run("invalid cover leaves an absent target absent", func(t *testing.T) {
		dir, home := t.TempDir(), t.TempDir()
		source, target := filepath.Join(dir, "play.org"), filepath.Join(dir, "absent.pdf")
		writeAppFile(t, source, "#+TITLE: Atomic\n* ACT ONE\n**** CÁIT\nBody.\n")
		writeAppFile(t, filepath.Join(dir, "script.yaml"), "folio:\n  cover-image:\n    path: missing.png\n    mode: background\n")
		_, diagnostic, err := officeRun(t, binary, t.TempDir(), officeEnv(t, home, nil), "convert", source, target)
		if err == nil || !strings.Contains(diagnostic, "cover") {
			t.Fatalf("expected cover failure, err=%v diagnostic=%s", err, diagnostic)
		}
		if _, err := os.Stat(target); !os.IsNotExist(err) {
			t.Fatalf("absent target was published: %v", err)
		}
	})

	t.Run("dodge overflow names the measured group and preserves target", func(t *testing.T) {
		dir, home := t.TempDir(), t.TempDir()
		source, target := filepath.Join(dir, "play.org"), filepath.Join(dir, "overflow.pdf")
		writeAppFile(t, source, "#+TITLE: "+strings.Repeat("Overflow title ", 120)+"\n* ACT ONE\n**** CÁIT\nBody.\n")
		writeW073PNG(t, filepath.Join(dir, "cover.png"), 100, 100, color.RGBA{R: 230, G: 20, B: 150, A: 255})
		writeAppFile(t, filepath.Join(dir, "script.yaml"), "folio:\n  cover-image:\n    path: cover.png\n    mode: inset\n    placement: dodge\n    size: 190mm\n")
		const sentinel = "existing target sentinel"
		writeAppFile(t, target, sentinel)
		_, diagnostic, err := officeRun(t, binary, t.TempDir(), officeEnv(t, home, nil), "convert", source, target)
		if err == nil || !strings.Contains(diagnostic, "upper") {
			t.Fatalf("expected upper-group overflow, err=%v diagnostic=%s", err, diagnostic)
		}
		if got := readAppFile(t, target); got != sentinel {
			t.Fatalf("target changed: %q", got)
		}
	})

	for _, command := range []string{"manuscript"} {
		t.Run("unsupported "+command+" consumer", func(t *testing.T) {
			dir, home := t.TempDir(), t.TempDir()
			source, target := filepath.Join(dir, "source.md"), filepath.Join(dir, command+".pdf")
			writeAppFile(t, source, "---\ntitle: Unsupported cover\n---\n\n## Section\n\nBody.\n")
			writeW073PNG(t, filepath.Join(dir, "cover.png"), 100, 50, color.RGBA{R: 230, G: 20, B: 150, A: 255})
			writeAppFile(t, filepath.Join(dir, "script.yaml"), "folio:\n  cover-image:\n    path: cover.png\n    mode: background\n")
			const sentinel = "existing target sentinel"
			writeAppFile(t, target, sentinel)
			_, diagnostic, err := officeRun(t, binary, t.TempDir(), officeEnv(t, home, nil), command, source, target)
			if err == nil || !strings.Contains(diagnostic, "cover") {
				t.Fatalf("expected unsupported cover failure, err=%v diagnostic=%s", err, diagnostic)
			}
			if got := readAppFile(t, target); got != sentinel {
				t.Fatalf("target changed: %q", got)
			}
		})
	}

	t.Run("unsupported letter consumer", func(t *testing.T) {
		dir, home := t.TempDir(), t.TempDir()
		source := filepath.Join(dir, "letters.org")
		target := filepath.Join(dir, "existing-two-theatre.pdf")
		writeAppFile(t, source, "#+AUTHOR: Example Author\n* Letters :letter:\n** Sender / Dublin :sender:\n** Submission :subject:\nHello.\n*** Recipient Two / Galway :to:\n**** Two Theatre :org:\n")
		writeW073PNG(t, filepath.Join(dir, "cover.png"), 100, 50, color.RGBA{R: 230, G: 20, B: 150, A: 255})
		writeAppFile(t, filepath.Join(dir, "script.yaml"), "folio:\n  cover-image:\n    path: cover.png\n    mode: background\n")
		const sentinel = "existing target sentinel"
		writeAppFile(t, target, sentinel)
		_, diagnostic, err := officeRun(t, binary, t.TempDir(), officeEnv(t, home, nil), "letter", source, "--dir", dir, "--prefix", "existing")
		if err == nil || !strings.Contains(diagnostic, "cover") {
			t.Fatalf("expected unsupported cover failure, err=%v diagnostic=%s", err, diagnostic)
		}
		if got := readAppFile(t, target); got != sentinel {
			t.Fatalf("target changed: %q", got)
		}
	})

	for _, extension := range []string{"pdf", "typ"} {
		t.Run("no-cover "+extension+" control", func(t *testing.T) {
			dir, home := t.TempDir(), t.TempDir()
			source, target := filepath.Join(dir, "play.org"), filepath.Join(dir, "control."+extension)
			writeAppFile(t, source, "#+TITLE: CONTROLTITLE\n* ACT ONE\n**** CÁIT\nBody.\n")
			_, diagnostic, err := officeRun(t, binary, t.TempDir(), officeEnv(t, home, nil), "convert", source, target)
			if err != nil {
				t.Fatalf("no-cover control: %v: %s", err, diagnostic)
			}
			if extension == "pdf" && !strings.Contains(w073PDFText(t, target), "CONTROLTITLE") {
				t.Fatal("no-cover PDF title changed")
			}
			if extension == "typ" && !strings.Contains(readAppFile(t, target), "CONTROLTITLE") {
				t.Fatal("no-cover Typst title changed")
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

func writeW073PNGHeader(t *testing.T, path string, width, height uint32) {
	t.Helper()
	data := make([]byte, 33)
	copy(data[:8], []byte{137, 80, 78, 71, 13, 10, 26, 10})
	binary.BigEndian.PutUint32(data[8:12], 13)
	copy(data[12:16], "IHDR")
	binary.BigEndian.PutUint32(data[16:20], width)
	binary.BigEndian.PutUint32(data[20:24], height)
	data[24], data[25], data[26], data[27], data[28] = 8, 2, 0, 0, 0
	binary.BigEndian.PutUint32(data[29:33], crc32.ChecksumIEEE(data[12:29]))
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
}
