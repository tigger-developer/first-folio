// ABOUTME: Verifies prose font discovery, permission warnings and oversize rejection at the CLI.
// ABOUTME: Uses temporary metadata variants of a bundled font without installing any faces.
package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRT063_8_OfficeLocalFontPermissions(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		for _, mode := range []string{"manuscript", "letter"} {
			for _, tc := range []struct {
				name, family      string
				rights            uint16
				present, embedded bool
			}{
				{"editable", "Fixture Publish", 8, true, true},
				{"installable", "Fixture Publish", 0, true, true},
				{"no-subsetting", "Fixture Publish", 0x108, true, true},
				{"restricted", "Fixture Publish", 2, true, false},
				{"preview-print-only", "Fixture Publish", 4, true, false},
				{"bitmap-only", "Fixture Publish", 0x208, true, false},
				{"missing", "Folio Missing Prose Font 063", 0, false, false},
			} {
				t.Run(format+"/"+mode+"/"+tc.name, func(t *testing.T) {
					t.Setenv("HOME", t.TempDir())
					fonts := t.TempDir()
					t.Setenv("TYPST_FONT_PATHS", fonts)
					want := officeTestFont(t, tc.rights)
					if tc.present {
						if err := os.WriteFile(filepath.Join(fonts, "custom.ttf"), want, 0600); err != nil {
							t.Fatal(err)
						}
					}
					dir := t.TempDir()
					cfg := "folio:\n  " + mode + ":\n    font:\n      family: " + tc.family + "\n"
					if mode == "manuscript" {
						cfg += "    title-page:\n      enabled: false\n    toc:\n      enabled: false\n"
					}
					proseSource(t, dir, "script.yaml", cfg)
					target := filepath.Join(dir, "output."+format)
					var args []string
					if mode == "manuscript" {
						src := proseSource(t, dir, "source.md", "LOCAL-FONT-MARKER\n")
						args = []string{"manuscript", src, target}
					} else {
						src := proseSource(t, dir, "letters.org", "* Letters :letter:\n** Subject :subject:\nLOCAL-FONT-MARKER\n*** Recipient :to:\n**** Theatre :org:\n")
						args = []string{"letter", src, "--format", format}
						target = filepath.Join(dir, "letter-theatre."+format)
					}
					status, _, stderr := runApp(t, args...)
					if status != 0 {
						t.Fatalf("font-policy export failed: %s", stderr)
					}
					pkg := prosePackage(t, target, format)
					layers := officeEffective(t, pkg, format, officeParagraph(t, officeBody(t, pkg, format), "LOCAL-FONT-MARKER"))
					if format == "docx" {
						officeAssertEffective(t, layers, "rFonts", "ascii", tc.family)
					} else {
						name, ok := officeEffectiveAttr(layers, "text-properties", "font-name")
						matched := false
						for _, member := range []string{"styles.xml", "content.xml"} {
							for _, face := range pkg[member].all("font-face") {
								if face.attr("name") == name && strings.Trim(face.attr("font-family"), "'\"") == tc.family {
									matched = true
								}
							}
						}
						if !ok || !matched {
							t.Error("configured family reference was lost")
						}
					}
					raw, err := os.ReadFile(target)
					if err != nil {
						t.Fatal(err)
					}
					count := 0
					for name, data := range officeFontMembers(t, raw) {
						if format == "odt" && bytes.Equal(data, want) {
							count++
						}
						if format == "docx" && strings.HasSuffix(name, ".odttf") && len(data) == len(want) && bytes.Equal(data[32:], want[32:]) {
							count++
						}
					}
					if (count == 1) != tc.embedded || count > 1 {
						t.Errorf("embedded payloads=%d; permitted=%v", count, tc.embedded)
					}
					if !tc.embedded && (!strings.Contains(stderr, tc.family) || !strings.Contains(stderr, "not embedded")) {
						t.Errorf("missing actionable font warning: %s", stderr)
					}
				})
			}
		}
	}
}

func TestRT063_8_OfficeOversizedFontPreservesTarget(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		t.Run(format, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			fonts := t.TempDir()
			t.Setenv("TYPST_FONT_PATHS", fonts)
			path := filepath.Join(fonts, "oversized.ttf")
			if err := os.WriteFile(path, officeTestFont(t, 8), 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.Truncate(path, (32<<20)+1); err != nil {
				t.Fatal(err)
			}
			dir := t.TempDir()
			proseSource(t, dir, "script.yaml", "folio:\n  manuscript:\n    font:\n      family: Fixture Publish\n")
			src := proseSource(t, dir, "source.md", "FONT-LIMIT-MARKER\n")
			target := proseSource(t, dir, "output."+format, "PRIOR-DESTINATION")
			status, _, stderr := runApp(t, "manuscript", src, target)
			if status == 0 || !strings.Contains(strings.ToLower(stderr), "font") || !strings.Contains(stderr, "32") {
				t.Errorf("oversized selected font must fail explicitly: status=%d stderr=%s", status, stderr)
			}
			raw, err := os.ReadFile(target)
			if err != nil || string(raw) != "PRIOR-DESTINATION" {
				t.Errorf("font overflow replaced target: %v", err)
			}
		})
	}
}
