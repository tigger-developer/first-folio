// ABOUTME: Verifies source configuration and recursive precedence through the manuscript CLI.
// ABOUTME: Compares rendered layouts and rejects invalid or ambiguous source configuration.
package manuscript

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const sourceConfigBody = "Alpha  \nBeta\n\nGamma\n"
const sourceConfigCommon = `folio:
  manuscript:
    page: a5
    margin: 15mm
    font:
      family: DejaVu Sans Mono
      size: 11pt
    line-spacing: 1.25
    paragraph-indent: 0pt
    paragraph-spacing: 5mm
    title-page:
      enabled: false
    toc:
      enabled: false
    page-header:
      enabled: false
    page-footer:
      enabled: false
`

// RT037.1: source font family survives local size overrides; local page wins for each edition.
func TestRT037SourceConfigMergedIntoEditionPDFs(t *testing.T) {
	binary := buildFontCLI(t)
	for _, page := range []string{"a4", "6x9in"} {
		t.Run(page, func(t *testing.T) {
			home, dir := t.TempDir(), t.TempDir()
			t.Setenv("HOME", home)
			writeFile(t, filepath.Join(home, ".config", "first-folio", "script.yaml"), "folio:\n  manuscript:\n    margin: 30mm\n    font:\n      family: Libertinus Serif\n      size: 9pt\n      weight: bold\n")
			input := filepath.Join(dir, "input.md")
			writeFile(t, input, "---\ntitle: SourceTitle\n"+sourceConfigCommon+"---\n"+sourceConfigBody)
			writeFile(t, filepath.Join(dir, "script.yaml"), fmt.Sprintf("folio:\n  manuscript:\n    page: %s\n    font:\n      size: 18pt\n", page))
			pdf := filepath.Join(dir, "source.pdf")
			commandOutput(t, exec.Command(binary, "manuscript", input, pdf))
			got := inlinePDFWords(t, pdf)
			reference := strings.Replace(sourceConfigCommon, "page: a5", "page: "+page, 1)
			reference = strings.Replace(reference, "size: 11pt", "size: 18pt", 1)
			writeFile(t, input, "---\ntitle: SourceTitle\n---\n"+sourceConfigBody)
			writeFile(t, filepath.Join(dir, "script.yaml"), reference)
			expected := filepath.Join(dir, "reference.pdf")
			commandOutput(t, exec.Command(binary, "manuscript", input, expected))
			want := inlinePDFWords(t, expected)
			for _, marker := range []string{"Alpha", "Beta", "Gamma"} {
				g, gok := got[marker]
				w, wok := want[marker]
				if !gok || !wok {
					t.Fatalf("missing %s in rendered output", marker)
				}
				if math.Abs(g.XMin-w.XMin) > 0.03 || math.Abs(g.XMax-w.XMax) > 0.03 || math.Abs(g.YMin-w.YMin) > 0.03 || math.Abs(g.YMax-w.YMax) > 0.03 {
					t.Errorf("%s: source configuration produced %+v, want %+v", marker, g, w)
				}
			}
			info := commandOutput(t, exec.Command("pdfinfo", pdf))
			size := "595.276 x 841.89"
			if page == "6x9in" {
				size = "432 x 648"
			}
			if !strings.Contains(info, size) {
				t.Errorf("wrong %s page dimensions: %s", page, info)
			}
		})
	}
}

// RT037.2: source style selects presets and siblings; local and CLI style still win.
func TestRT037SourceStylePrecedence(t *testing.T) {
	binary := buildFontCLI(t)
	for _, c := range []struct{ name, local, cli, want string }{
		{"source", "", "", "us"}, {"local", "british", "", "british"}, {"cli", "british", "us", "us"},
	} {
		t.Run(c.name, func(t *testing.T) {
			home, dir := t.TempDir(), t.TempDir()
			t.Setenv("HOME", home)
			writeFile(t, filepath.Join(home, ".config", "first-folio", "script.yaml"), "folio:\n  style: british\n")
			writeFile(t, filepath.Join(home, ".config", "first-folio", "script-british.yaml"), "folio:\n  manuscript:\n    margin: 40mm\n")
			writeFile(t, filepath.Join(home, ".config", "first-folio", "script-us.yaml"), "folio:\n  manuscript:\n    margin: 40mm\n")
			input := filepath.Join(dir, "input.md")
			writeFile(t, input, "---\nfolio:\n  manuscript:\n    style: us\n    margin: 15mm\n---\nBody.\n")
			local := "{}\n"
			if c.local != "" {
				local = "folio:\n  manuscript:\n    style: " + c.local + "\n"
			}
			writeFile(t, filepath.Join(dir, "script.yaml"), local)
			writeFile(t, filepath.Join(dir, "script-us.yaml"), "folio:\n  manuscript:\n    margin: 22mm\n")
			args := []string{"manuscript", input, filepath.Join(dir, "unused.pdf"), "--dry-run"}
			if c.cli != "" {
				args = append(args, "--style", c.cli)
			}
			output := commandOutput(t, exec.Command(binary, args...))
			margin := "15mm"
			if c.want == "us" {
				margin = "22mm"
			}
			if !strings.Contains(output, "style: "+c.want+"\n") || !strings.Contains(output, "margin: "+margin+"\n") {
				t.Fatal(output)
			}
		})
	}
}

