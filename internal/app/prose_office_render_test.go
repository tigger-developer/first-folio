// ABOUTME: Renders office fixtures headlessly with explicit field and index refresh.
// ABOUTME: Measures PDF evidence in private temporary profiles with bounded subprocesses.
package app

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io/fs"
	"math"
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
	buffer bytes.Buffer
	cancel context.CancelFunc
}

func (b *proseDiagnostic) Write(p []byte) (int, error) {
	if b.buffer.Len()+len(p) > 64<<10 {
		n, _ := b.buffer.Write(p[:(64<<10)-b.buffer.Len()])
		b.cancel()
		return n, fmt.Errorf("test subprocess exceeded 64 KiB diagnostics")
	}
	return b.buffer.Write(p)
}

func (b *proseDiagnostic) String() string { return b.buffer.String() }

func proseTool(t *testing.T, name string, args ...string) string {
	return proseToolDeadline(t, 120*time.Second, name, args...)
}

func proseToolDeadline(t *testing.T, timeout time.Duration, name string, args ...string) string {
	t.Helper()
	path, err := exec.LookPath(name)
	if err != nil {
		t.Fatalf("rendered prose regression requires %s on PATH: %v", name, err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
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
	return proseRenderPDFDeadline(t, source, 120*time.Second)
}

func proseRenderPDFDeadline(t *testing.T, source string, timeout time.Duration) string {
	t.Helper()
	dir, err := os.MkdirTemp(filepath.Dir(source), "render-")
	if err != nil {
		t.Fatal(err)
	}
	profile := filepath.Join(dir, "profile")
	// LibreOffice initializes its Basic library on first use. Complete that
	// initialization before installing the test-owned refresh helper.
	proseToolDeadline(t, timeout, "soffice", "-env:UserInstallation="+proseFileURL(profile), "--headless", "--terminate_after_init")
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
	proseToolDeadline(t, timeout, "soffice", "-env:UserInstallation="+proseFileURL(profile), "--headless", "macro:///Standard.Module1.RefreshFixture")
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
	root := proseRenderBudget(t, 50<<20)
	for _, format := range []string{"docx", "odt"} {
		t.Run(format, func(t *testing.T) {
			dir := proseRenderCase(t, root, format)
			t.Setenv("HOME", dir)
			raw, err := os.ReadFile("../../testdata/prose-office/typography.md")
			if err != nil {
				t.Fatal(err)
			}
			source := proseSource(t, dir, "source.md", string(raw))
			target := filepath.Join(dir, "output."+format)
			proseConvert(t, source, target)
			pdf := proseRenderPDF(t, target)
			proseFailurePages(t, pdf, dir, 2, "failed-page")
			xmlPath := filepath.Join(dir, "font-spans.xml")
			proseTool(t, "pdftohtml", "-xml", "-zoom", "1", "-noroundcoord", "-fontfullname", pdf, xmlPath)
			xmlRaw, err := os.ReadFile(xmlPath)
			if err != nil {
				t.Fatal(err)
			}
			tree := officeParseXML(t, xmlRaw)
			proseAssertFootnoteAnchor(t, tree)
			fonts := map[string]string{}
			for _, f := range tree.all("fontspec") {
				fonts[f.attr("id")] = f.attr("family")
			}
			for _, specimen := range []struct {
				marker, family string
				leftMM         float64
				fontPoints     float64
			}{
				{"OPENING-SAMPLE", "Iosevka-Publish", 20, 12},
				{"BODY-SAMPLE", "Iosevka-Publish", 30, 12},
				{"HEADING-SAMPLE", "Iosevka-Publish-Bold", -1, 16},
				{"QUOTE-SAMPLE", "Iosevka-Publish-Italic", 26, 10},
				{"CODE-SAMPLE: 9pt Iosevka Publish.", "Iosevka-Publish", 28, 9},
			} {
				found := 0
				for _, span := range tree.all("text") {
					if !strings.Contains(span.text(), specimen.marker) {
						continue
					}
					if strings.HasPrefix(specimen.marker, "CODE-SAMPLE:") && strings.TrimSpace(span.text()) != specimen.marker {
						continue
					}
					family := fonts[span.attr("font")]
					if at := strings.Index(family, "+"); at == 6 {
						family = family[7:]
					}
					if family != specimen.family {
						t.Errorf("%s rendered with %q; want %q", specimen.marker, family, specimen.family)
					}
					proseAssertRenderedIosevkaSize(t, span, specimen.fontPoints)
					if specimen.leftMM >= 0 {
						left, err := strconv.ParseFloat(span.attr("left"), 64)
						if err != nil {
							t.Fatal(err)
						}
						if math.Abs(left-specimen.leftMM*72/25.4) > 1 {
							t.Errorf("%s begins at %gpt; want %gmm from the page edge", specimen.marker, left, specimen.leftMM)
						}
					}
					found++
				}
				if found != 1 {
					t.Errorf("rendered marker %s appears %d times; want once", specimen.marker, found)
				}
			}
			proseAssertOpeningIntervals(t, tree)
			proseAssertHeaderGeometry(t, tree)
			proseAssertOrderedListBaseline(t, tree)
			if strings.Contains(tree.text(), "PRIVATE-SENTINEL") || !strings.Contains(tree.text(), "FINAL-SENTINEL") {
				t.Error("rendered content leaked private text or lost final content")
			}
		})
	}
}

func proseAssertOrderedListBaseline(t *testing.T, tree *officeXML) {
	t.Helper()
	for _, page := range tree.all("page") {
		for _, text := range page.all("text") {
			if !strings.Contains(text.text(), "ORDERED-FIRST") {
				continue
			}
			if strings.HasPrefix(text.text(), "1.") {
				proseAssertRenderedIosevkaSize(t, text, 12)
				return
			}
			for _, marker := range page.all("text") {
				if strings.TrimSpace(marker.text()) != "1." {
					continue
				}
				proseAssertRenderedIosevkaSize(t, marker, 12)
				top, err := strconv.ParseFloat(text.attr("top"), 64)
				if err != nil {
					t.Fatal(err)
				}
				markerTop, err := strconv.ParseFloat(marker.attr("top"), 64)
				if err != nil {
					t.Fatal(err)
				}
				if math.Abs(top-markerTop) > 1 {
					t.Errorf("ordered label and item text differ vertically by %gpt; want at most 1pt", markerTop-top)
				}
				return
			}
		}
	}
	t.Error("ordered list label was not measured")
}

func TestRT063_12_OfficeRenderedContentsReferences(t *testing.T) {
	root := proseRenderBudget(t, 100<<20)
	entryRE := regexp.MustCompile(`CHAPTER-([0-9]{2})[.\s]+([0-9]+)\s*$`)
	headingRE := regexp.MustCompile(`^CHAPTER-([0-9]{2})\s*$`)
	for _, format := range []string{"docx", "odt"} {
		t.Run(format, func(t *testing.T) {
			dir := proseRenderCase(t, root, format)
			t.Setenv("HOME", dir)
			raw, err := os.ReadFile("../../testdata/prose-office/book.md")
			if err != nil {
				t.Fatal(err)
			}
			src := proseSource(t, dir, "book.md", string(raw))
			target := filepath.Join(dir, "book."+format)
			proseConvert(t, src, target)
			pdf := proseRenderPDF(t, target)
			proseFailurePages(t, pdf, dir, 6, "failed-book-page")
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
			blank := false
			contentsOpening := true
			for i, page := range pages {
				physical := i + 1
				if contentsPages[physical] {
					want := "FRONT-ODD"
					if physical%2 == 0 {
						want = "FRONT-EVEN"
					}
					// DOCX special first-page bands may retain the odd variant.
					// Only the opening is exempt; later pages and ODT retain parity.
					if format == "docx" && contentsOpening && physical%2 == 0 && strings.Contains(page, "FRONT-ODD") {
						t.Logf("accepted DOCX opening-page limitation on physical page %d", physical)
						want = "FRONT-ODD"
					}
					contentsOpening = false
					if !strings.Contains(page, want) {
						t.Errorf("contents physical page %d has wrong opening/parity header; want %s", physical, want)
					}
				}
				if physical < firstBody && strings.TrimSpace(page) == "" {
					blank = true
				}
				if physical == firstBody {
					for _, forbidden := range []string{"BODY-ODD", "BODY-EVEN", "BODY-FOOT"} {
						if strings.Contains(page, forbidden) {
							t.Errorf("part opening page contains suppressed %s", forbidden)
						}
					}
				}
				if physical <= firstBody || strings.TrimSpace(page) == "" {
					continue
				}
				want, wrong := "BODY-ODD", "BODY-EVEN"
				if physical%2 == 0 {
					want, wrong = wrong, want
				}
				if !strings.Contains(page, want) || strings.Contains(page, wrong) || strings.Contains(page, "FRONT-") {
					t.Errorf("physical page %d has wrong front/body or odd/even running band; want %s", physical, want)
				}
				footer := fmt.Sprintf("BODY-FOOT %d", physical-firstBody+1)
				if !strings.Contains(strings.Join(strings.Fields(page), " "), footer) {
					t.Errorf("physical page %d missing displayed footer %s", physical, footer)
				}
			}
			if !blank {
				t.Error("configured blank after contents is missing or contains running matter")
			}
			xmlPath := filepath.Join(dir, "contents-spans.xml")
			proseTool(t, "pdftohtml", "-xml", "-zoom", "1", "-noroundcoord", "-fontfullname", pdf, xmlPath)
			xmlRaw, err := os.ReadFile(xmlPath)
			if err != nil {
				t.Fatal(err)
			}
			proseAssertContentsContinuation(t, officeParseXML(t, xmlRaw), contentsPages)
			for id, page := range actual {
				want := page - firstBody + 1
				if refs[id] != want {
					t.Errorf("chapter %s reference=%d; actual displayed page=%d (physical %d)", id, refs[id], want, page)
				}
			}
			if !strings.Contains(text, "FINAL-BOOK-SENTINEL") {
				t.Error("rendered book lost final content")
			}
		})
	}
}

func TestRT063_11_OfficeRenderedLetterMargins(t *testing.T) {
	root := proseRenderBudget(t, 50<<20)
	for _, format := range []string{"docx", "odt"} {
		t.Run(format, func(t *testing.T) {
			dir := proseRenderCase(t, root, format)
			t.Setenv("HOME", dir)
			for _, file := range []struct{ input, output string }{{"letters.org", "letters.org"}, {"letter.yaml", "script.yaml"}} {
				raw, err := os.ReadFile("../../testdata/prose-office/" + file.input)
				if err != nil {
					t.Fatal(err)
				}
				proseSource(t, dir, file.output, string(raw))
			}
			status, _, stderr := runApp(t, "letter", filepath.Join(dir, "letters.org"), "--format", format, "--to", "First")
			if status != 0 {
				t.Fatalf("letter export: %s", stderr)
			}
			target := filepath.Join(dir, "letter-first-theatre."+format)
			prosePackage(t, target, format)
			pdf := proseRenderPDF(t, target)
			proseFailurePages(t, pdf, dir, 2, "failed-letter-page")
			xmlPath := filepath.Join(dir, "letter-spans.xml")
			proseTool(t, "pdftohtml", "-xml", "-zoom", "1", "-noroundcoord", "-fontfullname", pdf, xmlPath)
			raw, err := os.ReadFile(xmlPath)
			if err != nil {
				t.Fatal(err)
			}
			tree := officeParseXML(t, raw)
			proseAssertLetterGaps(t, tree)
			fonts := map[string]string{}
			for _, font := range tree.all("fontspec") {
				fonts[font.attr("id")] = font.attr("family")
			}
			found := false
			for _, text := range tree.all("text") {
				if !strings.Contains(text.text(), "LETTER-BODY-SAMPLE") {
					continue
				}
				found = true
				x, err := strconv.ParseFloat(text.attr("left"), 64)
				if err != nil {
					t.Fatal(err)
				}
				// PDF text origin at the paragraph's unindented left margin.
				if math.Abs(x-30*72/25.4) > 1 {
					t.Errorf("letter body starts at %gpt; want 30mm from page edge", x)
				}
				family := fonts[text.attr("font")]
				if at := strings.Index(family, "+"); at == 6 {
					family = family[7:]
				}
				if family != "Iosevka-Publish" {
					t.Errorf("letter body font substituted: %s", family)
				}
				proseAssertRenderedIosevkaSize(t, text, 11)
			}
			if !found {
				t.Error("rendered letter body marker absent")
			}
			for _, bad := range []string{"LETTER-MUST-NOT-HAVE-", "Dear Sam"} {
				if strings.Contains(tree.text(), bad) {
					t.Errorf("rendered letter includes forbidden text %s", bad)
				}
			}
		})
	}
}

func proseAssertRenderedIosevkaSize(t *testing.T, span *officeXML, want float64) {
	t.Helper()
	height, err := strconv.ParseFloat(span.attr("height"), 64)
	if err != nil {
		t.Fatal(err)
	}
	// The bundled Publish faces expose a 1.25em PDF text box. Independent
	// Typst PDF measurements give 15/20/12.5/11.25pt at 12/16/10/9pt.
	// Unlike fontspec.size, text.height retains fractional precision.
	if math.Abs(height/1.25-want) > 0.25 {
		t.Errorf("%s has inferred font size %gpt; want %gpt", span.text(), height/1.25, want)
	}
}

func proseAssertOpeningIntervals(t *testing.T, tree *officeXML) {
	t.Helper()
	for _, page := range tree.all("page") {
		var opening []*officeXML
		collect := false
		for _, span := range page.all("text") {
			if strings.Contains(span.text(), "OPENING-SAMPLE") {
				collect = true
			}
			if strings.Contains(span.text(), "BODY-SAMPLE") && collect {
				if len(opening) < 2 {
					t.Error("opening specimen did not supply two measured lines")
					return
				}
				previous, _ := strconv.ParseFloat(opening[len(opening)-1].attr("top"), 64)
				current, err := strconv.ParseFloat(span.attr("top"), 64)
				if err != nil {
					t.Fatal(err)
				}
				want := 24.93 + 3*72/25.4
				if math.Abs(current-previous-want) > 1 {
					t.Errorf("paragraph interval %gpt; want %gpt", current-previous, want)
				}
				return
			}
			if !collect {
				continue
			}
			if len(opening) > 0 {
				previous, _ := strconv.ParseFloat(opening[len(opening)-1].attr("top"), 64)
				current, err := strconv.ParseFloat(span.attr("top"), 64)
				if err != nil {
					t.Fatal(err)
				}
				if math.Abs(current-previous-24.93) > 1 {
					t.Errorf("opening line interval %gpt; want 24.93pt", current-previous)
				}
			}
			opening = append(opening, span)
		}
	}
	t.Error("opening and following body paragraphs not found together for spacing measurement")
}

func proseAssertHeaderGeometry(t *testing.T, tree *officeXML) {
	t.Helper()
	for _, page := range tree.all("page") {
		headerCount := 0
		for _, span := range page.all("text") {
			top, err := strconv.ParseFloat(span.attr("top"), 64)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(span.text(), "HEADER-SAMPLE") {
				headerCount++
				// In the independent baseline, the 10pt Publish text-box
				// top is 9.65pt above its baseline at the configured edge.
				if math.Abs(top+9.65-20*72/25.4) > 1 {
					t.Errorf("header baseline %gpt; want 20mm from top edge", top+9.65)
				}
				proseAssertRenderedIosevkaSize(t, span, 10)
			}
			if strings.Contains(span.text(), "HEADING-SAMPLE") {
				// The native paragraph's first line box begins at 25mm. Its
				// ordinary glyphs occupy that box; forcing the cap top onto the
				// margin with raised characters breaks editing. Native-property
				// tests assert the exact margin and independent line interval.
				capTop := top + 3.68
				if capTop < 25*72/25.4-1 || capTop > 25*72/25.4+30 {
					t.Errorf("heading cap-box top %gpt lies outside its first native line box", capTop)
				}
			}
		}
		if headerCount != 1 {
			t.Errorf("page %s has %d header specimens; want one", page.attr("number"), headerCount)
		}
	}
}

func proseAssertLetterGaps(t *testing.T, tree *officeXML) {
	t.Helper()
	spans := tree.all("text")
	lookup := func(marker string) *officeXML {
		var found *officeXML
		for _, span := range spans {
			if strings.Contains(span.text(), marker) {
				if found != nil {
					t.Fatalf("ambiguous letter measurement marker %s", marker)
				}
				found = span
			}
		}
		if found == nil {
			t.Fatalf("letter measurement marker %s absent", marker)
		}
		return found
	}
	value := func(span *officeXML, attr string) float64 {
		v, err := strconv.ParseFloat(span.attr(attr), 64)
		if err != nil {
			t.Fatal(err)
		}
		return v
	}
	for _, gap := range []struct {
		name, from, to string
		mm             float64
	}{
		{"space-after-sender", "CONTACT-SAMPLE", "RECIPIENT-ONE", 8},
		{"space-after-recipient", "Cork", "16 October 2026", 6},
		{"space-after-date", "16 October 2026", "SUBJECT-SAMPLE", 5},
		{"space-after-subject", "SUBJECT-SAMPLE", "Dear Alex", 4},
		{"space-before-closing", "UNKNOWN-TOKEN-SAMPLE", "CLOSING-SAMPLE", 7},
		{"space-before-signoff", "CLOSING-SAMPLE", "SIGNOFF-SAMPLE", 9},
	} {
		// Both endpoints use 11pt Publish with equal text-box ascent. Their
		// top difference is consequently also their baseline difference.
		got := value(lookup(gap.to), "top") - value(lookup(gap.from), "top")
		want := 21.285 + gap.mm*72/25.4
		if math.Abs(got-want) > 1 {
			t.Errorf("%s baseline gap %gpt; want %gpt", gap.name, got, want)
		}
	}
	sender := lookup("SENDER-SAMPLE")
	senderLeft := value(sender, "left")
	// The 11pt sender uses a 16.885pt native line box beginning at 25mm.
	// Verify that ordinary glyphs stay within it, rather than requiring
	// raised-character formatting to pin their cap edge to the margin.
	capTop := value(sender, "top") + 2.53
	if capTop < 25*72/25.4-1 || capTop > 25*72/25.4+16.885 {
		t.Error("sender text lies outside its first native line box")
	}
	if math.Abs(value(lookup("CONTACT-SAMPLE"), "left")-senderLeft) > 1 {
		t.Error("sender/contact contents are not left-aligned together")
	}
	if senderLeft < 210*72/25.4/2 {
		t.Error("sender block is not on the right side")
	}
	right := 0.0
	for _, span := range spans {
		if value(span, "top") >= value(lookup("RECIPIENT-ONE"), "top") {
			continue
		}
		if edge := value(span, "left") + value(span, "width"); edge > right {
			right = edge
		}
	}
	if math.Abs(right-(210-25)*72/25.4) > 1 {
		t.Errorf("sender right edge %gpt; want 25mm right margin", right)
	}
}

// RT063.11 splits its 100 MiB aggregate allowance equally between its
// manuscript and letter cases; RT063.12 has its own 100 MiB allowance.
// All generated packages, private profiles, PDFs, XML and PNGs live below
// this root. Check the complete retained tree even when a child test fails.
func proseRenderBudget(t *testing.T, limit int64) string {
	t.Helper()
	root := t.TempDir()
	t.Cleanup(func() {
		var size int64
		err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				return nil
			}
			info, err := entry.Info()
			if err != nil {
				return err
			}
			size += info.Size()
			return nil
		})
		if err != nil {
			t.Errorf("measuring rendered artefacts: %v", err)
		}
		if size > limit {
			t.Errorf("rendered artefacts total %d bytes; bound is %d", size, limit)
		}
	})
	return root
}

