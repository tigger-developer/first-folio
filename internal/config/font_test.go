// ABOUTME: Regression tests for W039 unified font configuration.
// ABOUTME: Protects the exhaustive role inventory and same-path layer merging.
package config

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	folio "github.com/tigger-developer/first-folio"
	"gopkg.in/yaml.v3"
)

var fontRolePaths = []string{
	"folio.font",
	"folio.heading.font",
	"folio.title-page.title.font",
	"folio.title-page.subtitle.font",
	"folio.title-page.author.font",
	"folio.title-page.date.font",
	"folio.title-page.version.font",
	"folio.positioning.speech.speaker.font",
	"folio.positioning.speech.speech-instruction.font",
	"folio.positioning.speech.dialogue.font",
	"folio.positioning.stage-direction.font",
	"folio.positioning.transition.font",
	"folio.positioning.frontmatter.header.font",
	"folio.positioning.act-header.font",
	"folio.positioning.scene-header.font",
	"folio.letter.font",
	"folio.manuscript.font",
	"folio.manuscript.heading.font",
	"folio.manuscript.mono.font",
	"folio.manuscript.quoted-block.font",
	"folio.manuscript.title-page.title.font",
	"folio.manuscript.title-page.subtitle.font",
	"folio.manuscript.title-page.author.font",
	"folio.manuscript.title-page.date.font",
	"folio.manuscript.title-page.wordcount.font",
	"folio.manuscript.title-page.version.font",
	"folio.manuscript.title-page.contact.font",
	"folio.manuscript.page-header.font",
	"folio.manuscript.page-footer.font",
	"folio.manuscript.toc.font",
	"folio.manuscript.toc.heading.font",
	"folio.manuscript.copyright.font",
	"folio.manuscript.copyright.label.font",
}

var fontPropertyNames = []string{
	"family", "size", "weight", "stretch", "style", "letter-spacing",
}

// RT039.1: every public font role has the same complete six-property block.
func TestRT039_1EveryPublicFontRoleUsesSixPropertyBlock(t *testing.T) {
	for _, mode := range []Mode{ModeScript, ModeLetter, ModeManuscript} {
		cfg, err := Load(Options{Mode: mode, Home: t.TempDir(), LocalDir: t.TempDir()})
		if err != nil {
			t.Fatalf("Load(%s): %v", mode, err)
		}
		for _, path := range fontRolePaths {
			value, ok := cfg.Get(path)
			if !ok {
				t.Errorf("Load(%s) does not define %s", mode, path)
				continue
			}
			block, ok := value.(map[string]any)
			if !ok {
				t.Errorf("Load(%s) %s is %T, want mapping", mode, path, value)
				continue
			}
			if len(block) != len(fontPropertyNames) {
				t.Errorf("Load(%s) %s has %d properties, want %d: %#v", mode, path, len(block), len(fontPropertyNames), block)
			}
			for _, property := range fontPropertyNames {
				if value, found := block[property]; !found || value == nil || value == "" {
					t.Errorf("Load(%s) %s.%s is not populated", mode, path, property)
				}
			}
		}
	}
}

func TestFontReturnsCompleteValidatedValue(t *testing.T) {
	cfg, err := Load(Options{Mode: ModeScript, Home: t.TempDir(), LocalDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	font, err := cfg.Font("folio.font")
	if err != nil {
		t.Fatal(err)
	}
	want := Font{
		Family: "Libertinus Serif", Size: "12pt", Weight: "regular",
		Stretch: "100%", Style: "regular", LetterSpacing: "0em",
	}
	if font != want {
		t.Fatalf("Font(folio.font) = %#v, want %#v", font, want)
	}

	for _, test := range []struct {
		name string
		cfg  Config
		path string
	}{
		{"missing", Config{data: map[string]any{}}, "folio.font"},
		{"not mapping", Config{data: map[string]any{"folio": map[string]any{"font": "Serif"}}}, "folio.font"},
		{"not normalized", Config{data: map[string]any{"folio": map[string]any{"font": map[string]any{"family": 42}}}}, "folio.font"},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, err := test.cfg.Font(test.path); err == nil || !strings.Contains(err.Error(), test.path) {
				t.Fatalf("Font(%s) error = %v, want full path", test.path, err)
			}
		})
	}
}

