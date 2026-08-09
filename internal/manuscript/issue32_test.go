// ABOUTME: Regression tests for configurable TOC links and chapter numbering.
// ABOUTME: Exercises PDF annotations and multi-part numbering through the public command.
package manuscript_test

import (
	"bytes"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tigger-developer/first-folio/internal/app"
)

// RT-32.1: TOC links are present by default.
func TestRT_32_1_TOCLinksDefaultToEnabled(t *testing.T) {
	pdf := renderIssue32PDF(t, "")
	html := issue32PDFHTML(t, pdf)

	assertContains(t, html, `<a href=`)
}

// RT-32.2: the KDP-safe override removes PDF link annotations.
func TestRT_32_2_TOCLinksCanBeDisabled(t *testing.T) {
	pdf := renderIssue32PDF(t, "      links: false")
	html := issue32PDFHTML(t, pdf)

	assertNotContains(t, html, `<a href=`)
}

// RT-32.3: both link modes retain printable text and the document outline.
func TestRT_32_3_TOCLinkModesRemainPrintableAndOutlined(t *testing.T) {
	for _, tc := range []struct {
		name   string
		config string
	}{
		{name: "enabled"},
		{name: "disabled", config: "      links: false"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pdf := renderIssue32PDF(t, tc.config)
			text := pdfText(t, filepath.Dir(pdf), pdf)
			html := issue32PDFHTML(t, pdf)

			assertNotContains(t, text, "\u2060")
			assertContains(t, html, "<outline>")
			assertContains(t, html, "PART ONE")
		})
	}
}

// RT-32.4: links must be a YAML boolean.
func TestRT_32_4_TOCLinksRejectNonBooleanValues(t *testing.T) {
	dir := issue32Project(t, "      links: sometimes", "")
	stderr, status := runFolioForStatus("manuscript", filepath.Join(dir, "manuscript.md"), filepath.Join(dir, "out.typ"))

	if status == 0 {
		t.Fatal("non-boolean toc.links unexpectedly succeeded")
	}
	assertContains(t, stderr, "cannot unmarshal")
	assertContains(t, stderr, "bool")
}

// RT-32.5: chapter numbers are continuous across parts by default.
func TestRT_32_5_ChapterNumberingDefaultsToContinuous(t *testing.T) {
	typst := renderIssue32Typst(t, "", "1")

	assertContains(t, typst, `full: "1: Alpha"`)
	assertContains(t, typst, `full: "2: Beta"`)
	assertContains(t, typst, `full: "3: Gamma"`)
	assertContains(t, typst, `full: "4: Delta"`)
}

// RT-32.6: per-part remains an explicit reset mode.
func TestRT_32_6_ChapterNumberingCanResetPerPart(t *testing.T) {
	typst := renderIssue32Typst(t, "per-part", "1")

	assertContains(t, typst, `full: "1: Alpha"`)
	assertContains(t, typst, `full: "2: Beta"`)
	assertContains(t, typst, `full: "1: Gamma"`)
	assertContains(t, typst, `full: "2: Delta"`)
}

// RT-32.8 and RT-32.9: one- and two-segment formats render as specified.
func TestRT_32_8_9_ChapterNumberFormats(t *testing.T) {
	for _, tc := range []struct {
		format string
		want   string
	}{
		{format: "1", want: "1"},
		{format: "I", want: "I"},
		{format: "i", want: "i"},
		{format: "I.1", want: "II.1"},
		{format: "1.1", want: "2.1"},
		{format: "1.I", want: "2.I"},
	} {
		t.Run(tc.format, func(t *testing.T) {
			typst := renderIssue32Typst(t, "per-part", tc.format)
			assertContains(t, typst, `number: "`+tc.want+`"`)
			assertContains(t, typst, `full: "`+tc.want+`: Gamma"`)
		})
	}
}

