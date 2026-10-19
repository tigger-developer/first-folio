//go:build folio_acceptance

// ABOUTME: Performs ticket-scoped full-novel reader acceptance outside the regression pack.
// ABOUTME: Retains reproducible artefacts when given an OS-allocated acceptance directory.
package app

import (
	"context"
	"fmt"
	"io/fs"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/tigger-developer/first-folio/internal/testfixture"
)

func TestAT063_1_NovelScaleRenderedManuscripts(t *testing.T) {
	root := os.Getenv("FOLIO_W063_ACCEPTANCE_DIR")
	if root == "" {
		root = t.TempDir()
	}
	binary := officeBuild(t)
	started := time.Now()
	for _, input := range []string{"md", "org"} {
		for _, format := range []string{"docx", "odt"} {
			t.Run(input+"/"+format, func(t *testing.T) {
				dir, err := os.MkdirTemp(root, input+"-"+format+"-")
				if err != nil {
					t.Fatal(err)
				}
				t.Logf("acceptance artefacts: %s", dir)
				text, expected := testfixture.Novel(input)
				src := proseSource(t, dir, "novel."+input, text)
				proseSource(t, dir, "script.yaml", testfixture.NovelProfile)
				target := filepath.Join(dir, "novel."+format)
				ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
				defer cancel()
				cmd := exec.CommandContext(ctx, binary, "manuscript", src, target)
				cmd.Dir = dir
				cmd.Env = officeEnv(t, t.TempDir(), nil)
				diagnostic := proseDiagnostic{cancel: cancel}
				cmd.Stdout, cmd.Stderr = &diagnostic, &diagnostic
				if err := cmd.Run(); err != nil {
					t.Fatalf("novel CLI: %v: %s", err, diagnostic.String())
				}
				pkg := prosePackage(t, target, format)
				novelBodyOracle(t, officeBody(t, pkg, format).text(), expected)
				pdf := proseRenderPDFDeadline(t, target, 300*time.Second)
				textPath := filepath.Join(dir, "rendered.txt")
				proseTool(t, "pdftotext", "-layout", pdf, textPath)
				raw, err := os.ReadFile(textPath)
				if err != nil {
					t.Fatal(err)
				}
				novelBodyOracle(t, strings.Join(strings.Fields(string(raw)), " "), expected)
				novelNotesOracle(t, string(raw))
				checkNovelReferences(t, string(raw))
				xmlPath := filepath.Join(dir, "spans.xml")
				proseTool(t, "pdftohtml", "-xml", "-zoom", "1", "-noroundcoord", "-fontfullname", pdf, xmlPath)
				spans, err := os.ReadFile(xmlPath)
				if err != nil {
					t.Fatal(err)
				}
				tree := officeParseXML(t, spans)
				fonts := map[string]*officeXML{}
				for _, font := range tree.all("fontspec") {
					fonts[font.attr("id")] = font
				}
				for _, marker := range []string{"P00000", "P05000", "P09999"} {
					found := false
					for _, span := range tree.all("text") {
						if strings.Contains(span.text(), marker) {
							font := fonts[span.attr("font")]
							family := ""
							if font != nil {
								family = font.attr("family")
								if at := strings.Index(family, "+"); at == 6 {
									family = family[7:]
								}
							}
							if family != "Iosevka-Publish" && family != "IosevkaPublish" {
								t.Errorf("novel specimen %s has unexpected font", marker)
							}
							if font != nil {
								size, e := strconv.ParseFloat(font.attr("size"), 64)
								if e != nil || math.Abs(size-10) > 0.25 {
									t.Errorf("novel specimen %s font size %s, want 10pt", marker, font.attr("size"))
								}
							}
							left, e := strconv.ParseFloat(span.attr("left"), 64)
							if e != nil || math.Abs(left-20*72/25.4) > 1 {
								t.Errorf("novel specimen %s left %s, want 20mm body margin", marker, span.attr("left"))
							}
							found = true
						}
					}
					if !found {
						t.Errorf("no visible measured specimen %s", marker)
					}
				}
				pages := len(strings.Split(string(raw), "\f")) - 1
				for _, page := range []int{1, 3, 4, max(1, pages/2), max(1, pages-1), pages} {
					proseTool(t, "pdftoppm", "-f", strconv.Itoa(page), "-l", strconv.Itoa(page), "-png", "-r", "150", pdf, filepath.Join(dir, fmt.Sprintf("page-%04d", page)))
				}
			})
		}
	}
	var total int64
	if err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() {
			info, e := entry.Info()
			if e != nil {
				return e
			}
			total += info.Size()
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if total > 512<<20 {
		t.Errorf("novel artefact budget exceeded: %d", total)
	}
	if time.Since(started) > 150*time.Minute {
		t.Error("novel aggregate deadline exceeded")
	}
}

func checkNovelReferences(t *testing.T, text string) {
	t.Helper()
	entry := regexp.MustCompile(`^CH([0-9]{3})[.\s]+([0-9]+)$`)
	heading := regexp.MustCompile(`^CH([0-9]{3})$`)
	footer := regexp.MustCompile(`NOVEL-PAGE\s+([0-9]+)`)
	refs, actual := map[string]int{}, map[string]int{}
	for _, page := range strings.Split(text, "\f") {
		for _, line := range strings.Split(page, "\n") {
			line = strings.TrimSpace(line)
			if m := entry.FindStringSubmatch(line); m != nil {
				refs[m[1]], _ = strconv.Atoi(m[2])
			}
			if m := heading.FindStringSubmatch(line); m != nil {
				if f := footer.FindStringSubmatch(page); f != nil {
					actual[m[1]], _ = strconv.Atoi(f[1])
				}
			}
		}
	}
	if len(refs) != 50 || len(actual) != 50 {
		t.Fatalf("novel references incomplete: refs=%d headings=%d", len(refs), len(actual))
	}
	for id, page := range actual {
		if refs[id] != page {
			t.Errorf("chapter %s reference %d differs from displayed heading page %d", id, refs[id], page)
		}
	}
}
