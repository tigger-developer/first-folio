// ABOUTME: Regression tests for external ISBN barcode SVG generation.
// ABOUTME: Exercises sidecar modes and help through the manuscript command boundary.
package manuscript_test

import (
	"bytes"
	"encoding/xml"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tigger-developer/first-folio/internal/app"
)

// RT-28.1: file mode writes a valid SVG sidecar without embedding it.
func TestRT_28_1_FileModeWritesExternalBarcode(t *testing.T) {
	dir := barcodeProject(t, "file")
	output := filepath.Join(dir, "manuscript.typ")

	runFolio(t, "manuscript", filepath.Join(dir, "chapter.md"), output)

	assertValidSVG(t, filepath.Join(dir, "manuscript.barcode.svg"))
	assertNotContains(t, readTestFile(t, output), `#image(bytes(`)
}

// RT-28.2: render-and-file mode writes and embeds the same barcode format.
func TestRT_28_2_RenderAndFileModeWritesAndEmbedsBarcode(t *testing.T) {
	dir := barcodeProject(t, "render-and-file")
	output := filepath.Join(dir, "manuscript.typ")

	runFolio(t, "manuscript", filepath.Join(dir, "chapter.md"), output)

	assertValidSVG(t, filepath.Join(dir, "manuscript.barcode.svg"))
	assertContains(t, readTestFile(t, output), `#image(bytes(`)
}

// RT-28.3: public manuscript help documents external SVG generation.
func TestRT_28_3_HelpDocumentsExternalBarcode(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	status := app.Run([]string{"manuscript", "--help"}, strings.NewReader(""), &stdout, &stderr)
	if status != 0 {
		t.Fatalf("folio manuscript --help failed with status %d: %s", status, stderr.String())
	}

	help := stdout.String()
	assertContains(t, help, `isbn-barcode: file`)
	assertContains(t, help, `manuscript.barcode.svg`)
	assertContains(t, help, `folio manuscript manuscript.md manuscript.pdf`)
}

func barcodeProject(t *testing.T, mode string) string {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "chapter.md"), strings.Join([]string{
		"---",
		"title: Barcode Test",
		"author: Example Author",
		"---",
		"",
		"## Chapter One",
		"",
		"Body text.",
	}, "\n"))
	writeTestFile(t, filepath.Join(dir, "script.yaml"), strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    copyright:",
		"      enabled: true",
		"      isbn: \"978-0-000000-00-2\"",
		"      isbn-barcode: " + mode,
		"",
	}, "\n"))
	return dir
}

func runFolio(t *testing.T, args ...string) {
	t.Helper()
	var stderr bytes.Buffer
	status := app.Run(args, strings.NewReader(""), io.Discard, &stderr)
	if status != 0 {
		t.Fatalf("folio command failed with status %d: %s", status, stderr.String())
	}
}

func writeTestFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writing test file: %v", err)
	}
}

func readTestFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading test file: %v", err)
	}
	return string(data)
}

func assertValidSVG(t *testing.T, path string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading barcode SVG: %v", err)
	}
	var root struct {
		XMLName xml.Name
		Rects   []struct{} `xml:"rect"`
		Digits  []string   `xml:"text"`
	}
	if err := xml.Unmarshal(data, &root); err != nil {
		t.Fatalf("parsing barcode SVG: %v", err)
	}
	if root.XMLName.Local != "svg" {
		t.Fatalf("barcode sidecar root is %q, want svg", root.XMLName.Local)
	}
	if len(root.Rects) < 2 {
		t.Fatalf("barcode sidecar has %d rectangles, want background and bars", len(root.Rects))
	}
	if got := strings.Join(root.Digits, ""); got != "9780000000002" {
		t.Fatalf("barcode sidecar digits are %q, want ISBN digits", got)
	}
}

func assertContains(t *testing.T, value string, expected string) {
	t.Helper()
	if !strings.Contains(value, expected) {
		t.Fatalf("missing %q in:\n%s", expected, value)
	}
}

func assertNotContains(t *testing.T, value string, unexpected string) {
	t.Helper()
	if strings.Contains(value, unexpected) {
		t.Fatalf("unexpected %q in:\n%s", unexpected, value)
	}
}
