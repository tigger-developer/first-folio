// ABOUTME: Exercises explicit alignment configuration through the built CLI.
// ABOUTME: Checks rejection before output and grouped title-page PDF placement.
package manuscript

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

var alignmentFields = []string{
	"folio.title-page.title.align", "folio.title-page.subtitle.align", "folio.title-page.author.align",
	"folio.positioning.speech.align", "folio.positioning.speech.speaker.align",
	"folio.positioning.speech.speech-instruction.align", "folio.positioning.speech.dialogue.align",
	"folio.positioning.stage-direction.align", "folio.positioning.transition.align",
	"folio.positioning.frontmatter.header.align", "folio.positioning.act-header.align", "folio.positioning.scene-header.align",
	"folio.manuscript.title-page.title-block-align", "folio.manuscript.title-page.footer-align",
	"folio.manuscript.title-page.title.align", "folio.manuscript.title-page.subtitle.align",
	"folio.manuscript.title-page.author.align", "folio.manuscript.title-page.date.align",
	"folio.manuscript.title-page.wordcount.align", "folio.manuscript.title-page.version.align",
	"folio.manuscript.title-page.contact.align", "folio.manuscript.page-header.align",
	"folio.manuscript.page-footer.align", "folio.manuscript.part.align",
	"folio.manuscript.part.vertical-align", "folio.manuscript.chapter.align", "folio.manuscript.copyright.align",
}

// RT049.1: invalid resolved alignments fail before output, including inactive roles.
func TestRT049RejectInvalidAlignment(t *testing.T) {
	binary := buildFontCLI(t)
	for _, field := range alignmentFields {
		for _, value := range []any{"", "   ", nil, "diagonal", "left-right-center", "#panic()", true, 42, []any{"left"}, map[string]any{"value": "left"}} {
			t.Run(fmt.Sprintf("%s/%v", field, value), func(t *testing.T) {
				cfg := map[string]any{}
				setFontTestPath(cfg, field, value)
				dir := alignmentFixture(t, cfg)
				mode, source := "manuscript", "input.md"
				if !strings.HasPrefix(field, "folio.manuscript.") {
					mode, source = "convert", "play.org"
				}
				for _, ext := range []string{"typ", "pdf"} {
					output := filepath.Join(dir, "output."+ext)
					data, err := exec.Command(binary, mode, filepath.Join(dir, source), output).CombinedOutput()
					if err == nil {
						t.Errorf("accepted %s = %#v", field, value)
					}
					if !strings.Contains(string(data), field) || !strings.Contains(string(data), "expected") || !strings.Contains(string(data), fmt.Sprintf("%#v", value)) {
						t.Errorf("missing field and allowed-value guidance: %s", data)
					}
					if _, err := os.Stat(output); !os.IsNotExist(err) {
						t.Errorf("invalid input created output: %v", err)
					}
				}
			})
		}
	}
}

// RT049.2: explicit groups retain title flow and footer positions; US omissions are configurable.
func TestRT049GroupedTitlePage(t *testing.T) {
	binary := buildFontCLI(t)
	for _, style := range []string{"british", "us"} {
		for _, explicit := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/explicit=%v", style, explicit), func(t *testing.T) {
				cfg := map[string]any{}
				if explicit {
					for _, item := range []string{"title", "subtitle", "author", "date", "wordcount", "version"} {
						setFontTestPath(cfg, "folio.manuscript.title-page."+item+".align", "group")
					}
				}
				dir := alignmentFixture(t, cfg)
				pdf := filepath.Join(dir, "output.pdf")
				render := func() map[string]inlinePDFWord {
					commandOutput(t, exec.Command(binary, "manuscript", filepath.Join(dir, "input.md"), pdf, "--style", style))
					return inlinePDFWords(t, pdf)
				}
				words := render()
				for _, marker := range []string{"TitleProbe", "SubtitleProbe", "AuthorProbe", "CountProbe"} {
					word, ok := words[marker]
					if !ok {
						t.Fatalf("missing %s", marker)
					}
					if math.Abs((word.XMin+word.XMax)/2-297.638) > 2 {
						t.Errorf("%s is not centred: %+v", marker, word)
					}
				}
				if words["TitleProbe"].YMax >= words["SubtitleProbe"].YMin || words["SubtitleProbe"].YMax >= words["AuthorProbe"].YMin {
					t.Error("title group overlaps or loses reading order")
				}
				if words["CountProbe"].YMin < 700 {
					t.Error("word count is not at the bottom")
				}
				for _, marker := range []string{"DateProbe", "VersionProbe"} {
					_, present := words[marker]
					if present != (style == "british") {
						t.Errorf("%s presence=%v in %s", marker, present, style)
					}
				}
				if style == "us" {
					setFontTestPath(cfg, "folio.manuscript.title-page.include-date", true)
					setFontTestPath(cfg, "folio.manuscript.title-page.include-version", true)
					writeAlignmentConfig(t, dir, cfg)
					words = render()
				}
				for _, marker := range []string{"DateProbe", "VersionProbe"} {
					word, ok := words[marker]
					if !ok || word.YMin < 700 {
						t.Errorf("enabled %s missing from footer", marker)
					}
				}
				if math.Abs(words["DateProbe"].YMin-words["CountProbe"].YMin) > 2 || math.Abs(words["VersionProbe"].YMin-words["CountProbe"].YMin) > 2 {
					t.Error("footer items are not on the same row")
				}
				if words["VersionProbe"].XMin >= words["CountProbe"].XMin || words["DateProbe"].XMin <= words["CountProbe"].XMin {
					t.Error("footer group lost left/centre/right order")
				}
			})
		}
	}
}