// RT037.3: source values use the normal validation boundary and fail before output.
func TestRT037RejectInvalidSourceConfiguration(t *testing.T) {
	binary := buildFontCLI(t)
	for _, c := range []struct{ name, front, want string }{
		{"alignment", "folio:\n  manuscript:\n    chapter:\n      align: diagonal\n", "folio.manuscript.chapter.align"},
		{"spacing", "folio:\n  manuscript:\n    line-spacing: 1em\n", "folio.manuscript.line-spacing"},
		{"font", "folio:\n  manuscript:\n    font:\n      size: banana\n", "folio.manuscript.font.size"},
		{"scalar", "folio: wrong\n", "folio"}, {"null", "folio: null\n", "folio"},
		{"render", "render: false\n", "render"}, {"yaml", "folio: [\n", "frontmatter"},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			dir := t.TempDir()
			input := filepath.Join(dir, "input.md")
			writeFile(t, input, "---\n"+c.front+"---\nBody.\n")
			for _, ext := range []string{"typ", "pdf"} {
				output := filepath.Join(dir, "output."+ext)
				data, err := exec.Command(binary, "manuscript", input, output).CombinedOutput()
				if err == nil || !strings.Contains(string(data), c.want) {
					t.Errorf("expected %s rejection, got %s (%v)", c.want, data, err)
				}
				if _, err := os.Stat(output); !os.IsNotExist(err) {
					t.Errorf("invalid source created output: %v", err)
				}
			}
		})
	}
}

// RT037.4: only the first resolved Markdown file supplies configuration; metadata remains content-free.
func TestRT037FirstSourceAndMetadata(t *testing.T) {
	binary := buildFontCLI(t)
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	first, second := filepath.Join(dir, "01.md"), filepath.Join(dir, "02.md")
	writeFile(t, first, "---\ntitle: SourceTitle\nauthor: SourceAuthor\nfolio:\n  manuscript:\n    margin: 17mm\n    title-page:\n      enabled: true\n    toc:\n      enabled: false\n---\nFirstBody.\n")
	writeFile(t, second, "SecondBody.\n")
	target := filepath.Join(dir, "output.typ")
	commandOutput(t, exec.Command(binary, "manuscript", second, first, target))
	output := readFile(t, target)
	for _, value := range []string{"SourceTitle", "SourceAuthor", "FirstBody", "SecondBody", "margin: 17mm"} {
		if !strings.Contains(output, value) {
			t.Errorf("output missing %s", value)
		}
	}
	if strings.Contains(output, "folio:") {
		t.Error("configuration leaked into output")
	}
	for _, key := range []string{"folio", "render"} {
		writeFile(t, second, "---\n"+key+": {}\n---\nSecondBody.\n")
		rejected := filepath.Join(dir, "rejected-"+key+".typ")
		data, err := exec.Command(binary, "manuscript", first, second, rejected).CombinedOutput()
		if err == nil || !strings.Contains(string(data), "02.md") || !strings.Contains(string(data), "first") {
			t.Errorf("later source config not rejected clearly: %s (%v)", data, err)
		}
		if _, err := os.Stat(rejected); !os.IsNotExist(err) {
			t.Error("later configuration created output")
		}
	}
}

// RT037.3: validation sees the effective value after local overrides, not an obsolete source value.
func TestRT037LocalOverrideRepairsSourceValue(t *testing.T) {
	binary := buildFontCLI(t)
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	input := filepath.Join(dir, "input.md")
	writeFile(t, input, "---\nfolio:\n  manuscript:\n    chapter:\n      align: diagonal\nrender: {}\n---\nBody.\n")
	writeFile(t, filepath.Join(dir, "script.yaml"), "folio:\n  manuscript:\n    chapter:\n      align: left\n")
	commandOutput(t, exec.Command(binary, "manuscript", input, filepath.Join(dir, "output.pdf")))
}

// RT037.4: a later input may begin with a normal Markdown scene break.
func TestRT037LaterSceneBreakIsNotFrontmatter(t *testing.T) {
	binary := buildFontCLI(t)
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	first, second := filepath.Join(dir, "01.md"), filepath.Join(dir, "02.md")
	writeFile(t, first, "FirstBody.\n")
	for _, body := range []string{"---\n\nSecondBody.\n", "---\n\nSecondBody.\n\n---\n\nMoreBody.\n"} {
		writeFile(t, second, body)
		target := filepath.Join(dir, "output.typ")
		commandOutput(t, exec.Command(binary, "manuscript", first, second, target))
		if !strings.Contains(readFile(t, target), "SecondBody") {
			t.Fatal("later scene body missing")
		}
	}
}
