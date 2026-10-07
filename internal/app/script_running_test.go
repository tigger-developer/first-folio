// ABOUTME: Exercises script running matter through real multi-page Typst compilation.
// ABOUTME: Checks shared configuration, page state, geometry, and safe format substitution.
package app

import (
	"encoding/xml"
	"fmt"
	"math"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tigger-developer/first-folio/internal/config"
	"github.com/tigger-developer/first-folio/internal/play"
)

func runningConfig(t *testing.T, header, footer string) config.Config {
	t.Helper()
	dir := t.TempDir()
	font := `    font:
      family: "Libertinus Serif"
      size: "10pt"
      weight: "regular"
      stretch: "100%"
      style: "regular"
      letter-spacing: "0pt"
`
	writeAppFile(t, filepath.Join(dir, "script.yaml"), "folio:\n  page-header:\n"+font+header+"  page-footer:\n"+font+footer)
	cfg, err := config.Load(config.Options{Mode: config.ModeScript, Home: t.TempDir(), LocalDir: dir})
	if err != nil {
		t.Fatal(err)
	}
	return cfg
}

func runningPDF(t *testing.T, doc play.Document, cfg config.Config) []string {
	t.Helper()
	for _, tool := range []string{"typst", "pdftotext"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("%s is not installed", tool)
		}
	}
	dir := t.TempDir()
	target := filepath.Join(dir, "play.pdf")
	if err := renderPlayDocument(doc, cfg, target, false, false, nil); err != nil {
		t.Fatal(err)
	}
	raw, err := exec.Command("pdftotext", "-layout", target, "-").CombinedOutput()
	if err != nil {
		t.Fatalf("extracting PDF: %v: %s", err, raw)
	}
	return strings.Split(strings.TrimSuffix(string(raw), "\f"), "\f")
}

func runningActs(title string) play.Document {
	doc := play.Document{Metadata: map[string]string{"title": title, "author": "Author", "date": "DATEKEEP", "version": "VERSIONKEEP"}}
	for i := 1; i <= 3; i++ {
		doc.Events = append(doc.Events, play.Event{Kind: play.EventActHeader, Text: fmt.Sprintf("ACT %d", i)}, play.Event{Kind: play.EventStageDirection, Text: "Body text."})
	}
	return doc
}

func TestScriptRunningFrontmatterAndFacingPages(t *testing.T) {
	cfg := runningConfig(t, "    enabled: true\n    format: 'BODYEVEN'\n    alt-format: 'BODYODD'\n    frontmatter-format: 'FRONTEVEN'\n    alt-frontmatter-format: 'FRONTODD'\n", "    enabled: true\n    format: 'BE [page]'\n    alt-format: 'BO [page]'\n    frontmatter-format: 'FE [page]'\n    alt-frontmatter-format: 'FO [page]'\n")
	doc := runningActs("Front Play")
	intro := []play.Event{{Kind: play.EventIntroHeader, Text: "Introduction"}, {Kind: play.EventCharacterTableStart}, {Kind: play.EventCharacterTableRow, Name: "A", Text: "Cast member"}, {Kind: play.EventCharacterTableEnd}}
	for i := 0; i < 90; i++ {
		intro = append(intro, play.Event{Kind: play.EventIntroText, Text: "Introduction paragraph for repeated running matter."})
	}
	doc.Events = append(intro, doc.Events...)
	pages := runningPDF(t, doc, cfg)
	body := false
	frontPages := 0
	for i, page := range pages[1:] {
		pg := i + 2
		if strings.Contains(page, "ACT 1") {
			body = true
		}
		prefix, foot := "FRONTEVEN", "FE"
		if pg%2 == 1 {
			prefix, foot = "FRONTODD", "FO"
		}
		if body {
			prefix, foot = "BODYEVEN", "BE"
			if pg%2 == 1 {
				prefix, foot = "BODYODD", "BO"
			}
		} else {
			frontPages++
		}
		if !strings.Contains(page, prefix) || !strings.Contains(page, fmt.Sprintf("%s %d", foot, pg)) {
			t.Errorf("page %d wrong running region/parity: %s", pg, page)
		}
	}
	if frontPages < 2 || !body {
		t.Fatalf("fixture did not exercise multi-page frontmatter and body: %q", pages)
	}
}

