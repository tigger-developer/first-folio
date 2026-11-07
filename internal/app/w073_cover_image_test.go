// ABOUTME: Defines retained W073 cover-image PDF, source-limit, and Superscale regressions.
// ABOUTME: Uses synthetic local images and a controlled executable boundary without network access.
package app

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const w073MagentaThreshold = 180

func init() {
	if filepath.Base(os.Args[0]) != "superscale" {
		return
	}
	marker := os.Getenv("FOLIO_TEST_SUPERSCALE_MARKER")
	if marker != "" {
		_ = os.WriteFile(marker, []byte(strings.Join(os.Args[1:], "\n")), 0o600)
	}
	mode := os.Getenv("FOLIO_TEST_SUPERSCALE_MODE")
	if mode == "nonzero" {
		os.Exit(7)
	}
	if mode == "unusable" {
		os.Exit(0)
	}
	var outputDir, input string
	for i := 1; i < len(os.Args); i++ {
		if os.Args[i] == "-o" && i+1 < len(os.Args) {
			outputDir = os.Args[i+1]
			i++
			continue
		}
		if !strings.HasPrefix(os.Args[i], "-") {
			input = os.Args[i]
		}
	}
	if outputDir == "" || input == "" {
		os.Exit(8)
	}
	raw, err := os.ReadFile(input)
	if err != nil || os.MkdirAll(outputDir, 0o700) != nil || os.WriteFile(filepath.Join(outputDir, "upscaled.png"), raw, 0o600) != nil {
		os.Exit(9)
	}
	os.Exit(0)
}

func TestRT073_1CoverSourceLimitsUseExactBoundaries(t *testing.T) {
	valid := []struct {
		name          string
		bytes         int64
		width, height uint32
	}{
		{"minimum", 1, 1, 1},
		{"exact bytes", coverSourceByteLimit, 1, 1},
		{"exact axis", 1024, coverSourceAxisLimit, 1},
		{"exact pixels", 1024, 10_000, 10_000},
	}
	for _, tc := range valid {
		t.Run(tc.name, func(t *testing.T) {
			if err := validateCoverSourceLimits(tc.bytes, tc.width, tc.height); err != nil {
				t.Fatalf("valid boundary rejected: %v", err)
			}
		})
	}
	invalid := []struct {
		name          string
		bytes         int64
		width, height uint32
		want          string
	}{
		{"zero width", 1, 0, 1, "dimension"},
		{"zero height", 1, 1, 0, "dimension"},
		{"one byte over", coverSourceByteLimit + 1, 1, 1, "104857600"},
		{"one axis unit over", 1, coverSourceAxisLimit + 1, 1, "32768"},
		{"decoded pixels over", 1, 10_001, 10_000, "100000000"},
	}
	for _, tc := range invalid {
		t.Run(tc.name, func(t *testing.T) {
			err := validateCoverSourceLimits(tc.bytes, tc.width, tc.height)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected limit %s, got %v", tc.want, err)
			}
		})
	}
}

func TestRT073_2RenderedInsetUsesLongestAxisAndCreatesImageOnlyTitlePage(t *testing.T) {
	requireW073Tools(t)
	dir := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	source := filepath.Join(dir, "play.org")
	cover := filepath.Join(dir, "cover.png")
	target := filepath.Join(dir, "play.pdf")
	writeAppFile(t, source, "* ACT ONE\n**** CÁIT\nBody marker.\n")
	writeW073PNG(t, cover, 400, 100, color.RGBA{R: 230, G: 20, B: 150, A: 255})
	writeAppFile(t, filepath.Join(dir, "script.yaml"), `folio:
  cover-image:
    path: cover.png
    mode: inset
    placement: overlay
    size: 50mm
`)
	status, stdout, stderr := runApp(t, "convert", source, target)
	if status != 0 {
		t.Fatalf("status %d stdout=%s stderr=%s", status, stdout, stderr)
	}
	raster := rasterW073Page(t, target, 1)
	box, ok := w073ColourBounds(raster)
	if !ok {
		t.Fatal("title page contains no cover-image pixels")
	}
	width, height := box.Dx(), box.Dy()
	if width < 137 || width > 147 || height < 31 || height > 41 {
		t.Fatalf("inset bounds=%v (%dx%d), want approximately 50mm x 12.5mm at 72dpi", box, width, height)
	}
	page := raster.Bounds()
	if absW073((box.Min.X+box.Max.X)-(page.Min.X+page.Max.X)) > 6 || absW073((box.Min.Y+box.Max.Y)-(page.Min.Y+page.Max.Y)) > 6 {
		t.Fatalf("inset is not centred: image=%v page=%v", box, page)
	}
	if !w073PDFHasPage(t, target, 2) {
		t.Fatal("image-only cover did not create a dedicated title page before the body")
	}
}

