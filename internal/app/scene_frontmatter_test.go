// ABOUTME: Protects scene-level introductory paragraphs under render.frontmatter.
// ABOUTME: Rendered documents may hide them while semantic text conversions retain them.
package app

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var sceneFrontmatterSources = map[string]string{
	"org": `* ACT ONE
** Scene One
SCENE-INTRO-MARKER
*** A visible stage direction.
**** CÁIT
Visible dialogue.
`,
	"md": `# Scene frontmatter

## ACT ONE

### Scene One

SCENE-INTRO-MARKER

*A visible stage direction.*

**CÁIT:**
Visible dialogue.
`,
}

func TestRT069_1DisabledFrontmatterHidesSceneParagraphFromRenderedOutputs(t *testing.T) {
	for sourceFormat, sourceText := range sceneFrontmatterSources {
		for _, frontmatter := range []bool{false, true} {
			for _, format := range []string{"pdf", "docx", "odt"} {
				t.Run(sourceFormat+"/"+format+"/frontmatter="+map[bool]string{false: "false", true: "true"}[frontmatter], func(t *testing.T) {
					t.Setenv("HOME", t.TempDir())
					dir := t.TempDir()
					source := filepath.Join(dir, "play."+sourceFormat)
					target := filepath.Join(dir, "play."+format)
					writeAppFile(t, source, sourceText)
					writeAppFile(t, filepath.Join(dir, "script.yaml"), "render:\n  frontmatter: "+map[bool]string{false: "false", true: "true"}[frontmatter]+"\n")

					status, stdout, stderr := runApp(t, "convert", source, target)
					if status != 0 {
						t.Fatalf("status %d; stdout=%q stderr=%q", status, stdout, stderr)
					}
					var rendered string
					if format == "pdf" {
						command := exec.Command("pdftotext", target, "-")
						raw, err := command.CombinedOutput()
						if err != nil {
							t.Fatalf("pdftotext: %v: %s", err, raw)
						}
						rendered = string(raw)
					} else {
						raw, err := os.ReadFile(target)
						if err != nil {
							t.Fatal(err)
						}
						rendered = officeBody(t, officePackage(t, raw), format).text()
					}
					if visible := strings.Contains(rendered, "SCENE-INTRO-MARKER"); visible != frontmatter {
						t.Fatalf("scene frontmatter visibility in %s output = %t, want %t: %q", format, visible, frontmatter, rendered)
					}
					for _, marker := range []string{"Scene One", "visible stage direction", "Visible dialogue"} {
						if !strings.Contains(strings.ToLower(rendered), strings.ToLower(marker)) {
							t.Errorf("%s output lost %q: %q", format, marker, rendered)
						}
					}
				})
			}
		}
	}
}

func TestRT069_2TextConversionsRetainSceneParagraph(t *testing.T) {
	for sourceFormat, sourceText := range sceneFrontmatterSources {
		for _, format := range []string{"md", "fountain"} {
			t.Run(sourceFormat+"/"+format, func(t *testing.T) {
				t.Setenv("HOME", t.TempDir())
				dir := t.TempDir()
				source := filepath.Join(dir, "source."+sourceFormat)
				target := filepath.Join(dir, "output."+format)
				writeAppFile(t, source, sourceText)
				writeAppFile(t, filepath.Join(dir, "script.yaml"), "render:\n  frontmatter: false\n")

				status, stdout, stderr := runApp(t, "convert", source, target)
				if status != 0 {
					t.Fatalf("status %d; stdout=%q stderr=%q", status, stdout, stderr)
				}
				if converted := readAppFile(t, target); !strings.Contains(converted, "SCENE-INTRO-MARKER") {
					t.Fatalf("%s conversion dropped scene frontmatter: %q", format, converted)
				}
			})
		}
	}
}