func TestScriptRunningMatterRepeatsAndSuppressesTitle(t *testing.T) {
	cfg := runningConfig(t, "    enabled: true\n    format: 'RUNHEAD [title] [author]'\n", "    enabled: true\n    format: 'RUNFOOT [page]/[total-pages]'\n")
	pages := runningPDF(t, runningActs("Named Play"), cfg)
	if !strings.Contains(strings.Join(pages[1:], ""), "RUNHEAD") {
		t.Fatalf("no running headers on body pages: %q", pages)
	}
	if len(pages) != 4 {
		t.Fatalf("got %d pages, want title plus three acts: %q", len(pages), pages)
	}
	if strings.Contains(pages[0], "RUNHEAD") || strings.Contains(pages[0], "RUNFOOT") {
		t.Fatalf("running matter on title: %s", pages[0])
	}
	if !strings.Contains(pages[0], "DATEKEEP") || !strings.Contains(pages[0], "VERSIONKEEP") {
		t.Fatalf("lost title metadata: %s", pages[0])
	}
	for i, page := range pages[1:] {
		if !strings.Contains(page, "RUNHEAD Named Play Author") {
			t.Errorf("page %d missing repeated header: %s", i+2, page)
		}
		if !strings.Contains(page, fmt.Sprintf("RUNFOOT %d/4", i+2)) {
			t.Errorf("page %d missing physical page/total: %s", i+2, page)
		}
	}
}

// PDF bbox coordinates expose the rendered geometry without relying on source strings.
type runningBBox struct {
	Pages []struct {
		Width  float64 `xml:"width,attr"`
		Height float64 `xml:"height,attr"`
		Words  []struct {
			Text   string  `xml:",chardata"`
			X      float64 `xml:"xMin,attr"`
			Y      float64 `xml:"yMin,attr"`
			Bottom float64 `xml:"yMax,attr"`
		} `xml:"word"`
	} `xml:"body>doc>page"`
}

func runningBounds(t *testing.T, doc play.Document, cfg config.Config) runningBBox {
	t.Helper()
	target := filepath.Join(t.TempDir(), "bounds.pdf")
	if err := renderPlayDocument(doc, cfg, target, false, false, nil); err != nil {
		t.Fatal(err)
	}
	raw, err := exec.Command("pdftotext", "-bbox", target, "-").CombinedOutput()
	if err != nil {
		t.Fatalf("extracting geometry: %v: %s", err, raw)
	}
	var bbox runningBBox
	if err := xml.Unmarshal(raw, &bbox); err != nil {
		t.Fatal(err)
	}
	return bbox
}

func TestScriptRunningGeometryAndPairAlign(t *testing.T) {
	for _, unit := range []string{"mm", "em"} {
		t.Run(unit, func(t *testing.T) {
			distance, padding, expected := "20mm", "10mm", 20*72/25.4
			if unit == "em" {
				distance, padding, expected = "4em", "2em", 48
			}
			cfg := runningConfig(t, "    enabled: true\n    format: 'HEADGEOM'\n    align: left-right\n    distance-from-edge: "+distance+"\n    content-padding-after: "+padding+"\n", "    enabled: true\n    format: 'FOOTGEOM'\n    align: right-left\n    distance-from-edge: "+distance+"\n    content-padding-after: "+padding+"\n")
			bbox := runningBounds(t, runningActs("Geometry"), cfg)
			if len(bbox.Pages) != 4 {
				t.Fatalf("got %d geometry pages", len(bbox.Pages))
			}
			for i, page := range bbox.Pages[1:] {
				found := 0
				for _, word := range page.Words {
					switch word.Text {
					case "HEADGEOM":
						found++
						if math.Abs(word.Bottom-expected) > 3 {
							t.Errorf("page %d header bottom %.2f, want %.2f", i+2, word.Bottom, expected)
						}
						if (i%2 == 0 && word.X > page.Width/2) || (i%2 == 1 && word.X < page.Width/2) {
							t.Errorf("page %d header pair alignment x=%.2f", i+2, word.X)
						}
					case "ACT":
						bodyTop := expected + 10*72/25.4
						if unit == "em" {
							bodyTop = expected + 24
						}
						// PDF glyph bounds can extend above Typst's font layout box.
						if word.Y < bodyTop-5 || word.Y > bodyTop+12 {
							t.Errorf("page %d body top %.2f, want near %.2f", i+2, word.Y, bodyTop)
						}
					case "FOOTGEOM":
						found++
						if math.Abs((page.Height-word.Y)-expected) > 3 {
							t.Errorf("page %d footer distance %.2f, want %.2f", i+2, page.Height-word.Y, expected)
						}
						if (i%2 == 0 && word.X < page.Width/2) || (i%2 == 1 && word.X > page.Width/2) {
							t.Errorf("page %d footer pair alignment x=%.2f", i+2, word.X)
						}
					}
				}
				if found != 2 {
					t.Errorf("page %d missing running geometry words", i+2)
				}
			}
		})
	}
}