// RT039.1: every path/property combination survives the public configuration merge independently.
func TestRT039_1EveryPublicFontPropertyCanBeConfiguredIndependently(t *testing.T) {
	distinctive := map[string]string{
		"family": "W039 Distinctive", "size": "13.5pt", "weight": "650",
		"stretch": "117.5%", "style": "oblique", "letter-spacing": "-0.03em",
	}
	for _, path := range fontRolePaths {
		mode := modeForFontPath(path)
		for _, property := range fontPropertyNames {
			t.Run(strings.TrimPrefix(path, "folio.")+"/"+property, func(t *testing.T) {
				home := t.TempDir()
				project := t.TempDir()
				baseline, err := Load(Options{Mode: mode, Home: home, LocalDir: project})
				if err != nil {
					t.Fatal(err)
				}
				otherRoles := map[string]any{}
				for _, otherPath := range fontRolePaths {
					if otherPath == path {
						continue
					}
					otherRoles[otherPath], _ = baseline.Get(otherPath)
				}

				overlay := map[string]any{}
				setPath(overlay, path+"."+property, distinctive[property])
				raw, err := yaml.Marshal(overlay)
				if err != nil {
					t.Fatal(err)
				}
				writeYAML(t, filepath.Join(project, "script.yaml"), string(raw))

				configured, err := Load(Options{Mode: mode, Home: home, LocalDir: project})
				if err != nil {
					t.Fatal(err)
				}
				assertValue(t, configured, path+"."+property, distinctive[property])
				for otherPath, before := range otherRoles {
					after, _ := configured.Get(otherPath)
					if !reflect.DeepEqual(after, before) {
						t.Fatalf("changing %s.%s altered other role %s", path, property, otherPath)
					}
				}
			})
		}
	}
}

// RT039.2: the signed definition file is the installed base and every mode starts from it.
func TestRT039_2BritishBaseIsNormativeSharedRuntimeBase(t *testing.T) {
	attachedRaw, err := os.ReadFile("../../specs/039-unified-font-configuration/british.yaml")
	if err != nil {
		t.Fatal(err)
	}
	installedRaw, err := folio.Assets.ReadFile("presets/british.yaml")
	if err != nil {
		t.Fatal(err)
	}
	attached, err := parseYAML("attached british.yaml", attachedRaw)
	if err != nil {
		t.Fatal(err)
	}
	installed, err := parseYAML("installed british.yaml", installedRaw)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(installed, attached) {
		t.Fatal("installed presets/british.yaml differs from the signed normative YAML")
	}
	for _, mode := range []Mode{ModeScript, ModeLetter, ModeManuscript} {
		home := t.TempDir()
		project := t.TempDir()
		plain, err := Load(Options{Mode: mode, Home: home, LocalDir: project})
		if err != nil {
			t.Fatal(err)
		}
		writeYAML(t, filepath.Join(project, "script.yaml"), string(installedRaw))
		copied, err := Load(Options{Mode: mode, Home: home, LocalDir: project})
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(plain.data, installed) || !reflect.DeepEqual(copied.data, installed) {
			t.Errorf("Load(%s) does not resolve to the shared British base", mode)
		}
	}
}

// RT039.3: partial overrides merge only against the same role in lower layers.
func TestRT039_3PartialFontBlockUsesOnlySameRoleLowerLayer(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	writeYAML(t, filepath.Join(project, "script.yaml"), `folio:
  manuscript:
    font:
      family: Body Family Must Not Leak
    page-header:
      font:
        stretch: 125%
`)
	cfg, err := Load(Options{Mode: ModeManuscript, Home: home, LocalDir: project})
	if err != nil {
		t.Fatal(err)
	}
	assertValue(t, cfg, "folio.manuscript.page-header.font.family", "Libertinus Sans")
	assertValue(t, cfg, "folio.manuscript.page-header.font.stretch", "125%")
	assertValue(t, cfg, "folio.manuscript.font.family", "Body Family Must Not Leak")
}

// RT039.4: shipped overlays contain no leaf that repeats the effective British value.
func TestRT039_4StyleOverlaysContainOnlyDifferences(t *testing.T) {
	baseRaw, err := folio.Assets.ReadFile("presets/british.yaml")
	if err != nil {
		t.Fatal(err)
	}
	base, err := parseYAML("presets/british.yaml", baseRaw)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"presets/us.yaml", "presets/screenplay.yaml"} {
		raw, err := folio.Assets.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		overlay, err := parseYAML(name, raw)
		if err != nil {
			t.Fatal(err)
		}
		for path, value := range leafValues(overlay, "") {
			if current, ok := (Config{data: base}).Get(path); ok && reflect.DeepEqual(current, value) {
				t.Errorf("%s repeats British value at %s: %#v", name, path, value)
			}
		}
	}
}

