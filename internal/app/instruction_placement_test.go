// ABOUTME: Measures config-selected speaker instruction placement in actual PDFs.
// ABOUTME: Protects British dialogue-column placement and comma removal in both styles.
package app

import (
	"encoding/xml"
	"math"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestSpeakerInstructionPlacement(t *testing.T) {
	for _, tool := range []string{"typst", "pdftotext"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("requires %s", tool)
		}
	}
	binary := officeBuild(t)
	for _, tc := range []struct{ name, style, override string }{
		{"us-inline", "us", ""}, {"british-column", "british", ""},
		{"us-separate", "us", "folio:\n  positioning:\n    speech:\n      speech-instruction:\n        placement: new-line\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			src, pdf := filepath.Join(dir, "play.org"), filepath.Join(dir, "play.pdf")
			writeAppFile(t, src, "* Act One\n** Scene One\n**** Bob, cheerfully\nJust boiled.\n")
			if tc.override != "" {
				writeAppFile(t, filepath.Join(dir, "script.yaml"), tc.override)
			}
			env := officeEnv(t, dir, map[string]string{"TYPST_IGNORE_SYSTEM_FONTS": "true", "TYPST_FONT_PATHS": dir})
			if _, stderr, err := officeRun(t, binary, dir, env, "convert", src, pdf, "--style="+tc.style); err != nil {
				t.Fatalf("render: %v: %s", err, stderr)
			}
			raw, err := exec.Command("pdftotext", "-bbox", pdf, "-").CombinedOutput()
			if err != nil {
				t.Fatalf("extract: %v: %s", err, raw)
			}
			type word struct {
				Text  string  `xml:",chardata"`
				Left  float64 `xml:"xMin,attr"`
				Right float64 `xml:"xMax,attr"`
				Top   float64 `xml:"yMin,attr"`
			}
			var doc struct {
				Words []word `xml:"body>doc>page>word"`
			}
			if err := xml.Unmarshal(raw, &doc); err != nil {
				t.Fatal(err)
			}
			found := map[string]word{}
			for _, w := range doc.Words {
				if strings.Contains(w.Text, "BOB,") {
					t.Errorf("separator comma remains: %s", w.Text)
				}
				key := strings.TrimSuffix(w.Text, ":")
				if key == "BOB" || key == "(cheerfully)" || key == "Just" {
					found[key] = w
				}
			}
			if len(found) != 3 {
				t.Fatalf("missing expected words: %v", found)
			}
			cue, instruction, dialogue := found["BOB"], found["(cheerfully)"], found["Just"]
			switch tc.name {
			case "us-inline":
				if math.Abs(cue.Top-instruction.Top) > 2 || instruction.Left <= cue.Right {
					t.Errorf("instruction not beside cue: %v %v", cue, instruction)
				}
			case "british-column":
				if math.Abs(instruction.Left-dialogue.Left) > 1 || math.Abs(cue.Top-instruction.Top) > 2 || dialogue.Top <= instruction.Top {
					t.Errorf("British placement changed: %v %v %v", cue, instruction, dialogue)
				}
			case "us-separate":
				if instruction.Top <= cue.Top || dialogue.Top <= instruction.Top {
					t.Errorf("new-line override ignored: %v %v %v", cue, instruction, dialogue)
				}
			}
		})
	}
}
