// ABOUTME: Exercises the bounded DOCX spike through the installed-shaped CLI.
// ABOUTME: Checks editable structures, layout, local-only output and safe publication.
package app

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestSpikeDOCXSourceStyleMatrix(t *testing.T) {
	binary := spikeDOCXBuild(t)
	for _, source := range []string{".org", ".md", ".fountain"} {
		for _, style := range []string{"british", "us", "screenplay"} {
			t.Run(source+"/"+style, func(t *testing.T) {
				pkg := officeExport(t, binary, source, "docx", style)
				body := officeBody(t, pkg, "docx")
				previous := -1
				for _, text := range []string{"Matrix Play", "Example Author", "ACT ONE", "SCENE ONE", "Night.", "CÁIT", "softly", "Hello.", "Second line."} {
					at := strings.Index(body.text(), text)
					if at <= previous {
						t.Fatalf("missing or out of order %q in %q", text, body.text())
					}
					previous = at
				}
				if got := len(body.all("tbl")); (style == "british") != (got > 0) {
					t.Fatalf("speech tables %d for %s", got, style)
				}
				pageWidth, margin := "11906", "1417"
				if style == "us" {
					pageWidth, margin = "12240", "1440"
				}
				officeAssertAttr(t, body, "pgSz", "w", pageWidth)
				officeAssertAttr(t, body, "pgMar", "left", margin)
				officeAssertStyleUsed(t, pkg, "docx", "Folio Dialogue")
			})
		}
	}
}

func TestSpikeDOCXRichContentAndConfig(t *testing.T) {
	binary := spikeDOCXBuild(t)
	dir := t.TempDir()
	src := filepath.Join(dir, "rich.org")
	text := "#+TITLE: Rich & <editable>\n#+SCHEMA: https://example.invalid/schema\n* CHARACTERS\n| CÁIT | A visitor |\n* Introduction\nOpening summary.\n* ACT ONE\n** SCENE ONE\n*** Night direction.\n**** CÁIT (quietly)\n**Bold text** and *italic text* and ***combined text*** and _underlined text_ and =literal text=.[fn:note]\n*** Prop card :prop:\n***** CUT TO\n[fn:note] Note content. **Bold note** and *italic note*.\n"
	writeAppFile(t, src, text)
	home := t.TempDir()
	writeAppFile(t, filepath.Join(home, ".config", "first-folio", "script.yaml"), "folio:\n  positioning:\n    speech:\n      dialogue:\n        font:\n          family: Example Family\n          size: 13pt\n          style: italic\n          stretch: 90%\n          letter-spacing: 0.2pt\n")
	configuration := "folio:\n  page: a5\n  margin: 18mm\n  positioning:\n    speech:\n      dialogue:\n        font:\n          weight: bold\n"
	writeAppFile(t, filepath.Join(dir, "script.yaml"), configuration)
	raw, stderr, err := officeRun(t, binary, dir, officeEnv(t, home, nil), "convert", src, "--to", "docx")
	if err != nil {
		t.Fatalf("%v: %s", err, stderr)
	}
	pkg := officePackage(t, raw)
	body := officeBody(t, pkg, "docx")
	for _, want := range []string{"Rich & <editable>", "A visitor", "Opening summary.", "Night direction.", "Bold text", "italic text", "combined text", "underlined text", "literal text", "Prop card", "CUT TO"} {
		if !strings.Contains(body.text(), want) {
			t.Errorf("missing %q", want)
		}
	}
	for _, property := range []string{"b", "i", "u"} {
		if len(body.all(property)) == 0 {
			t.Errorf("missing inline %s", property)
		}
	}
	officeAssertNoteLink(t, pkg, "docx")
	officeAssertInlineNoteFormatting(t, pkg, "docx", pkg["word/footnotes.xml"])
	officeAssertTableRatio(t, pkg, "docx", body, "A visitor", 0.3)
	dialogue := officeStyle(t, pkg, "docx", "Folio Dialogue")
	officeAssertAttr(t, dialogue, "rFonts", "ascii", "Example Family")
	officeAssertAttr(t, dialogue, "sz", "val", "26")
	officeAssertAttr(t, dialogue, "w", "val", "90")
	officeAssertAttr(t, dialogue, "spacing", "val", "4")
	officeAssertAttr(t, body, "pgSz", "w", "8391")
	if strings.Contains(body.text(), "example.invalid/schema") {
		t.Fatal("visible schema")
	}
	if readAppFile(t, src) != text || readAppFile(t, filepath.Join(dir, "script.yaml")) != configuration {
		t.Fatal("input changed")
	}
	for toggle, absent := range map[string]string{"stage-directions": "Night direction.", "frontmatter": "Opening summary.", "character-table": "A visitor", "footnotes": "Note content.", "transitions": "CUT TO"} {
		t.Run(toggle, func(t *testing.T) {
			writeAppFile(t, filepath.Join(dir, "script.yaml"), "render:\n  "+toggle+": false\n")
			raw, stderr, err := officeRun(t, binary, dir, officeEnv(t, home, nil), "convert", src, "--to", "docx")
			if err != nil {
				t.Fatalf("%v: %s", err, stderr)
			}
			for name, node := range officePackage(t, raw) {
				if strings.Contains(node.text(), absent) {
					t.Errorf("hidden text in %s", name)
				}
			}
		})
	}
}

