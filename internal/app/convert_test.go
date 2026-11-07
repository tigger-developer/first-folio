// ABOUTME: Characterizes text conversion through the public Go application boundary.
// ABOUTME: Covers files, stdout, config metadata, Unicode, warnings, and invalid input.
package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tigger-developer/first-folio/internal/config"
)

func TestConvertTextFormats(t *testing.T) {
	dir := t.TempDir()
	home := t.TempDir()
	t.Setenv("HOME", home)
	orgPath := filepath.Join(dir, "play.org")
	mdPath := filepath.Join(dir, "play.md")
	writeAppFile(t, orgPath, `#+TITLE: Samhain
#+AUTHOR: Tadhg O'Brien
* ACT ONE
** Scene One
**** CÁIT quietly
Hello.
`)

	status, stdout, stderr := runApp(t, "convert", orgPath, mdPath)
	if status != 0 {
		t.Fatalf("status %d\nstdout:%s\nstderr:%s", status, stdout, stderr)
	}
	markdown := readAppFile(t, mdPath)
	for _, fragment := range []string{"# Samhain", "## ACT ONE", "**CÁIT:** *(quietly)*", "Hello."} {
		if !strings.Contains(markdown, fragment) {
			t.Errorf("Markdown missing %q:\n%s", fragment, markdown)
		}
	}

	status, stdout, stderr = runApp(t, "convert", mdPath, "--to", "org")
	if status != 0 || stderr != "" {
		t.Fatalf("status %d stderr %q", status, stderr)
	}
	for _, fragment := range []string{"#+TITLE: Samhain", "* ACT ONE", "**** CÁIT quietly"} {
		if !strings.Contains(stdout, fragment) {
			t.Errorf("Org stdout missing %q:\n%s", fragment, stdout)
		}
	}
}

func TestConvertConfigMetadataOverride(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	writeAppFile(t, filepath.Join(dir, "play.org"), "#+TITLE: Source Title\n* ACT ONE\n**** CÁIT\nHello.\n")
	writeAppFile(t, filepath.Join(dir, "script.yaml"), "title: Config Title\nfolio:\n  default-format: md\n")

	status, stdout, stderr := runApp(t, "convert", filepath.Join(dir, "play.org"))
	if status != 0 || stderr != "" {
		t.Fatalf("status %d stderr %q", status, stderr)
	}
	if !strings.Contains(stdout, "# Config Title") || strings.Contains(stdout, "# Source Title") {
		t.Fatalf("config metadata not applied:\n%s", stdout)
	}
}

func TestOrgDottedFrontmatterUsesSourceConfigurationLayer(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	source := filepath.Join(dir, "play.org")
	writeAppFile(t, source, `#+FOLIO.DEFAULT-FORMAT: md
#+RENDER.CHARACTER-TABLE: false
#+TITLE: Source configuration
* CHARACTERS
| CÁIT | A speaker |
* ACT ONE
**** CÁIT
Hello.
`)

	status, stdout, stderr := runApp(t, "convert", source)
	if status != 0 || stderr != "" {
		t.Fatalf("status %d stderr %q", status, stderr)
	}
	if !strings.Contains(stdout, "# Source configuration") || strings.Contains(stdout, "A speaker") {
		t.Fatalf("Org source configuration did not select Markdown and hide the character table:\n%s", stdout)
	}
	if strings.Contains(stdout, "FOLIO.DEFAULT-FORMAT") || strings.Contains(stdout, "RENDER.CHARACTER-TABLE") {
		t.Fatalf("Org configuration frontmatter leaked into document output:\n%s", stdout)
	}
}

func TestOrgDottedFrontmatterWarnsAndIgnoresUnknownFolioField(t *testing.T) {
	for _, field := range []string{"FOLIO.DEFAULT-FROMAT", "FOLIO-DEFAULT-FORMAT"} {
		t.Run(field, func(t *testing.T) {
			dir := t.TempDir()
			t.Setenv("HOME", t.TempDir())
			source := filepath.Join(dir, "play.org")
			target := filepath.Join(dir, "play.md")
			writeAppFile(t, source, "#+"+field+`: md
#+TITLE: Warning example
* ACT ONE
**** CÁIT
Hello.
`)

			status, _, stderr := runApp(t, "convert", source, target)
			if status != 0 {
				t.Fatalf("status %d stderr %q", status, stderr)
			}
			want := source + `:1: unrecognized FOLIO frontmatter field "` + strings.ToLower(field) + `"`
			if !strings.Contains(stderr, want) {
				t.Fatalf("stderr %q does not contain %q", stderr, want)
			}
			markdown := readAppFile(t, target)
			if strings.Contains(markdown, field) {
				t.Fatalf("unknown Org configuration frontmatter leaked into document output:\n%s", markdown)
			}
		})
	}
}

