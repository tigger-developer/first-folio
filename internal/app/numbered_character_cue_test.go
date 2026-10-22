// ABOUTME: Checks numbered Org speaker cues through public script conversion.
// ABOUTME: Keeps numeral suffixes in character names and real directions separate.
package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRT066_1NumberedOrgCharacterCues(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	source := filepath.Join(dir, "play.org")
	writeAppFile(t, source, "* CHARACTERS\n| PARENT 2 | A parent |\n| PARENT 3 | Another parent |\n| PARENT 27 | A third parent |\n| MARY | A neighbour |\n* ACT ONE\n**** PARENT 2\n14 minutes late.\n**** PARENT 3 (to Fiona)\nWhere is she?\n**** PARENT 27\nHere.\n**** MARY (IGNORING the question)\nNot himself anyhow.\n**** PARENT 3 (ASKING the question)\nWhere is he?\n**** JACK softly\nQuiet now.\n")

	for _, format := range []string{"typ", "docx"} {
		t.Run(format, func(t *testing.T) {
			target := filepath.Join(dir, "play."+format)
			status, stdout, stderr := runApp(t, "convert", source, target)
			if status != 0 {
				t.Fatalf("status %d\nstdout:%s\nstderr:%s", status, stdout, stderr)
			}
			if format == "typ" {
				visible := readAppFile(t, target)
				for _, want := range []string{
					`#dialogue("PARENT 2")`,
					`#dialogue("PARENT 3", direction: "to Fiona")`,
					`#dialogue("PARENT 27")`,
					`#dialogue("MARY", direction: "IGNORING the question")`,
					`#dialogue("PARENT 3", direction: "ASKING the question")`,
					`#dialogue("JACK", direction: "softly")`,
				} {
					if !strings.Contains(visible, want) {
						t.Errorf("missing %q in Typst output", want)
					}
				}
				return
			}
			raw, err := os.ReadFile(target)
			if err != nil {
				t.Fatal(err)
			}
			body := officeBody(t, officePackage(t, raw), "docx")
			for _, want := range []string{"PARENT 2:", "PARENT 3:", "PARENT 27:", "MARY:", "(to Fiona)", "(IGNORING the question)", "(ASKING the question)", "14 minutes late.", "Where is she?", "Here.", "Not himself anyhow."} {
				if !strings.Contains(body.text(), want) {
					t.Errorf("missing %q in DOCX body", want)
				}
			}
			for _, unwanted := range []string{"(2)", "(3 (to Fiona)", "(27)", "MARY (IGNORING:", "PARENT 3 (ASKING:"} {
				if strings.Contains(body.text(), unwanted) {
					t.Errorf("spurious direction %q in DOCX body", unwanted)
				}
			}
		})
	}
}
