//go:build folio_acceptance

// ABOUTME: Exercises W074 through a freshly built executable and rendered DOCX title pages.
// ABOUTME: Measures native package, geometry, backdrop, editability and atomic publication evidence.
package app

import (
	"archive/zip"
	"bytes"
	"image"
	"image/color"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func requireW074Tools(t *testing.T) {
	t.Helper()
	for _, tool := range []string{"soffice", "pdftoppm", "pdftotext"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Fatalf("W074 acceptance requires %s on PATH: %v", tool, err)
		}
	}
}

func w074PackageMember(t *testing.T, path, name string) []byte {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	z, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range z.File {
		if file.Name != name {
			continue
		}
		reader, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		data, readErr := io.ReadAll(reader)
		closeErr := reader.Close()
		if readErr != nil || closeErr != nil {
			t.Fatalf("read %s: %v / %v", name, readErr, closeErr)
		}
		return data
	}
	t.Fatalf("missing package member %s", name)
	return nil
}

func TestAT074_1PublicDOCXConversionPreservesTitleImageSemantics(t *testing.T) {
	requireW074Tools(t)
	binary := officeBuild(t)

	t.Run("background overlay and exact title spacing", func(t *testing.T) {
		dir := t.TempDir()
		source, target := filepath.Join(dir, "play.org"), filepath.Join(dir, "play.docx")
		writeW073CropMarkerPNG(t, filepath.Join(dir, "cover.png"), 300, 500)
		writeAppFile(t, source, "#+TITLE: BACKGROUNDTITLE\n#+SUBTITLE: BACKGROUNDSUBTITLE\n#+AUTHOR: BACKGROUNDAUTHOR\n#+DATE: BACKGROUNDDATE\n#+VERSION: BACKGROUNDVERSION\n* ACT ONE\n**** CÁIT\nBODYMARKER.\n")
		writeAppFile(t, filepath.Join(dir, "script.yaml"), `folio:
  title-page:
    title:
      position: 40%
    subtitle:
      space-before: 4mm
  cover-image:
    path: cover.png
    mode: background
    placement: overlay
    overlay:
      blur: 2mm
      padding: 5mm
      wash:
        treatment: dark
        opacity: 70%
`)
		_, stderr, err := officeRun(t, binary, dir, officeEnv(t, t.TempDir(), nil), "convert", source, target)
		if err != nil {
			t.Fatalf("convert: %v: %s", err, stderr)
		}
		document := string(w074PackageMember(t, target, "word/document.xml"))
		for _, marker := range []string{"BACKGROUNDTITLE", "BACKGROUNDSUBTITLE", "BACKGROUNDAUTHOR", "BACKGROUNDDATE", "BACKGROUNDVERSION", "BODYMARKER"} {
			if !strings.Contains(document, marker) {
				t.Errorf("editable document text lacks %s", marker)
			}
		}
		pdf := proseRenderPDF(t, target)
		page := rasterW073Page(t, pdf, 1)
		for _, point := range []image.Point{{2, 2}, {page.Bounds().Max.X - 3, 2}, {2, page.Bounds().Max.Y - 3}, {page.Bounds().Max.X - 3, page.Bounds().Max.Y - 3}} {
			if !isW073Magenta(page.At(point.X, point.Y)) {
				t.Errorf("background does not cover page corner %v", point)
			}
		}
		title := w073WordBounds(t, pdf, "BACKGROUNDTITLE")
		subtitle := w073WordBounds(t, pdf, "BACKGROUNDSUBTITLE")
		margin := 25 * 72 / 25.4
		wantTitleTop := margin + (float64(page.Bounds().Dy())-2*margin)*0.4
		if absoluteFloatW073(title.yMin-wantTitleTop) > 2*72/25.4 {
			t.Errorf("literal 40%% title top=%.2fpt, want %.2fpt within 2mm", title.yMin, wantTitleTop)
		}
		wantGap := 4 * 72 / 25.4
		if absoluteFloatW073((subtitle.yMin-title.yMax)-wantGap) > 2*72/25.4 {
			t.Errorf("subtitle visible gap=%.2fpt, want %.2fpt within 2mm", subtitle.yMin-title.yMax, wantGap)
		}
		washPoint := image.Pt(80, int((title.yMax+subtitle.yMin)/2))
		washPixel := color.RGBAModel.Convert(page.At(washPoint.X, washPoint.Y)).(color.RGBA)
		if washPixel.R > 120 || washPixel.G > 40 || washPixel.B > 90 {
			t.Errorf("title wash is not visibly dark at %v: got %#v", washPoint, washPixel)
		}
		date := w073WordBounds(t, pdf, "BACKGROUNDDATE")
		bottomWashPoint := image.Pt(180, int(date.yMin))
		bottomWashPixel := color.RGBAModel.Convert(page.At(bottomWashPoint.X, bottomWashPoint.Y)).(color.RGBA)
		if bottomWashPixel.R > 120 || bottomWashPixel.G > 40 || bottomWashPixel.B > 90 {
			t.Errorf("bottom metadata wash is not visibly dark at %v: got %#v", bottomWashPoint, bottomWashPixel)
		}
	})

	t.Run("landscape inset with solid backdrop", func(t *testing.T) {
		dir := t.TempDir()
		source, target := filepath.Join(dir, "play.org"), filepath.Join(dir, "play.docx")
		writeW073PNG(t, filepath.Join(dir, "cover.png"), 400, 200, color.RGBA{R: 230, G: 20, B: 150, A: 255})
		writeAppFile(t, source, "#+TITLE: SOLIDTITLE\n#+AUTHOR: SOLIDAUTHOR\n#+DATE: SOLIDDATE\n#+VERSION: SOLIDVERSION\n* ACT ONE\n**** CÁIT\nBody.\n")
		writeAppFile(t, filepath.Join(dir, "script.yaml"), `folio:
  title-page:
    author:
      position: bottom-center
  cover-image:
    path: cover.png
    mode: inset
    placement: dodge
    size: 80mm
    dodge:
      padding: 5mm
    backdrop:
      mode: solid
      solid:
        color: "#2a2524"
      wash:
        treatment: dark
        opacity: 0%
`)
		_, stderr, err := officeRun(t, binary, dir, officeEnv(t, t.TempDir(), nil), "convert", source, target)
		if err != nil {
			t.Fatalf("convert: %v: %s", err, stderr)
		}
		pdf := proseRenderPDF(t, target)
		page := rasterW073Page(t, pdf, 1)
		box, ok := w073ColourBounds(page)
		if !ok || absW073(box.Dx()-227) > 6 || absW073(box.Dy()-113) > 6 {
			t.Fatalf("landscape inset=%v found=%t, want approximately 80mm x 40mm", box, ok)
		}
		for _, point := range []image.Point{{4, 4}, {page.Bounds().Max.X - 5, page.Bounds().Max.Y - 5}} {
			if !w079NearRGB(page.At(point.X, point.Y), 0x2a, 0x25, 0x24, 5) {
				t.Errorf("solid backdrop at %v=%v", point, page.At(point.X, point.Y))
			}
		}
		for _, marker := range []string{"SOLIDTITLE", "SOLIDAUTHOR", "SOLIDDATE", "SOLIDVERSION"} {
			if !strings.Contains(w073PDFText(t, pdf), marker) {
				t.Errorf("rendered editable text lacks %s", marker)
			}
		}
		date := w073WordBounds(t, pdf, "SOLIDDATE")
		author := w073WordBounds(t, pdf, "SOLIDAUTHOR")
		version := w073WordBounds(t, pdf, "SOLIDVERSION")
		if date.xMax >= author.xMin || author.xMax >= version.xMin || absoluteFloatW073(date.yMin-author.yMin) > 2*72/25.4 || absoluteFloatW073(author.yMin-version.yMin) > 2*72/25.4 {
			t.Errorf("bottom metadata is not one left-centre-right row: date=%#v author=%#v version=%#v", date, author, version)
		}
	})

	t.Run("portrait inset with prepared image backdrop", func(t *testing.T) {
		dir := t.TempDir()
		source, target := filepath.Join(dir, "play.org"), filepath.Join(dir, "play.docx")
		writeW079PatternPNG(t, filepath.Join(dir, "cover.png"), 300, 600)
		writeAppFile(t, source, `#+TITLE: IMAGETITLE
#+SUBTITLE: IMAGESUBTITLE
#+AUTHOR: IMAGEAUTHOR
#+DATE: IMAGEDATE
#+VERSION: IMAGEVERSION
#+FOLIO.TITLE-PAGE.TITLE.POSITION: 40%
#+FOLIO.COVER-IMAGE.PATH: cover.png
#+FOLIO.COVER-IMAGE.MODE: inset
#+FOLIO.COVER-IMAGE.PLACEMENT: dodge
#+FOLIO.COVER-IMAGE.SIZE: 80mm
#+FOLIO.COVER-IMAGE.DODGE.PADDING: 5mm
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
			t.Fatalf("convert: %v: %s", err, stderr)
		}
		if !strings.Contains(string(w074PackageMember(t, target, "[Content_Types].xml")), "image/jpeg") {
			t.Fatal("prepared JPEG backdrop content type absent")
		}
		pdf := proseRenderPDF(t, target)
		page := rasterW073Page(t, pdf, 1)
		bounds := page.Bounds()
		for _, point := range []image.Point{{5, 5}, {bounds.Max.X - 6, 5}, {5, bounds.Max.Y - 6}, {bounds.Max.X - 6, bounds.Max.Y - 6}} {
			r, g, b, _ := page.At(point.X, point.Y).RGBA()
			if r>>8 > 248 && g>>8 > 248 && b>>8 > 248 {
				t.Errorf("image backdrop left white page edge at %v", point)
			}
		}
		upper := w073WordBounds(t, pdf, "IMAGETITLE")
		author := w073WordBounds(t, pdf, "IMAGEAUTHOR")
		lower := w073WordBounds(t, pdf, "IMAGEVERSION")
		imageBox, ok := w073ColourBounds(page)
		if !ok || upper.yMax >= float64(imageBox.Min.Y) || author.yMin <= float64(imageBox.Max.Y) || lower.yMin <= author.yMax {
			t.Fatalf("dodge overlap or grouping error image=%v found=%t upper=%#v author=%#v lower=%#v", imageBox, ok, upper, author, lower)
		}
	})
}

func TestAT074_2DOCXPublicationIsAtomic(t *testing.T) {
	binary := officeBuild(t)
	for _, tc := range []struct {
		name, path, config string
	}{
		{"missing image", "folio.cover-image.path", "path: missing.png\n    mode: background\n"},
		{"invalid solid colour", "folio.cover-image.backdrop.solid.color", "path: cover.png\n    mode: inset\n    size: 80mm\n    backdrop:\n      mode: solid\n      solid:\n        color: '#abc'\n      wash:\n        treatment: light\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			source, target := filepath.Join(dir, "play.org"), filepath.Join(dir, "play.docx")
			writeAppFile(t, source, "#+TITLE: Atomic\n* ACT ONE\n**** CÁIT\nBody.\n")
			writeW073PNG(t, filepath.Join(dir, "cover.png"), 400, 200, color.RGBA{R: 230, G: 20, B: 150, A: 255})
			writeAppFile(t, filepath.Join(dir, "script.yaml"), "folio:\n  cover-image:\n    "+tc.config)
			const sentinel = "W074 sentinel target\n"
			writeAppFile(t, target, sentinel)
			_, stderr, err := officeRun(t, binary, dir, officeEnv(t, t.TempDir(), nil), "convert", source, target)
			if err == nil || !strings.Contains(stderr, tc.path) {
				t.Fatalf("error=%v stderr=%q, want %s", err, stderr, tc.path)
			}
			raw, readErr := os.ReadFile(target)
			if readErr != nil || string(raw) != sentinel {
				t.Fatalf("target changed: %q / %v", raw, readErr)
			}
		})
	}

	t.Run("no-image control", func(t *testing.T) {
		dir := t.TempDir()
		source, target := filepath.Join(dir, "play.org"), filepath.Join(dir, "play.docx")
		writeAppFile(t, source, "#+TITLE: CONTROLTITLE\n* ACT ONE\n**** CÁIT\nBody.\n")
		_, stderr, err := officeRun(t, binary, dir, officeEnv(t, t.TempDir(), nil), "convert", source, target)
		if err != nil {
			t.Fatalf("no-image control: %v: %s", err, stderr)
		}
		document := string(w074PackageMember(t, target, "word/document.xml"))
		if strings.Contains(document, "wp:anchor") || !strings.Contains(document, "CONTROLTITLE") {
			t.Fatalf("no-image title structure changed")
		}
	})
}