func TestScriptRunningDisabledAndUntitled(t *testing.T) {
	for _, header := range []bool{false, true} {
		for _, footer := range []bool{false, true} {
			t.Run(fmt.Sprintf("header=%t/footer=%t", header, footer), func(t *testing.T) {
				cfg := runningConfig(t, fmt.Sprintf("    enabled: %t\n    format: 'HEADON'\n", header), fmt.Sprintf("    enabled: %t\n    format: 'FOOTON [page]/[total-pages]'\n", footer))
				pages := runningPDF(t, runningActs(""), cfg)
				if len(pages) != 3 {
					t.Fatalf("untitled document got %d pages: %q", len(pages), pages)
				}
				for i, page := range pages {
					if strings.Contains(page, "HEADON") != header || strings.Contains(page, "FOOTON") != footer {
						t.Errorf("page %d disabled role ignored: %s", i+1, page)
					}
					if footer && !strings.Contains(page, fmt.Sprintf("FOOTON %d/3", i+1)) {
						t.Errorf("untitled counter wrong: %s", page)
					}
					if !footer {
						for _, line := range strings.Split(page, "\n") {
							if strings.TrimSpace(line) == fmt.Sprint(i+1) {
								t.Errorf("automatic page number survived disabled footer: %s", page)
							}
						}
					}
				}
			})
		}
	}
}

func TestScriptRunningBodyBoundaryWithoutAct(t *testing.T) {
	for _, kind := range []play.EventKind{play.EventSceneHeader, play.EventCharacter, play.EventStageDirection} {
		t.Run(string(kind), func(t *testing.T) {
			cfg := runningConfig(t, "    enabled: true\n    format: 'BODYFIRST'\n    frontmatter-format: 'FRONTFIRST'\n", "    enabled: true\n    format: 'BFOOT'\n    frontmatter-format: 'FFOOT'\n")
			doc := play.Document{Metadata: map[string]string{}, Events: []play.Event{{Kind: kind, Name: "SPEAKER", Text: "Body starts here."}, {Kind: play.EventDialogue, Text: "Speech."}}}
			pages := runningPDF(t, doc, cfg)
			if len(pages) != 1 || !strings.Contains(pages[0], "BODYFIRST") || !strings.Contains(pages[0], "BFOOT") || strings.Contains(pages[0], "FRONTFIRST") {
				t.Fatalf("first body page classified as frontmatter: %q", pages)
			}
		})
	}
}

