// ABOUTME: Verifies frontmatter paragraph font inheritance and isolation.
// ABOUTME: Exercises the public loader and all six validated font properties.
package config

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestFrontmatterParagraphFont(t *testing.T) {
	const path = "folio.positioning.frontmatter.paragraph.font"
	for _, style := range []string{"british", "us", "screenplay"} {
		t.Run(style, func(t *testing.T) {
			home, dir := t.TempDir(), t.TempDir()
			load := func() Config {
				c, e := Load(Options{Home: home, LocalDir: dir, CLI: map[string]any{"style": style}})
				if e != nil {
					t.Fatal(e)
				}
				return c
			}
			baseline := load()
			root, e := baseline.Font("folio.font")
			if e != nil {
				t.Fatal(e)
			}
			f, e := baseline.Font(path)
			if e != nil || f != root {
				t.Fatalf("default paragraph = %+v, %v; want %+v", f, e, root)
			}
			writeYAML(t, filepath.Join(home, ".config/first-folio/script.yaml"), "folio:\n  positioning:\n    frontmatter:\n      paragraph:\n        font:\n          style: italic\n")
			writeYAML(t, filepath.Join(dir, "script.yaml"), "folio:\n  font:\n    family: Inherited Family\n    weight: bold\n    stretch: 110%\n    letter-spacing: 0.1pt\n  positioning:\n    frontmatter:\n      paragraph:\n        font:\n          size: 9pt\n")
			c := load()
			f, e = c.Font(path)
			want := Font{Family: "Inherited Family", Size: "9pt", Weight: "bold", Stretch: "110%", Style: "italic", LetterSpacing: "0.1pt"}
			if e != nil || f != want {
				t.Fatalf("partial/layered paragraph = %+v, %v; want %+v", f, e, want)
			}
			for _, other := range []string{"folio.positioning.speech.dialogue.font", "folio.positioning.frontmatter.header.font"} {
				before, _ := baseline.Font(other)
				after, _ := c.Font(other)
				if before != after {
					t.Fatalf("changed unrelated role %s", other)
				}
			}
			root, _ = c.Font("folio.font")
			if root.Size == "9pt" || root.Style == "italic" {
				t.Fatal("paragraph override leaked into root")
			}
		})
	}
	for property, value := range map[string]string{"family": "''", "size": "0pt", "weight": "heavy", "stretch": "-5%", "style": "bold", "letter-spacing": "1px", "unknown": "value"} {
		t.Run("invalid/"+property, func(t *testing.T) {
			dir := t.TempDir()
			writeYAML(t, filepath.Join(dir, "script.yaml"), "folio:\n  positioning:\n    frontmatter:\n      paragraph:\n        font:\n          "+property+": "+value+"\n")
			_, err := Load(Options{Home: t.TempDir(), LocalDir: dir})
			if err == nil || !strings.Contains(err.Error(), path+"."+property) {
				t.Fatalf("invalid property not rejected with path: %v", err)
			}
		})
	}
}