func TestSpikeDOCXSelectionAndFailures(t *testing.T) {
	binary := spikeDOCXBuild(t)
	dir := t.TempDir()
	src := filepath.Join(dir, "play.org")
	writeAppFile(t, src, officeSources[".org"])
	env := officeEnv(t, t.TempDir(), map[string]string{"PATH": t.TempDir()})
	// No external executable is available: export must remain a local transform.
	raw, stderr, err := officeRun(t, binary, dir, env, "convert", src, "--to", "docx")
	if err != nil {
		t.Fatalf("%v: %s", err, stderr)
	}
	officePackage(t, raw)
	target := filepath.Join(dir, "play.docx")
	writeAppFile(t, filepath.Join(dir, "script.yaml"), "folio:\n  default-format: docx\n")
	raw, stderr, err = officeRun(t, binary, dir, env, "convert", src)
	if err != nil {
		t.Fatalf("default: %v: %s", err, stderr)
	}
	officePackage(t, raw)
	_, stderr, err = officeRun(t, binary, dir, env, "convert", src, target, "--to", "md")
	if err != nil {
		t.Fatalf("target precedence: %v: %s", err, stderr)
	}
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	officePackage(t, data)
	_, stderr, err = officeRun(t, binary, dir, env, "convert", target, "--to", "md")
	if err == nil || !strings.Contains(stderr, "write-only") {
		t.Fatalf("DOCX input: %v %s", err, stderr)
	}
	for _, content := range []string{"", "* ACT ONE\n**** CÁIT\nHello.\n"} {
		writeAppFile(t, src, content)
		raw, stderr, err = officeRun(t, binary, dir, env, "convert", src, "--to", "docx")
		if err != nil {
			t.Fatalf("empty/untitled: %v %s", err, stderr)
		}
		officePackage(t, raw)
	}
	for _, cfg := range []string{"folio:\n  page: impossible\n", "folio:\n  margin: 900mm\n", "folio:\n  positioning:\n    speech:\n      dialogue:\n        font:\n          weight: 650\n"} {
		writeAppFile(t, src, officeSources[".org"])
		writeAppFile(t, filepath.Join(dir, "script.yaml"), cfg)
		writeAppFile(t, target, "OLD TARGET")
		raw, stderr, err = officeRun(t, binary, dir, env, "convert", src, target)
		if err == nil || len(raw) != 0 || stderr == "" {
			t.Fatalf("unsafe config accepted: %v %s", err, stderr)
		}
		if readAppFile(t, target) != "OLD TARGET" {
			t.Fatal("old target changed")
		}
	}
	writeAppFile(t, filepath.Join(dir, "script.yaml"), "{}\n")
	link := filepath.Join(dir, "link.docx")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	_, stderr, err = officeRun(t, binary, dir, env, "convert", src, link)
	if err == nil || !strings.Contains(stderr, "symlink") {
		t.Fatalf("symlink: %v %s", err, stderr)
	}
	before := readAppFile(t, src)
	_, stderr, err = officeRun(t, binary, dir, env, "convert", src, "--to", "docx")
	if err != nil {
		t.Fatalf("recovery: %v %s", err, stderr)
	}
	if readAppFile(t, src) != before {
		t.Fatal("source modified")
	}
	writeAppFile(t, src, "* ACT ONE\n**** CÁIT\n<tag> & ![remote](https://example.invalid/secret) $(command).\n")
	raw, stderr, err = officeRun(t, binary, dir, env, "convert", src, "--to", "docx")
	if err != nil {
		t.Fatalf("literal content: %v %s", err, stderr)
	}
	pkg := officePackage(t, raw)
	if !strings.Contains(officeBody(t, pkg, "docx").text(), "<tag> & ![remote]") {
		t.Fatal("text became markup")
	}
	for name, node := range pkg {
		if strings.Contains(name, "vba") || len(node.all("drawing")) > 0 {
			t.Fatal("unexpected active content")
		}
		for _, rel := range node.all("Relationship") {
			if rel.attr("TargetMode") == "External" {
				t.Fatal("external relationship")
			}
		}
	}
	writeAppFile(t, src, "* ACT ONE\n**** CÁIT\ninvalid\x01text\n")
	raw, stderr, err = officeRun(t, binary, dir, env, "convert", src, "--to", "docx")
	if err == nil || len(raw) != 0 || !strings.Contains(stderr, "XML") {
		t.Fatalf("invalid XML input: %v %s", err, stderr)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".folio-docx-") {
			t.Fatal("staging file leaked")
		}
	}
}

