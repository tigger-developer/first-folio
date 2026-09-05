// ABOUTME: Regression tests for promoting complete-line Markdown code spans to manuscript code blocks.
// ABOUTME: Protects ordinary inline code, mixed-content lines, and existing fenced code from reclassification.
package manuscript

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestStandaloneMarkdownCodeSpanUsesCodeBlockLayout(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "chapter.md"), strings.Join([]string{
		"## Chapter 1",
		"",
		"This is `inline_code` here",
		"",
		"```",
		`echo "This is a code block"`,
		"```",
		"",
		"`echo \"This is also a code block`",
		"",
		"  `echo \"Whitespace still makes this a code block\"`  ",
		"",
		"~~~",
		`echo "This fenced block uses tildes"`,
		"~~~",
		"",
		"`echo \"This is not a code block\"` # trailing prose",
		"",
	}, "\n"))

	output := filepath.Join(dir, "manuscript.typ")
	runManuscriptDirect(t, filepath.Join(dir, "chapter.md"), output)
	typst := readFile(t, output)

	assertContains(t, typst, "This is `inline_code` here")
	assertContains(t, typst, strings.Join([]string{
		"```",
		`echo "This is a code block"`,
		"```",
	}, "\n"))
	assertContains(t, typst, strings.Join([]string{
		"```",
		`echo "This is also a code block`,
		"```",
	}, "\n"))
	assertContains(t, typst, strings.Join([]string{
		"```",
		`echo "Whitespace still makes this a code block"`,
		"```",
	}, "\n"))
	assertContains(t, typst, strings.Join([]string{
		"```",
		`echo "This fenced block uses tildes"`,
		"```",
	}, "\n"))
	assertContains(t, typst, "`echo \"This is not a code block\"` \\# trailing prose")
}

func TestStandaloneMarkdownCodeSpanPreprocessingIsSourceAware(t *testing.T) {
	input := strings.Join([]string{
		"  `whole line`  ",
		"before `inline` after",
		"`code` # trailing prose",
		"``code with ` tick``",
		"    `indented code`",
		"```",
		"`already fenced`",
		"```",
		"~~~",
		"`inside tilde fence`",
		"~~~",
		"`unclosed",
		"",
	}, "\n")
	want := strings.Join([]string{
		"```",
		"whole line",
		"```",
		"before `inline` after",
		"`code` # trailing prose",
		"```",
		"code with ` tick",
		"```",
		"    `indented code`",
		"```",
		"`already fenced`",
		"```",
		"~~~",
		"`inside tilde fence`",
		"~~~",
		"`unclosed",
		"",
	}, "\n")

	if got := promoteStandaloneMarkdownCodeSpans(input); got != want {
		t.Fatalf("standalone code preprocessing mismatch\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestStandaloneOrgCodeSpanUsesCanonicalCodeBlockLayout(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "chapter.org"), strings.Join([]string{
		"* Chapter 1",
		"",
		"=echo from org=",
		"",
		"Text with =inline code= here.",
		"",
	}, "\n"))

	output := filepath.Join(dir, "manuscript.typ")
	runManuscriptDirect(t, filepath.Join(dir, "chapter.org"), output)
	typst := readFile(t, output)

	assertContains(t, typst, strings.Join([]string{
		"```",
		"echo from org",
		"```",
	}, "\n"))
	assertContains(t, typst, "Text with `inline code` here.")
}
