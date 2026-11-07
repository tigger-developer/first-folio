// ABOUTME: Defines W079 strict backdrop configuration and source-projection regressions.
// ABOUTME: Exercises mode switching after cascade merge without changing existing cover defaults.
package config

import (
	"strings"
	"testing"
)

func TestRT079_1BackdropConfigurationResolvesModesDefaultsAndInactiveBranches(t *testing.T) {
	cases := []struct {
		name     string
		backdrop map[string]any
		want     CoverImage
	}{
		{
			name: "image defaults retain inactive solid branch",
			backdrop: map[string]any{
				"mode":  "image",
				"solid": map[string]any{"color": "#2a2524"},
				"wash":  map[string]any{"treatment": "dark"},
			},
			want: CoverImage{Path: "cover.png", Mode: "inset", Placement: "overlay", Size: "80mm", ImageTreatment: "light", ImageOpacity: 1, DodgePadding: "1em", OverlayBlur: "0mm", OverlayWashTreatment: "light", OverlayPadding: "0em", BackdropMode: "image", BackdropImageBlur: "8mm", BackdropSolidColor: "#2a2524", BackdropWashTreatment: "dark"},
		},
		{
			name: "solid accepts lower image branch",
			backdrop: map[string]any{
				"mode":  "solid",
				"image": map[string]any{"blur": "12mm"},
				"solid": map[string]any{"color": "#A0b1C2"},
				"wash":  map[string]any{"treatment": "light", "opacity": "37%"},
			},
			want: CoverImage{Path: "cover.png", Mode: "inset", Placement: "overlay", Size: "80mm", ImageTreatment: "light", ImageOpacity: 1, DodgePadding: "1em", OverlayBlur: "0mm", OverlayWashTreatment: "light", OverlayPadding: "0em", BackdropMode: "solid", BackdropImageBlur: "12mm", BackdropSolidColor: "#A0b1C2", BackdropWashTreatment: "light", BackdropWashOpacity: 0.37},
		},
		{
			name: "none disables populated lower branches",
			backdrop: map[string]any{
				"mode":  "none",
				"image": map[string]any{"blur": "not-a-length"},
				"solid": map[string]any{"color": "not-a-colour"},
				"wash":  map[string]any{"treatment": "unused", "opacity": "not-a-percentage"},
			},
			want: CoverImage{Path: "cover.png", Mode: "inset", Placement: "overlay", Size: "80mm", ImageTreatment: "light", ImageOpacity: 1, DodgePadding: "1em", OverlayBlur: "0mm", OverlayWashTreatment: "light", OverlayPadding: "0em", BackdropMode: "none", BackdropImageBlur: "not-a-length", BackdropSolidColor: "not-a-colour", BackdropWashTreatment: "unused"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := Load(Options{Mode: ModeScript, Home: t.TempDir(), LocalDir: t.TempDir(), Source: map[string]any{"folio": map[string]any{"cover-image": map[string]any{
				"path": "cover.png", "mode": "inset", "size": "80mm", "backdrop": tc.backdrop,
			}}}})
			if err != nil {
				t.Fatalf("loading backdrop: %v", err)
			}
			got, enabled, err := cfg.CoverImage()
			if err != nil || !enabled || got != tc.want {
				t.Fatalf("CoverImage() = %#v, %t, %v; want %#v, true, nil", got, enabled, err, tc.want)
			}
		})
	}

	cfg, err := Load(Options{Mode: ModeScript, Home: t.TempDir(), LocalDir: t.TempDir(), Source: map[string]any{"folio": map[string]any{"cover-image": map[string]any{
		"path": "cover.png", "mode": "inset", "size": "80mm",
	}}}})
	if err != nil {
		t.Fatal(err)
	}
	got, enabled, err := cfg.CoverImage()
	if err != nil || !enabled || got.BackdropMode != "" {
		t.Fatalf("absent backdrop changed cover plan: %#v, %t, %v", got, enabled, err)
	}
}