func TestRT073_2NoCoverControlAndDodgeOverflow(t *testing.T) {
	requireW073Tools(t)
	t.Run("no cover", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("HOME", t.TempDir())
		source, target := filepath.Join(dir, "play.org"), filepath.Join(dir, "play.pdf")
		writeAppFile(t, source, "#+TITLE: Control\n* ACT ONE\n**** CÁIT\nBody.\n")
		status, _, stderr := runApp(t, "convert", source, target)
		if status != 0 || !w073PDFHasPage(t, target, 1) {
			t.Fatalf("no-cover control status=%d stderr=%s", status, stderr)
		}
	})
	t.Run("dodge overflow", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("HOME", t.TempDir())
		source, target := filepath.Join(dir, "play.org"), filepath.Join(dir, "play.pdf")
		cover := filepath.Join(dir, "cover.png")
		writeW073PNG(t, cover, 100, 100, color.RGBA{R: 230, G: 20, B: 150, A: 255})
		writeAppFile(t, source, "#+TITLE: "+strings.Repeat("Very long title ", 80)+"\n#+AUTHOR: "+strings.Repeat("Long author ", 80)+"\n* ACT ONE\n**** CÁIT\nBody.\n")
		writeAppFile(t, filepath.Join(dir, "script.yaml"), "folio:\n  cover-image:\n    path: cover.png\n    mode: inset\n    placement: dodge\n    size: 190mm\n")
		status, _, stderr := runApp(t, "convert", source, target)
		if status == 0 || (!strings.Contains(stderr, "upper") && !strings.Contains(stderr, "lower") && !strings.Contains(stderr, "cover")) {
			t.Fatalf("overflow status=%d stderr=%s", status, stderr)
		}
	})
}

func TestRT073_3SuperscaleInvocationAndFallbackAreDeterministic(t *testing.T) {
	requireW073Tools(t)
	for _, tc := range []struct {
		name, value, mode string
		fake, warning     bool
	}{
		{"disabled", "null", "", false, false},
		{"missing", "default", "", false, true},
		{"nonzero", "2x", "nonzero", true, true},
		{"unusable", "4x", "unusable", true, true},
		{"usable", "8x", "success", true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			home, tempRoot := t.TempDir(), t.TempDir()
			toolDir := controlledW073Tools(t, tc.fake)
			t.Setenv("HOME", home)
			t.Setenv("TMPDIR", tempRoot)
			t.Setenv("PATH", toolDir)
			t.Setenv("FOLIO_TEST_SUPERSCALE_MODE", tc.mode)
			marker := filepath.Join(dir, "superscale.marker")
			t.Setenv("FOLIO_TEST_SUPERSCALE_MARKER", marker)
			source, target := filepath.Join(dir, "play.org"), filepath.Join(dir, "play.pdf")
			cover := filepath.Join(dir, "cover image.png")
			writeW073PNG(t, cover, 100, 50, color.RGBA{R: 230, G: 20, B: 150, A: 255})
			writeAppFile(t, source, "#+TITLE: Superscale\n* ACT ONE\n**** CÁIT\nBody.\n")
			writeAppFile(t, filepath.Join(dir, "script.yaml"), "folio:\n  cover-image:\n    path: 'cover image.png'\n    mode: inset\n    placement: overlay\n    size: 50mm\n    superscale: "+tc.value+"\n")
			status, _, stderr := runApp(t, "convert", source, target)
			if status != 0 {
				t.Fatalf("status=%d stderr=%s", status, stderr)
			}
			if got := strings.Count(stderr, "folio.cover-image.superscale"); got != map[bool]int{false: 0, true: 1}[tc.warning] {
				t.Fatalf("warning count=%d want warning=%t: %s", got, tc.warning, stderr)
			}
			_, markerErr := os.Stat(marker)
			if tc.fake != (markerErr == nil) {
				t.Fatalf("superscale invoked=%t want=%t: %v", markerErr == nil, tc.fake, markerErr)
			}
			if tc.fake && tc.mode == "success" {
				args := readAppFile(t, marker)
				if !strings.Contains(args, "--no-face-enhance") || !strings.Contains(args, "--scale\n8") || !strings.Contains(args, "cover image.png") {
					t.Fatalf("unexpected superscale argv: %q", args)
				}
			}
			assertNoW073TemporaryOutput(t, tempRoot)
		})
	}
}

