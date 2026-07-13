// ABOUTME: Characterizes template-backed script Typst and PDF output through the Go CLI.
// ABOUTME: Covers configured layout, Unicode, hostile characters, and real compilation.
package app

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestConvertTypstUsesConfiguredTemplate(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	source := filepath.Join(dir, "play.org")
	target := filepath.Join(dir, "play.typ")
	writeAppFile(t, source, `#+TITLE: Syntax # [Test]
#+AUTHOR: Taḋg Paul
* ACT ONE
** Scene One
*** A sign reads #OPEN [NOW].
**** CÁIT quietly
Price is $5 and path is C:\tmp.[fn:cost]
[fn:cost] A **bold** note.
`)
	writeAppFile(t, filepath.Join(dir, "script.yaml"), `folio:
  font: Libertinus Serif
  font-size: 11pt
  page: a4
  margin: 20mm
  positioning:
    speech:
      space-before: 2em
      speaker:
        bold: false
      dialogue:
        wrap-indent: 8em
`)

	status, stdout, stderr := runApp(t, "convert", source, target)
	if status != 0 {
		t.Fatalf("status %d\nstdout:%s\nstderr:%s", status, stdout, stderr)
	}
	typst := readAppFile(t, target)
	for _, fragment := range []string{
		`#set page(paper: "a4", margin: 20mm`,
		`#set text(font: "Libertinus Serif", size: 11pt`,
		`columns: (8em, 1fr)`,
		`weight: "regular"`,
		`Syntax \# \[Test\]`,
		`Price is \$5 and path is C:\\tmp.`,
		`#footnote[A *bold* note.]`,
	} {
		if !strings.Contains(typst, fragment) {
			t.Errorf("Typst missing %q:\n%s", fragment, typst)
		}
	}
}

func TestConvertPDFCompiles(t *testing.T) {
	if _, err := exec.LookPath("typst"); err != nil {
		t.Skip("typst is not installed")
	}
	dir := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	source := filepath.Join(dir, "play.md")
	target := filepath.Join(dir, "play.pdf")
	writeAppFile(t, source, "# Samhain\n\n*by Taḋg Paul*\n\n## ACT ONE\n\n**CÁIT:**\nHello.\n")

	status, stdout, stderr := runApp(t, "convert", source, target)
	if status != 0 {
		t.Fatalf("status %d\nstdout:%s\nstderr:%s", status, stdout, stderr)
	}
	raw, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(raw, []byte("%PDF")) || len(raw) < 1000 {
		t.Fatalf("invalid PDF output: %d bytes", len(raw))
	}
}

func TestScriptSourceFormatsRasterizeInBritishAndUSStyles(t *testing.T) {
	toolHome := os.Getenv("HOME")
	for _, tool := range []string{"typst", "pdf-to-png"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("%s is not installed", tool)
		}
	}
	sources := map[string]string{
		"org": `#+TITLE: Matrix Play
#+AUTHOR: Example Author
* ACT ONE
** Scene One
*** Night.
**** CÁIT quietly
Hello.
`,
		"md": `# Matrix Play

*by Example Author*

## ACT ONE

### Scene One

*Night.*

**CÁIT:** *(quietly)*
Hello.
`,
		"fountain": `Title: Matrix Play
Author: Example Author

> **ACT ONE** <

.SCENE ONE

Night.

CÁIT
(quietly)
Hello.
`,
	}
	extensions := map[string]string{"org": ".org", "md": ".md", "fountain": ".fountain"}
	for _, style := range []string{"british", "us"} {
		for format, sourceText := range sources {
			t.Run(style+"/"+format, func(t *testing.T) {
				dir := t.TempDir()
				t.Setenv("HOME", t.TempDir())
				source := filepath.Join(dir, "play"+extensions[format])
				target := filepath.Join(dir, style+"-"+format+".pdf")
				writeAppFile(t, source, sourceText)

				status, stdout, stderr := runApp(t, "convert", source, target, "--style", style)
				if status != 0 {
					t.Fatalf("status %d\nstdout:%s\nstderr:%s", status, stdout, stderr)
				}
				raw, err := os.ReadFile(target)
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.HasPrefix(raw, []byte("%PDF")) || len(raw) < 1000 {
					t.Fatalf("invalid PDF output: %d bytes", len(raw))
				}

				cmd := exec.Command("pdf-to-png", filepath.Base(target), "120")
				cmd.Dir = dir
				cmd.Env = []string{"HOME=" + toolHome}
				for _, value := range os.Environ() {
					if !strings.HasPrefix(value, "HOME=") {
						cmd.Env = append(cmd.Env, value)
					}
				}
				if output, err := cmd.CombinedOutput(); err != nil {
					t.Fatalf("rasterizing %s/%s: %v\n%s", style, format, err, output)
				}
				images, err := filepath.Glob(filepath.Join(dir, style+"-"+format+"-*.png"))
				if err != nil || len(images) == 0 {
					t.Fatalf("%s/%s raster output: %v, %v", style, format, images, err)
				}
				for _, image := range images {
					info, err := os.Stat(image)
					if err != nil || info.Size() == 0 {
						t.Errorf("empty raster output %s: %v", image, err)
					}
				}
			})
		}
	}
}
