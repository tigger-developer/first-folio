// ABOUTME: Defines W054 public executable regressions for formatted office exports.
// ABOUTME: Checks generated native content/layout and safe failures without inspecting source.
package app

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

type officeFailingWriter struct{}

func (officeFailingWriter) Write([]byte) (int, error) {
	return 0, fmt.Errorf("synthetic stream failure")
}

func TestRT054_3OfficeStreamAndResourceLimits(t *testing.T) {
	binary := officeBuild(t)
	dir := t.TempDir()
	src := filepath.Join(dir, "play.org")
	writeAppFile(t, src, officeSources[".org"])
	for _, format := range []string{"docx", "odt"} {
		_, stderr, err := officeRun(t, binary, dir, officeEnv(t, t.TempDir(), nil), "convert", src, filepath.Join(dir, "absent", "output."+format))
		if err == nil || stderr == "" {
			t.Fatalf("unwritable target: %v %s", err, stderr)
		}
	}
	t.Setenv("HOME", t.TempDir())
	var stderr bytes.Buffer
	status := Run([]string{"convert", src, "--to", "docx"}, strings.NewReader(""), officeFailingWriter{}, &stderr)
	if status == 0 || !strings.Contains(stderr.String(), "synthetic stream failure") {
		t.Errorf("stream failure lost: status=%d %s", status, stderr.String())
	}
	writeAppFile(t, src, strings.Repeat("A", (32<<20)+1))
	_, diagnostic, err := officeRun(t, binary, dir, officeEnv(t, t.TempDir(), nil), "convert", src, "--to", "docx")
	if err == nil || !strings.Contains(diagnostic, "32 MiB") {
		t.Errorf("source limit not enforced: %v %s", err, diagnostic)
	}
}

