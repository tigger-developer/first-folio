// ABOUTME: Exercises compact script titles through real Typst and PDF output.
// ABOUTME: Protects metadata, running matter, act pagination, and mode boundaries.
package app

import (
	"fmt"
	"github.com/tigger-developer/first-folio/internal/config"
	"github.com/tigger-developer/first-folio/internal/play"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func scriptTitlePDFPages(t *testing.T, sourceText, extension, settings string) []string {
	t.Helper()
	for _, tool := range []string{"typst", "pdftotext"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("%s required: %v", tool, err)
		}
	}
	dir := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	source, target := filepath.Join(dir, "source"+extension), filepath.Join(dir, "out.pdf")
	writeAppFile(t, source, sourceText)
	if settings != "" {
		writeAppFile(t, filepath.Join(dir, "script.yaml"), settings)
	}
	status, stdout, stderr := runApp(t, "convert", source, target)
	if status != 0 {
		t.Fatalf("status %d stdout: %s stderr: %s", status, stdout, stderr)
	}
	raw, err := exec.Command("pdftotext", "-layout", target, "-").CombinedOutput()
	if err != nil {
		t.Fatalf("extract PDF: %v: %s", err, raw)
	}
	return strings.Split(strings.TrimSuffix(string(raw), "\f"), "\f")
}

const scriptTitleSource = `#+TITLE: Submission Play
#+SUBTITLE: A Short Play
#+AUTHOR: Example Author
#+DATE: 7 October 2026
#+VERSION: Draft Two
* ACT ONE
**** CÁIT
Opening dialogue.
* ACT TWO
**** CÁIT
Later dialogue.
`

func TestCompactScriptTitleSharesFirstPDFPage(t *testing.T) {
	dedicated := scriptTitlePDFPages(t, scriptTitleSource, ".org", "")
	compact := scriptTitlePDFPages(t, scriptTitleSource, ".org", `folio:
  title-page:
    enabled: false
    skip-header: false
    skip-footer: false
  page-header:
    enabled: true
    format: "Running [title] / [author] / [page]"
  page-footer:
    format: "Running footer [page]"
`)
	if len(dedicated) != 3 || len(compact) != 2 {
		t.Fatalf("dedicated pages=%d compact pages=%d; compact: %q", len(dedicated), len(compact), compact)
	}
	assertTextContains(t, compact[0], []string{"Submission Play", "A Short Play", "by Example Author", "7 October 2026", "Draft Two", "ACT ONE", "Opening dialogue.", "Running Submission Play / Example Author / 1", "Running footer 1"})
	assertTextContains(t, compact[1], []string{"ACT TWO", "Later dialogue."})
	for _, text := range []string{"ACT ONE", "Opening dialogue."} {
		if strings.Contains(dedicated[0], text) {
			t.Errorf("dedicated title page contains body %q", text)
		}
	}
	if strings.Contains(compact[0], "ACT TWO") {
		t.Error("later act failed to break")
	}
	for _, text := range []string{"7 October 2026", "Draft Two"} {
		if strings.Count(strings.Join(compact, ""), text) != 1 {
			t.Errorf("metadata %q duplicated", text)
		}
		if strings.Index(compact[0], text) > strings.Index(compact[0], "ACT ONE") {
			t.Errorf("metadata %q not above body", text)
		}
	}
}

func TestScriptTitleRunningMatterFirstPage(t *testing.T) {
	for _, dedicated := range []bool{true, false} {
		for _, skipHeader := range []bool{true, false} {
			for _, skipFooter := range []bool{true, false} {
				t.Run(fmt.Sprintf("dedicated=%t/header=%t/footer=%t", dedicated, skipHeader, skipFooter), func(t *testing.T) {
					settings := fmt.Sprintf(`folio:
  title-page:
    enabled: %t
    skip-header: %t
    skip-footer: %t
  page-header:
    enabled: true
    format: "Running header [page]"
  page-footer:
    enabled: true
    format: "Running footer [page]"
`, dedicated, skipHeader, skipFooter)
					pages := scriptTitlePDFPages(t, scriptTitleSource, ".org", settings)
					wantPages := 2
					if dedicated {
						wantPages = 3
					}
					if len(pages) != wantPages {
						t.Fatalf("got %d pages, want %d: %q", len(pages), wantPages, pages)
					}
					for _, matter := range []struct {
						text string
						skip bool
					}{{"Running header", skipHeader}, {"Running footer", skipFooter}} {
						if got := strings.Contains(pages[0], matter.text+" 1"); got == matter.skip {
							t.Errorf("page one %s presence=%t, skip=%t: %q", matter.text, got, matter.skip, pages[0])
						}
						for i := 1; i < len(pages); i++ {
							assertTextContains(t, pages[i], []string{fmt.Sprintf("%s %d", matter.text, i+1)})
						}
					}
					assertTextContains(t, pages[0], []string{"7 October 2026", "Draft Two"})
					for _, metadata := range []string{"7 October 2026", "Draft Two"} {
						if strings.Count(strings.Join(pages, ""), metadata) != 1 {
							t.Errorf("metadata %q must appear once", metadata)
						}
					}
				})
			}
		}
	}
}