func TestRT079_1BackdropConfigurationRejectsActiveInvalidAndUnknownValues(t *testing.T) {
	cases := []struct {
		name, path string
		mode       string
		backdrop   map[string]any
		coverMode  string
	}{
		{"unknown backdrop key", "folio.cover-image.backdrop.typo", "image", map[string]any{"typo": true, "wash": map[string]any{"treatment": "light"}}, "inset"},
		{"unknown image key", "folio.cover-image.backdrop.image.typo", "image", map[string]any{"image": map[string]any{"typo": true}, "wash": map[string]any{"treatment": "light"}}, "inset"},
		{"unknown solid key", "folio.cover-image.backdrop.solid.typo", "solid", map[string]any{"solid": map[string]any{"color": "#000000", "typo": true}, "wash": map[string]any{"treatment": "light"}}, "inset"},
		{"unknown wash key", "folio.cover-image.backdrop.wash.typo", "solid", map[string]any{"solid": map[string]any{"color": "#000000"}, "wash": map[string]any{"treatment": "light", "typo": true}}, "inset"},
		{"missing mode", "folio.cover-image.backdrop.mode", "", map[string]any{"wash": map[string]any{"treatment": "light"}}, "inset"},
		{"unsupported mode", "folio.cover-image.backdrop.mode", "gradient", map[string]any{"wash": map[string]any{"treatment": "light"}}, "inset"},
		{"background cover", "folio.cover-image.backdrop", "solid", map[string]any{"solid": map[string]any{"color": "#000000"}, "wash": map[string]any{"treatment": "light"}}, "background"},
		{"image missing treatment", "folio.cover-image.backdrop.wash.treatment", "image", map[string]any{}, "inset"},
		{"image zero blur", "folio.cover-image.backdrop.image.blur", "image", map[string]any{"image": map[string]any{"blur": "0mm"}, "wash": map[string]any{"treatment": "light"}}, "inset"},
		{"image wrong blur unit", "folio.cover-image.backdrop.image.blur", "image", map[string]any{"image": map[string]any{"blur": "2em"}, "wash": map[string]any{"treatment": "light"}}, "inset"},
		{"solid missing colour", "folio.cover-image.backdrop.solid.color", "solid", map[string]any{"wash": map[string]any{"treatment": "light"}}, "inset"},
		{"solid malformed colour", "folio.cover-image.backdrop.solid.color", "solid", map[string]any{"solid": map[string]any{"color": "#abc"}, "wash": map[string]any{"treatment": "light"}}, "inset"},
		{"invalid active treatment", "folio.cover-image.backdrop.wash.treatment", "solid", map[string]any{"solid": map[string]any{"color": "#000000"}, "wash": map[string]any{"treatment": "contrast"}}, "inset"},
		{"invalid active opacity", "folio.cover-image.backdrop.wash.opacity", "solid", map[string]any{"solid": map[string]any{"color": "#000000"}, "wash": map[string]any{"treatment": "light", "opacity": "101%"}}, "inset"},
		{"inactive image must remain a mapping", "folio.cover-image.backdrop.image", "none", map[string]any{"image": []any{"8mm"}}, "inset"},
		{"active image must remain a mapping", "folio.cover-image.backdrop.image", "image", map[string]any{"image": "8mm", "wash": map[string]any{"treatment": "light"}}, "inset"},
		{"active solid must remain a mapping", "folio.cover-image.backdrop.solid", "solid", map[string]any{"solid": "#000000", "wash": map[string]any{"treatment": "light"}}, "inset"},
		{"active wash must remain a mapping", "folio.cover-image.backdrop.wash", "solid", map[string]any{"solid": map[string]any{"color": "#000000"}, "wash": "light"}, "inset"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			backdrop := map[string]any{}
			for key, value := range tc.backdrop {
				backdrop[key] = value
			}
			if tc.mode != "" {
				backdrop["mode"] = tc.mode
			}
			cover := map[string]any{"path": "cover.png", "mode": tc.coverMode, "backdrop": backdrop}
			if tc.coverMode == "inset" {
				cover["size"] = "80mm"
			}
			_, err := Load(Options{Mode: ModeScript, Home: t.TempDir(), LocalDir: t.TempDir(), Source: map[string]any{"folio": map[string]any{"cover-image": cover}}})
			if err == nil || !strings.Contains(err.Error(), tc.path) {
				t.Fatalf("expected %s validation error, got %v", tc.path, err)
			}
		})
	}
	_, err := Load(Options{Mode: ModeScript, Home: t.TempDir(), LocalDir: t.TempDir(), Source: map[string]any{"folio": map[string]any{"cover-image": map[string]any{
		"path": "cover.png", "mode": "inset", "size": "80mm", "backdrop": []any{"solid"},
	}}}})
	if err == nil || !strings.Contains(err.Error(), "folio.cover-image.backdrop") {
		t.Fatalf("expected backdrop mapping error, got %v", err)
	}
}

func TestRT079_1HigherLayerNoneDisablesLowerBackdropWithoutRestatingBranches(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir+"/script.yaml", `folio:
  cover-image:
    path: cover.png
    mode: inset
    size: 80mm
    backdrop:
      mode: image
      image:
        blur: 12mm
      wash:
        treatment: dark
        opacity: 25%
`)
	cfg, err := Load(Options{Mode: ModeScript, Home: t.TempDir(), LocalDir: dir, Source: map[string]any{"folio": map[string]any{"cover-image": map[string]any{
		"backdrop": map[string]any{"mode": "none"},
	}}}})
	if err != nil {
		t.Fatalf("loading cascade disablement: %v", err)
	}
	cover, enabled, err := cfg.CoverImage()
	if err != nil || !enabled {
		t.Fatalf("CoverImage() = %#v, %t, %v", cover, enabled, err)
	}
	if cover.BackdropMode != "none" || cover.BackdropImageBlur != "12mm" || cover.BackdropWashTreatment != "dark" || cover.BackdropWashOpacity != 0.25 {
		t.Fatalf("cascade disablement lost or activated lower branches: %#v", cover)
	}
}

func TestRT079_1BackdropLeavesAreAvailableToOrgSourceConfiguration(t *testing.T) {
	paths, err := ScriptSourcePaths()
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{
		"folio.cover-image.backdrop.mode",
		"folio.cover-image.backdrop.image.blur",
		"folio.cover-image.backdrop.solid.color",
		"folio.cover-image.backdrop.wash.treatment",
		"folio.cover-image.backdrop.wash.opacity",
	} {
		if _, ok := paths[path]; !ok {
			t.Errorf("%s is absent from Org source configuration paths", path)
		}
	}
}