func TestSpikeDOCXBrokenOutputStream(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	src := filepath.Join(dir, "play.org")
	writeAppFile(t, src, officeSources[".org"])
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := reader.Close(); err != nil {
		t.Fatal(err)
	}
	defer writer.Close()
	var stderr bytes.Buffer
	status := Run([]string{"convert", src, "--to", "docx"}, strings.NewReader(""), writer, &stderr)
	if status == 0 {
		t.Fatal("broken stdout reported success")
	}
	if !strings.Contains(stderr.String(), "broken pipe") {
		t.Fatalf("expected write failure, got %q", stderr.String())
	}
}

func TestSpikeDOCXTitleRunningMatterAndLimits(t *testing.T) {
	binary := spikeDOCXBuild(t)
	dir := t.TempDir()
	src := filepath.Join(dir, "play.org")
	env := officeEnv(t, t.TempDir(), nil)
	for _, compact := range []bool{false, true} {
		content := "#+TITLE: Page test\n* ACT ONE\n**** CÁIT\nHello.\n* ACT TWO\n**** LIAM\nEnd.\n"
		writeAppFile(t, src, content)
		enabled := "true"
		if compact {
			enabled = "false"
		}
		writeAppFile(t, filepath.Join(dir, "script.yaml"), "folio:\n  title-page:\n    enabled: "+enabled+"\n  page-header:\n    enabled: true\n    format: 'EVEN [title] [page]'\n    alt-format: 'ODD [author] [total-pages]'\n    align: left-right\n    frontmatter-format: 'INTRO'\n")
		raw, stderr, err := officeRun(t, binary, dir, env, "convert", src, "--to", "docx")
		if err != nil {
			t.Fatalf("%v %s", err, stderr)
		}
		if !strings.Contains(stderr, "frontmatter-format") {
			t.Fatal("missing approximation diagnostic")
		}
		pkg := officePackage(t, raw)
		body := officeBody(t, pkg, "docx")
		first := officeParagraph(t, body, "ACT ONE")
		second := officeParagraph(t, body, "ACT TWO")
		count := len(first.all("pageBreakBefore"))
		want := 1
		if compact {
			want = 0
		}
		if count != want {
			t.Errorf("first act has %d page-break properties, want %d", count, want)
		}
		if len(second.all("pageBreakBefore")) != 1 {
			t.Error("second act break lost")
		}
		if strings.Contains(pkg["word/header3.xml"].text(), "ODD") {
			t.Error("first-page header not suppressed")
		}
		if !strings.Contains(pkg["word/header1.xml"].text(), "ODD") || !strings.Contains(pkg["word/header2.xml"].text(), "EVEN") {
			t.Error("facing headers absent")
		}
		for _, part := range []string{"word/header1.xml", "word/header2.xml", "word/footer1.xml"} {
			if len(pkg[part].all("instrText")) == 0 {
				t.Errorf("missing live page field in %s", part)
			}
		}
	}
	writeAppFile(t, filepath.Join(dir, "script.yaml"), "{}\n")
	writeAppFile(t, src, strings.Repeat("a", (32<<20)+1))
	raw, stderr, err := officeRun(t, binary, dir, env, "convert", src, "--to", "docx")
	if err == nil || len(raw) != 0 || !strings.Contains(stderr, "32 MiB") {
		t.Fatalf("source bound: %v %s", err, stderr)
	}
	writeAppFile(t, src, "* ACT ONE\n**** CÁIT\n"+strings.Repeat("**x** ", 100001))
	raw, stderr, err = officeRun(t, binary, dir, env, "convert", src, "--to", "docx")
	if err == nil || len(raw) != 0 || !strings.Contains(stderr, "100000") {
		t.Fatalf("inline bound: %v %s", err, stderr)
	}
	writeAppFile(t, src, officeSources[".org"])
	alias := filepath.Join(dir, "alias.docx")
	if err := os.Link(src, alias); err != nil {
		t.Fatal(err)
	}
	_, stderr, err = officeRun(t, binary, dir, env, "convert", src, alias)
	if err == nil || !strings.Contains(stderr, "same file") {
		t.Fatalf("source alias: %v %s", err, stderr)
	}
	if readAppFile(t, src) != officeSources[".org"] {
		t.Fatal("source damaged through alias")
	}
}