func TestScriptRunningHostileStringsAndKnownTokens(t *testing.T) {
	cfg := runningConfig(t, "    enabled: true\n    format: '[title]|[author]|[part][part-number][part-prefix][part-full][chapter][chapter-number][chapter-prefix][chapter-full]|[unknown] # $ @ * _ \\\n'\n", "    enabled: true\n    format: '[page] of [total-pages]'\n")
	doc := runningActs(`Hostile # [page] $ @ * _ \\`)
	doc.Metadata["author"] = `Writer [total-pages] #panic("unsafe")`
	pages := runningPDF(t, doc, cfg)
	for _, page := range pages[1:] {
		for _, want := range []string{"Hostile # [page] $ @", `Writer [total-pages] #panic(“unsafe”)`, "||[unknown] # $ @"} {
			if !strings.Contains(page, want) {
				t.Errorf("literal format/metadata missing %q: %s", want, page)
			}
		}
		if strings.Contains(page, "[chapter]") || strings.Contains(page, "[part]") {
			t.Errorf("manuscript placeholders were not empty: %s", page)
		}
	}
}

func TestScriptRunningSourceFormatsAndStyles(t *testing.T) {
	sources := map[string]string{
		"org":      "#+TITLE: Matrix Play\n#+AUTHOR: Writer\n* ACT ONE\n*** Night.\n* ACT TWO\n*** Morning.\n",
		"md":       "# Matrix Play\n\n*by Writer*\n\n## ACT ONE\n\n*Night.*\n\n## ACT TWO\n\n*Morning.*\n",
		"fountain": "Title: Matrix Play\nAuthor: Writer\n\n> **ACT ONE** <\n\nNight.\n\n> **ACT TWO** <\n\nMorning.\n",
	}
	for _, style := range []string{"british", "us", "screenplay"} {
		for format, source := range sources {
			t.Run(style+"/"+format, func(t *testing.T) {
				dir := t.TempDir()
				t.Setenv("HOME", t.TempDir())
				cfg := runningConfig(t, "    enabled: true\n    format: 'MATRIXHEADER [title]'\n", "    enabled: true\n    format: 'MATRIXFOOTER [page]/[total-pages]'\n")
				var configData map[string]any
				if err := cfg.Decode(&configData); err != nil {
					t.Fatal(err)
				}
				folioData := configData["folio"].(map[string]any)
				configData = map[string]any{"folio": map[string]any{"page-header": folioData["page-header"], "page-footer": folioData["page-footer"]}}
				cfg, err := config.Load(config.Options{Mode: config.ModeScript, Home: t.TempDir(), LocalDir: dir, Source: configData, CLI: map[string]any{"style": style}})
				if err != nil {
					t.Fatal(err)
				}
				parsed, err := play.ParseFormat(format)
				if err != nil {
					t.Fatal(err)
				}
				doc, _, err := play.Parse(parsed, source, "matrix."+format)
				if err != nil {
					t.Fatal(err)
				}
				pages := runningPDF(t, doc, cfg)
				if len(pages) != 3 {
					t.Fatalf("want title plus two body pages, got %q", pages)
				}
				for i, page := range pages[1:] {
					if !strings.Contains(page, "MATRIXHEADER Matrix Play") || !strings.Contains(page, fmt.Sprintf("MATRIXFOOTER %d/3", i+2)) {
						t.Errorf("running matter missing: %s", page)
					}
				}
			})
		}
	}
}

func TestScriptRunningExplicitEmptyFrontmatterAlternate(t *testing.T) {
	cfg := runningConfig(t, "    enabled: true\n    format: 'BODYHEADER'\n    frontmatter-format: 'EVENFRONT'\n    alt-frontmatter-format: ''\n", "    enabled: true\n    format: 'BODYFOOT'\n    frontmatter-format: 'EVENFOOT'\n    alt-frontmatter-format: ''\n")
	doc := play.Document{Metadata: map[string]string{}, Events: []play.Event{{Kind: play.EventIntroHeader, Text: "Introduction"}}}
	for i := 0; i < 90; i++ {
		doc.Events = append(doc.Events, play.Event{Kind: play.EventIntroText, Text: "A long introduction paragraph to fill frontmatter pages."})
	}
	pages := runningPDF(t, doc, cfg)
	if len(pages) < 2 {
		t.Fatal("fixture needs at least two frontmatter pages")
	}
	for i, page := range pages {
		want := (i+1)%2 == 0
		if strings.Contains(page, "EVENFRONT") != want || strings.Contains(page, "EVENFOOT") != want {
			t.Errorf("page %d explicit blank alternate ignored: %s", i+1, page)
		}
	}
}

