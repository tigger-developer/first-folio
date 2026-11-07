// ABOUTME: Defines W079 retained image-preparation and rendered text-colour regressions.
// ABOUTME: Verifies the derived backdrop independently from the unchanged sharp inset path.
package app

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tigger-developer/first-folio/internal/config"
)

func TestRT079_1BackdropImagePreparationBoundsCropAndDoesNotUpscale(t *testing.T) {
	cases := []struct {
		name          string
		width, height int
		wantMax       int
	}{
		{"large source is capped", 1000, 1000, 512},
		{"small source is not enlarged", 200, 100, 100},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			source := filepath.Join(dir, "play.org")
			cover := filepath.Join(dir, "cover.png")
			writeAppFile(t, source, "* ACT ONE\n**** CÁIT\nBody.\n")
			writeW079PatternPNG(t, cover, tc.width, tc.height)
			cfg := loadW079CoverConfig(t, dir, map[string]any{
				"mode": "image", "image": map[string]any{"blur": "8mm"}, "wash": map[string]any{"treatment": "dark"},
			})
			plan, cleanup, err := prepareCoverForRender(cfg, source, nil)
			if err != nil {
				t.Fatalf("preparing cover: %v", err)
			}
			defer cleanup()
			if plan.BackdropPath == "" || plan.BackdropPath == plan.SourcePath {
				t.Fatalf("backdrop path = %q, sharp path = %q; want an independent derivative", plan.BackdropPath, plan.SourcePath)
			}
			if plan.OriginalSourcePath != cover {
				t.Fatalf("original source = %q, want %q", plan.OriginalSourcePath, cover)
			}
			backdrop := decodeW073PNG(t, plan.BackdropPath)
			bounds := backdrop.Bounds()
			longest := maxInt(bounds.Dx(), bounds.Dy())
			if longest > 512 || longest != tc.wantMax {
				t.Fatalf("backdrop dimensions %dx%d have longest axis %d, want %d", bounds.Dx(), bounds.Dy(), longest, tc.wantMax)
			}
			ratio := float64(bounds.Dx()) / float64(bounds.Dy())
			pageRatio := plan.PageWidthMM / plan.PageHeightMM
			if ratio < pageRatio-0.01 || ratio > pageRatio+0.01 {
				t.Fatalf("backdrop ratio %.4f, want page ratio %.4f", ratio, pageRatio)
			}
		})
	}

	dir := t.TempDir()
	source := filepath.Join(dir, "play.org")
	cover := filepath.Join(dir, "cover.png")
	writeAppFile(t, source, "* ACT ONE\n**** CÁIT\nBody.\n")
	writeW079PatternPNG(t, cover, 300, 200)
	cfg := loadW079CoverConfig(t, dir, map[string]any{
		"mode": "solid", "solid": map[string]any{"color": "#2a2524"}, "wash": map[string]any{"treatment": "light"},
	})
	plan, cleanup, err := prepareCoverForRender(cfg, source, nil)
	if err != nil {
		t.Fatalf("preparing solid cover: %v", err)
	}
	defer cleanup()
	if plan.BackdropPath != "" {
		t.Fatalf("solid backdrop produced image derivative %q", plan.BackdropPath)
	}
	noneCfg := loadW079CoverConfig(t, dir, map[string]any{
		"mode": "none", "image": map[string]any{"blur": "12mm"}, "solid": map[string]any{"color": "#2a2524"}, "wash": map[string]any{"treatment": "dark"},
	})
	nonePlan, noneCleanup, err := prepareCoverForRender(noneCfg, source, nil)
	if err != nil {
		t.Fatalf("preparing disabled backdrop: %v", err)
	}
	defer noneCleanup()
	if nonePlan.BackdropPath != "" {
		t.Fatalf("mode none produced image derivative %q", nonePlan.BackdropPath)
	}
}

