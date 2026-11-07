//go:build folio_acceptance

// ABOUTME: Exercises W079 through a freshly built folio executable and rendered PDF output.
// ABOUTME: Keeps visual and atomic-failure evidence ticket-specific rather than in make test.
package app

import (
	"image"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAT079_1RenderedInsetBackdropsFollowPDFContract(t *testing.T) {
	requireW073Tools(t)
	binary := officeBuild(t)

	t.Run("dotted Org image backdrop", func(t *testing.T) {
		dir := t.TempDir()
		source := filepath.Join(dir, "play.org")
		target := filepath.Join(dir, "play.pdf")
		writeW079PatternPNG(t, filepath.Join(dir, "cover.png"), 900, 600)
		writeAppFile(t, source, `#+TITLE: IMAGETITLE
#+SUBTITLE: IMAGESUBTITLE
#+AUTHOR: IMAGEAUTHOR
#+DATE: IMAGEDATE
#+VERSION: IMAGEVERSION
#+FOLIO.COVER-IMAGE.PATH: cover.png
#+FOLIO.COVER-IMAGE.MODE: inset
#+FOLIO.COVER-IMAGE.PLACEMENT: dodge
#+FOLIO.COVER-IMAGE.SIZE: 80mm
#+FOLIO.COVER-IMAGE.BACKDROP.MODE: image
#+FOLIO.COVER-IMAGE.BACKDROP.IMAGE.BLUR: 8mm
#+FOLIO.COVER-IMAGE.BACKDROP.WASH.TREATMENT: dark
#+FOLIO.COVER-IMAGE.BACKDROP.WASH.OPACITY: 15%
* ACT ONE
**** CÁIT
Body.
`)
		_, stderr, err := officeRun(t, binary, dir, officeEnv(t, t.TempDir(), nil), "convert", source, target)
		if err != nil {
			t.Fatalf("image backdrop conversion: %v: %s", err, stderr)
		}
		raster := rasterW073Page(t, target, 1)
		bounds := raster.Bounds()
		for _, point := range []image.Point{{5, 5}, {bounds.Max.X - 6, 5}, {5, bounds.Max.Y - 6}, {bounds.Max.X - 6, bounds.Max.Y - 6}} {
			r, g, b, _ := raster.At(point.X, point.Y).RGBA()
			if r>>8 > 248 && g>>8 > 248 && b>>8 > 248 {
				t.Errorf("image backdrop left white page edge at %v", point)
			}
		}
		centre := image.Rect(bounds.Max.X/2-100, bounds.Max.Y/2-65, bounds.Max.X/2+100, bounds.Max.Y/2+65)
		corner := image.Rect(0, 0, 120, 120)
		sharpVariation := w079RegionVariation(raster, centre)
		backdropVariation := w079RegionVariation(raster, corner)
		if sharpVariation < backdropVariation*1.8 || sharpVariation < 20 {
			t.Fatalf("sharp inset variation %.2f does not exceed blurred backdrop %.2f", sharpVariation, backdropVariation)
		}
		for _, marker := range []string{"IMAGETITLE", "IMAGESUBTITLE", "IMAGEAUTHOR", "IMAGEDATE", "IMAGEVERSION"} {
			if got := w079PDFTextColour(t, target, marker); got != "#ffffff" {
				t.Errorf("%s colour = %s, want white", marker, got)
			}
		}
	})

	t.Run("YAML solid backdrop", func(t *testing.T) {
		dir := t.TempDir()
		source := filepath.Join(dir, "play.org")
		target := filepath.Join(dir, "play.pdf")
		writeW073PNG(t, filepath.Join(dir, "cover.png"), 400, 200, color.RGBA{R: 235, G: 30, B: 145, A: 255})
		writeAppFile(t, source, "#+TITLE: SOLIDTITLE\n#+SUBTITLE: SOLIDSUBTITLE\n#+AUTHOR: SOLIDAUTHOR\n#+DATE: SOLIDDATE\n#+VERSION: SOLIDVERSION\n* ACT ONE\n**** CÁIT\nBody.\n")
		writeAppFile(t, filepath.Join(dir, "script.yaml"), `folio:
  title-page:
    author:
      position: bottom-center
  cover-image:
    path: cover.png
    mode: inset
    placement: dodge
    size: 80mm
    backdrop:
      mode: solid
      solid:
        color: "#2a2524"
      wash:
        treatment: light
        opacity: 0%
`)
		_, stderr, err := officeRun(t, binary, dir, officeEnv(t, t.TempDir(), nil), "convert", source, target)
		if err != nil {
			t.Fatalf("solid backdrop conversion: %v: %s", err, stderr)
		}
		raster := rasterW073Page(t, target, 1)
		bounds := raster.Bounds()
		for _, point := range []image.Point{{5, 5}, {bounds.Max.X - 6, 5}, {5, bounds.Max.Y - 6}, {bounds.Max.X - 6, bounds.Max.Y - 6}} {
			if !w079NearRGB(raster.At(point.X, point.Y), 0x2a, 0x25, 0x24, 3) {
				t.Errorf("solid backdrop pixel at %v = %v, want #2a2524", point, raster.At(point.X, point.Y))
			}
		}
		for _, marker := range []string{"SOLIDTITLE", "SOLIDSUBTITLE", "SOLIDAUTHOR", "SOLIDDATE", "SOLIDVERSION"} {
			if got := w079PDFTextColour(t, target, marker); got != "#000000" {
				t.Errorf("%s colour = %s, want black", marker, got)
			}
		}
	})
}

func TestAT079_2InvalidBackdropInputPreservesTarget(t *testing.T) {
	requireW073Tools(t)
	binary := officeBuild(t)
	cases := []struct {
		name, output, want, config string
	}{
		{"background cover mode", "pdf", "folio.cover-image.backdrop", `    mode: background
    backdrop:
      mode: solid
      solid:
        color: "#2a2524"
      wash:
        treatment: light
`},
		{"solid without colour", "pdf", "folio.cover-image.backdrop.solid.color", `    mode: inset
    size: 80mm
    backdrop:
      mode: solid
      wash:
        treatment: light
`},
		{"malformed active colour", "pdf", "folio.cover-image.backdrop.solid.color", `    mode: inset
    size: 80mm
    backdrop:
      mode: solid
      solid:
        color: "#abc"
      wash:
        treatment: light
`},
		{"invalid active image blur", "pdf", "folio.cover-image.backdrop.image.blur", `    mode: inset
    size: 80mm
    backdrop:
      mode: image
      image:
        blur: 1em
      wash:
        treatment: light
`},
		{"unknown nested key", "pdf", "folio.cover-image.backdrop.wash.typo", `    mode: inset
    size: 80mm
    backdrop:
      mode: solid
      solid:
        color: "#2a2524"
      wash:
        treatment: light
        typo: true
`},
		{"unsupported output", "docx", "folio.cover-image.backdrop", `    mode: inset
    size: 80mm
    backdrop:
      mode: solid
      solid:
        color: "#2a2524"
      wash:
        treatment: light
`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			source := filepath.Join(dir, "play.org")
			target := filepath.Join(dir, "play."+tc.output)
			writeW073PNG(t, filepath.Join(dir, "cover.png"), 400, 200, color.RGBA{R: 235, G: 30, B: 145, A: 255})
			writeAppFile(t, source, "#+TITLE: Title\n* ACT ONE\n**** CÁIT\nBody.\n")
			writeAppFile(t, filepath.Join(dir, "script.yaml"), "folio:\n  cover-image:\n    path: cover.png\n    placement: overlay\n"+tc.config)
			const sentinel = "W079 sentinel target\n"
			writeAppFile(t, target, sentinel)
			_, stderr, err := officeRun(t, binary, dir, officeEnv(t, t.TempDir(), nil), "convert", source, target)
			if err == nil || !strings.Contains(stderr, tc.want) {
				t.Fatalf("error = %v, stderr = %q; want path %s", err, stderr, tc.want)
			}
			raw, readErr := os.ReadFile(target)
			if readErr != nil || string(raw) != sentinel {
				t.Fatalf("target changed after invalid input: bytes=%q err=%v", raw, readErr)
			}
		})
	}
}

func w079RegionVariation(img image.Image, region image.Rectangle) float64 {
	region = region.Intersect(img.Bounds())
	if region.Dx() < 2 || region.Dy() < 2 {
		return 0
	}
	total, samples := 0.0, 0
	for y := region.Min.Y; y < region.Max.Y; y++ {
		for x := region.Min.X + 1; x < region.Max.X; x++ {
			r1, g1, b1, _ := img.At(x-1, y).RGBA()
			r2, g2, b2, _ := img.At(x, y).RGBA()
			total += float64(absW079(int(r1>>8)-int(r2>>8))+absW079(int(g1>>8)-int(g2>>8))+absW079(int(b1>>8)-int(b2>>8))) / 3
			samples++
		}
	}
	return total / float64(samples)
}

func w079NearRGB(value color.Color, red, green, blue, tolerance int) bool {
	r, g, b, _ := value.RGBA()
	return absW079(int(r>>8)-red) <= tolerance && absW079(int(g>>8)-green) <= tolerance && absW079(int(b>>8)-blue) <= tolerance
}

func absW079(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
