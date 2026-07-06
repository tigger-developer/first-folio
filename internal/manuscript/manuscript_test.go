// ABOUTME: Regression tests for manuscript rendering behaviour.
// ABOUTME: Exercises the Go CLI path with temporary Markdown and org projects.
package manuscript

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestMarkdownManuscriptCLIProducesTypstContract(t *testing.T) {
	root := testProjectRoot(t)
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "part1", "ch02.md"), markdownChapterTwo())
	writeFile(t, filepath.Join(dir, "part1", "ch01.md"), markdownChapterOne())

	output := filepath.Join(dir, "out.typ")
	runManuscript(t, root, filepath.Join(dir, "part?", "ch??.md"), output)
	typst := readFile(t, output)

	assertContains(t, typst, `paper: "a4"`)
	assertContains(t, typst, `margin: 20mm`)
	assertContains(t, typst, `Contents`)
	assertContains(t, typst, `#outline(title: none)`)
	assertContains(t, typst, `font: "Libertinus Serif"`)
	assertContains(t, typst, `font: "Libertinus Sans"`)
	assertContains(t, typst, `#folio-part(first: true)[PART ONE]`)
	assertContains(t, typst, `#folio-chapter(first: false)[Chapter 1]`)
	assertContains(t, typst, `#folio-chapter(first: false)[Chapter 2]`)
	assertContains(t, typst, `#folio-scene-break()`)
	assertContains(t, typst, `#text(font: "Libertinus Mono")[watch]`)
	assertNotContains(t, typst, `Private planning`)
	assertBefore(t, typst, `Chapter 1`, `Chapter 2`)
}

func TestOrgManuscriptCLIProducesTypstContract(t *testing.T) {
	root := testProjectRoot(t)
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "part1", "ch01.org"), orgChapterOne())
	writeFile(t, filepath.Join(dir, "part1", "ch02.org"), orgChapterTwo())

	output := filepath.Join(dir, "out.typ")
	runManuscript(t, root, filepath.Join(dir, "part?", "ch??.org"), output)
	typst := readFile(t, output)

	assertContains(t, typst, `About Time`)
	assertContains(t, typst, `#folio-part(first: true)[PART ONE]`)
	assertContains(t, typst, `#folio-chapter(first: false)[Chapter 1]`)
	assertContains(t, typst, `#folio-chapter(first: false)[Chapter 2]`)
	assertContains(t, typst, `#folio-scene-break()`)
	assertNotContains(t, typst, `Private planning`)
	assertBefore(t, typst, `Chapter 1`, `Chapter 2`)
}

func TestUSManuscriptOverridesBritishWithoutChangingPageSize(t *testing.T) {
	root := testProjectRoot(t)
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "ch01.md"), markdownChapterOne())

	output := filepath.Join(dir, "us.typ")
	runManuscript(t, root, "--style", "us", filepath.Join(dir, "ch01.md"), output)
	typst := readFile(t, output)

	assertContains(t, typst, `paper: "a4"`)
	assertNotContains(t, typst, `us-letter`)
	assertContains(t, typst, `font: "Libertinus Mono"`)
	assertContains(t, typst, `first-line-indent: 12.7mm`)
	assertContains(t, typst, `leading: 2em`)
	assertContains(t, typst, `margin: 25mm`)
}

func TestTOCCanBeDisabledByConfig(t *testing.T) {
	root := testProjectRoot(t)
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "script.yaml"), "folio:\n  manuscript:\n    toc:\n      enabled: false\n")
	writeFile(t, filepath.Join(dir, "ch01.md"), markdownChapterOne())

	output := filepath.Join(dir, "out.typ")
	runManuscript(t, root, filepath.Join(dir, "ch01.md"), output)
	typst := readFile(t, output)

	assertNotContains(t, typst, `#outline(title: none)`)
}