func proseRenderCase(t *testing.T, root, name string) string {
	t.Helper()
	dir := filepath.Join(root, name)
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	return dir
}

func proseFailurePages(t *testing.T, pdf, dir string, pages int, name string) {
	t.Helper()
	t.Cleanup(func() {
		if t.Failed() {
			proseTool(t, "pdftoppm", "-f", "1", "-l", strconv.Itoa(pages), "-png", "-r", "150", pdf, filepath.Join(dir, name))
		}
	})
}

func proseAssertContentsContinuation(t *testing.T, tree *officeXML, contentsPages map[int]bool) {
	t.Helper()
	first := 0
	for page := range contentsPages {
		if first == 0 || page < first {
			first = page
		}
	}
	checked := 0
	for _, page := range tree.all("page") {
		number, err := strconv.Atoi(page.attr("number"))
		if err != nil {
			t.Fatal(err)
		}
		if !contentsPages[number] || number == first {
			continue
		}
		found := false
		for _, text := range page.all("text") {
			if !strings.Contains(text.text(), "CHAPTER-") && !strings.Contains(text.text(), "SECTION-") {
				continue
			}
			found = true
			top, err := strconv.ParseFloat(text.attr("top"), 64)
			if err != nil {
				t.Fatal(err)
			}
			// Base TOC uses 12pt Publish: cap-box top is 2.76pt below
			// the PDF text-box top. Reserve 20mm margin plus 10mm pad.
			if top+2.76 < 30*72/25.4-1 {
				t.Errorf("contents continuation page %d intrudes into reserved top band at %gpt", number, top+2.76)
			}
		}
		if !found {
			t.Errorf("no measurable entries on continuation page %d", number)
		}
		checked++
	}
	if checked == 0 {
		t.Error("no contents continuation page measured")
	}
}

