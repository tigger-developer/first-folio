// ABOUTME: Measures configured script font stretch in rendered PDF geometry.
// ABOUTME: Protects speaker, instruction, and dialogue scaling beyond Typst source text.
package app

import (
	"encoding/xml"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/tigger-developer/first-folio/internal/config"
	"github.com/tigger-developer/first-folio/internal/play"
)

func TestRT067_1ScriptPDFAppliesGeometricFontStretch(t *testing.T) {
	for _, tool := range []string{"typst", "pdftotext"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("requires %s", tool)
		}
	}
	dir := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	source := filepath.Join(dir, "play.org")
	writeAppFile(t, source, "* ACT ONE\n**** SPEAKERSCALE (INSTRUCTIONSCALE)\nDIALOGUESCALE =MONOSCALE=\n")

	render := func(name string, speaker, instruction, dialogue, mono int) map[string]float64 {
		t.Helper()
		writeAppFile(t, filepath.Join(dir, "script.yaml"), "folio:\n  mono:\n    font:\n      family: Iosevka Publish\n      stretch: "+strconv.Itoa(mono)+"%\n  positioning:\n    speech:\n      speaker:\n        font:\n          family: Iosevka Publish\n          stretch: "+strconv.Itoa(speaker)+"%\n      speech-instruction:\n        font:\n          family: Iosevka Publish\n          stretch: "+strconv.Itoa(instruction)+"%\n      dialogue:\n        font:\n          family: Iosevka Publish\n          stretch: "+strconv.Itoa(dialogue)+"%\n")
		pdf := filepath.Join(dir, name+".pdf")
		status, stdout, stderr := runApp(t, "convert", source, pdf)
		if status != 0 {
			t.Fatalf("status %d\nstdout:%s\nstderr:%s", status, stdout, stderr)
		}
		raw, err := exec.Command("pdftotext", "-bbox", pdf, "-").CombinedOutput()
		if err != nil {
			t.Fatalf("extracting PDF bounds: %v: %s", err, raw)
		}
		var bounds struct {
			Words []struct {
				Text  string  `xml:",chardata"`
				Left  float64 `xml:"xMin,attr"`
				Right float64 `xml:"xMax,attr"`
			} `xml:"body>doc>page>word"`
		}
		if err := xml.Unmarshal(raw, &bounds); err != nil {
			t.Fatal(err)
		}
		widths := map[string]float64{}
		for _, word := range bounds.Words {
			key := strings.Trim(word.Text, "():")
			if strings.HasSuffix(key, "SCALE") {
				widths[key] = word.Right - word.Left
			}
		}
		for _, key := range []string{"SPEAKERSCALE", "INSTRUCTIONSCALE", "DIALOGUESCALE", "MONOSCALE"} {
			if widths[key] <= 0 {
				t.Fatalf("missing %s bounds in %v", key, widths)
			}
		}
		return widths
	}

	baseline := render("baseline", 100, 100, 100, 100)
	stretched := render("stretched", 50, 200, 75, 75)
	for _, tc := range []struct {
		key       string
		want, tol float64
	}{
		{"SPEAKERSCALE", 0.5, 0.08},
		{"INSTRUCTIONSCALE", 2, 0.12},
		{"DIALOGUESCALE", 0.75, 0.08},
		{"MONOSCALE", 0.75, 0.08},
	} {
		ratio := stretched[tc.key] / baseline[tc.key]
		if ratio < tc.want-tc.tol || ratio > tc.want+tc.tol {
			t.Errorf("%s width ratio %.3f, want %.2f +/- %.2f", tc.key, ratio, tc.want, tc.tol)
		}
	}
}

func TestRT067_2EveryScriptFontRoleUsesGeometricStretch(t *testing.T) {
	paths := []string{
		"font", "mono.font", "heading.font", "page-header.font", "page-footer.font",
		"title-page.title.font", "title-page.subtitle.font", "title-page.author.font", "title-page.date.font", "title-page.version.font",
		"positioning.speech.speaker.font", "positioning.speech.speech-instruction.font", "positioning.speech.dialogue.font",
		"positioning.stage-direction.font", "positioning.transition.font", "positioning.frontmatter.header.font",
		"positioning.frontmatter.paragraph.font", "positioning.act-header.font", "positioning.scene-header.font",
	}
	source := map[string]any{}
	set := func(path string, value any) {
		parts := strings.Split(path, ".")
		current := source
		for _, part := range parts[:len(parts)-1] {
			next, ok := current[part].(map[string]any)
			if !ok {
				next = map[string]any{}
				current[part] = next
			}
			current = next
		}
		current[parts[len(parts)-1]] = value
	}
	set("folio.page-header.enabled", true)
	set("folio.page-footer.enabled", true)
	for i, path := range paths {
		set("folio."+path+".stretch", fmt.Sprintf("%d%%", 101+i))
	}
	cfg, err := config.Load(config.Options{Mode: config.ModeScript, Home: t.TempDir(), LocalDir: t.TempDir(), Source: source})
	if err != nil {
		t.Fatal(err)
	}
	doc := play.Document{Metadata: map[string]string{"title": "TITLE", "subtitle": "SUBTITLE", "author": "AUTHOR", "date": "DATE", "version": "VERSION"}, Events: []play.Event{
		{Kind: play.EventIntroHeader, Text: "INTROHEAD"}, {Kind: play.EventIntroText, Text: "INTROBODY"},
		{Kind: play.EventActHeader, Text: "ACT"}, {Kind: play.EventSceneHeader, Text: "SCENE"},
		{Kind: play.EventStageDirection, Text: "DIRECTION"}, {Kind: play.EventCharacter, Name: "SPEAKER", Direction: "INSTRUCTION"},
		{Kind: play.EventDialogue, Text: "DIALOGUE `MONO`"}, {Kind: play.EventTransition, Text: "TRANSITION"},
	}}
	typst, err := renderPlayTypst(doc, cfg, play.FormatOrg)
	if err != nil {
		t.Fatal(err)
	}
	for i, path := range paths {
		ratio := fmt.Sprintf("%d%%", 101+i)
		if !strings.Contains(typst, "folio-font-stretch("+ratio) {
			t.Errorf("%s does not route %s through geometric scaling", path, ratio)
		}
		if strings.Contains(typst, "stretch: "+ratio) {
			t.Errorf("%s still delegates %s to Typst native stretch", path, ratio)
		}
	}
}