func TestScriptTitleRunningMatterDefaultsAndSharedDisabled(t *testing.T) {
	for _, dedicated := range []bool{true, false} {
		for _, shared := range []bool{true, false} {
			t.Run(fmt.Sprintf("dedicated=%t/shared=%t", dedicated, shared), func(t *testing.T) {
				switches := ""
				if !shared {
					switches = "    skip-header: false\n    skip-footer: false\n"
				}
				settings := fmt.Sprintf(`folio:
  title-page:
    enabled: %t
%s  page-header:
    enabled: %t
    format: "Shared header [page]"
  page-footer:
    enabled: %t
    format: "Shared footer [page]"
`, dedicated, switches, shared, shared)
				pages := scriptTitlePDFPages(t, scriptTitleSource, ".org", settings)
				assertTextContains(t, pages[0], []string{"7 October 2026", "Draft Two"})
				for i, page := range pages {
					for _, matter := range []string{"Shared header", "Shared footer"} {
						want := shared && i > 0
						if got := strings.Contains(page, matter); got != want {
							t.Errorf("page %d %s presence=%t, want %t", i+1, matter, got, want)
						}
					}
				}
			})
		}
	}
}

func TestScriptTitleRunningMatterFrontmatterBranches(t *testing.T) {
	for _, dedicated := range []bool{true, false} {
		t.Run(fmt.Sprintf("dedicated=%t", dedicated), func(t *testing.T) {
			source := strings.Replace(scriptTitleSource, "* ACT ONE", "* Introduction\n"+strings.Repeat("Introductory text fills the frontmatter page.\n\n", 70)+"* ACT ONE", 1)
			settings := fmt.Sprintf(`folio:
  title-page:
    enabled: %t
    skip-header: false
    skip-footer: false
  page-header:
    enabled: true
    format: "Body even [page]"
    alt-format: "Body odd [page]"
    frontmatter-format: "Front even [page]"
    alt-frontmatter-format: "Front odd [page]"
  page-footer:
    format: "Footer even [page]"
    alt-format: "Footer odd [page]"
    frontmatter-format: "Preface even [page]"
    alt-frontmatter-format: "Preface odd [page]"
`, dedicated)
			pages := scriptTitlePDFPages(t, source, ".org", settings)
			assertTextContains(t, pages[0], []string{"Front odd 1", "Preface odd 1", "7 October 2026", "Draft Two"})
			if len(pages) < 4 {
				t.Fatalf("need multiple frontmatter and body pages, got %d", len(pages))
			}
			assertTextContains(t, pages[1], []string{"Front even 2", "Preface even 2"})
			for i, page := range pages {
				if strings.Contains(page, "ACT TWO") {
					parity := "odd"
					if (i+1)%2 == 0 {
						parity = "even"
					}
					assertTextContains(t, page, []string{fmt.Sprintf("Body %s %d", parity, i+1), fmt.Sprintf("Footer %s %d", parity, i+1)})
				}
			}
		})
	}
}

func TestCompactScriptTitleSkipsPhysicalFirstPageOnly(t *testing.T) {
	source := strings.Replace(scriptTitleSource, "* ACT ONE", "* Introduction\n"+strings.Repeat("Introductory text fills the frontmatter page.\n\n", 70)+"* ACT ONE", 1)
	pages := scriptTitlePDFPages(t, source, ".org", `folio:
  title-page:
    enabled: false
  page-header:
    enabled: true
    format: "Physical header [page]"
  page-footer:
    format: "Physical footer [page]"
`)
	if len(pages) < 4 {
		t.Fatalf("need multiple pages, got %d", len(pages))
	}
	for i, page := range pages {
		for _, matter := range []string{"Physical header", "Physical footer"} {
			if got := strings.Contains(page, matter); got != (i > 0) {
				t.Errorf("page %d %s presence=%t", i+1, matter, got)
			}
		}
	}
	assertTextContains(t, pages[1], []string{"Introductory text"})
}