// RT-32.10: composite numbers appear in both body headings and compiled TOC text.
func TestRT_32_10_CompositeNumbersReachBodyAndTOC(t *testing.T) {
	requireIssue32PDFTools(t)
	dir := issue32Project(t, "", strings.Join([]string{
		"    chapter:",
		"      show-number: true",
		"      prefix: \"\"",
		"      separator: \": \"",
		"      number-reset: per-part",
		"      number-format: \"I.1\"",
	}, "\n"))
	pdf := filepath.Join(dir, "manuscript.pdf")
	runFolio(t, "manuscript", filepath.Join(dir, "manuscript.md"), pdf)
	text := pdfText(t, dir, pdf)

	if strings.Count(text, "II.1: Gamma") < 2 {
		t.Fatalf("composite chapter number did not appear in TOC and body:\n%s", text)
	}
}

// RT-32.11: unsupported chapter number patterns fail clearly.
func TestRT_32_11_InvalidChapterNumberFormatIsRejected(t *testing.T) {
	dir := issue32Project(t, "", strings.Join([]string{
		"    chapter:",
		"      number-format: \"I-1\"",
	}, "\n"))
	stderr, status := runFolioForStatus("manuscript", filepath.Join(dir, "manuscript.md"), filepath.Join(dir, "out.typ"))

	if status == 0 {
		t.Fatal("invalid chapter.number-format unexpectedly succeeded")
	}
	assertContains(t, stderr, "chapter.number-format")
}

func renderIssue32PDF(t *testing.T, tocConfig string) string {
	t.Helper()
	requireIssue32PDFTools(t)
	dir := issue32Project(t, tocConfig, "")
	pdf := filepath.Join(dir, "manuscript.pdf")
	runFolio(t, "manuscript", filepath.Join(dir, "manuscript.md"), pdf)
	return pdf
}

func renderIssue32Typst(t *testing.T, reset string, format string) string {
	t.Helper()
	chapter := []string{
		"    chapter:",
		"      show-number: true",
		"      prefix: \"\"",
		"      separator: \": \"",
		"      number-format: \"" + format + "\"",
	}
	if reset != "" {
		chapter = append(chapter, "      number-reset: "+reset)
	}
	dir := issue32Project(t, "      enabled: false", strings.Join(chapter, "\n"))
	typstPath := filepath.Join(dir, "manuscript.typ")
	runFolio(t, "manuscript", filepath.Join(dir, "manuscript.md"), typstPath)
	return readTestFile(t, typstPath)
}

func issue32Project(t *testing.T, tocConfig string, manuscriptConfig string) string {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "manuscript.md"), strings.Join([]string{
		"---",
		"title: Numbering Test",
		"author: Example Author",
		"---",
		"",
		"# PART ONE",
		"",
		"## Alpha",
		"",
		"First chapter.",
		"",
		"## Beta",
		"",
		"Second chapter.",
		"",
		"# PART TWO",
		"",
		"## Gamma",
		"",
		"Third chapter.",
		"",
		"## Delta",
		"",
		"Fourth chapter.",
	}, "\n"))
	if tocConfig != "" || manuscriptConfig != "" {
		config := []string{"folio:", "  manuscript:"}
		if tocConfig != "" {
			config = append(config, "    toc:", tocConfig)
		}
		if manuscriptConfig != "" {
			config = append(config, manuscriptConfig)
		}
		writeTestFile(t, filepath.Join(dir, "script.yaml"), strings.Join(config, "\n")+"\n")
	}
	return dir
}

func issue32PDFHTML(t *testing.T, pdf string) string {
	t.Helper()
	requireIssue32PDFTools(t)
	xmlPath := filepath.Join(t.TempDir(), "manuscript.xml")
	return commandText(t, exec.Command("pdftohtml", "-xml", "-hidden", pdf, xmlPath)) + readTestFile(t, xmlPath)
}

func requireIssue32PDFTools(t *testing.T) {
	t.Helper()
	for _, name := range []string{"typst", "pdftotext", "pdftohtml"} {
		if _, err := exec.LookPath(name); err != nil {
			t.Skipf("%s is not installed", name)
		}
	}
}

func runFolioForStatus(args ...string) (string, int) {
	var stderr bytes.Buffer
	status := app.Run(args, strings.NewReader(""), io.Discard, &stderr)
	return stderr.String(), status
}
