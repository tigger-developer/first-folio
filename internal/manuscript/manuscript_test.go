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
	assertContains(t, typst, `leading: 1.5em`)
	assertContains(t, typst, `spacing: 1.5em`)
	assertContains(t, typst, `author\@example.invalid`)
	assertContains(t, typst, `+353 1 000 0000`)
	assertContains(t, typst, `6 July 2026`)
	assertContains(t, typst, `above: if it.level == 1 { 0.5em } else { 0pt }`)
	assertContains(t, typst, `if it.level == 1 and true`)
	assertContains(t, typst, `footer: none`)
	assertContains(t, typst, `#folio-part(first: true)[PART ONE]`)
	assertContains(t, typst, `#folio-chapter(first: false)[Chapter 1]`)
	assertContains(t, typst, `#folio-chapter(first: false)[Chapter 2]`)
	assertContains(t, typst, `#folio-scene-break()`)
	assertContains(t, typst, `#text(font: "Libertinus Mono", size: 10pt, weight: "bold")[watch]`)
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

	assertContains(t, typst, `The Glass Orchard`)
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
	assertContains(t, typst, `size: 10pt`)
	assertContains(t, typst, `font: "Menlo"`)
	assertContains(t, typst, `size: 9pt`)
	assertContains(t, typst, `weight: "bold"`)
	assertContains(t, typst, `first-line-indent: 12.7mm`)
	assertContains(t, typst, `leading: 2em`)
	assertContains(t, typst, `spacing: 2em`)
	assertContains(t, typst, `margin: 25mm`)
	assertContains(t, typst, `author\@example.invalid`)
	assertContains(t, typst, `+353 1 000 0000`)
	assertContains(t, typst, `#stack(`)
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

func TestTOCPartGapBeforeCanBeConfigured(t *testing.T) {
	root := testProjectRoot(t)
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "script.yaml"), "folio:\n  manuscript:\n    toc:\n      part-gap-before: 1.25em\n")
	writeFile(t, filepath.Join(dir, "ch01.md"), markdownChapterOne())

	output := filepath.Join(dir, "out.typ")
	runManuscript(t, root, filepath.Join(dir, "ch01.md"), output)
	typst := readFile(t, output)

	assertContains(t, typst, `above: if it.level == 1 { 1.25em } else { 0pt }`)
}

func TestTOCPartBoldCanBeDisabled(t *testing.T) {
	root := testProjectRoot(t)
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "script.yaml"), "folio:\n  manuscript:\n    toc:\n      part-bold: false\n")
	writeFile(t, filepath.Join(dir, "ch01.md"), markdownChapterOne())

	output := filepath.Join(dir, "out.typ")
	runManuscript(t, root, filepath.Join(dir, "ch01.md"), output)
	typst := readFile(t, output)

	assertContains(t, typst, `if it.level == 1 and false`)
}

func TestRenderDateUsesConfiguredGoLayoutForISOInput(t *testing.T) {
	if got := renderDate("2026-07-06", "2 January 2006"); got != "6 July 2026" {
		t.Fatalf("unexpected British date render: %q", got)
	}
	if got := renderDate("2026-07-06", "January 2, 2006"); got != "July 6, 2026" {
		t.Fatalf("unexpected US date render: %q", got)
	}
	if got := renderDate("July 2026", "2 January 2006"); got != "July 2026" {
		t.Fatalf("non-ISO dates should render as written, got %q", got)
	}
}

func TestTitlePageDateFormatCanBeConfigured(t *testing.T) {
	root := testProjectRoot(t)
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "script.yaml"), "folio:\n  manuscript:\n    date-format: 2006/01/02\n")
	writeFile(t, filepath.Join(dir, "ch01.md"), markdownChapterOne())

	output := filepath.Join(dir, "out.typ")
	runManuscript(t, root, filepath.Join(dir, "ch01.md"), output)
	typst := readFile(t, output)

	assertContains(t, typst, `2026\/07\/06`)
	assertNotContains(t, typst, `6 July 2026`)
}

