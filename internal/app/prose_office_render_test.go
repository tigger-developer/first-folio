// ABOUTME: Renders office fixtures headlessly with explicit field and index refresh.
// ABOUTME: Measures PDF evidence in private temporary profiles with bounded subprocesses.
package app

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

type proseDiagnostic struct {
	bytes.Buffer
	cancel context.CancelFunc
}

func (b *proseDiagnostic) Write(p []byte) (int, error) {
	if b.Len()+len(p) > 64<<10 {
		n, _ := b.Buffer.Write(p[:(64<<10)-b.Len()])
		b.cancel()
		return n, fmt.Errorf("test subprocess exceeded 64 KiB diagnostics")
	}
	return b.Buffer.Write(p)
}

func proseTool(t *testing.T, name string, args ...string) string {
	t.Helper()
	path, err := exec.LookPath(name)
	if err != nil {
		t.Fatalf("rendered prose regression requires %s on PATH: %v", name, err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, path, args...)
	var output proseDiagnostic
	output.cancel = cancel
	cmd.Stdout, cmd.Stderr = &output, &output
	cmd.WaitDelay = 2 * time.Second
	if err := cmd.Run(); err != nil {
		t.Fatalf("%s failed: %v\n%s", name, err, output.String())
	}
	return output.String()
}

func proseFileURL(path string) string {
	return (&url.URL{Scheme: "file", Path: filepath.ToSlash(path)}).String()
}

// The helper is installed only in a test-owned LibreOffice profile. It is never
// put inside an exported document. Source-document macro execution is disabled.
const proseRefreshBasic = `Sub RefreshFixture
 On Error GoTo Failed
 Dim load(2) As New com.sun.star.beans.PropertyValue
 load(0).Name = "Hidden"
 load(0).Value = True
 load(1).Name = "MacroExecutionMode"
 load(1).Value = com.sun.star.document.MacroExecMode.NEVER_EXECUTE
 load(2).Name = "UpdateDocMode"
 load(2).Value = com.sun.star.document.UpdateDocMode.FULL_UPDATE
 Dim doc As Object, indexes As Object, i As Integer
 doc = StarDesktop.loadComponentFromURL("SOURCE_URL", "_blank", 0, load())
 doc.TextFields.refresh()
 indexes = doc.getDocumentIndexes()
 For i = 0 To indexes.getCount() - 1
  indexes.getByIndex(i).update()
 Next i
 doc.refresh()
 Dim filter(0) As New com.sun.star.beans.PropertyValue
 filter(0).Name = "IsSkipEmptyPages"
 filter(0).Value = False
 Dim props(1) As New com.sun.star.beans.PropertyValue
 props(0).Name = "FilterName"
 props(0).Value = "writer_pdf_Export"
 props(1).Name = "FilterData"
 props(1).Value = filter()
 doc.storeToURL("PDF_URL", props())
 doc.close(True)
 StarDesktop.terminate()
 Exit Sub
Failed:
 Open "ERROR_PATH" For Output As #1
 Print #1, Error$
 Close #1
 StarDesktop.terminate()
End Sub
`

func proseRenderPDF(t *testing.T, source string) string {
	t.Helper()
	dir := t.TempDir()
	profile := filepath.Join(dir, "profile")
	basic := filepath.Join(profile, "user", "basic")
	if err := os.MkdirAll(filepath.Join(basic, "Standard"), 0700); err != nil {
		t.Fatal(err)
	}
	list := `<library:libraries xmlns:library="http://openoffice.org/2000/library"><library:library library:name="Standard" library:link="false"/></library:libraries>`
	lib := `<library:library xmlns:library="http://openoffice.org/2000/library" library:name="Standard" library:readonly="false" library:passwordprotected="false">`
	files := map[string]string{
		"script.xlc": list, "dialog.xlc": list,
		"Standard/script.xlb": lib + `<library:element library:name="Module1"/></library:library>`,
		"Standard/dialog.xlb": lib + `</library:library>`,
	}
	pdf := filepath.Join(dir, "rendered.pdf")
	errorPath := filepath.Join(dir, "refresh-error.txt")
	code := strings.NewReplacer("SOURCE_URL", proseFileURL(source), "PDF_URL", proseFileURL(pdf), "ERROR_PATH", strings.ReplaceAll(errorPath, `"`, `""`)).Replace(proseRefreshBasic)
	var escaped bytes.Buffer
	if err := xml.EscapeText(&escaped, []byte(code)); err != nil {
		t.Fatal(err)
	}
	files["Standard/Module1.xba"] = `<script:module xmlns:script="http://openoffice.org/2000/script" script:name="Module1" script:language="StarBasic">` + escaped.String() + `</script:module>`
	for name, contents := range files {
		if err := os.WriteFile(filepath.Join(basic, name), []byte(contents), 0600); err != nil {
			t.Fatal(err)
		}
	}
	proseTool(t, "soffice", "-env:UserInstallation="+proseFileURL(profile), "--headless", "macro:///Standard.Module1.RefreshFixture")
	if raw, err := os.ReadFile(errorPath); err == nil {
		t.Fatalf("LibreOffice refresh failed: %s", raw)
	} else if !os.IsNotExist(err) {
		t.Fatal(err)
	}
	info, err := os.Stat(pdf)
	if err != nil || info.Size() == 0 {
		t.Fatalf("LibreOffice did not publish a PDF: %v", err)
	}
	return pdf
}

func TestRT063_11_OfficeRenderedManuscriptFonts(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		t.Run(format, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			dir := t.TempDir()
			raw, err := os.ReadFile("../../testdata/prose-office/typography.md")
			if err != nil {
				t.Fatal(err)
			}
			source := proseSource(t, dir, "source.md", string(raw))
			target := filepath.Join(dir, "output."+format)
			proseConvert(t, source, target)
			pdf := proseRenderPDF(t, target)
			xmlPath := filepath.Join(dir, "font-spans.xml")
			proseTool(t, "pdftohtml", "-xml", "-zoom", "1", "-noroundcoord", "-fontfullname", pdf, xmlPath)
			xmlRaw, err := os.ReadFile(xmlPath)
			if err != nil {
				t.Fatal(err)
			}
			tree := officeParseXML(t, xmlRaw)
			fonts := map[string]string{}
			for _, f := range tree.all("fontspec") {
				fonts[f.attr("id")] = f.attr("family")
			}
			for _, specimen := range []struct{ marker, family string }{
				{"OPENING-SAMPLE", "Iosevka-Publish"},
				{"HEADING-SAMPLE", "Iosevka-Publish-Bold-Extended"},
				{"QUOTE-SAMPLE", "Iosevka-Publish-Extended-Italic"},
				{"CODE-SAMPLE: 9pt", "Iosevka-Publish"},
			} {
				found := false
				for _, span := range tree.all("text") {
					if !strings.Contains(span.text(), specimen.marker) {
						continue
					}
					family := fonts[span.attr("font")]
					if at := strings.Index(family, "+"); at == 6 {
						family = family[7:]
					}
					if family != specimen.family {
						t.Errorf("%s rendered with %q; want %q", specimen.marker, family, specimen.family)
					}
					found = true
				}
				if !found {
					t.Errorf("missing rendered marker %s", specimen.marker)
				}
			}
			if strings.Contains(tree.text(), "PRIVATE-SENTINEL") || !strings.Contains(tree.text(), "FINAL-SENTINEL") {
				t.Error("rendered content leaked private text or lost final content")
			}
			if t.Failed() {
				proseTool(t, "pdftoppm", "-f", "1", "-l", "2", "-png", "-r", "150", pdf, filepath.Join(dir, "failed-page"))
			}
		})
	}
}