func TestSpikeDOCXTerminalGuard(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("uses macOS PTY command")
	}
	binary := spikeDOCXBuild(t)
	dir := t.TempDir()
	src := filepath.Join(dir, "play.org")
	writeAppFile(t, src, officeSources[".org"])
	for _, force := range []bool{false, true} {
		args := []string{"-q", "/dev/null", binary, "convert", src, "--to", "docx"}
		if force {
			args = append(args, "--force")
		}
		command := exec.Command("/usr/bin/script", args...)
		command.Env = officeEnv(t, t.TempDir(), nil)
		raw, err := command.CombinedOutput()
		if err != nil {
			exit, ok := err.(*exec.ExitError)
			if force || !ok || exit.ExitCode() != 1 {
				t.Fatalf("PTY: %v %.200s", err, raw)
			}
		}
		if force {
			if !bytes.Contains(raw, []byte("PK")) {
				t.Fatal("forced binary absent")
			}
		} else if !bytes.Contains(raw, []byte("Refusing to write to terminal")) {
			t.Fatalf("terminal guard absent: %.200s", raw)
		}
	}
}

func TestSpikeDOCXContinuousSceneTable(t *testing.T) {
	binary := spikeDOCXBuild(t)
	dir := t.TempDir()
	src := filepath.Join(dir, "scene.org")
	writeAppFile(t, src, "#+TITLE: Editable scenes\n* ACT ONE\n** SCENE ONE\n*** First direction.\n*** Second direction.\n*** Third direction.\n**** CÁIT (quietly)\nFirst speech.\n*** Between speeches.\n**** LIAM\nSecond speech.\n*** A card :prop:\n***** BLACKOUT\n** SCENE TWO\n*** Next scene direction.\n**** CÁIT\nLast speech.\n")
	for _, style := range []string{"british", "us", "screenplay"} {
		t.Run(style, func(t *testing.T) {
			raw, stderr, err := officeRun(t, binary, dir, officeEnv(t, t.TempDir(), nil), "convert", src, "--to", "docx", "--style", style)
			if err != nil {
				t.Fatalf("%v %s", err, stderr)
			}
			pkg := officePackage(t, raw)
			body := officeBody(t, pkg, "docx")
			speaker := officeStyle(t, pkg, "docx", "Folio Speaker")
			if got := len(speaker.all("keepNext")); (style == "british" && got != 0) || (style != "british" && got != 1) {
				t.Fatalf("speaker keep-next joins table rows or separates a paragraph cue: %d", got)
			}
			tables := body.all("tbl")
			if style != "british" {
				if len(tables) != 0 {
					t.Fatal("separate-line format gained a scene table")
				}
				return
			}
			if len(tables) != 2 {
				t.Fatalf("got %d tables, want one per scene", len(tables))
			}
			for i, wants := range [][]int{{1, 1, 1, 2, 1, 2, 1, 1}, {1, 2}} {
				rows := tables[i].all("tr")
				if len(rows) != len(wants) {
					t.Fatalf("scene %d has %d rows, want %d", i, len(rows), len(wants))
				}
				for j, want := range wants {
					cells := rows[j].all("tc")
					if len(cells) != want {
						t.Fatalf("scene %d row %d has %d cells, want %d", i, j, len(cells), want)
					}
					if want == 1 {
						officeAssertAttr(t, cells[0], "gridSpan", "val", "2")
					}
				}
			}
			for _, child := range body.Children {
				if child.Name.Local == "p" && strings.Contains(child.text(), "direction.") {
					t.Fatal("stage direction left outside the scene table")
				}
			}
		})
	}
}