func TestRT063_11_OfficeRenderedPageFieldTypography(t *testing.T) {
	root := proseRenderBudget(t, 50<<20)
	for _, format := range []string{"docx", "odt"} {
		t.Run(format, func(t *testing.T) {
			dir := proseRenderCase(t, root, format)
			t.Setenv("HOME", dir)
			proseSource(t, dir, "script.yaml", "folio:\n  page-header:\n    enabled: true\n    font: {family: Iosevka Publish, size: 10pt, style: italic, weight: regular, stretch: 100%, letter-spacing: 0pt}\n    format: 'HEADER-FIELD [page]'\n    frontmatter-format: 'HEADER-FIELD [page]'\n    alt-format: 'HEADER-FIELD [page]'\n    alt-frontmatter-format: 'HEADER-FIELD [page]'\n  page-footer:\n    enabled: true\n    font: {family: Iosevka Publish, size: 9pt, style: regular, weight: bold, stretch: 100%, letter-spacing: 0pt}\n    format: 'FOOTER-FIELD [total-pages]'\n    frontmatter-format: 'FOOTER-FIELD [total-pages]'\n    alt-format: 'FOOTER-FIELD [total-pages]'\n    alt-frontmatter-format: 'FOOTER-FIELD [total-pages]'\n  manuscript:\n    title-page: {enabled: false}\n    toc: {enabled: false}\n    copyright: {enabled: false}\n")
			src := proseSource(t, dir, "source.md", "Body text.\n")
			target := filepath.Join(dir, "output."+format)
			proseConvert(t, src, target)
			pdf := proseRenderPDF(t, target)
			xmlPath := filepath.Join(dir, "fields.xml")
			proseTool(t, "pdftohtml", "-xml", "-zoom", "1", "-noroundcoord", "-fontfullname", pdf, xmlPath)
			raw, err := os.ReadFile(xmlPath)
			if err != nil {
				t.Fatal(err)
			}
			tree := officeParseXML(t, raw)
			fonts := map[string]string{}
			for _, font := range tree.all("fontspec") {
				fonts[font.attr("id")] = font.attr("family")
			}
			for _, band := range []struct {
				marker, face string
				size         float64
			}{{"HEADER-FIELD", "Italic", 10}, {"FOOTER-FIELD", "Bold", 9}} {
				var label *officeXML
				for _, span := range tree.all("text") {
					if strings.Contains(span.text(), band.marker) {
						label = span
						break
					}
				}
				if label == nil {
					t.Fatalf("missing %s", band.marker)
				}
				labelTop, _ := strconv.ParseFloat(label.attr("top"), 64)
				found := false
				for _, span := range tree.all("text") {
					top, _ := strconv.ParseFloat(span.attr("top"), 64)
					if math.Abs(top-labelTop) > 30 || !strings.Contains(span.text(), "1") {
						continue
					}
					found = true
					if !strings.Contains(fonts[span.attr("font")], "Iosevka-Publish-"+band.face) {
						t.Errorf("%s field font=%s", band.marker, fonts[span.attr("font")])
					}
					proseAssertRenderedIosevkaSize(t, span, band.size)
					if math.Abs(top-labelTop) > 0.3 {
						t.Errorf("%s field baseline differs from label by %.3fpt", band.marker, top-labelTop)
					}
				}
				if !found {
					t.Errorf("missing refreshed field beside %s", band.marker)
				}
			}
			proseFailurePages(t, pdf, dir, 1, "failed-fields")
		})
	}
}

func proseAssertFootnoteAnchor(t *testing.T, tree *officeXML) {
	t.Helper()
	var anchor *officeXML
	for _, span := range tree.all("text") {
		if strings.Contains(span.text(), "NOTE-ANCHOR") {
			anchor = span
			break
		}
	}
	if anchor == nil {
		t.Fatal("rendered anchor text missing")
	}
	top, _ := strconv.ParseFloat(anchor.attr("top"), 64)
	height, _ := strconv.ParseFloat(anchor.attr("height"), 64)
	found := false
	for _, span := range tree.all("text") {
		if strings.TrimSpace(span.text()) != "1" {
			continue
		}
		y, _ := strconv.ParseFloat(span.attr("top"), 64)
		h, _ := strconv.ParseFloat(span.attr("height"), 64)
		if math.Abs(y-top) > 20 {
			continue
		}
		found = true
		if y+h >= top+height-1 {
			t.Errorf("note reference bottom %g is not above body bottom %g", y+h, top+height)
		}
	}
	if !found {
		t.Fatal("rendered footnote reference missing beside anchor")
	}
}