func TestRT054_1OfficeContentMatrix(t *testing.T) {
	if _, err := exec.LookPath("pandoc"); err != nil {
		t.Fatal("office regression requires Pandoc", err)
	}
	binary := officeBuild(t)
	for _, format := range []string{"docx", "odt"} {
		for _, style := range []string{"british", "us", "screenplay"} {
			for source := range officeSources {
				t.Run(format+"/"+style+source, func(t *testing.T) {
					pkg := officeExport(t, binary, source, format, style)
					body := officeBody(t, pkg, format)
					text := body.text()
					previous := -1
					for _, want := range []string{"Matrix Play", "Example Author", "ACT ONE", "SCENE ONE", "Night.", "CÁIT", "softly", "Hello.", "Second line."} {
						at := strings.Index(text, want)
						if at <= previous {
							t.Errorf("missing/out-of-order %q in %q", want, text)
						}
						previous = at
					}
					if len(body.all("drawing"))+len(body.all("image")) != 0 {
						t.Fatal("script flattened into images")
					}
				})
			}
		}
	}
	t.Run("rich-and-filtered", func(t *testing.T) {
		fixtures := map[string]string{
			".org": "#+TITLE: Rich Script\n#+SCHEMA: https://example.invalid/schema\n* CHARACTERS\n| CÁIT | A visitor |\n* Introduction\nOpening summary.\n* ACT ONE\n** SCENE ONE\n*** Night direction.\n**** CÁIT (quietly)\n**Bold text** and *italic text* and ***combined text*** and _underlined text_ and =literal text=.[fn:note]\n*** Prop card :prop:\n***** CUT TO\n[fn:note] Note content. **Bold note** and *italic note*.\n",
			".md":  "---\nschema: https://example.invalid/schema\n---\n# Rich Script\n\n| Name | Description |\n| --- | --- |\n| CÁIT | A visitor |\n\n## Introduction\nOpening summary.\n\n## ACT ONE\n### SCENE ONE\n*Night direction.*\n**CÁIT:** *(quietly)*\n**Bold text** and *italic text* and ***combined text*** and _underlined text_ and `literal text`.[^note]\n***\"Prop card\"***\n> CUT TO\n[^note]: Note content. **Bold note** and *italic note*.\n",
		}
		for extension, rich := range fixtures {
			for _, format := range []string{"docx", "odt"} {
				t.Run(extension+"/"+format, func(t *testing.T) {
					dir := t.TempDir()
					src := filepath.Join(dir, "rich"+extension)
					writeAppFile(t, src, rich)
					target := filepath.Join(dir, "rich."+format)
					env := officeEnv(t, t.TempDir(), nil)
					_, stderr, err := officeRun(t, binary, dir, env, "convert", src, target)
					if err != nil {
						t.Fatalf("%v: %s", err, stderr)
					}
					raw, err := os.ReadFile(target)
					if err != nil {
						t.Fatal(err)
					}
					pkg := officePackage(t, raw)
					body := officeBody(t, pkg, format)
					for _, want := range []string{"A visitor", "Opening summary.", "Night direction.", "Bold text", "italic text", "combined text", "underlined text", "literal text", "Prop card", "CUT TO"} {
						if !strings.Contains(body.text(), want) {
							t.Errorf("%s missing %q", format, want)
						}
					}
					tables := len(body.all("tbl")) + len(body.all("table"))
					if tables == 0 {
						t.Fatal("cast table absent")
					}
					note := body
					if format == "docx" {
						note = pkg["word/footnotes.xml"]
					}
					if note == nil || !strings.Contains(note.text(), "Note content.") {
						t.Fatal("native footnote missing")
					}
					officeAssertNoteLink(t, pkg, format)
					officeAssertInlineNoteFormatting(t, pkg, format, note)
					officeAssertProp(t, pkg, format, body)
					officeAssertTableRatio(t, pkg, format, body, "A visitor", 0.3)
					if strings.Contains(body.text(), "example.invalid/schema") {
						t.Fatal("visible schema")
					}
					if format == "docx" {
						for _, kind := range []string{"b", "i", "u"} {
							if len(body.all(kind)) == 0 {
								t.Error("missing inline", kind)
							}
						}
					} else {
						if len(body.all("span")) < 4 {
							t.Error("inline styles missing")
						}
					}
					unchanged, err := os.ReadFile(src)
					if err != nil || string(unchanged) != rich {
						t.Fatal("source modified", err)
					}
					for toggle, absent := range map[string]string{"stage-directions": "Night direction.", "frontmatter": "Opening summary.", "character-table": "A visitor", "footnotes": "Note content.", "transitions": "CUT TO"} {
						cfg := "render:\n  " + toggle + ": false\n"
						writeAppFile(t, filepath.Join(dir, "script.yaml"), cfg)
						_, stderr, err := officeRun(t, binary, dir, env, "convert", src, target)
						if err != nil {
							t.Fatalf("toggle %s: %v %s", toggle, err, stderr)
						}
						raw, err = os.ReadFile(target)
						if err != nil {
							t.Fatal(err)
						}
						pkg = officePackage(t, raw)
						allText := ""
						for _, n := range pkg {
							allText += n.text()
						}
						if strings.Contains(allText, absent) {
							t.Errorf("hidden %s still present", toggle)
						}
						got, err := os.ReadFile(filepath.Join(dir, "script.yaml"))
						if err != nil || string(got) != cfg {
							t.Fatal("config modified", err)
						}
					}
				})
			}
		}
	})
	for _, content := range []string{"", "* ACT ONE\n**** CÁIT\nHello.\n"} {
		for _, format := range []string{"docx", "odt"} {
			dir := t.TempDir()
			src := filepath.Join(dir, "empty.org")
			writeAppFile(t, src, content)
			raw, stderr, err := officeRun(t, binary, dir, officeEnv(t, t.TempDir(), nil), "convert", src, "--to", format)
			if err != nil {
				t.Fatalf("empty/untitled %s: %v %s", format, err, stderr)
			}
			officePackage(t, raw)
		}
	}
}