func TestInvalidInputsFailClearly(t *testing.T) {
	root := testProjectRoot(t)
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "ch01.md"), markdownChapterOne())
	writeFile(t, filepath.Join(dir, "ch01.org"), orgChapterOne())
	writeFile(t, filepath.Join(dir, "bad.fountain"), "Title: Bad\n\n# Section\n")

	assertCommandFails(t, root, []string{
		filepath.Join(dir, "missing*.md"),
		filepath.Join(dir, "out.typ"),
	}, "no files match glob")
	assertCommandFails(t, root, []string{
		filepath.Join(dir, "ch01.md"),
		filepath.Join(dir, "ch01.org"),
		filepath.Join(dir, "out.typ"),
	}, "must not mix Markdown and org-mode")
	assertCommandFails(t, root, []string{
		filepath.Join(dir, "bad.fountain"),
		filepath.Join(dir, "out.typ"),
	}, "accepts only Markdown or org-mode")
}

func TestQuotedHomeGlobExpandsDeterministically(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeFile(t, filepath.Join(home, "notes", "about-time-nove", "part1", "ch02.md"), markdownChapterTwo())
	writeFile(t, filepath.Join(home, "notes", "about-time-nove", "part1", "ch01.md"), markdownChapterOne())

	inputs, err := ResolveInputs([]string{"~/notes/about-time-nove/part?/ch??.md"})
	if err != nil {
		t.Fatalf("resolving quoted home glob: %v", err)
	}
	if len(inputs.Paths) != 2 {
		t.Fatalf("expected 2 paths, got %d", len(inputs.Paths))
	}
	assertBefore(t, strings.Join(inputs.Paths, "\n"), "ch01.md", "ch02.md")
}

func TestCLIHelpVersionAndDryRun(t *testing.T) {
	root := testProjectRoot(t)
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "ch01.md"), markdownChapterOne())
	output := filepath.Join(dir, "out.typ")

	help := runManuscriptOutput(t, root, "--help")
	assertContains(t, help, `Usage: folio manuscript`)
	assertContains(t, help, `--dry-run`)

	version := runManuscriptOutput(t, root, "--version")
	assertContains(t, version, `folio-manuscript`)

	dryRun := runManuscriptOutput(t, root, "--dry-run", filepath.Join(dir, "ch01.md"), output)
	assertContains(t, dryRun, `format: markdown`)
	assertContains(t, dryRun, `style: british`)
	if _, err := os.Stat(output); err == nil {
		t.Fatalf("dry run wrote output target")
	}
}

func TestPublicFolioDispatcherReachesGoManuscriptHelp(t *testing.T) {
	root := testProjectRoot(t)
	cmd := exec.Command(filepath.Join(root, "bin", "folio"), "manuscript", "--help")
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "FIRST_FOLIO_ROOT="+root)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("folio manuscript --help failed: %v\n%s", err, string(out))
	}
	assertContains(t, string(out), `Usage: folio manuscript`)
}

func TestPDFRenderHasTOCAndNoBlankPageBeforePart(t *testing.T) {
	requireTool(t, "typst")
	requireTool(t, "pdfinfo")
	requireTool(t, "pdftotext")

	root := testProjectRoot(t)
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "part1", "ch01.md"), markdownChapterOne())
	writeFile(t, filepath.Join(dir, "part1", "ch02.md"), markdownChapterTwo())

	output := filepath.Join(dir, "out.pdf")
	runManuscript(t, root, filepath.Join(dir, "part?", "ch??.md"), output)

	info := commandOutput(t, exec.Command("pdfinfo", output))
	assertContains(t, info, `A4`)

	textPath := filepath.Join(dir, "out.txt")
	commandOutput(t, exec.Command("pdftotext", "-layout", output, textPath))
	pdfText := readFile(t, textPath)
	assertContains(t, pdfText, `Contents`)
	assertContains(t, pdfText, `PART ONE`)
	assertContains(t, pdfText, `Chapter 1`)
	assertContains(t, pdfText, `Tadhg Paul / About Time / 1`)
	assertContains(t, pdfText, `Tadhg Paul / About Time / 2`)
	assertBefore(t, pdfText, `Contents`, `PART ONE`)
	assertBefore(t, pdfText, `PART ONE`, `Chapter 1`)
}