func TestCompactScriptTitleKeepsIntroWithFirstAct(t *testing.T) {
	source := strings.Replace(scriptTitleSource, "* ACT ONE", "* Introduction\nA short introduction.\n* ACT ONE", 1)
	pages := scriptTitlePDFPages(t, source, ".org", "folio:\n  title-page:\n    enabled: false\n")
	if len(pages) != 2 {
		t.Fatalf("got %d pages, want 2: %q", len(pages), pages)
	}
	assertTextContains(t, pages[0], []string{"Submission Play", "A short introduction.", "ACT ONE", "Opening dialogue."})
	assertTextContains(t, pages[1], []string{"ACT TWO", "Later dialogue."})
}

func TestCompactScriptTitleAllFormatsAndStyles(t *testing.T) {
	for format, source := range scriptRoundTripSources {
		for _, style := range []string{"british", "us", "screenplay"} {
			t.Run(string(format)+"/"+style, func(t *testing.T) {
				pages := scriptTitlePDFPages(t, source, scriptFormatExtensions[format], "folio:\n  style: "+style+"\n  title-page:\n    enabled: false\n")
				if len(pages) != 1 {
					t.Fatalf("got %d pages, want 1: %q", len(pages), pages)
				}
				assertTextContains(t, pages[0], []string{"Matrix Play", "Tadhg O’Brien", "ACT ONE", "Hello."})
				if style == "screenplay" {
					assertTextContains(t, pages[0], []string{"Written", "by"})
				}
			})
		}
	}
}

func TestCompactScriptTitleMissingMetadata(t *testing.T) {
	for _, metadata := range []string{"", "#+AUTHOR: Lone Author\n#+DATE: A Date\n#+VERSION: A Version\n", "#+SUBTITLE: Lone Subtitle\n"} {
		t.Run(metadata, func(t *testing.T) {
			pages := scriptTitlePDFPages(t, metadata+"* ACT ONE\n**** CÁIT\nOpening dialogue.\n", ".org", "folio:\n  title-page:\n    enabled: false\n")
			if len(pages) != 1 {
				t.Fatalf("got %d pages, want 1", len(pages))
			}
			assertTextContains(t, pages[0], []string{"ACT ONE", "Opening dialogue."})
			if strings.Contains(metadata, "Lone Author") {
				assertTextContains(t, pages[0], []string{"by Lone Author", "A Date", "A Version"})
			}
			if strings.Contains(metadata, "Lone Subtitle") {
				assertTextContains(t, pages[0], []string{"Lone Subtitle"})
			}
		})
	}
}

func TestCompactScriptTitleEscapingAndFontRoles(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	source, target := filepath.Join(dir, "source.org"), filepath.Join(dir, "out.typ")
	text := "#+TITLE: Hash #title [literal]\n#+SUBTITLE: Sub #value\n#+AUTHOR: Name #author\n#+DATE: Date #date\n#+VERSION: Version #version\n* ACT ONE\n**** CÁIT\nOpening dialogue.\n"
	settings := `folio:
  title-page:
    enabled: false
    title:
      font:
        size: 21pt
    subtitle:
      space-before: 0.7em
      font:
        size: 13pt
    author:
      prefix: "Written\nby\n"
      space-before: 0.9em
      font:
        size: 11pt
    date:
      position: bottom-right
      font:
        size: 9pt
    version:
      position: bottom-left
      font:
        size: 8pt
`
	writeAppFile(t, source, text)
	writeAppFile(t, filepath.Join(dir, "script.yaml"), settings)
	status, stdout, stderr := runApp(t, "convert", source, target)
	if status != 0 {
		t.Fatalf("status %d stdout: %s stderr: %s", status, stdout, stderr)
	}
	output := readAppFile(t, target)
	for _, role := range []struct{ size, value string }{{"21pt", "Hash #title [literal]"}, {"13pt", "Sub #value"}, {"11pt", "Name #author"}, {"9pt", "Date #date"}, {"8pt", "Version #version"}} {
		if !strings.Contains(output, "size: "+role.size) || !strings.Contains(output, escapeTypstContent(role.value)) {
			t.Errorf("missing font or escaped content: %#v", role)
		}
	}
	assertTextContains(t, output, []string{"#v(0.7em)", "#v(0.9em)", ")[Written]", ")[by]"})
	for _, forbidden := range []string{"#v(30%)", "#v(40%)", "#pagebreak()", "#pagebreak(weak: true)"} {
		if strings.Contains(output, forbidden) {
			t.Errorf("compact output contains %q", forbidden)
		}
	}
	pages := scriptTitlePDFPages(t, text, ".org", settings)
	if len(pages) != 1 {
		t.Fatalf("got %d pages, want 1", len(pages))
	}
	assertTextContains(t, pages[0], []string{"Hash #title [literal]", "Sub #value", "Written", "by", "Name #author", "Date #date", "Version #version", "Opening dialogue."})
}

