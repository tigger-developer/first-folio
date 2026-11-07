// ABOUTME: Defines retained W073 cover-image PDF, source-limit, and Superscale regressions.
// ABOUTME: Uses synthetic local images and a controlled executable boundary without network access.
package app

import (
	"html"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/tigger-developer/first-folio/internal/config"
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
		if (os.Args[i] == "-o" || os.Args[i] == "--scale") && i+1 < len(os.Args) {
			if os.Args[i] == "-o" {
				outputDir = os.Args[i+1]
			}
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
	if _, err := os.Stat(input); err != nil || os.MkdirAll(outputDir, 0o700) != nil {
		os.Exit(9)
	}
	output, err := os.Create(filepath.Join(outputDir, "upscaled.png"))
	if err != nil {
		os.Exit(9)
	}
	upscaled := image.NewRGBA(image.Rect(0, 0, 200, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 200; x++ {
			upscaled.SetRGBA(x, y, color.RGBA{R: 20, G: 190, B: 210, A: 255})
		}
	}
	if err := png.Encode(output, upscaled); err != nil || output.Close() != nil {
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

func TestRT073_1CoverPlanResolvesSourcePathFormatsAndGeometry(t *testing.T) {
	root := t.TempDir()
	sourceDir := filepath.Join(root, "source")
	configDir := filepath.Join(root, "configuration")
	workDir := filepath.Join(root, "working")
	for _, dir := range []string{sourceDir, configDir, workDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	source := filepath.Join(sourceDir, "play.org")
	writeAppFile(t, source, "* ACT ONE\n**** CÁIT\nBody.\n")
	writeW073PNG(t, filepath.Join(sourceDir, "portrait.png"), 200, 400, color.RGBA{R: 230, G: 20, B: 150, A: 255})
	writeW073JPEG(t, filepath.Join(sourceDir, "landscape.jpg"), 1000, 250, color.RGBA{R: 230, G: 20, B: 150, A: 255})
	for _, tc := range []struct {
		name, path, size      string
		wantWidth, wantHeight float64
	}{
		{"portrait PNG", "portrait.png", "80mm", 40, 80},
		{"landscape JPEG", "landscape.jpg", "50mm", 50, 12.5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := config.Load(config.Options{Mode: config.ModeScript, Home: t.TempDir(), LocalDir: configDir, Source: map[string]any{"folio": map[string]any{"cover-image": map[string]any{"path": tc.path, "mode": "inset", "size": tc.size}}}})
			if err != nil {
				t.Fatal(err)
			}
			oldWorking, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}
			if err := os.Chdir(workDir); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = os.Chdir(oldWorking) })
			plan, err := planCoverImage(cfg, source)
			if err != nil {
				t.Fatalf("planning: %v", err)
			}
			if plan.SourcePath != filepath.Join(sourceDir, tc.path) || plan.RenderedWidthMM != tc.wantWidth || plan.RenderedHeightMM != tc.wantHeight {
				t.Fatalf("plan=%#v, want source-relative %.1fx%.1fmm", plan, tc.wantWidth, tc.wantHeight)
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
	for _, tc := range []struct {
		name, source, want string
	}{
		{"upper dodge overflow", "#+TITLE: " + strings.Repeat("Very long title ", 100) + "\n", "upper"},
		{"lower dodge overflow", "#+AUTHOR: " + strings.Repeat("Very long author ", 100) + "\n", "lower"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			t.Setenv("HOME", t.TempDir())
			source, target := filepath.Join(dir, "play.org"), filepath.Join(dir, "play.pdf")
			cover := filepath.Join(dir, "cover.png")
			writeW073PNG(t, cover, 100, 100, color.RGBA{R: 230, G: 20, B: 150, A: 255})
			writeAppFile(t, source, tc.source+"* ACT ONE\n**** CÁIT\nBody.\n")
			writeAppFile(t, filepath.Join(dir, "script.yaml"), "folio:\n  cover-image:\n    path: cover.png\n    mode: inset\n    placement: dodge\n    size: 190mm\n")
			status, _, stderr := runApp(t, "convert", source, target)
			if status == 0 || !strings.Contains(stderr, tc.want) {
				t.Fatalf("%s overflow status=%d stderr=%s", tc.want, status, stderr)
			}
		})
	}
}

func TestRT073_2DodgeFitAcceptsExactHeightAndRejectsPointOneOver(t *testing.T) {
	for _, group := range []string{"upper", "lower"} {
		if err := validateDodgeGroupFit(group, 72, 72); err != nil {
			t.Fatalf("%s exact-height group rejected: %v", group, err)
		}
		err := validateDodgeGroupFit(group, 72.1, 72)
		if err == nil || !strings.Contains(err.Error(), group) {
			t.Fatalf("%s 0.1pt overflow = %v", group, err)
		}
	}
}

func TestRT073_2CoverVariantsApplyImageTreatmentAndPhysicalBlur(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "checker.png")
	img := image.NewRGBA(image.Rect(0, 0, 21, 21))
	for y := 0; y < 21; y++ {
		for x := 0; x < 21; x++ {
			value := uint8(0)
			if (x+y)%2 == 0 {
				value = 255
			}
			img.SetRGBA(x, y, color.RGBA{R: value, G: value, B: value, A: 255})
		}
	}
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(file, img); err != nil || file.Close() != nil {
		t.Fatalf("writing checker: %v", err)
	}
	plan := &coverImagePlan{
		Config:     config.CoverImage{ImageTreatment: "dark", ImageOpacity: 0.5, OverlayBlur: "2mm"},
		SourcePath: path, PixelWidth: 21, PixelHeight: 21, RenderedWidthMM: 10, RenderedHeightMM: 10,
	}
	prepared, err := prepareCoverVariants(plan, dir)
	if err != nil {
		t.Fatalf("preparing variants: %v", err)
	}
	treated := decodeW073PNG(t, prepared.ImagePath)
	blurred := decodeW073PNG(t, prepared.BlurredPath)
	r, _, _, _ := treated.At(0, 0).RGBA()
	if r>>8 < 120 || r>>8 > 135 {
		t.Fatalf("dark treatment pixel=%d, want approximately 128", r>>8)
	}
	r, _, _, _ = blurred.At(10, 10).RGBA()
	if r>>8 < 45 || r>>8 > 90 {
		t.Fatalf("blurred checker centre=%d, want smoothed treated value", r>>8)
	}
}

func TestRT073_2OverlayAndDodgeUseLiveMeasuredText(t *testing.T) {
	requireW073Tools(t)
	t.Run("overlay", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("HOME", t.TempDir())
		source, target := filepath.Join(dir, "play.org"), filepath.Join(dir, "overlay.pdf")
		writeAppFile(t, source, "#+TITLE: LIVEOVERLAY\n* ACT ONE\n**** CÁIT\nBody.\n")
		writeW073PNG(t, filepath.Join(dir, "cover.png"), 600, 900, color.RGBA{R: 230, G: 20, B: 150, A: 255})
		writeAppFile(t, filepath.Join(dir, "script.yaml"), "folio:\n  cover-image:\n    path: cover.png\n    mode: background\n    overlay:\n      blur: 2mm\n      padding: 5mm\n      wash:\n        treatment: dark\n        opacity: 80%\n")
		status, _, stderr := runApp(t, "convert", source, target)
		if status != 0 {
			t.Fatalf("status=%d stderr=%s", status, stderr)
		}
		if !strings.Contains(w073PDFText(t, target), "LIVEOVERLAY") {
			t.Fatal("overlay title is not live PDF text")
		}
		page := rasterW073Page(t, target, 1)
		if _, ok := w073ColourBounds(page); !ok {
			t.Fatal("overlay background image absent")
		}
	})

	for _, position := range []string{"center", "third"} {
		t.Run("dodge "+position, func(t *testing.T) {
			dir := t.TempDir()
			t.Setenv("HOME", t.TempDir())
			source, target := filepath.Join(dir, "play.org"), filepath.Join(dir, "dodge.pdf")
			writeAppFile(t, source, "#+TITLE: UPPERREGRESSION\n#+AUTHOR: LOWERREGRESSION\n* ACT ONE\n**** CÁIT\nBody.\n")
			writeW073PNG(t, filepath.Join(dir, "cover.png"), 400, 200, color.RGBA{R: 230, G: 20, B: 150, A: 255})
			writeAppFile(t, filepath.Join(dir, "script.yaml"), "folio:\n  title-page:\n    title:\n      position: "+position+"\n  cover-image:\n    path: cover.png\n    mode: inset\n    placement: dodge\n    size: 100mm\n")
			status, _, stderr := runApp(t, "convert", source, target)
			if status != 0 {
				t.Fatalf("status=%d stderr=%s", status, stderr)
			}
			page := rasterW073Page(t, target, 1)
			imageBox, ok := w073ColourBounds(page)
			if !ok {
				t.Fatal("dodge image absent")
			}
			upper := w073WordBounds(t, target, "UPPERREGRESSION")
			lower := w073WordBounds(t, target, "LOWERREGRESSION")
			if upper.yMax >= float64(imageBox.Min.Y) || lower.yMin <= float64(imageBox.Max.Y) {
				t.Fatalf("dodge overlap position=%s image=%v upper=%#v lower=%#v", position, imageBox, upper, lower)
			}
		})
	}
}

func TestRT073_3SuperscaleInvocationAndFallbackAreDeterministic(t *testing.T) {
	requireW073Tools(t)
	for _, tc := range []struct {
		name, value, mode, scale string
		fake, warning, generated bool
	}{
		{"disabled", "null", "", "", false, false, false},
		{"missing", "default", "", "", false, true, false},
		{"nonzero", "2x", "nonzero", "2", true, true, false},
		{"unusable", "4x", "unusable", "4", true, true, false},
		{"usable default", "default", "success", "", true, false, true},
		{"usable 2x", "2x", "success", "2", true, false, true},
		{"usable 4x", "4x", "success", "4", true, false, true},
		{"usable 8x", "8x", "success", "8", true, false, true},
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
			cover := filepath.Join(dir, "cover [literal] $ image.png")
			writeW073PNG(t, cover, 100, 50, color.RGBA{R: 230, G: 20, B: 150, A: 255})
			writeAppFile(t, source, "#+TITLE: Superscale\n* ACT ONE\n**** CÁIT\nBody.\n")
			writeAppFile(t, filepath.Join(dir, "script.yaml"), "folio:\n  cover-image:\n    path: 'cover [literal] $ image.png'\n    mode: inset\n    placement: overlay\n    size: 50mm\n    superscale: "+tc.value+"\n")
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
			if tc.fake {
				args := readAppFile(t, marker)
				if !strings.Contains(args, "--no-face-enhance") || !strings.Contains(args, "cover [literal] $ image.png") {
					t.Fatalf("unexpected superscale argv: %q", args)
				}
				if tc.scale == "" && strings.Contains(args, "--scale") {
					t.Fatalf("default invocation unexpectedly set scale: %q", args)
				}
				if tc.scale != "" && !strings.Contains(args, "--scale\n"+tc.scale+"\n") {
					t.Fatalf("scale %s missing from argv: %q", tc.scale, args)
				}
			}
			page := rasterW073Page(t, target, 1)
			if tc.generated != w073ContainsCyan(page) {
				t.Fatalf("generated-image pixels=%t want=%t", w073ContainsCyan(page), tc.generated)
			}
			assertNoW073TemporaryOutput(t, tempRoot)
		})
	}
}

