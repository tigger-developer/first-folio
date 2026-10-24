// ABOUTME: Measures geometric letter and inline monospace stretch in rendered PDFs.
// ABOUTME: Ensures equal nested ratios apply once without Typst width-face selection.
package letter

import (
	"encoding/xml"
	"math"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/tigger-developer/first-folio/internal/config"
)

func TestRT067LetterAndNestedMonoStretchGeometrically(t *testing.T) {
	for _, tool := range []string{"typst", "pdftotext"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("requires %s", tool)
		}
	}
	render := func(name, ratio string) map[string]float64 {
		t.Helper()
		cfg, err := config.Load(config.Options{Mode: config.ModeLetter, Home: t.TempDir(), LocalDir: t.TempDir(), Source: map[string]any{
			"folio": map[string]any{
				"letter": map[string]any{"font": map[string]any{"family": "Iosevka Publish", "stretch": ratio}},
				"mono":   map[string]any{"font": map[string]any{"family": "Iosevka Publish", "stretch": ratio}},
			},
		}})
		if err != nil {
			t.Fatal(err)
		}
		typst, err := RenderTypst(Letter{Subject: "Stretch", Body: "LETTERWIDTH =MONOWIDTH="}, cfg)
		if err != nil {
			t.Fatal(err)
		}
		pdf := filepath.Join(t.TempDir(), name+".pdf")
		if err := CompilePDF(typst, pdf); err != nil {
			t.Fatal(err)
		}
		raw, err := exec.Command("pdftotext", "-bbox", pdf, "-").CombinedOutput()
		if err != nil {
			t.Fatalf("bbox: %v: %s", err, raw)
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
			if word.Text == "LETTERWIDTH" || word.Text == "MONOWIDTH" {
				widths[word.Text] = word.Right - word.Left
			}
		}
		return widths
	}
	baseline := render("baseline", "100%")
	stretched := render("stretched", "150%")
	for _, marker := range []string{"LETTERWIDTH", "MONOWIDTH"} {
		if baseline[marker] == 0 || stretched[marker] == 0 {
			t.Fatalf("missing %s bounds: baseline=%v stretched=%v", marker, baseline, stretched)
		}
		if ratio := stretched[marker] / baseline[marker]; math.Abs(ratio-1.5) > 0.08 {
			t.Errorf("%s width ratio %.3f, want 1.5", marker, ratio)
		}
	}
}