func TestMarkdownFrontmatterAndHeadingContract(t *testing.T) {
	doc, err := parseMarkdown(strings.Join([]string{
		"---",
		"title: The Glass Orchard",
		"subtitle: A Novel",
		"author: Example Author",
		"date: 2026-07-06",
		"version: Draft 4",
		"wordcount: about 90,000 words",
		"phone: +353 1 000 0000",
		"---",
		"",
		"# PART ONE",
		"",
		"## Chapter 1",
		"",
		"Body text.",
	}, "\n"))
	if err != nil {
		t.Fatalf("parsing markdown frontmatter: %v", err)
	}

	if doc.Metadata.Title != "The Glass Orchard" || doc.Metadata.Date != "2026-07-06" || doc.Metadata.WordCount != "about 90,000 words" {
		t.Fatalf("frontmatter metadata not parsed: %#v", doc.Metadata)
	}
	if doc.Blocks[0].Kind != "part" || doc.Blocks[0].Text != "PART ONE" {
		t.Fatalf("H1 should be manuscript part, got %#v", doc.Blocks[0])
	}
	if doc.Blocks[1].Kind != "chapter" || doc.Blocks[1].Text != "Chapter 1" {
		t.Fatalf("H2 should be manuscript chapter, got %#v", doc.Blocks[1])
	}

	canonical := RenderMarkdown(doc)
	assertContains(t, canonical, "---\n")
	assertContains(t, canonical, "title: \"The Glass Orchard\"")
	assertContains(t, canonical, "date: \"2026-07-06\"")
	assertContains(t, canonical, "wordcount: \"about 90,000 words\"")
	assertContains(t, canonical, "# PART ONE")
	assertContains(t, canonical, "## Chapter 1")
	assertNotContains(t, canonical, "| Metadata | Value |")
}

func TestMarkdownFrontmatterScalarValuesBecomeStrings(t *testing.T) {
	doc, err := parseMarkdown(strings.Join([]string{
		"---",
		"title: The Glass Orchard",
		"date: 2026-07-06",
		"wordcount: 90000",
		"---",
		"",
		"## Chapter 1",
		"",
		"Body text.",
	}, "\n"))
	if err != nil {
		t.Fatalf("parsing markdown frontmatter: %v", err)
	}
	if doc.Metadata.Date != "2026-07-06" {
		t.Fatalf("date should be rendered as ISO string, got %q", doc.Metadata.Date)
	}
	if doc.Metadata.WordCount != "90000" {
		t.Fatalf("numeric wordcount should be coerced to string, got %q", doc.Metadata.WordCount)
	}
}

func TestMarkdownBodyDashDialogueDoesNotOverrideVersion(t *testing.T) {
	doc, err := parseMarkdown(strings.Join([]string{
		"---",
		"title: The Glass Orchard",
		"version: Draft 4",
		"---",
		"",
		"## Chapter 17",
		"",
		"--- I didn't even ---",
		"",
		"--- Out.",
	}, "\n"))
	if err != nil {
		t.Fatalf("parsing markdown: %v", err)
	}
	if doc.Metadata.Version != "Draft 4" {
		t.Fatalf("body dialogue should not override version, got %q", doc.Metadata.Version)
	}
	assertContains(t, RenderMarkdown(doc), "--- I didn't even ---")
}

func TestContactNameIsOptionalAndDoesNotFallbackToAuthor(t *testing.T) {
	root := testProjectRoot(t)
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "ch01.md"), strings.Join([]string{
		"---",
		"title: The Glass Orchard",
		"author: Example Author",
		"contact-name: Example Agent",
		"email: agent@example.invalid",
		"---",
		"",
		"## Chapter 1",
		"",
		"Body text.",
	}, "\n"))

	output := filepath.Join(dir, "out.typ")
	runManuscript(t, root, filepath.Join(dir, "ch01.md"), output)
	typst := readFile(t, output)

	assertContains(t, typst, `Example Agent`)
	assertContains(t, typst, `agent\@example.invalid`)
	assertNotContains(t, typst, `size: 10pt, weight: "regular")[Example Author]`)
}

func TestAuthorAttributionDefaultsEmptyAndCanPrefixAuthor(t *testing.T) {
	root := testProjectRoot(t)
	dir := t.TempDir()

	withoutAttribution := filepath.Join(dir, "without.md")
	writeFile(t, withoutAttribution, strings.Join([]string{
		"---",
		"title: The Glass Orchard",
		"author: Example Author",
		"---",
		"",
		"## Chapter 1",
		"",
		"Body text.",
	}, "\n"))
	withoutOutput := filepath.Join(dir, "without.typ")
	runManuscript(t, root, withoutAttribution, withoutOutput)
	withoutTypst := readFile(t, withoutOutput)
	assertContains(t, withoutTypst, `)[Example Author]`)
	assertNotContains(t, withoutTypst, `)[by Example Author]`)
	assertNotContains(t, withoutTypst, `)[ Example Author]`)

	withAttribution := filepath.Join(dir, "with.md")
	writeFile(t, withAttribution, strings.Join([]string{
		"---",
		"title: The Glass Orchard",
		"author: Example Author",
		"attribution: by",
		"---",
		"",
		"## Chapter 1",
		"",
		"Body text.",
	}, "\n"))
	withOutput := filepath.Join(dir, "with.typ")
	runManuscript(t, root, withAttribution, withOutput)
	assertContains(t, readFile(t, withOutput), `)[by Example Author]`)
}