func TestRT063_12_OfficeRenderedContentsReferences(t *testing.T) {
	entryRE := regexp.MustCompile(`CHAPTER-([0-9]{2})[.\s]+([0-9]+)\s*$`)
	headingRE := regexp.MustCompile(`^CHAPTER-([0-9]{2})\s*$`)
	for _, format := range []string{"docx", "odt"} {
		t.Run(format, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			dir := t.TempDir()
			raw, err := os.ReadFile("../../testdata/prose-office/book.md")
			if err != nil {
				t.Fatal(err)
			}
			src := proseSource(t, dir, "book.md", string(raw))
			target := filepath.Join(dir, "book."+format)
			proseConvert(t, src, target)
			pdf := proseRenderPDF(t, target)
			text := proseTool(t, "pdftotext", "-layout", pdf, "-")
			refs, actual := map[string]int{}, map[string]int{}
			contentsPages := map[int]bool{}
			firstBody := 0
			pages := strings.Split(text, "\f")
			for i, page := range pages {
				for _, line := range strings.Split(page, "\n") {
					line = strings.TrimSpace(line)
					if line == "PART-SAMPLE" {
						firstBody = i + 1
					}
					if match := entryRE.FindStringSubmatch(line); match != nil {
						n, err := strconv.Atoi(match[2])
						if err != nil {
							t.Fatal(err)
						}
						refs[match[1]] = n
						contentsPages[i+1] = true
					} else if match := headingRE.FindStringSubmatch(line); match != nil {
						if _, exists := actual[match[1]]; exists {
							t.Errorf("duplicate body heading %s", match[1])
						}
						actual[match[1]] = i + 1
					}
				}
			}
			if len(refs) != 40 || len(actual) != 40 || len(contentsPages) < 2 {
				t.Errorf("expected 40 live chapter references and a multi-page contents: refs=%d headings=%d contents-pages=%d", len(refs), len(actual), len(contentsPages))
			}
			if firstBody == 0 || firstBody%2 != 1 {
				t.Errorf("body part must start on a physical right-hand page; got %d", firstBody)
			}
			for id, page := range actual {
				want := page - firstBody + 1
				if refs[id] != want {
					t.Errorf("chapter %s reference=%d; actual displayed page=%d (physical %d)", id, refs[id], want, page)
				}
			}
			if !strings.Contains(text, "FINAL-BOOK-SENTINEL") {
				t.Error("rendered book lost final content")
			}
			if t.Failed() {
				proseTool(t, "pdftoppm", "-f", "1", "-l", "6", "-png", "-r", "150", pdf, filepath.Join(dir, "failed-book-page"))
			}
		})
	}
}