func TestRT079_1OrgAndYAMLBackdropValuesResolveIdentically(t *testing.T) {
	dir := t.TempDir()
	org, warnings, err := orgScriptSourceConfig(`#+FOLIO.COVER-IMAGE.PATH: cover.png
#+FOLIO.COVER-IMAGE.MODE: inset
#+FOLIO.COVER-IMAGE.PLACEMENT: overlay
#+FOLIO.COVER-IMAGE.SIZE: 80mm
#+FOLIO.COVER-IMAGE.BACKDROP.MODE: solid
#+FOLIO.COVER-IMAGE.BACKDROP.SOLID.COLOR: #2a2524
#+FOLIO.COVER-IMAGE.BACKDROP.WASH.TREATMENT: dark
#+FOLIO.COVER-IMAGE.BACKDROP.WASH.OPACITY: 25%
* ACT ONE
**** CÁIT
Body.
`, filepath.Join(dir, "play.org"))
	if err != nil || len(warnings) != 0 {
		t.Fatalf("Org backdrop projection: warnings=%v err=%v", warnings, err)
	}
	yaml := map[string]any{"folio": map[string]any{"cover-image": map[string]any{
		"path": "cover.png", "mode": "inset", "placement": "overlay", "size": "80mm",
		"backdrop": map[string]any{"mode": "solid", "solid": map[string]any{"color": "#2a2524"}, "wash": map[string]any{"treatment": "dark", "opacity": "25%"}},
	}}}
	load := func(source map[string]any) config.CoverImage {
		cfg, err := config.Load(config.Options{Mode: config.ModeScript, Home: t.TempDir(), LocalDir: dir, Source: source})
		if err != nil {
			t.Fatalf("loading projected source: %v", err)
		}
		cover, enabled, err := cfg.CoverImage()
		if err != nil || !enabled {
			t.Fatalf("typed cover: %#v, %t, %v", cover, enabled, err)
		}
		return cover
	}
	if got, want := load(org), load(yaml); got != want {
		t.Fatalf("Org cover %#v differs from YAML cover %#v", got, want)
	}
}

func TestRT079_1BackdropDerivativeUsesOriginalRatherThanSelectedSharpAsset(t *testing.T) {
	dir := t.TempDir()
	original := filepath.Join(dir, "original.png")
	selected := filepath.Join(dir, "superscaled.png")
	writeW073PNG(t, original, 120, 120, color.RGBA{R: 230, G: 20, B: 140, A: 128})
	writeW073PNG(t, selected, 240, 120, color.RGBA{R: 20, G: 190, B: 220, A: 255})
	plan := &coverImagePlan{
		Config: config.CoverImage{
			Mode: "inset", Size: "80mm", ImageTreatment: "dark", ImageOpacity: 0.5,
			OverlayBlur: "0mm", BackdropMode: "image", BackdropImageBlur: "8mm", BackdropWashTreatment: "dark",
		},
		SourcePath: selected, OriginalSourcePath: original,
		PixelWidth: 240, PixelHeight: 120, RenderedWidthMM: 80, RenderedHeightMM: 40,
		PageWidthMM: 210, PageHeightMM: 297,
	}
	variants, err := prepareCoverVariants(plan, dir)
	if err != nil {
		t.Fatalf("preparing variants: %v", err)
	}
	if variants.BackdropPath == "" {
		t.Fatal("image backdrop derivative was not produced")
	}
	img := decodeW073PNG(t, variants.BackdropPath)
	r, g, b, a := img.At(img.Bounds().Dx()/2, img.Bounds().Dy()/2).RGBA()
	if int(r>>8) < 52 || int(r>>8) > 64 || int(g>>8) > 8 || int(b>>8) < 30 || int(b>>8) > 42 || a>>8 != 255 {
		t.Fatalf("backdrop centre rgba(%d,%d,%d,%d) does not reflect original transparent source with dark 50%% treatment", r>>8, g>>8, b>>8, a>>8)
	}
}