func runManuscript(t *testing.T, root string, args ...string) {
	t.Helper()
	_ = runManuscriptOutput(t, root, args...)
}

func runManuscriptOutput(t *testing.T, root string, args ...string) string {
	t.Helper()
	cmdArgs := append([]string{"run", filepath.Join(root, "cmd", "folio-manuscript")}, args...)
	cmd := exec.Command("go", cmdArgs...)
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "FIRST_FOLIO_ROOT="+root)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("folio manuscript failed: %v\n%s", err, string(out))
	} else {
		return string(out)
	}
	return ""
}

func assertCommandFails(t *testing.T, root string, args []string, want string) {
	t.Helper()
	cmdArgs := append([]string{"run", filepath.Join(root, "cmd", "folio-manuscript")}, args...)
	cmd := exec.Command("go", cmdArgs...)
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "FIRST_FOLIO_ROOT="+root)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected command to fail")
	}
	assertContains(t, string(out), want)
}

func testProjectRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("working directory: %v", err)
	}
	for dir := wd; ; dir = filepath.Dir(dir) {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("cannot find project root")
		}
	}
}

func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("creating fixture directory: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writing fixture: %v", err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	return string(data)
}

func assertContains(t *testing.T, haystack string, needle string) {
	t.Helper()
	if !strings.Contains(haystack, needle) {
		t.Fatalf("missing %q in:\n%s", needle, haystack)
	}
}

func assertNotContains(t *testing.T, haystack string, needle string) {
	t.Helper()
	if strings.Contains(haystack, needle) {
		t.Fatalf("unexpected %q in:\n%s", needle, haystack)
	}
}

func assertBefore(t *testing.T, haystack string, first string, second string) {
	t.Helper()
	firstIndex := strings.Index(haystack, first)
	secondIndex := strings.Index(haystack, second)
	if firstIndex < 0 || secondIndex < 0 || firstIndex >= secondIndex {
		t.Fatalf("expected %q before %q in:\n%s", first, second, haystack)
	}
}

func requireTool(t *testing.T, name string) {
	t.Helper()
	if _, err := exec.LookPath(name); err != nil {
		t.Skipf("%s not installed", name)
	}
}

func commandOutput(t *testing.T, cmd *exec.Cmd) string {
	t.Helper()
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s failed: %v\n%s", cmd.String(), err, string(out))
	}
	return string(out)
}

func markdownChapterOne() string {
	return strings.Join([]string{
		"# About Time",
		"",
		"**A Novel**",
		"",
		"*by Tadhg Paul*",
		"",
		"--- Draft 4 | July 2026 ---",
		"",
		"## PART ONE",
		"",
		"### Chapter 1",
		"",
		"The rain had been falling since Tuesday, though nobody could agree which Tuesday had started it.",
		"",
		"Mairead found the `watch` under the loose floorboard.",
		"",
		"***",
		"",
		"By noon, the hands had moved backwards twice.",
		"",
		"### Notes <!-- noexport -->",
		"",
		"Private planning must not appear.",
	}, "\n")
}

func markdownChapterTwo() string {
	return strings.Join([]string{
		"### Chapter 2",
		"",
		"The first rule of time travel was that nobody should do it before breakfast.",
	}, "\n")
}

func orgChapterOne() string {
	return strings.Join([]string{
		"#+TITLE: About Time",
		"#+SUBTITLE: A Novel",
		"#+AUTHOR: Tadhg Paul",
		"#+DATE: July 2026",
		"#+VERSION: Draft 4",
		"#+WORDCOUNT: 80000",
		"",
		"* PART ONE",
		"** Chapter 1",
		"The rain had been falling since Tuesday.",
		"",
		"-----",
		"",
		"By noon, the hands had moved backwards twice.",
		"",
		"*** Notes :noexport:",
		"Private planning must not appear.",
	}, "\n")
}

func orgChapterTwo() string {
	return strings.Join([]string{
		"** Chapter 2",
		"The first rule of time travel was that nobody should do it before breakfast.",
	}, "\n")
}