func TestSpikeDOCXTableImmediatelyAfterTitle(t *testing.T) {
	binary := spikeDOCXBuild(t)
	dir := t.TempDir()
	src := filepath.Join(dir, "no-heading.org")
	writeAppFile(t, src, "#+TITLE: No scene heading\n**** CÁIT\nHello.\n*** First direction.\n*** Second direction.\n*** Third direction.\n")
	raw, stderr, err := officeRun(t, binary, dir, officeEnv(t, t.TempDir(), nil), "convert", src, "--to", "docx")
	if err != nil {
		t.Fatalf("%v %s", err, stderr)
	}
	body := officeBody(t, officePackage(t, raw), "docx")
	tables := body.all("tbl")
	if len(tables) != 1 || len(tables[0].all("tr")) != 4 {
		t.Fatal("headingless scene is not one four-row table")
	}
	if len(body.all("pageBreakBefore")) != 1 || len(tables[0].all("pageBreakBefore")) != 0 {
		t.Fatal("title break must be on a paragraph outside the table")
	}
}

func TestSpikeDOCXRelativeRootFont(t *testing.T) {
	binary := spikeDOCXBuild(t)
	dir := t.TempDir()
	src := filepath.Join(dir, "relative.org")
	writeAppFile(t, src, "* Introduction\nBody text.\n")
	writeAppFile(t, filepath.Join(dir, "script.yaml"), "folio:\n  font:\n    size: 1.5em\n")
	raw, stderr, err := officeRun(t, binary, dir, officeEnv(t, t.TempDir(), nil), "convert", src, "--to", "docx")
	if err != nil {
		t.Fatalf("%v %s", err, stderr)
	}
	style := officeAssertStyleUsed(t, officePackage(t, raw), "docx", "Folio Body")
	officeAssertAttr(t, style, "sz", "val", "36")
}

// Keep executable coverage outside Go's per-test temporary directory when requested.
// GOFLAGS=-cover instruments officeBuild; this hook affects only test subprocesses.
func spikeDOCXBuild(t *testing.T) string {
	t.Helper()
	if dir := os.Getenv("FOLIO_SPIKE_COVERDIR"); dir != "" {
		t.Setenv("GOCOVERDIR", dir)
	}
	return officeBuild(t)
}