func TestMarkdownInlineMarkupAndLiteralDelimitersRenderToTypst(t *testing.T) {
	root := testProjectRoot(t)
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "ch01.md"), strings.Join([]string{
		"---",
		"title: The Glass Orchard",
		"author: Example Author",
		"---",
		"",
		"## Chapter 1",
		"",
		"Kevin thought, _.",
		"",
		"The form had three blanks: _, _, and _.",
		"",
		"Dialogue begins --- like this -- then continues with **bold**, *italic*, and `kevin_murray`.",
		"",
		"// JOKES ABOUT HALLOWEEN",
		"",
		"```",
		"living_room:",
		"  north_wall: 4.20m",
		"```",
	}, "\n"))

	output := filepath.Join(dir, "out.typ")
	runManuscript(t, root, filepath.Join(dir, "ch01.md"), output)
	typst := readFile(t, output)

	assertContains(t, typst, `Kevin thought, \_.`)
	assertContains(t, typst, `The form had three blanks: \_, \_, and \_.`)
	assertContains(t, typst, `Dialogue begins — like this – then continues with *bold*, _italic_, and #text(font: "Libertinus Mono", size: 10pt, weight: "bold")[kevin\_murray].`)
	assertContains(t, typst, `\/\/ JOKES ABOUT HALLOWEEN`)
	assertContains(t, typst, `font: "Libertinus Mono"`)
	assertContains(t, typst, `size: 10pt`)
	assertContains(t, typst, `weight: "bold"`)
	assertContains(t, typst, `#folio-code[living\_room:`)
	assertContains(t, typst, `north\_wall: 4.20m]`)
}

func TestRenderInlineMarkup(t *testing.T) {
	got := renderInlineMarkup("Dialogue begins --- like this -- then continues with **bold**, *italic*, and `kevin_murray`.", "Libertinus Mono", "10pt", "bold")
	want := `Dialogue begins — like this – then continues with *bold*, _italic_, and #text(font: "Libertinus Mono", size: 10pt, weight: "bold")[kevin\_murray].`
	if got != want {
		t.Fatalf("unexpected inline render\nwant: %s\n got: %s", want, got)
	}

	doc, err := parseMarkdown(strings.Join([]string{
		"## Chapter 1",
		"",
		"Dialogue begins --- like this -- then continues with **bold**, *italic*, and `kevin_murray`.",
	}, "\n"))
	if err != nil {
		t.Fatalf("parsing markdown: %v", err)
	}
	canonicalDoc, err := parseMarkdown(RenderMarkdown(doc))
	if err != nil {
		t.Fatalf("parsing canonical markdown: %v", err)
	}
	got = renderInlineMarkup(canonicalDoc.Blocks[1].Text, "Libertinus Mono", "10pt", "bold")
	if got != want {
		t.Fatalf("unexpected canonical inline render\ncanonical: %s\nwant: %s\n got: %s", RenderMarkdown(doc), want, got)
	}
}

func TestOrgInlineMarkupAndSectionBreakRenderThroughCanonicalMarkdown(t *testing.T) {
	root := testProjectRoot(t)
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "ch01.org"), strings.Join([]string{
		"#+TITLE: The Glass Orchard",
		"#+AUTHOR: Example Author",
		"",
		"** Chapter 1",
		"Dialogue begins --- like this -- then continues with *bold*, /italic/, and `kevin_murray`.",
		"",
		"-----",
		"",
		"_____",
		"",
		"Kevin thought, _.",
	}, "\n"))

	output := filepath.Join(dir, "out.typ")
	runManuscript(t, root, filepath.Join(dir, "ch01.org"), output)
	typst := readFile(t, output)

	assertContains(t, typst, `Dialogue begins — like this – then continues with *bold*, _italic_, and #text(font: "Libertinus Mono", size: 10pt, weight: "bold")[kevin\_murray].`)
	assertContains(t, typst, `#folio-scene-break()`)
	assertContains(t, typst, `Kevin thought, \_.`)
}

func TestOrgManuscriptRenderingUsesCanonicalMarkdown(t *testing.T) {
	dir := t.TempDir()
	mdInput := filepath.Join(dir, "source.md")
	orgInput := filepath.Join(dir, "source.org")
	writeFile(t, mdInput, markdownChapterOne())
	writeFile(t, orgInput, orgChapterOne())

	mdDoc, err := Parse("markdown", readFile(t, mdInput))
	if err != nil {
		t.Fatalf("parsing markdown: %v", err)
	}
	orgDoc, err := Parse("org", readFile(t, orgInput))
	if err != nil {
		t.Fatalf("parsing org: %v", err)
	}
	if RenderMarkdown(mdDoc) != RenderMarkdown(orgDoc) {
		t.Fatalf("Markdown and org did not canonicalize to the same Markdown\n--- markdown ---\n%s\n--- org ---\n%s", RenderMarkdown(mdDoc), RenderMarkdown(orgDoc))
	}
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

	shortHelp := runManuscriptOutput(t, root, "-h")
	assertContains(t, shortHelp, `Usage: folio manuscript`)
	assertContains(t, shortHelp, `--dry-run`)

	longHelp := runManuscriptOutput(t, root, "--help")
	assertContains(t, longHelp, `Usage: folio manuscript`)
	assertContains(t, longHelp, `--dry-run`)

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
	assertContains(t, pdfText, `Example Author / The Glass Orchard / 1`)
	assertContains(t, pdfText, `Example Author / The Glass Orchard / 2`)
	assertBefore(t, pdfText, `Contents`, `PART ONE`)
	assertBefore(t, pdfText, `PART ONE`, `Chapter 1`)
}