func TestRT054_2OfficeNativeFormatting(t *testing.T) {
	binary := officeBuild(t)
	for _, format := range []string{"docx", "odt"} {
		t.Run(format, func(t *testing.T) {
			dir := t.TempDir()
			home := t.TempDir()
			src := filepath.Join(dir, "play.org")
			writeAppFile(t, src, officeSources[".org"])
			writeAppFile(t, filepath.Join(home, ".config", "first-folio", "script.yaml"), "folio:\n  positioning:\n    speech:\n      dialogue:\n        font:\n          family: Example Family\n          size: 13pt\n          style: italic\n          stretch: 90%\n          letter-spacing: 0.2pt\n")
			writeAppFile(t, filepath.Join(dir, "script.yaml"), "folio:\n  page: a5\n  margin: 18mm\n  positioning:\n    speech:\n      dialogue:\n        font:\n          weight: bold\n    scene-header:\n      space-before: 10pt\n      space-after: 5pt\n")
			raw, stderr, err := officeRun(t, binary, dir, officeEnv(t, home, nil), "convert", src, "--to", format, "--font-size", "11pt")
			if err != nil {
				t.Fatalf("%v %s", err, stderr)
			}
			pkg := officePackage(t, raw)
			body := officeBody(t, pkg, format)
			dialogue := officeStyle(t, pkg, format, "Folio Dialogue")
			if format == "docx" {
				officeAssertAttr(t, dialogue, "rFonts", "ascii", "Example Family")
				officeAssertAttr(t, dialogue, "sz", "val", "26")
				officeAssertAttr(t, dialogue, "w", "val", "90")
				officeAssertAttr(t, dialogue, "spacing", "val", "4")
				if len(dialogue.all("b")) == 0 || len(dialogue.all("i")) == 0 {
					t.Error("missing inherited bold/italic")
				}
				officeAssertAttr(t, body, "pgSz", "w", "8391")
				officeAssertAttr(t, body, "pgMar", "left", "1020")
				scene := officeStyle(t, pkg, format, "Folio Scene")
				officeAssertAttr(t, scene, "spacing", "before", "200")
				officeAssertAttr(t, scene, "spacing", "after", "100")
			} else {
				officeAssertAttr(t, dialogue, "text-properties", "font-name", "Example Family")
				officeAssertAttr(t, dialogue, "text-properties", "font-size", "13pt")
				officeAssertAttr(t, dialogue, "text-properties", "font-weight", "bold")
				officeAssertAttr(t, dialogue, "text-properties", "font-style", "italic")
				officeAssertAttr(t, dialogue, "text-properties", "text-scale", "90%")
				officeAssertAttr(t, dialogue, "text-properties", "letter-spacing", "0.2pt")
				officeAssertAttr(t, pkg["styles.xml"], "page-layout-properties", "page-width", "148mm")
				officeAssertAttr(t, pkg["styles.xml"], "page-layout-properties", "margin-left", "18mm")
			}
			if len(body.all("tbl"))+len(body.all("table")) == 0 {
				t.Fatal("British same-line speech has no native columns")
			}
			for _, style := range []string{"us", "screenplay"} {
				pkg = officeExport(t, binary, ".org", format, style)
				body = officeBody(t, pkg, format)
				if len(body.all("tbl"))+len(body.all("table")) != 0 {
					t.Error("separate-line speech unexpectedly tabulated")
				}
			}
		})
	}
}

func TestRT054_2OfficeTitlesAndRunningMatter(t *testing.T) {
	binary := officeBuild(t)
	for _, format := range []string{"docx", "odt"} {
		for _, compact := range []bool{false, true} {
			t.Run(format+map[bool]string{true: "/compact", false: "/dedicated"}[compact], func(t *testing.T) {
				dir := t.TempDir()
				src := filepath.Join(dir, "play.org")
				writeAppFile(t, src, "#+TITLE: Title\n#+SUBTITLE: Subtitle\n#+AUTHOR: Author\n#+DATE: A Date\n#+VERSION: Version\n* Introduction\nSome introduction.\n"+strings.Repeat("Summary text.\n", 30)+"* ACT ONE\n** SCENE ONE\n**** CÁIT\n"+strings.Repeat("Long dialogue line.\n", 80)+"* ACT TWO\n**** CÁIT\nEnd.\n")
				writeAppFile(t, filepath.Join(dir, "script.yaml"), "folio:\n  title-page:\n    enabled: "+map[bool]string{true: "false", false: "true"}[compact]+"\n    skip-header: true\n    skip-footer: false\n  page-header:\n    enabled: true\n    format: 'BODY-EVEN [title] [page]'\n    alt-format: 'BODY-ODD [author] [total-pages]'\n    frontmatter-format: 'INTRO-EVEN'\n    alt-frontmatter-format: ''\n    align: left-right\n  page-footer:\n    format: 'FOOT [page]/[total-pages] [unknown] [chapter]'\n    frontmatter-format: 'INTRO-FOOT-EVEN'\n    alt-frontmatter-format: ''\n")
				raw, stderr, err := officeRun(t, binary, dir, officeEnv(t, t.TempDir(), nil), "convert", src, "--to", format)
				if err != nil {
					t.Fatalf("%v %s", err, stderr)
				}
				pkg := officePackage(t, raw)
				body := officeBody(t, pkg, format)
				officeAssertEmptyIntroOdd(t, pkg, format)
				for _, s := range []string{"Title", "Subtitle", "Author", "ACT TWO", "End."} {
					if !strings.Contains(body.text(), s) {
						t.Errorf("missing title/body %s", s)
					}
				}
				all := ""
				for _, n := range pkg {
					all += n.text()
				}
				for _, s := range []string{"A Date", "Version", "BODY-EVEN", "BODY-ODD", "INTRO-EVEN", "FOOT", "[unknown]"} {
					if !strings.Contains(all, s) {
						t.Errorf("missing running variant %s", s)
					}
				}
				if strings.Contains(all, "[chapter]") {
					t.Error("manuscript token leaked")
				}
				if format == "docx" {
					if len(body.all("sectPr")) == 0 {
						t.Error("no native section properties")
					}
					fieldText := ""
					for _, n := range pkg {
						for _, f := range n.all("instrText") {
							fieldText += f.text()
						}
					}
					if !strings.Contains(fieldText, "PAGE") || !strings.Contains(fieldText, "NUMPAGES") {
						t.Error("native page fields absent")
					}
				}
				if format == "odt" {
					styles := pkg["styles.xml"]
					if len(styles.all("master-page")) == 0 {
						t.Error("no master pages")
					}
					if len(styles.all("page-number")) == 0 || len(styles.all("page-count")) == 0 {
						t.Error("native page fields absent")
					}
				}
			})
		}
	}
}