func TestOrgDottedCoverImageFrontmatterUsesExistingYAMLContract(t *testing.T) {
	source := `#+FOLIO.COVER-IMAGE.PATH: artwork/cover.jpg
#+FOLIO.COVER-IMAGE.MODE: inset
#+FOLIO.COVER-IMAGE.PLACEMENT: overlay
#+FOLIO.COVER-IMAGE.SIZE: 80mm
#+FOLIO.COVER-IMAGE.IMAGE.TREATMENT: light
#+FOLIO.COVER-IMAGE.IMAGE.OPACITY: 100%
#+FOLIO.COVER-IMAGE.SUPERSCALE: null
#+FOLIO.COVER-IMAGE.OVERLAY.BLUR: 0mm
#+FOLIO.COVER-IMAGE.OVERLAY.WASH.TREATMENT: dark
#+FOLIO.COVER-IMAGE.OVERLAY.WASH.OPACITY: 30%
#+FOLIO.COVER-IMAGE.OVERLAY.PADDING: 1em
#+TITLE: Cover example
* ACT ONE
`

	layer, warnings, err := orgScriptSourceConfig(source, "cover.org")
	if err != nil || len(warnings) != 0 {
		t.Fatalf("extracting Org cover configuration: warnings=%v err=%v", warnings, err)
	}
	cfg, err := config.Load(config.Options{Mode: config.ModeScript, Home: t.TempDir(), LocalDir: t.TempDir(), Source: layer})
	if err != nil {
		t.Fatalf("loading projected Org cover configuration: %v", err)
	}
	cover, enabled, err := cfg.CoverImage()
	if err != nil || !enabled {
		t.Fatalf("cover enabled=%v err=%v", enabled, err)
	}
	if cover.Path != "artwork/cover.jpg" || cover.Mode != "inset" || cover.Size != "80mm" || cover.ImageOpacity != 1 || cover.OverlayWashOpacity != 0.3 {
		t.Fatalf("cover = %#v", cover)
	}
}

func TestOrgSourceConfigurationRemainsBelowLocalConfiguration(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	source := filepath.Join(dir, "play.org")
	writeAppFile(t, source, `#+FOLIO.DEFAULT-FORMAT: md
#+RENDER.CHARACTER-TABLE: false
#+TITLE: Cascade example
* CHARACTERS
| CÁIT | A speaker |
* ACT ONE
**** CÁIT
Hello.
`)
	writeAppFile(t, filepath.Join(dir, "script.yaml"), "render:\n  character-table: true\n")

	status, stdout, stderr := runApp(t, "convert", source)
	if status != 0 || stderr != "" {
		t.Fatalf("status %d stderr %q", status, stderr)
	}
	if !strings.Contains(stdout, "A speaker") {
		t.Fatalf("local configuration did not override Org source frontmatter:\n%s", stdout)
	}
}

func TestOrgConfigurationPrefixIsBoundedWithoutBoundingTheDocumentBody(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "large.org")
	writeAppFile(t, source, "#+FOLIO.DEFAULT-FORMAT: md\n* ACT ONE\n"+strings.Repeat("Body text. ", orgConfigurationPrefixLimit/8))
	prefix, err := readOrgConfigurationPrefix(source)
	if err != nil {
		t.Fatalf("reading bounded frontmatter before a large body: %v", err)
	}
	if !strings.Contains(prefix, "FOLIO.DEFAULT-FORMAT") {
		t.Fatalf("configuration prefix missing supported field: %q", prefix)
	}

	oversized := filepath.Join(dir, "oversized-frontmatter.org")
	writeAppFile(t, oversized, strings.Repeat("#+TITLE: frontmatter padding\n", orgConfigurationPrefixLimit/10))
	if _, err := readOrgConfigurationPrefix(oversized); err == nil || !strings.Contains(err.Error(), "Org configuration frontmatter exceeds") {
		t.Fatalf("oversized Org frontmatter error = %v", err)
	}
}

func TestConvertInvalidInputs(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	binary := filepath.Join(dir, "binary.org")
	writeAppFile(t, binary, "text\x00binary")
	text := filepath.Join(dir, "text.org")
	writeAppFile(t, text, "* ACT ONE\n")

	tests := []struct {
		name string
		args []string
		want string
	}{
		{"missing", []string{"convert", filepath.Join(dir, "missing.org"), "out.md"}, "file not found"},
		{"binary", []string{"convert", binary, "out.md"}, "appears to be a binary file"},
		{"unknown target", []string{"convert", text, "out.xyz"}, "Unrecognised file extension"},
		{"no source", []string{"convert"}, "no source file specified"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, _, stderr := runApp(t, tt.args...)
			if status == 0 || !strings.Contains(stderr, tt.want) {
				t.Fatalf("status %d, stderr %q; want %q", status, stderr, tt.want)
			}
		})
	}
}

func runApp(t *testing.T, args ...string) (int, string, string) {
	t.Helper()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	status := Run(args, strings.NewReader(""), &stdout, &stderr)
	return status, stdout.String(), stderr.String()
}

func writeAppFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readAppFile(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}
