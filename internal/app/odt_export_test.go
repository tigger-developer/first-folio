// ABOUTME: Verifies configured ODT export through the public conversion command.
// ABOUTME: Covers editable layout and safe failure without altering DOCX expectations.
package app

import (
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestRT058_1ODTSourceStyleMatrix(t *testing.T) {
	binary := officeBuild(t)
	for _, source := range []string{".org", ".md", ".fountain"} {
		for _, style := range []string{"british", "us", "screenplay"} {
			t.Run(source+"/"+style, func(t *testing.T) {
				pkg := officeExport(t, binary, source, "odt", style)
				body := officeBody(t, pkg, "odt")
				officeAssertStyleUsed(t, pkg, "odt", "Folio Dialogue")
				for name := range pkg {
					if name == "META-INF/manifest.xml" {
						continue
					}
					found := false
					for _, entry := range pkg["META-INF/manifest.xml"].all("file-entry") {
						if entry.attr("full-path") == name {
							found = true
						}
					}
					if !found {
						t.Errorf("manifest omits %s", name)
					}
				}

				previous := -1
				for _, text := range []string{"Matrix Play", "Example Author", "ACT ONE", "SCENE ONE", "Night.", "CÁIT", "softly", "Hello.", "Second line."} {
					at := strings.Index(body.text(), text)
					if at <= previous {
						t.Fatalf("missing/out of order %q in %q", text, body.text())
					}
					previous = at
				}
				if (len(body.all("table")) > 0) != (style == "british") {
					t.Fatal("wrong speech structure")
				}
				pageWidth := "595.3pt"
				if style == "us" {
					pageWidth = "612pt"
				}
				officeAssertAttr(t, pkg["styles.xml"], "page-layout-properties", "page-width", pageWidth)
			})
		}
	}
}

func TestRT058_2ODTConfigurationAndContent(t *testing.T) {
	binary := officeBuild(t)
	dir := t.TempDir()
	src := filepath.Join(dir, "play.org")
	writeAppFile(t, src, "#+TITLE: Rich & <editable>\n* CHARACTERS\n| CÁIT | A visitor |\n* Introduction\nOpening summary.\n* ACT ONE\n** SCENE ONE\n*** First direction.\n*** Second direction.\n**** CÁIT (quietly)\n**Bold text** and /italic text/ and _underlined text_ and =literal text=.[fn:note]\n*** Between speeches.\n**** LIAM\nLast words.\n***** BLACKOUT\n[fn:note] Note content. **Bold note**.\n* ACT TWO\n**** CÁIT\nEnd.\n")
	writeAppFile(t, filepath.Join(dir, "script.yaml"), "folio:\n  page: a5\n  margin: 18mm\n  positioning:\n    speech:\n      dialogue:\n        wrap-indent: 30mm\n        font:\n          family: DejaVu Sans Mono\n          size: 13pt\n          weight: bold\n          stretch: 90%\n          letter-spacing: 0.2pt\n  page-header:\n    enabled: true\n    format: 'EVEN [page]'\n    alt-format: 'ODD [total-pages]'\n    frontmatter-format: INTRO\n")
	raw, stderr, err := officeRun(t, binary, dir, officeEnv(t, t.TempDir(), nil), "convert", src, "--to", "odt")
	if err != nil {
		t.Fatalf("%v %s", err, stderr)
	}
	pkg := officePackage(t, raw)
	body := officeBody(t, pkg, "odt")
	styles := pkg["styles.xml"]
	for _, text := range []string{"Rich & <editable>", "A visitor", "Opening summary.", "Bold text", "italic text", "underlined text", "literal text", "Note content.", "BLACKOUT"} {
		if !strings.Contains(body.text(), text) {
			t.Errorf("missing %q", text)
		}
	}
	if len(body.all("note")) != 1 {
		t.Fatal("native footnote absent")
	}
	merged := 0
	for _, c := range body.all("table-cell") {
		if c.attr("number-columns-spanned") == "2" {
			merged++
		}
	}
	// Pandoc expresses column widths as relative ODF widths. Resolve the active
	// table's columns rather than requiring a particular physical-unit spelling.
	for _, table := range body.all("table") {
		if !strings.Contains(table.text(), "Last words.") {
			continue
		}
		widths := []float64{}
		for _, column := range table.all("table-column") {
			for _, style := range pkg["content.xml"].all("style") {
				if style.attr("name") != column.attr("style-name") {
					continue
				}
				for _, property := range style.all("table-column-properties") {
					width, e := strconv.ParseFloat(strings.TrimSuffix(property.attr("rel-column-width"), "*"), 64)
					if e != nil {
						t.Fatal(e)
					}
					widths = append(widths, width)
				}
			}
		}
		if len(widths) != 2 || math.Abs(widths[0]/(widths[0]+widths[1])-30.0/112) > 0.001 {
			t.Fatalf("configured 30mm cue column lost: %v", widths)
		}
	}
	if !officeBreakBefore(t, pkg, "odt", "ACT TWO") {
		t.Fatal("second act lost its effective page break")
	}
	if merged != 4 {

		t.Fatalf("merged direction/transition rows = %d", merged)
	}
	officeAssertAttr(t, styles, "text-properties", "font-family", "DejaVu Sans Mono")
	officeAssertAttr(t, styles, "text-properties", "text-scale", "90%")
	officeAssertAttr(t, styles, "text-properties", "letter-spacing", "0.2pt")
	officeAssertAttr(t, styles, "page-layout-properties", "page-width", "419.55pt")
	officeAssertAttr(t, styles, "paragraph-properties", "break-before", "page")
	if len(styles.all("page-number")) == 0 || len(styles.all("page-count")) == 0 {
		t.Fatal("live fields absent")
	}
	if !strings.Contains(stderr, "frontmatter") {
		t.Fatal("missing limitation diagnostic")
	}
	for _, toggle := range []string{"stage-directions", "frontmatter", "character-table", "footnotes", "transitions"} {
		writeAppFile(t, filepath.Join(dir, "script.yaml"), "render:\n  "+toggle+": false\n")
		raw, stderr, err = officeRun(t, binary, dir, officeEnv(t, t.TempDir(), nil), "convert", src, "--to", "odt")
		if err != nil {
			t.Fatalf("%s: %v %s", toggle, err, stderr)
		}
		absent := map[string]string{"stage-directions": "First direction.", "frontmatter": "Opening summary.", "character-table": "A visitor", "footnotes": "Note content.", "transitions": "BLACKOUT"}[toggle]
		if strings.Contains(officeBody(t, officePackage(t, raw), "odt").text(), absent) {
			t.Errorf("visible %s", toggle)
		}
	}
}

func TestRT058_3ODTSelectionAndPublication(t *testing.T) {
	binary := officeBuild(t)
	dir := t.TempDir()
	src := filepath.Join(dir, "play.org")
	target := filepath.Join(dir, "play.odt")
	writeAppFile(t, src, officeSources[".org"])
	env := officeEnv(t, t.TempDir(), nil)
	writeAppFile(t, filepath.Join(dir, "script.yaml"), "folio:\n  default-format: odt\n")
	raw, stderr, err := officeRun(t, binary, dir, env, "convert", src)
	if err != nil {
		t.Fatalf("default: %v %s", err, stderr)
	}
	officePackage(t, raw)
	_, stderr, err = officeRun(t, binary, dir, env, "convert", src, target, "--to", "md")
	if err != nil {
		t.Fatalf("extension: %v %s", err, stderr)
	}
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	officePackage(t, data)
	_, stderr, err = officeRun(t, binary, dir, env, "convert", target, "--to", "md")
	if err == nil || !strings.Contains(stderr, "write-only") {
		t.Fatalf("input: %v %s", err, stderr)
	}
	writeAppFile(t, target, "OLD TARGET")
	_, stderr, err = officeRun(t, binary, dir, officeEnv(t, t.TempDir(), map[string]string{"PATH": t.TempDir()}), "convert", src, target)
	if err == nil || !strings.Contains(stderr, "pandoc") || readAppFile(t, target) != "OLD TARGET" {
		t.Fatalf("missing tool: %v %s", err, stderr)
	}
	for _, kind := range []string{"symlink", "hardlink"} {
		path := filepath.Join(dir, kind+".odt")
		if kind == "symlink" {
			err = os.Symlink(target, path)
		} else {
			err = os.Link(src, path)
		}
		if err != nil {
			t.Fatal(err)
		}
		_, stderr, err = officeRun(t, binary, dir, env, "convert", src, path)
		if err == nil {
			t.Fatalf("accepted %s: %s", kind, stderr)
		}
	}
	if readAppFile(t, src) != officeSources[".org"] {
		t.Fatal("source changed")
	}
}