func writeW073JPEG(t *testing.T, path string, width, height int, fill color.RGBA) {
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
	if err := jpeg.Encode(file, img, &jpeg.Options{Quality: 95}); err != nil {
		_ = file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
}

func w073ContainsCyan(img image.Image) bool {
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			if r>>8 < 60 && g>>8 > 150 && b>>8 > 170 {
				return true
			}
		}
	}
	return false
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

func writeW073CropMarkerPNG(t *testing.T, path string, width, height int) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		fill := color.RGBA{R: 230, G: 20, B: 150, A: 255}
		if y < height/20 || y >= height-height/20 {
			fill = color.RGBA{R: 20, G: 200, B: 30, A: 255}
		}
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

func decodeW073PNG(t *testing.T, path string) image.Image {
	t.Helper()
	file, err := os.Open(path)
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
	box := image.Rectangle{Min: image.Pt(bounds.Max.X, bounds.Max.Y), Max: image.Pt(bounds.Min.X, bounds.Min.Y)}
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

func w073PDFText(t *testing.T, path string) string {
	t.Helper()
	raw, err := exec.Command("pdftotext", path, "-").CombinedOutput()
	if err != nil {
		t.Fatalf("extracting PDF text: %v: %s", err, raw)
	}
	return string(raw)
}

type wordBoundsW073 struct {
	xMin, yMin, xMax, yMax float64
}

func (b wordBoundsW073) centerY() float64 { return (b.yMin + b.yMax) / 2 }

var wordREW073 = regexp.MustCompile(`<word xMin="([0-9.]+)" yMin="([0-9.]+)" xMax="([0-9.]+)" yMax="([0-9.]+)">([^<]+)</word>`)

func w073WordBounds(t *testing.T, path, marker string) wordBoundsW073 {
	t.Helper()
	raw, err := exec.Command("pdftotext", "-bbox", path, "-").CombinedOutput()
	if err != nil {
		t.Fatalf("extracting PDF word bounds: %v: %s", err, raw)
	}
	for _, match := range wordREW073.FindAllStringSubmatch(string(raw), -1) {
		if html.UnescapeString(match[5]) != marker {
			continue
		}
		values := [4]float64{}
		for i := range values {
			values[i], err = strconv.ParseFloat(match[i+1], 64)
			if err != nil {
				t.Fatalf("parsing PDF bound %q: %v", match[i+1], err)
			}
		}
		return wordBoundsW073{xMin: values[0], yMin: values[1], xMax: values[2], yMax: values[3]}
	}
	t.Fatalf("word %q not found in PDF bounds", marker)
	return wordBoundsW073{}
}

func absoluteFloatW073(value float64) float64 {
	if value < 0 {
		return -value
	}
	return value
}