func TestScriptRunningDefaultCompilation(t *testing.T) {
	cfg, err := config.Load(config.Options{Mode: config.ModeScript, Home: t.TempDir(), LocalDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Bool("folio.page-header.enabled", true) || !cfg.Bool("folio.page-footer.enabled", false) {
		t.Fatal("shared defaults should disable header and enable footer")
	}
	pages := runningPDF(t, runningActs("Default Play"), cfg)
	if len(pages) != 4 {
		t.Fatalf("default script got %d pages", len(pages))
	}
	for i, page := range pages[1:] {
		if strings.Contains(page, "Default Play") {
			t.Errorf("default header enabled: %s", page)
		}
		if !strings.Contains(page, fmt.Sprint(i+2)) {
			t.Errorf("default footer missing on page %d: %s", i+2, page)
		}
	}
}

func TestScriptRunningFirstSceneBeforeLaterAct(t *testing.T) {
	cfg := runningConfig(t, "    enabled: true\n    format: 'DRAMATICBODY'\n    frontmatter-format: 'STILLFRONT'\n", "    enabled: true\n")
	doc := runningActs("")
	doc.Events = append([]play.Event{{Kind: play.EventSceneHeader, Text: "First scene"}, {Kind: play.EventStageDirection, Text: "The drama has begun."}}, doc.Events...)
	pages := runningPDF(t, doc, cfg)
	if !strings.Contains(pages[0], "DRAMATICBODY") || strings.Contains(pages[0], "STILLFRONT") {
		t.Fatalf("scene before later act misclassified: %s", pages[0])
	}
}

func TestScriptRunningMatterContinuesAcrossAutomaticPageBreaks(t *testing.T) {
	cfg := runningConfig(t, "    enabled: true\n    format: 'CONTINUEDHEADER'\n", "    enabled: true\n    format: 'CONTINUEDFOOTER [page]/[total-pages]'\n")
	doc := play.Document{Metadata: map[string]string{"title": "Continuous Play"}, Events: []play.Event{{Kind: play.EventActHeader, Text: "ACT ONE"}}}
	for i := 0; i < 100; i++ {
		doc.Events = append(doc.Events, play.Event{Kind: play.EventCharacter, Name: "SPEAKER"}, play.Event{Kind: play.EventDialogue, Text: "Dialogue flowing onto subsequent pages without an explicit act break."})
	}
	pages := runningPDF(t, doc, cfg)
	if len(pages) < 4 {
		t.Fatal("fixture did not paginate automatically")
	}
	for i, page := range pages[1:] {
		if !strings.Contains(page, "CONTINUEDHEADER") || !strings.Contains(page, fmt.Sprintf("CONTINUEDFOOTER %d/%d", i+2, len(pages))) {
			t.Errorf("page %d running matter missing: %s", i+2, page)
		}
	}
}

func TestScriptRunningExplicitBlankFrontmatter(t *testing.T) {
	cfg := runningConfig(t, "    enabled: true\n    format: 'BODYONLYHEADER'\n    frontmatter-format: ''\n", "    enabled: true\n    format: 'BODYONLYFOOTER'\n    frontmatter-format: ''\n")
	doc := runningActs("Blank Front")
	doc.Events = append([]play.Event{{Kind: play.EventIntroHeader, Text: "Introduction"}, {Kind: play.EventIntroText, Text: "Frontmatter should carry no running matter."}}, doc.Events...)
	pages := runningPDF(t, doc, cfg)
	if len(pages) != 5 {
		t.Fatalf("want title, intro, three acts: %q", pages)
	}
	if strings.Contains(pages[1], "BODYONLYHEADER") || strings.Contains(pages[1], "BODYONLYFOOTER") {
		t.Errorf("explicit blank frontmatter ignored: %s", pages[1])
	}
	if !strings.Contains(pages[2], "BODYONLYHEADER") || !strings.Contains(pages[2], "BODYONLYFOOTER") {
		t.Errorf("first body page lost running matter: %s", pages[2])
	}
}