func TestRT054_3OfficeSelectionAndFailure(t *testing.T) {
	binary := officeBuild(t)
	dir := t.TempDir()
	src := filepath.Join(dir, "play.org")
	writeAppFile(t, src, officeSources[".org"])
	for _, format := range []string{"docx", "odt"} {
		raw, stderr, err := officeRun(t, binary, dir, officeEnv(t, t.TempDir(), nil), "convert", src, "--to", format)
		if err != nil || stderr != "" {
			t.Fatalf("stdout %s: %v %s", format, err, stderr)
		}
		officePackage(t, raw)
		input := filepath.Join(dir, "input."+format)
		if err := os.WriteFile(input, raw, 0o600); err != nil {
			t.Fatal(err)
		}
		_, stderr, err = officeRun(t, binary, dir, officeEnv(t, t.TempDir(), nil), "convert", input, "--to", "md")
		if err == nil || !strings.Contains(stderr, "write-only") {
			t.Errorf("office input should identify write-only: %v %s", err, stderr)
		}
		cfg := "folio:\n  default-format: " + format + "\n"
		writeAppFile(t, filepath.Join(dir, "script.yaml"), cfg)
		raw, stderr, err = officeRun(t, binary, dir, officeEnv(t, t.TempDir(), nil), "convert", src)
		if err != nil {
			t.Fatalf("default: %v %s", err, stderr)
		}
		officePackage(t, raw)
		target := filepath.Join(dir, "out."+format)
		_, stderr, err = officeRun(t, binary, dir, officeEnv(t, t.TempDir(), nil), "convert", src, target, "--to", "md")
		if err != nil {
			t.Fatalf("extension precedence: %v %s", err, stderr)
		}
		data, err := os.ReadFile(target)
		if err != nil {
			t.Fatal(err)
		}
		officePackage(t, data)
	}
	writeAppFile(t, filepath.Join(dir, "script.yaml"), "folio:\n  positioning:\n    speech:\n      dialogue:\n        font:\n          weight: 650\n")
	target := filepath.Join(dir, "kept.docx")
	writeAppFile(t, target, "OLD TARGET")
	raw, stderr, err := officeRun(t, binary, dir, officeEnv(t, t.TempDir(), nil), "convert", src, target)
	if err == nil || len(raw) != 0 || !strings.Contains(stderr, "folio.positioning.speech.dialogue.font.weight") {
		t.Errorf("unsupported formatting diagnostic: %v %s", err, stderr)
	}
	data, err := os.ReadFile(target)
	if err != nil || string(data) != "OLD TARGET" {
		t.Fatal("target damaged", err)
	}
	writeAppFile(t, filepath.Join(dir, "script.yaml"), "folio:\n  page: unsupported-paper\n")
	_, stderr, err = officeRun(t, binary, dir, officeEnv(t, t.TempDir(), nil), "convert", src, target)
	if err == nil || !strings.Contains(stderr, "folio.page") {
		t.Errorf("paper diagnostic: %v %s", err, stderr)
	}
	writeAppFile(t, filepath.Join(dir, "script.yaml"), "folio:\n  page-header:\n    enabled: false\n    font:\n      weight: 650\n")
	raw, stderr, err = officeRun(t, binary, dir, officeEnv(t, t.TempDir(), nil), "convert", src, "--to", "docx")
	if err != nil {
		t.Fatalf("inactive role: %v %s", err, stderr)
	}
	officePackage(t, raw)
	link := filepath.Join(dir, "link.docx")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	_, stderr, err = officeRun(t, binary, dir, officeEnv(t, t.TempDir(), nil), "convert", src, link)
	if err == nil || !strings.Contains(stderr, "symlink") {
		t.Errorf("symlink accepted: %v %s", err, stderr)
	}
	tools := t.TempDir()
	testBinary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(testBinary, filepath.Join(tools, "pandoc")); err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"missing", "exit", "malformed", "timeout"} {
		t.Run(mode, func(t *testing.T) {
			extra := map[string]string{"PATH": tools, "FOLIO_TEST_PANDOC_FAILURE": mode}
			if mode == "missing" {
				extra = map[string]string{"PATH": t.TempDir()}
			}
			env := officeEnv(t, t.TempDir(), extra)
			writeAppFile(t, target, "OLD TARGET")
			start := time.Now()
			raw, stderr, err := officeRun(t, binary, dir, env, "convert", src, target)
			if err == nil || len(raw) != 0 || stderr == "" {
				t.Fatalf("failure %s not reported: %v %s", mode, err, stderr)
			}
			if mode == "timeout" && (time.Since(start) < 110*time.Second || time.Since(start) > 130*time.Second) {
				t.Error("120-second deadline not applied")
			}
			got, err := os.ReadFile(target)
			if err != nil || string(got) != "OLD TARGET" {
				t.Fatal("target damaged", err)
			}
			for _, e := range env {
				if strings.HasPrefix(e, "TMPDIR=") {
					entries, err := os.ReadDir(strings.TrimPrefix(e, "TMPDIR="))
					if err != nil || len(entries) != 0 {
						t.Fatal("scratch leaked", entries, err)
					}
				}
			}
		})
	}
	hostile := "#+TITLE: <tag> & \"quote\"\n* ACT ONE\n**** CÁIT\n![remote](https://example.invalid/secret) <script>literal</script> #raw $(command).\n"
	writeAppFile(t, src, hostile)
	raw, stderr, err = officeRun(t, binary, dir, officeEnv(t, t.TempDir(), nil), "convert", src, "--to", "docx")
	if err != nil {
		t.Fatalf("hostile text: %v %s", err, stderr)
	}
	pkg := officePackage(t, raw)
	if !strings.Contains(officeBody(t, pkg, "docx").text(), "<script>literal</script>") {
		t.Error("source text became markup")
	}
	for name, n := range pkg {
		if strings.Contains(name, "vba") || len(n.all("image")) > 0 {
			t.Error("active/media package member", name)
		}
	}
}

