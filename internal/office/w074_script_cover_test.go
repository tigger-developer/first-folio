// ABOUTME: Defines retained native DOCX title-image projection and failure regressions.
// ABOUTME: Inspects package media, relationships, anchors, layers and editable title text.
package office

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"regexp"
	"strings"
	"testing"

	"github.com/tigger-developer/first-folio/internal/config"
	"github.com/tigger-developer/first-folio/internal/play"
)

func w074OfficeConfig(t *testing.T) config.Config {
	t.Helper()
	cfg, err := config.Load(config.Options{
		Mode:     config.ModeScript,
		Home:     t.TempDir(),
		LocalDir: t.TempDir(),
		Source: map[string]any{"folio": map[string]any{
			"title-page": map[string]any{
				"title":    map[string]any{"position": "40%"},
				"subtitle": map[string]any{"space-before": "4mm"},
				"author":   map[string]any{"position": "bottom-center"},
			},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	return cfg
}

func w074Document() play.Document {
	return play.Document{
		Metadata: map[string]string{
			"title": "W074 TITLE", "subtitle": "W074 SUBTITLE", "author": "W074 AUTHOR",
			"date": "W074 DATE", "version": "W074 VERSION",
		},
		Events: []play.Event{{Kind: play.EventActHeader, Text: "ACT ONE"}, {Kind: play.EventStageDirection, Text: "BODY MARKER"}},
	}
}

func w074Asset(t *testing.T, format string, width, height int, fill color.RGBA) ScriptCoverAsset {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.SetRGBA(x, y, fill)
		}
	}
	var output bytes.Buffer
	var err error
	if format == "jpeg" {
		err = jpeg.Encode(&output, img, &jpeg.Options{Quality: 85})
	} else {
		err = png.Encode(&output, img)
	}
	if err != nil {
		t.Fatal(err)
	}
	return ScriptCoverAsset{Data: output.Bytes(), Format: format, PixelWidth: uint32(width), PixelHeight: uint32(height)}
}

func w074PackageFiles(t *testing.T, raw []byte) map[string][]byte {
	t.Helper()
	files, err := readPackage(raw, defaultPackageLimits())
	if err != nil {
		t.Fatal(err)
	}
	return files
}

func TestRT074_1NativeDOCXPackageRepresentsSharedImagePlan(t *testing.T) {
	sharp := w074Asset(t, "png", 400, 200, color.RGBA{R: 215, G: 40, B: 120, A: 255})
	blurred := w074Asset(t, "png", 400, 200, color.RGBA{R: 150, G: 120, B: 140, A: 255})
	backdrop := w074Asset(t, "jpeg", 842, 1191, color.RGBA{R: 55, G: 65, B: 75, A: 255})

	cases := []struct {
		name              string
		cover             ScriptCover
		wantMedia         []string
		wantDrawingNames  []string
		wantDocumentToken []string
	}{
		{
			name:              "background overlay",
			cover:             ScriptCover{Config: config.CoverImage{Mode: "background", Placement: "overlay", ImageTreatment: "light", ImageOpacity: 1, OverlayBlur: "2mm", OverlayWashTreatment: "dark", OverlayWashOpacity: 0.35, OverlayPadding: "5mm"}, Sharp: sharp, Blurred: blurred},
			wantMedia:         []string{"word/media/folio-cover-sharp.png", "word/media/folio-cover-blurred.png"},
			wantDrawingNames:  []string{"Folio cover background", "Folio title treatment", "Folio title text"},
			wantDocumentToken: []string{"a:srcRect", "wp:anchor", "relativeFrom=\"page\""},
		},
		{
			name:              "inset dodge with current placement semantics",
			cover:             ScriptCover{Config: config.CoverImage{Mode: "inset", Placement: "dodge", Size: "80mm", ImageTreatment: "light", ImageOpacity: 1, DodgePadding: "5mm"}, Sharp: sharp, RenderedWidthMM: 80, RenderedHeightMM: 40},
			wantMedia:         []string{"word/media/folio-cover-sharp.png"},
			wantDrawingNames:  []string{"Folio cover inset", "Folio upper title text", "Folio bottom metadata text"},
			wantDocumentToken: []string{"W074 TITLE", "W074 SUBTITLE", "W074 AUTHOR", "W074 DATE", "W074 VERSION"},
		},
		{
			name:              "image backdrop",
			cover:             ScriptCover{Config: config.CoverImage{Mode: "inset", Placement: "dodge", Size: "80mm", ImageTreatment: "light", ImageOpacity: 1, DodgePadding: "1em", BackdropMode: "image", BackdropImageBlur: "8mm", BackdropWashTreatment: "dark", BackdropWashOpacity: 0.15}, Sharp: sharp, Backdrop: backdrop, RenderedWidthMM: 80, RenderedHeightMM: 40},
			wantMedia:         []string{"word/media/folio-cover-sharp.png", "word/media/folio-cover-backdrop.jpg"},
			wantDrawingNames:  []string{"Folio backdrop image", "Folio backdrop wash", "Folio cover inset"},
			wantDocumentToken: []string{"FFFFFF", "W074 TITLE", "BODY MARKER"},
		},
		{
			name:              "solid backdrop",
			cover:             ScriptCover{Config: config.CoverImage{Mode: "inset", Placement: "overlay", Size: "80mm", ImageTreatment: "light", ImageOpacity: 1, OverlayBlur: "0mm", OverlayWashTreatment: "light", OverlayPadding: "0em", BackdropMode: "solid", BackdropSolidColor: "#2a2524", BackdropWashTreatment: "light", BackdropWashOpacity: 0.25}, Sharp: sharp, RenderedWidthMM: 80, RenderedHeightMM: 40},
			wantMedia:         []string{"word/media/folio-cover-sharp.png"},
			wantDrawingNames:  []string{"Folio backdrop solid", "Folio backdrop wash", "Folio cover inset"},
			wantDocumentToken: []string{"2A2524", "000000", "W074 TITLE"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			raw, warnings, err := ExportScriptWithCover(w074Document(), w074OfficeConfig(t), play.FormatOrg, "docx", &tc.cover)
			if err != nil {
				t.Fatalf("ExportScriptWithCover(): %v", err)
			}
			if len(warnings) != 0 {
				t.Fatalf("unexpected warnings: %v", warnings)
			}
			if err := validatePackage(raw, "docx", defaultPackageLimits()); err != nil {
				t.Fatalf("generated package: %v", err)
			}
			files := w074PackageFiles(t, raw)
			for _, name := range tc.wantMedia {
				if len(files[name]) == 0 {
					t.Errorf("missing embedded media %s", name)
				}
			}
			document := string(files["word/document.xml"])
			rels := string(files["word/_rels/document.xml.rels"])
			for _, name := range tc.wantDrawingNames {
				if !strings.Contains(document, name) {
					t.Errorf("missing drawing name %q", name)
				}
			}
			for _, token := range tc.wantDocumentToken {
				if !strings.Contains(document, token) {
					t.Errorf("missing document token %q", token)
				}
			}
			for _, media := range tc.wantMedia {
				target := strings.TrimPrefix(media, "word/")
				if !strings.Contains(rels, "Target=\""+target+"\"") {
					t.Errorf("missing relationship target %s", target)
				}
			}
			ids := regexp.MustCompile(`<wp:docPr[^>]+id="([0-9]+)"`).FindAllStringSubmatch(document, -1)
			seen := map[string]bool{}
			for _, match := range ids {
				if seen[match[1]] {
					t.Errorf("duplicate drawing id %s", match[1])
				}
				seen[match[1]] = true
			}
			if strings.Contains(string(raw), "source/cover") {
				t.Error("source path leaked into DOCX package")
			}
		})
	}
}

func TestRT074_2DOCXPackagingRejectsInvalidImageStateSafely(t *testing.T) {
	sharp := w074Asset(t, "png", 400, 200, color.RGBA{R: 215, G: 40, B: 120, A: 255})
	cases := []struct {
		name, want string
		cover      ScriptCover
	}{
		{"missing sharp asset", "folio.cover-image.path", ScriptCover{Config: config.CoverImage{Mode: "background"}}},
		{"unsupported sharp format", "format", ScriptCover{Config: config.CoverImage{Mode: "background"}, Sharp: ScriptCoverAsset{Data: []byte("gif"), Format: "gif", PixelWidth: 1, PixelHeight: 1}}},
		{"zero sharp dimensions", "dimension", ScriptCover{Config: config.CoverImage{Mode: "background"}, Sharp: ScriptCoverAsset{Data: sharp.Data, Format: "png"}}},
		{"inset geometry missing", "geometry", ScriptCover{Config: config.CoverImage{Mode: "inset", Size: "80mm"}, Sharp: sharp}},
		{"image backdrop missing", "backdrop", ScriptCover{Config: config.CoverImage{Mode: "inset", Size: "80mm", BackdropMode: "image"}, Sharp: sharp, RenderedWidthMM: 80, RenderedHeightMM: 40}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			raw, _, err := ExportScriptWithCover(w074Document(), w074OfficeConfig(t), play.FormatOrg, "docx", &tc.cover)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error=%v, want %q", err, tc.want)
			}
			if len(raw) != 0 {
				t.Fatalf("invalid plan returned %d publishable bytes", len(raw))
			}
		})
	}

	raw, _, err := ExportScript(w074Document(), w074OfficeConfig(t), play.FormatOrg, "docx")
	if err != nil {
		t.Fatalf("no-image control: %v", err)
	}
	files := w074PackageFiles(t, raw)
	if strings.Contains(string(files["word/document.xml"]), "wp:anchor") {
		t.Fatal("no-image control contains a cover drawing")
	}
	for name := range files {
		if strings.HasPrefix(name, "word/media/") {
			t.Fatalf("no-image control contains %s", name)
		}
	}
}