func writeW073PNG(t *testing.T, path string, width, height int, fill color.RGBA) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.SetRGBA(x, y, fill)
		}
	}
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(file, img); err != nil {
		_ = file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
}

func requireW073Tools(t *testing.T) {
	t.Helper()
	for _, tool := range []string{"typst", "pdftoppm", "pdfinfo"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("%s is not installed", tool)
		}
	}
}

func controlledW073Tools(t *testing.T, fakeSuperscale bool) string {
	t.Helper()
	dir := t.TempDir()
	for _, tool := range []string{"typst", "pdftoppm", "pdfinfo"} {
		path, err := exec.LookPath(tool)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(path, filepath.Join(dir, tool)); err != nil {
			t.Fatal(err)
		}
	}
	if fakeSuperscale {
		self, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		raw, err := os.ReadFile(self)
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(dir, "superscale")
		if err := os.WriteFile(path, raw, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func rasterW073Page(t *testing.T, pdf string, page int) image.Image {
	t.Helper()
	prefix := filepath.Join(t.TempDir(), "page")
	cmd := exec.Command("pdftoppm", "-f", stringIntW073(page), "-l", stringIntW073(page), "-singlefile", "-r", "72", "-png", pdf, prefix)
	if raw, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("rasterizing page %d: %v: %s", page, err, raw)
	}
	file, err := os.Open(prefix + ".png")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	img, err := png.Decode(file)
	if err != nil {
		t.Fatal(err)
	}
	return img
}

func w073ColourBounds(img image.Image) (image.Rectangle, bool) {
	bounds := img.Bounds()
	box := image.Rect(bounds.Max.X, bounds.Max.Y, bounds.Min.X, bounds.Min.Y)
	found := false
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			if int(r>>8) < w073MagentaThreshold || int(g>>8) > 80 || int(b>>8) < 100 {
				continue
			}
			found = true
			if x < box.Min.X {
				box.Min.X = x
			}
			if y < box.Min.Y {
				box.Min.Y = y
			}
			if x+1 > box.Max.X {
				box.Max.X = x + 1
			}
			if y+1 > box.Max.Y {
				box.Max.Y = y + 1
			}
		}
	}
	return box, found
}

func w073PDFHasPage(t *testing.T, pdf string, page int) bool {
	t.Helper()
	raw, err := exec.Command("pdfinfo", pdf).CombinedOutput()
	if err != nil {
		t.Fatalf("pdfinfo: %v: %s", err, raw)
	}
	return strings.Contains(string(raw), "Pages:") && w073PageCount(string(raw)) >= page
}

func w073PageCount(info string) int {
	for _, line := range strings.Split(info, "\n") {
		if strings.HasPrefix(line, "Pages:") {
			value := strings.TrimSpace(strings.TrimPrefix(line, "Pages:"))
			count := 0
			for _, r := range value {
				if r < '0' || r > '9' {
					break
				}
				count = count*10 + int(r-'0')
			}
			return count
		}
	}
	return 0
}

func stringIntW073(value int) string {
	if value == 0 {
		return "0"
	}
	var digits [20]byte
	i := len(digits)
	for value > 0 {
		i--
		digits[i] = byte('0' + value%10)
		value /= 10
	}
	return string(digits[i:])
}

func absW073(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

func assertNoW073TemporaryOutput(t *testing.T, root string) {
	t.Helper()
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "folio-cover-") {
			t.Fatalf("temporary cover output remains: %s", entry.Name())
		}
	}
}