func TestRT079_2BackdropAndOverlayTreatmentsResolveAtRenderedTextBounds(t *testing.T) {
	requireW073Tools(t)
	if _, err := exec.LookPath("pdftohtml"); err != nil {
		t.Skip("pdftohtml is not installed")
	}
	cases := []struct {
		name, placement, backdropTreatment, overlayTreatment, want string
	}{
		{"dodge light uses black", "dodge", "light", "", "#000000"},
		{"dodge dark uses white", "dodge", "dark", "", "#ffffff"},
		{"overlay light overrides dark backdrop", "overlay", "dark", "light", "#000000"},
		{"overlay dark overrides light backdrop", "overlay", "light", "dark", "#ffffff"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			t.Setenv("HOME", t.TempDir())
			source := filepath.Join(dir, "play.org")
			target := filepath.Join(dir, "play.pdf")
			writeW073PNG(t, filepath.Join(dir, "cover.png"), 400, 200, color.RGBA{R: 120, G: 80, B: 180, A: 255})
			writeAppFile(t, source, "#+TITLE: TITLEMARKER\n#+SUBTITLE: SUBTITLEMARKER\n#+AUTHOR: AUTHORMARKER\n#+VERSION: VERSIONMARKER\n#+DATE: DATEMARKER\n* ACT ONE\n**** CÁIT\nBody.\n")
			overlay := ""
			if tc.overlayTreatment != "" {
				overlay = fmt.Sprintf("    overlay:\n      wash:\n        treatment: %s\n        opacity: 25%%\n", tc.overlayTreatment)
			}
			writeAppFile(t, filepath.Join(dir, "script.yaml"), fmt.Sprintf(`folio:
  title-page:
    author:
      position: bottom-center
  cover-image:
    path: cover.png
    mode: inset
    placement: %s
    size: 80mm
%s    backdrop:
      mode: solid
      solid:
        color: "#2a2524"
      wash:
        treatment: %s
        opacity: 0%%
`, tc.placement, overlay, tc.backdropTreatment))
			status, _, stderr := runApp(t, "convert", source, target)
			if status != 0 {
				t.Fatalf("convert status %d: %s", status, stderr)
			}
			for _, marker := range []string{"TITLEMARKER", "SUBTITLEMARKER", "AUTHORMARKER", "VERSIONMARKER", "DATEMARKER"} {
				if got := w079PDFTextColour(t, target, marker); got != tc.want {
					t.Errorf("%s colour = %s, want %s", marker, got, tc.want)
				}
			}
		})
	}
}

func loadW079CoverConfig(t *testing.T, dir string, backdrop map[string]any) config.Config {
	t.Helper()
	cfg, err := config.Load(config.Options{Mode: config.ModeScript, Home: t.TempDir(), LocalDir: dir, Source: map[string]any{"folio": map[string]any{"cover-image": map[string]any{
		"path": "cover.png", "mode": "inset", "placement": "overlay", "size": "80mm", "backdrop": backdrop,
	}}}})
	if err != nil {
		t.Fatalf("loading cover config: %v", err)
	}
	return cfg
}

func writeW079PatternPNG(t *testing.T, path string, width, height int) {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			cell := ((x / 8) + (y / 8)) % 2
			if cell == 0 {
				img.SetNRGBA(x, y, color.NRGBA{R: 235, G: 30, B: 145, A: 255})
			} else {
				img.SetNRGBA(x, y, color.NRGBA{R: 25, G: 180, B: 220, A: 255})
			}
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

func w079PDFTextColour(t *testing.T, pdf, marker string) string {
	t.Helper()
	xmlPath := filepath.Join(t.TempDir(), "text.xml")
	proseTool(t, "pdftohtml", "-f", "1", "-l", "1", "-xml", "-zoom", "1", "-noroundcoord", pdf, xmlPath)
	raw, err := os.ReadFile(xmlPath)
	if err != nil {
		t.Fatal(err)
	}
	tree := officeParseXML(t, raw)
	fonts := map[string]string{}
	for _, font := range tree.all("fontspec") {
		fonts[font.attr("id")] = strings.ToLower(font.attr("color"))
	}
	for _, node := range tree.all("text") {
		if strings.Contains(node.text(), marker) {
			return fonts[node.attr("font")]
		}
	}
	t.Fatalf("marker %q not found in page-one PDF text XML", marker)
	return ""
}
