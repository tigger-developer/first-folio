// ABOUTME: Defines W073 cover-image configuration validation and strict-key regression cases.
// ABOUTME: Protects the new block without making established configuration namespaces strict.
package config

import (
	"strings"
	"testing"
)

func TestRT073_1CoverImageConfigurationKeysAreValidated(t *testing.T) {
	valid := map[string]any{"folio": map[string]any{"cover-image": map[string]any{
		"path": "cover.png", "mode": "inset", "placement": "overlay", "size": "80mm", "superscale": "4x",
		"image":   map[string]any{"treatment": "dark", "opacity": "75%"},
		"overlay": map[string]any{"blur": "2mm", "padding": "1em", "wash": map[string]any{"treatment": "light", "opacity": "40%"}},
	}}}
	if _, err := Load(Options{Mode: ModeScript, Home: t.TempDir(), LocalDir: t.TempDir(), Source: valid}); err != nil {
		t.Fatalf("valid cover-image configuration: %v", err)
	}

	invalid := []struct {
		name, path string
		cover      map[string]any
	}{
		{"unknown direct key", "folio.cover-image.typo", map[string]any{"path": "cover.png", "mode": "background", "typo": true}},
		{"unknown image key", "folio.cover-image.image.typo", map[string]any{"path": "cover.png", "mode": "background", "image": map[string]any{"typo": true}}},
		{"unknown dodge key", "folio.cover-image.dodge.typo", map[string]any{"path": "cover.png", "mode": "inset", "placement": "dodge", "size": "80mm", "dodge": map[string]any{"typo": true}}},
		{"unknown overlay key", "folio.cover-image.overlay.typo", map[string]any{"path": "cover.png", "mode": "background", "overlay": map[string]any{"typo": true}}},
		{"unknown wash key", "folio.cover-image.overlay.wash.typo", map[string]any{"path": "cover.png", "mode": "background", "overlay": map[string]any{"wash": map[string]any{"typo": true}}}},
		{"missing mode", "folio.cover-image.mode", map[string]any{"path": "cover.png"}},
		{"invalid mode", "folio.cover-image.mode", map[string]any{"path": "cover.png", "mode": "stretch"}},
		{"background size", "folio.cover-image.size", map[string]any{"path": "cover.png", "mode": "background", "size": "80mm"}},
		{"inset missing size", "folio.cover-image.size", map[string]any{"path": "cover.png", "mode": "inset"}},
		{"background dodge", "folio.cover-image.placement", map[string]any{"path": "cover.png", "mode": "background", "placement": "dodge"}},
		{"invalid opacity", "folio.cover-image.image.opacity", map[string]any{"path": "cover.png", "mode": "background", "image": map[string]any{"opacity": "101%"}}},
		{"negative opacity", "folio.cover-image.image.opacity", map[string]any{"path": "cover.png", "mode": "background", "image": map[string]any{"opacity": "-1%"}}},
		{"invalid image treatment", "folio.cover-image.image.treatment", map[string]any{"path": "cover.png", "mode": "background", "image": map[string]any{"treatment": "auto"}}},
		{"invalid blur unit", "folio.cover-image.overlay.blur", map[string]any{"path": "cover.png", "mode": "background", "overlay": map[string]any{"blur": "1em"}}},
		{"negative blur", "folio.cover-image.overlay.blur", map[string]any{"path": "cover.png", "mode": "background", "overlay": map[string]any{"blur": "-0.1mm"}}},
		{"invalid padding unit", "folio.cover-image.overlay.padding", map[string]any{"path": "cover.png", "mode": "background", "overlay": map[string]any{"padding": "1pt"}}},
		{"negative padding", "folio.cover-image.overlay.padding", map[string]any{"path": "cover.png", "mode": "background", "overlay": map[string]any{"padding": "-0.1em"}}},
		{"invalid dodge padding unit", "folio.cover-image.dodge.padding", map[string]any{"path": "cover.png", "mode": "inset", "placement": "dodge", "size": "80mm", "dodge": map[string]any{"padding": "1pt"}}},
		{"zero dodge padding", "folio.cover-image.dodge.padding", map[string]any{"path": "cover.png", "mode": "inset", "placement": "dodge", "size": "80mm", "dodge": map[string]any{"padding": "0em"}}},
		{"negative dodge padding", "folio.cover-image.dodge.padding", map[string]any{"path": "cover.png", "mode": "inset", "placement": "dodge", "size": "80mm", "dodge": map[string]any{"padding": "-0.1mm"}}},
		{"dodge block with overlay placement", "folio.cover-image.dodge", map[string]any{"path": "cover.png", "mode": "inset", "placement": "overlay", "size": "80mm", "dodge": map[string]any{"padding": "1em"}}},
		{"invalid wash treatment", "folio.cover-image.overlay.wash.treatment", map[string]any{"path": "cover.png", "mode": "background", "overlay": map[string]any{"wash": map[string]any{"treatment": "contrast"}}}},
		{"invalid wash opacity", "folio.cover-image.overlay.wash.opacity", map[string]any{"path": "cover.png", "mode": "background", "overlay": map[string]any{"wash": map[string]any{"opacity": "101%"}}}},
		{"dodge treatment", "folio.cover-image.overlay", map[string]any{"path": "cover.png", "mode": "inset", "placement": "dodge", "size": "80mm", "overlay": map[string]any{"blur": "1mm"}}},
		{"invalid superscale", "folio.cover-image.superscale", map[string]any{"path": "cover.png", "mode": "background", "superscale": "16x"}},
	}
	for _, tc := range invalid {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Load(Options{Mode: ModeScript, Home: t.TempDir(), LocalDir: t.TempDir(), Source: map[string]any{"folio": map[string]any{"cover-image": tc.cover}}})
			if err == nil || !strings.Contains(err.Error(), tc.path) {
				t.Fatalf("expected %s validation error, got %v", tc.path, err)
			}
		})
	}

	if _, err := Load(Options{Mode: ModeScript, Home: t.TempDir(), LocalDir: t.TempDir(), Source: map[string]any{"folio": map[string]any{"unrelated-future-key": true}}}); err != nil {
		t.Fatalf("cover-image strictness leaked outside its block: %v", err)
	}
}