func TestRT054_3OfficeTerminalGuard(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("platform PTY command currently macOS-specific")
	}
	binary := officeBuild(t)
	dir := t.TempDir()
	src := filepath.Join(dir, "play.org")
	writeAppFile(t, src, officeSources[".org"])
	for _, format := range []string{"docx", "odt"} {
		cmd := exec.Command("/usr/bin/script", "-q", "/dev/null", binary, "convert", src, "--to", format)
		cmd.Env = officeEnv(t, t.TempDir(), nil)
		raw, _ := cmd.CombinedOutput()
		// script itself need not propagate its child's status; inspect the child diagnostic.
		if !bytes.Contains(raw, []byte("Refusing to write to terminal")) {
			t.Fatalf("terminal guard absent: %s", raw)
		}
		cmd = exec.Command("/usr/bin/script", "-q", "/dev/null", binary, "convert", src, "--to", format, "--force")
		cmd.Env = officeEnv(t, t.TempDir(), nil)
		raw, err := cmd.CombinedOutput()
		if err != nil || !bytes.Contains(raw, []byte("PK")) || bytes.Contains(raw, []byte("Refusing")) {
			t.Fatalf("force failed: %v %.100s", err, raw)
		}
	}
}

func TestRT054_4ExistingScriptRoutesDoNotRequirePandoc(t *testing.T) {
	binary := officeBuild(t)
	dir := t.TempDir()
	src := filepath.Join(dir, "play.org")
	writeAppFile(t, src, officeSources[".org"])
	tools := t.TempDir()
	typst, err := exec.LookPath("typst")
	if err != nil {
		t.Fatal("preservation requires Typst", err)
	}
	if err := os.Symlink(typst, filepath.Join(tools, "typst")); err != nil {
		t.Fatal(err)
	}
	for _, ext := range []string{"md", "org", "fountain", "typ", "pdf"} {
		_, stderr, err := officeRun(t, binary, dir, officeEnv(t, t.TempDir(), map[string]string{"PATH": tools}), "convert", src, filepath.Join(dir, "output."+ext))
		if err != nil {
			t.Errorf("%s depended on Pandoc: %v %s", ext, err, stderr)
		}
	}
}