// RT049.3: accepted alignment forms and empty content remain usable.
func TestRT049AcceptedAlignmentAndEmptyContent(t *testing.T) {
	binary := buildFontCLI(t)
	cases := []struct {
		field  string
		values []string
	}{
		{"folio.manuscript.title-page.title.align", []string{"left", "center", "right", "top-left", "top-center", "top-right", "center-left", "center-center", "center-right", "bottom-left", "bottom-center", "bottom-right", "group"}},
		{"folio.manuscript.page-header.align", []string{"left", "center", "right", "left-right", "right-left", "left-center", "right-center", "center-left", "center-right", "left-left", "center-center", "right-right"}},
		{"folio.manuscript.part.vertical-align", []string{"top", "center", "bottom", "middle", "horizon"}},
		{"folio.manuscript.chapter.align", []string{"left", "center", "right"}},
	}
	for _, c := range cases {
		for _, value := range c.values {
			t.Run(c.field+"/"+value, func(t *testing.T) {
				cfg := map[string]any{}
				setFontTestPath(cfg, c.field, value)
				for _, path := range []string{"folio.manuscript.copyright.publisher", "folio.manuscript.copyright.isbn", "folio.positioning.speech.speaker.prefix", "folio.positioning.speech.dialogue.suffix"} {
					setFontTestPath(cfg, path, "")
				}
				dir := alignmentFixture(t, cfg)
				commandOutput(t, exec.Command(binary, "manuscript", filepath.Join(dir, "input.md"), filepath.Join(dir, "output.typ")))
			})
		}
	}
}

func alignmentFixture(t *testing.T, cfg map[string]any) string {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "input.md"), "---\ntitle: TitleProbe\nsubtitle: SubtitleProbe\nauthor: AuthorProbe\ndate: DateProbe\nversion: VersionProbe\nwordcount: CountProbe\n---\nBodyProbe\n")
	writeFile(t, filepath.Join(dir, "play.org"), "#+TITLE: TitleProbe\n* ACT ONE\n** Scene One\n**** ALEX\nBodyProbe.\n")
	setFontTestPath(cfg, "folio.manuscript.toc.enabled", false)
	setFontTestPath(cfg, "folio.manuscript.page-header.enabled", false)
	setFontTestPath(cfg, "folio.manuscript.page-footer.enabled", false)
	setFontTestPath(cfg, "folio.manuscript.title-page.author.attribution", "")
	writeAlignmentConfig(t, dir, cfg)
	return dir
}

func writeAlignmentConfig(t *testing.T, dir string, cfg map[string]any) {
	t.Helper()
	data, err := yaml.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(dir, "script.yaml"), string(data))
}

// RT049.1: group belongs only to the six grouped title-page items.
func TestRT049GroupRejectedOutsideTitleItems(t *testing.T) {
	binary := buildFontCLI(t)
	for _, field := range alignmentFields {
		if strings.HasPrefix(field, "folio.manuscript.title-page.") && strings.HasSuffix(field, ".align") && !strings.Contains(field, ".contact.") {
			continue
		}
		t.Run(field, func(t *testing.T) {
			cfg := map[string]any{}
			setFontTestPath(cfg, field, "group")
			dir := alignmentFixture(t, cfg)
			data, err := exec.Command(binary, "manuscript", filepath.Join(dir, "input.md"), filepath.Join(dir, "output.typ")).CombinedOutput()
			if err == nil || !strings.Contains(string(data), field) {
				t.Fatalf("expected rejection naming %s: %s (%v)", field, data, err)
			}
		})
	}
}

// RT049.3: null optional content and prefix/suffix strings are not alignment errors.
func TestRT049NullOptionalContent(t *testing.T) {
	binary := buildFontCLI(t)
	cfg := map[string]any{}
	for _, path := range []string{"folio.manuscript.copyright.publisher", "folio.manuscript.copyright.isbn", "folio.positioning.speech.speaker.prefix", "folio.positioning.speech.dialogue.suffix"} {
		setFontTestPath(cfg, path, nil)
	}
	dir := alignmentFixture(t, cfg)
	for _, command := range []struct{ mode, source string }{{"manuscript", "input.md"}, {"convert", "play.org"}} {
		commandOutput(t, exec.Command(binary, command.mode, filepath.Join(dir, command.source), filepath.Join(dir, command.mode+".pdf")))
	}
}