// RT039.5: malformed blocks and retired properties fail with the full offending path.
func TestRT039_5RejectsInvalidFontConfiguration(t *testing.T) {
	tests := []struct {
		name, yaml, path string
	}{
		{"retired scalar", "folio:\n  font: Old Family\n", "folio.font"},
		{"retired size", "folio:\n  font-size: 12pt\n", "folio.font-size"},
		{"retired heading", "folio:\n  heading-font: Old Heading\n", "folio.heading-font"},
		{"retired bool", "folio:\n  positioning:\n    stage-direction:\n      italic: true\n", "folio.positioning.stage-direction.italic"},
		{"unknown property", "folio:\n  font:\n    decoration: underline\n", "folio.font.decoration"},
		{"null property", "folio:\n  font:\n    family: null\n", "folio.font.family"},
		{"empty property", "folio:\n  font:\n    style: \"\"\n", "folio.font.style"},
		{"invalid size", "folio:\n  font:\n    size: 0pt\n", "folio.font.size"},
		{"invalid weight", "folio:\n  font:\n    weight: 900.5\n", "folio.font.weight"},
		{"invalid stretch", "folio:\n  font:\n    stretch: Inf\n", "folio.font.stretch"},
		{"invalid style", "folio:\n  font:\n    style: bold\n", "folio.font.style"},
		{"invalid tracking", "folio:\n  font:\n    letter-spacing: 2px\n", "folio.font.letter-spacing"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			project := t.TempDir()
			writeYAML(t, filepath.Join(project, "script.yaml"), test.yaml)
			_, err := Load(Options{Mode: ModeScript, Home: t.TempDir(), LocalDir: project})
			if err == nil || !strings.Contains(err.Error(), test.path) {
				t.Fatalf("Load error = %v, want path %s", err, test.path)
			}
		})
	}
}

func TestRT039_5FontPropertyDomainBoundaries(t *testing.T) {
	tests := []struct {
		property string
		accepted []any
		rejected []any
	}{
		{"family", []any{" Serif "}, []any{nil, "", "   ", []any{"Serif"}, map[string]any{"name": "Serif"}}},
		{"size", []any{".5pt", "1mm", "2.25cm", "3in", "1em"}, []any{"0pt", "-1pt", "12", "2px", 12}},
		{"weight", []any{"thin", "extralight", "light", "regular", "medium", "semibold", "bold", "extrabold", "black", 100, 900, "643"}, []any{"heavy", 99, 901, 643.5, []any{400}}},
		{"stretch", []any{".5", "117.5%", 1, 117.5}, []any{"0", "-1", "100pt", 0, -1, math.Inf(1), []any{100}}},
		{"style", []any{" regular ", "ITALIC", "oblique"}, []any{"bold", "", nil, []any{"italic"}}},
		{"letter-spacing", []any{"0pt", "-0.03em", "+1mm", ".5cm", "2in"}, []any{"0", "1px", 0, nil, []any{"0em"}}},
	}
	for _, test := range tests {
		for _, value := range test.accepted {
			t.Run(test.property+"/accept/"+fmt.Sprint(value), func(t *testing.T) {
				if _, err := validateFontProperty("folio.font."+test.property, test.property, value); err != nil {
					t.Fatalf("accepted value %q rejected: %v", value, err)
				}
			})
		}
		for _, value := range test.rejected {
			t.Run(test.property+"/reject/"+fmt.Sprint(value), func(t *testing.T) {
				if _, err := validateFontProperty("folio.font."+test.property, test.property, value); err == nil {
					t.Fatalf("rejected value %q accepted", value)
				}
			})
		}
	}
}

func TestRT039_5IncompleteBuiltInFontBlockIsRejected(t *testing.T) {
	block := map[string]any{
		"family": "Serif", "size": "12pt", "weight": "regular",
		"stretch": "100%", "style": "regular",
	}
	err := validateFontBlock("folio.font", block)
	if err == nil || !strings.Contains(err.Error(), "folio.font.letter-spacing") {
		t.Fatalf("validateFontBlock error = %v, want missing full path", err)
	}
}

func modeForFontPath(path string) Mode {
	switch {
	case strings.HasPrefix(path, "folio.manuscript."):
		return ModeManuscript
	case path == "folio.letter.font":
		return ModeLetter
	default:
		return ModeScript
	}
}

func leafValues(node map[string]any, prefix string) map[string]any {
	result := map[string]any{}
	for key, value := range node {
		path := key
		if prefix != "" {
			path = prefix + "." + key
		}
		if child, ok := value.(map[string]any); ok {
			for childPath, childValue := range leafValues(child, path) {
				result[childPath] = childValue
			}
			continue
		}
		result[path] = value
	}
	return result
}

func ExampleFont() {
	fmt.Println(len(FontRolePaths) * len(fontPropertyNames))
	// Output: 198
}
