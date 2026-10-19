// ABOUTME: Tests parser mode semantics and complete novel-scale office exports.
// ABOUTME: Uses public commands and an independent synthetic source/content inventory.
package app

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tigger-developer/first-folio/internal/play"
	"github.com/tigger-developer/first-folio/internal/testfixture"
)

func TestRT063_13_SharedSourceRecognition(t *testing.T) {
	for _, format := range []play.Format{play.FormatOrg, play.FormatMarkdown} {
		heading, table := "* Act One\n** Scene One\n", "| Name | Role |\n| ALICE | Writer |\n"
		if format == play.FormatMarkdown {
			heading = "# Title\n\n## Act One\n### Scene One\n"
		} else {
			heading = "* CHARACTERS\n"
		}
		doc, _, err := play.Parse(format, heading+table, "synthetic")
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, event := range doc.Events {
			if event.Kind == play.EventCharacterTableRow && event.Name == "ALICE" {
				found = true
			}
		}
		if !found {
			t.Errorf("%s convert parsing lost typed cast row", format)
		}
	}
	TestRT063_13_ModeSpecificSemantics(t)
}

func TestRT063_13_ModeSpecificSemantics(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		for _, input := range []string{"md", "org"} {
			t.Run(format+"/"+input, func(t *testing.T) {
				t.Setenv("HOME", t.TempDir())
				dir := t.TempDir()
				body := "## Heading\n\n**BOLD** *ITALIC* `*LITERAL*`\n\n| Name | Role |\n|---|---|\n| Alice | Writer |\n"
				if input == "org" {
					body = "** Heading\n\n*BOLD* /ITALIC/ =*LITERAL*=\n\n| Name | Role |\n|------+------|\n| Alice | Writer |\n"
				}
				proseSource(t, dir, "script.yaml", "folio:\n  manuscript:\n    title-page:\n      enabled: false\n    toc:\n      enabled: false\n")
				pkg := proseConvert(t, proseSource(t, dir, "source."+input, body), filepath.Join(dir, "output."+format))
				text := officeBody(t, pkg, format).text()
				for _, want := range []string{"Heading", "BOLD", "ITALIC", "*LITERAL*", "Alice", "Writer"} {
					if !strings.Contains(text, want) {
						t.Errorf("missing %q in %s", want, text)
					}
				}
				for _, spec := range []struct{ marker, docx, odt, attr, value string }{{"BOLD", "b", "font-weight", "font-weight", "bold"}, {"ITALIC", "i", "font-style", "font-style", "italic"}} {
					layers := proseInlineLayers(t, pkg, format, spec.marker)
					if format == "docx" {
						active := false
						for _, layer := range layers {
							for _, flag := range layer.all(spec.docx) {
								value := flag.attr("val")
								if value != "0" && value != "false" {
									active = true
								}
							}
						}
						if !active {
							t.Errorf("%s lost its native %s emphasis", spec.marker, spec.docx)
						}
					} else {
						officeAssertEffective(t, layers, "text-properties", spec.attr, spec.value)
					}
				}
			})
		}
	}
}

func TestRT063_13_NovelScaleDOCXWithoutPandoc(t *testing.T)   { novelOfficeCases(t, "docx") }
func TestRT063_13_NovelScaleODTConfiguredWriter(t *testing.T) { novelOfficeCases(t, "odt") }

func novelOfficeCases(t *testing.T, format string) {
	t.Helper()
	for _, input := range []string{"md", "org"} {
		t.Run(input, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			dir := t.TempDir()
			source, expected := testfixture.Novel(input)
			if len(strings.Fields(source)) < 120000 {
				t.Fatal("novel fixture is too short")
			}
			proseSource(t, dir, "script.yaml", testfixture.NovelProfile)
			path := proseSource(t, dir, "novel."+input, source)
			target := filepath.Join(dir, "novel."+format)
			novelConversion(t, path, target, format)
			pkg := prosePackage(t, target, format)
			novelBodyOracle(t, officeBody(t, pkg, format).text(), expected)
			var all strings.Builder
			for _, file := range pkg {
				all.WriteString(file.text())
			}
			novelNotesOracle(t, all.String())
			tableName, listName := "tbl", "numPr"
			if format == "odt" {
				tableName, listName = "table", "list"
			}
			body := officeBody(t, pkg, format)
			if len(body.all(tableName)) < 3 || len(body.all(listName)) < 3 {
				t.Error("novel structured tables/lists were flattened")
			}
		})
	}
}

func novelNotesOracle(t *testing.T, text string) {
	t.Helper()
	for _, i := range []int{0, 5000, 9999} {
		marker := fmt.Sprintf("NOTE%05d", i)
		if strings.Count(text, marker) != 1 {
			t.Errorf("native note missing or duplicated: %s", marker)
		}
	}
}

func novelConversion(t *testing.T, source, target, format string) {
	t.Helper()
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	bin := t.TempDir()
	for _, name := range []string{"pandoc", "typst"} {
		if err := os.Symlink(self, filepath.Join(bin, name)); err != nil {
			t.Fatal(err)
		}
	}
	mode := "forbid"
	realPandoc := ""
	if format == "odt" {
		mode = "writer-only"
		realPandoc, err = exec.LookPath("pandoc")
		if err != nil {
			t.Fatal(err)
		}
	}
	marker := filepath.Join(bin, "pandoc-invoked")
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, self, "-test.run=^TestProseDependencyChild$", "--", "manuscript", source, target)
	cmd.Env = officeEnv(t, t.TempDir(), map[string]string{"PATH": bin, "FOLIO_TEST_PROSE_CHILD": "1", "FOLIO_TEST_PROSE_TOOL": mode, "FOLIO_TEST_REAL_PANDOC": realPandoc, "FOLIO_TEST_PROSE_PANDOC_MARKER": marker, "FOLIO_TEST_PROSE_TYPST_MARKER": filepath.Join(bin, "typst-invoked")})
	diagnostic := proseDiagnostic{cancel: cancel}
	cmd.Stdout, cmd.Stderr = &diagnostic, &diagnostic
	if err := cmd.Run(); err != nil {
		t.Fatalf("novel %s command: %v: %s", format, err, diagnostic.String())
	}
	if format == "docx" {
		if _, err := os.Stat(marker); !os.IsNotExist(err) {
			t.Error("novel DOCX invoked Pandoc")
		}
	}
}

func novelBodyOracle(t *testing.T, body string, expected []string) {
	t.Helper()
	last := 0
	for _, paragraph := range expected {
		if strings.Count(body, paragraph) != 1 {
			t.Fatalf("paragraph missing or duplicated: %q", paragraph)
		}
		at := strings.Index(body[last:], paragraph)
		if at < 0 {
			t.Fatalf("paragraph out of order: %q", paragraph)
		}
		last += at + len(paragraph)
	}
	if strings.Contains(body, "PRIVATE-NOVEL-MARKER") {
		t.Error("private novel content leaked")
	}
	for i := 1; i <= 50; i++ {
		if !strings.Contains(body, fmt.Sprintf("CH%03d", i)) {
			t.Errorf("chapter %d lost", i)
		}
	}
	for _, i := range []int{0, 5000, 9999} {
		for _, prefix := range []string{"BOLD", "ITALIC", "CODE", "LINK", "QUOTE", "LIST", "NESTED", "TABLE", "CELL"} {
			if !strings.Contains(body, fmt.Sprintf("%s%05d", prefix, i)) {
				t.Errorf("structured specimen %s%d lost", prefix, i)
			}
		}
	}
}