func TestInvalidScriptTitleSwitchDoesNotWriteOutput(t *testing.T) {
	for _, field := range []string{"enabled", "skip-header", "skip-footer"} {
		for _, value := range []string{"null", "\"false\"", "{}"} {
			t.Run(field+"/"+value, func(t *testing.T) {
				dir := t.TempDir()
				t.Setenv("HOME", t.TempDir())
				source, target := filepath.Join(dir, "source.org"), filepath.Join(dir, "out.typ")
				writeAppFile(t, source, scriptTitleSource)
				writeAppFile(t, filepath.Join(dir, "script.yaml"), "folio:\n  title-page:\n    "+field+": "+value+"\n")
				status, _, stderr := runApp(t, "convert", source, target)
				if status == 0 || !strings.Contains(stderr, "folio.title-page."+field+" must be a boolean") {
					t.Fatalf("status=%d stderr=%s", status, stderr)
				}
				if _, err := os.Stat(target); !os.IsNotExist(err) {
					t.Fatalf("invalid switch wrote output: %v", err)
				}
			})
		}
	}
}

func TestCompactScriptTitleSuppressesOnlyFirstActBreakAfterOpeningDirection(t *testing.T) {
	source := strings.Replace(scriptTitleSource, "* ACT ONE", "*** Opening stage direction.\n* ACT ONE", 1)
	pages := scriptTitlePDFPages(t, source, ".org", "folio:\n  title-page:\n    enabled: false\n")
	if len(pages) != 2 {
		t.Fatalf("got %d pages, want 2: %q", len(pages), pages)
	}
	assertTextContains(t, pages[0], []string{"Submission Play", "Opening stage direction.", "ACT ONE", "Opening dialogue."})
	assertTextContains(t, pages[1], []string{"ACT TWO", "Later dialogue."})
}

func TestCompactScriptTitleFirstActBreakAfterBodyStart(t *testing.T) {
	cfg, err := config.Load(config.Options{Mode: config.ModeScript, Home: t.TempDir(), LocalDir: t.TempDir(), Source: map[string]any{"folio": map[string]any{"title-page": map[string]any{"enabled": false}}}})
	if err != nil {
		t.Fatal(err)
	}
	doc := play.Document{Metadata: map[string]string{"title": "Opening"}, Events: []play.Event{
		{Kind: play.EventStageDirection, Text: "Opening direction."},
		{Kind: play.EventActHeader, Text: "ACT ONE"},
		{Kind: play.EventDialogue, Text: "Opening dialogue."},
		{Kind: play.EventActHeader, Text: "ACT TWO"},
	}}
	source, err := renderPlayTypst(doc, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Count(source, "#pagebreak(weak: true)"); got != 1 {
		t.Fatalf("got %d act breaks, want later act only", got)
	}
	firstAct := strings.Index(source, "#act-header[ACT ONE]")
	if strings.Index(source, "#pagebreak(weak: true)") < firstAct {
		t.Fatal("first act must remain with opening body")
	}
}

func TestCompactScriptTitleAllowsNaturalPagination(t *testing.T) {
	source := "#+TITLE: Long Submission\n#+AUTHOR: Example Author\n* ACT ONE\n"
	source += strings.Repeat("**** CÁIT\nA deliberately repeated line of dialogue to fill the script pages.\n", 70)
	pages := scriptTitlePDFPages(t, source, ".org", "folio:\n  title-page:\n    enabled: false\n")
	if len(pages) < 2 {
		t.Fatalf("long script did not paginate: %d pages", len(pages))
	}
	assertTextContains(t, pages[0], []string{"Long Submission", "Example Author", "ACT ONE", "A deliberately repeated line"})
	if strings.Count(strings.Join(pages, ""), "Long Submission") != 1 {
		t.Error("compact heading repeated on continuation pages")
	}
}

func TestScriptTitleExplicitTruePreservesDefaultPDF(t *testing.T) {
	defaults := scriptTitlePDFPages(t, scriptTitleSource, ".org", "")
	enabled := scriptTitlePDFPages(t, scriptTitleSource, ".org", "folio:\n  title-page:\n    enabled: true\n")
	if strings.Join(defaults, "\f") != strings.Join(enabled, "\f") {
		t.Fatal("explicit enabled=true changed dedicated title output")
	}
}