func TestDummyMarkdownAndOrgExamplesRenderSamePDFText(t *testing.T) {
	requireTool(t, "typst")
	requireTool(t, "pdftotext")

	root := testProjectRoot(t)
	assertExampleTypstMatches(t, root, "british")
	assertExampleTypstMatches(t, root, "us")

	britishMarkdown := renderExamplePDFText(t, root, "british", "dummy-manuscript.md")
	britishOrg := renderExamplePDFText(t, root, "british", "dummy-manuscript.org")
	if britishMarkdown != britishOrg {
		t.Fatalf("British Markdown and org examples produced different PDF text\n--- markdown ---\n%s\n--- org ---\n%s", britishMarkdown, britishOrg)
	}

	usMarkdown := renderExamplePDFText(t, root, "us", "dummy-manuscript.md")
	usOrg := renderExamplePDFText(t, root, "us", "dummy-manuscript.org")
	if usMarkdown != usOrg {
		t.Fatalf("US Markdown and org examples produced different PDF text\n--- markdown ---\n%s\n--- org ---\n%s", usMarkdown, usOrg)
	}
}

func assertExampleTypstMatches(t *testing.T, root string, style string) {
	t.Helper()
	dir := t.TempDir()
	mdTypst := filepath.Join(dir, "md-"+style+".typ")
	orgTypst := filepath.Join(dir, "org-"+style+".typ")
	runManuscript(t, root, "--style", style, filepath.Join(root, "examples", "dummy-manuscript.md"), mdTypst)
	runManuscript(t, root, "--style", style, filepath.Join(root, "examples", "dummy-manuscript.org"), orgTypst)
	if readFile(t, mdTypst) != readFile(t, orgTypst) {
		t.Fatalf("%s Markdown and org examples produced different Typst\n--- markdown ---\n%s\n--- org ---\n%s", style, readFile(t, mdTypst), readFile(t, orgTypst))
	}
}

func renderExamplePDFText(t *testing.T, root string, style string, name string) string {
	t.Helper()
	dir := t.TempDir()
	output := filepath.Join(dir, strings.TrimSuffix(name, filepath.Ext(name))+"-"+style+".pdf")
	input := filepath.Join(root, "examples", name)
	runManuscript(t, root, "--style", style, input, output)

	textPath := filepath.Join(dir, strings.TrimSuffix(name, filepath.Ext(name))+"-"+style+".txt")
	commandOutput(t, exec.Command("pdftotext", "-layout", output, textPath))
	return readFile(t, textPath)
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
		"---",
		"title: The Glass Orchard",
		"subtitle: A Novel",
		"author: Example Author",
		"date: 2026-07-06",
		"version: Draft 4",
		"wordcount: 90000",
		"contact-name: Example Agent",
		"address: 100 Example Street / Sample City / Exampleland",
		"phone: +353 1 000 0000",
		"email: author@example.invalid",
		"website: https://example.invalid",
		"---",
		"",
		"# PART ONE",
		"",
		"## Chapter 1",
		"",
		"The rain had been falling since Tuesday, though nobody could agree which Tuesday had started it.",
		"",
		"Mira found the `watch` under the loose floorboard.",
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
		"## Chapter 2",
		"",
		"The first rule of time travel was that nobody should do it before breakfast.",
	}, "\n")
}

func orgChapterOne() string {
	return strings.Join([]string{
		"#+TITLE: The Glass Orchard",
		"#+SUBTITLE: A Novel",
		"#+AUTHOR: Example Author",
		"#+DATE: 2026-07-06",
		"#+VERSION: Draft 4",
		"#+WORDCOUNT: 90000",
		"#+CONTACT-NAME: Example Agent",
		"#+ADDRESS: 100 Example Street / Sample City / Exampleland",
		"#+PHONE: +353 1 000 0000",
		"#+EMAIL: author@example.invalid",
		"#+WEBSITE: https://example.invalid",
		"",
		"* PART ONE",
		"** Chapter 1",
		"The rain had been falling since Tuesday, though nobody could agree which Tuesday had started it.",
		"",
		"Mira found the `watch` under the loose floorboard.",
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