func TestRT073_1CoverImageResolvesTypedDefaultsAndBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name  string
		cover map[string]any
		want  CoverImage
	}{
		{
			name:  "background defaults and zero boundaries",
			cover: map[string]any{"path": "cover.png", "mode": "background", "image": map[string]any{"opacity": "0%"}, "overlay": map[string]any{"blur": "0mm", "padding": "0em", "wash": map[string]any{"opacity": "0%"}}},
			want:  CoverImage{Path: "cover.png", Mode: "background", Placement: "overlay", ImageTreatment: "light", ImageOpacity: 0, Superscale: "", DodgePadding: "1em", OverlayBlur: "0mm", OverlayWashTreatment: "light", OverlayWashOpacity: 0, OverlayPadding: "0em"},
		},
		{
			name: "inset explicit upper boundaries",
			cover: map[string]any{"path": "cover.jpg", "mode": "inset", "placement": "overlay", "size": "80mm", "superscale": "default",
				"image":   map[string]any{"treatment": "dark", "opacity": "100%"},
				"overlay": map[string]any{"blur": "2.5mm", "padding": "1.5mm", "wash": map[string]any{"treatment": "dark", "opacity": "100%"}}},
			want: CoverImage{Path: "cover.jpg", Mode: "inset", Placement: "overlay", Size: "80mm", ImageTreatment: "dark", ImageOpacity: 1, Superscale: "default", DodgePadding: "1em", OverlayBlur: "2.5mm", OverlayWashTreatment: "dark", OverlayWashOpacity: 1, OverlayPadding: "1.5mm"},
		},
		{
			name:  "dodge explicit millimetre padding",
			cover: map[string]any{"path": "cover.jpg", "mode": "inset", "placement": "dodge", "size": "80mm", "dodge": map[string]any{"padding": "3.5mm"}},
			want:  CoverImage{Path: "cover.jpg", Mode: "inset", Placement: "dodge", Size: "80mm", ImageTreatment: "light", ImageOpacity: 1, Superscale: "", DodgePadding: "3.5mm", OverlayBlur: "0mm", OverlayWashTreatment: "light", OverlayWashOpacity: 0, OverlayPadding: "0em"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := Load(Options{Mode: ModeScript, Home: t.TempDir(), LocalDir: t.TempDir(), Source: map[string]any{"folio": map[string]any{"cover-image": tc.cover}}})
			if err != nil {
				t.Fatalf("loading configuration: %v", err)
			}
			got, enabled, err := cfg.CoverImage()
			if err != nil || !enabled || got != tc.want {
				t.Fatalf("CoverImage() = %#v, %t, %v; want %#v, true, nil", got, enabled, err, tc.want)
			}
		})
	}

	cfg, err := Load(Options{Mode: ModeScript, Home: t.TempDir(), LocalDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if got, enabled, err := cfg.CoverImage(); err != nil || enabled || got != (CoverImage{}) {
		t.Fatalf("absent CoverImage() = %#v, %t, %v", got, enabled, err)
	}
}

func TestRT073_1DodgePaddingIsAvailableToOrgSourceConfiguration(t *testing.T) {
	paths, err := ScriptSourcePaths()
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := paths["folio.cover-image.dodge.padding"]; !ok {
		t.Fatal("folio.cover-image.dodge.padding is absent from Org source configuration paths")
	}
	if _, ok := paths["folio.title-page.author.position"]; !ok {
		t.Fatal("folio.title-page.author.position is absent from Org source configuration paths")
	}
}

func TestRT081_2DisabledCoverBypassesActiveFieldValidation(t *testing.T) {
	paths, pathsErr := ScriptSourcePaths()
	if pathsErr != nil {
		t.Fatal(pathsErr)
	}
	if _, ok := paths["folio.cover-image.enabled"]; !ok {
		t.Fatal("folio.cover-image.enabled is absent from Org source configuration paths")
	}
	cfg, err := Load(Options{Mode: ModeScript, Home: t.TempDir(), LocalDir: t.TempDir(), Source: map[string]any{"folio": map[string]any{"cover-image": map[string]any{"enabled": false, "path": 42, "mode": "invalid"}}}})
	if err != nil {
		t.Fatal(err)
	}
	if got, enabled, coverErr := cfg.CoverImage(); coverErr != nil || enabled || got != (CoverImage{}) {
		t.Fatalf("disabled CoverImage() = %#v, %t, %v; want zero, false, nil", got, enabled, coverErr)
	}
}
