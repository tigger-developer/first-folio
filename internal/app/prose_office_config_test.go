// ABOUTME: Proves configuration layering and invalid-value rejection through office artefacts.
// ABOUTME: Uses explicit independent expected values for global, source, local, style and CLI layers.
package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRT063_4_OfficeConfigurationLayering(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		for _, tc := range []struct {
			name, local, cli string
			margin           float64
		}{
			{"source-style", "", "", 22},
			{"local-style", "british", "", 15},
			{"cli-style", "british", "us", 22},
		} {
			t.Run(format+"/"+tc.name, func(t *testing.T) {
				home, dir := t.TempDir(), t.TempDir()
				t.Setenv("HOME", home)
				global := filepath.Join(home, ".config", "first-folio")
				if err := os.MkdirAll(global, 0700); err != nil {
					t.Fatal(err)
				}
				proseSource(t, global, "script.yaml", "folio:\n  style: british\n  manuscript:\n    margin: 30mm\n    font:\n      family: Iosevka Publish Extended\n      size: 9pt\n")
				proseSource(t, global, "script-us.yaml", "folio:\n  manuscript:\n    margin: 40mm\n")
				src := proseSource(t, dir, "source.md", "---\ntitle: SOURCE-TITLE\nfolio:\n  manuscript:\n    style: us\n    margin: 15mm\n    page: a5\n    font:\n      family: Iosevka Publish\n      size: 11pt\n    title-page:\n      enabled: false\n    toc:\n      enabled: false\n---\nLAYERED-FONT-MARKER\n")
				local := "folio:\n  manuscript:\n    page: 6x9in\n    font:\n      size: 13pt\n"
				if tc.local != "" {
					local += "    style: " + tc.local + "\n"
				}
				proseSource(t, dir, "script.yaml", local)
				proseSource(t, dir, "script-us.yaml", "folio:\n  manuscript:\n    margin: 22mm\n")
				target := filepath.Join(dir, "output."+format)
				args := []string{"manuscript", src, target}
				if tc.cli != "" {
					args = append(args, "--style", tc.cli)
				}
				status, _, stderr := runApp(t, args...)
				if status != 0 {
					t.Fatalf("layered export: %s", stderr)
				}
				pkg := prosePackage(t, target, format)
				body := officeBody(t, pkg, format)
				layers := officeEffective(t, pkg, format, officeParagraph(t, body, "LAYERED-FONT-MARKER"))
				if format == "docx" {
					officeAssertEffective(t, layers, "rFonts", "ascii", "Iosevka Publish")
					officeAssertEffective(t, layers, "sz", "val", "26")
				} else {
					officeAssertEffective(t, layers, "text-properties", "font-size", "13pt")
				}
				page := []*officeXML{body}
				if format == "odt" {
					page = []*officeXML{pkg["styles.xml"]}
				}
				officeAssertLength(t, page, format, "pgSz", "w", "page-layout-properties", "page-width", 432)
				officeAssertLength(t, page, format, "pgSz", "h", "page-layout-properties", "page-height", 648)
				officeAssertLength(t, page, format, "pgMar", "top", "page-layout-properties", "margin-top", tc.margin*72/25.4)
			})
		}
	}
}

func TestRT063_4_OfficeInvalidEffectiveConfiguration(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		for _, tc := range []struct{ name, yaml, path string }{
			{"font", "font:\n      size: banana", "folio.manuscript.font.size"},
			{"alignment", "chapter:\n      align: diagonal", "folio.manuscript.chapter.align"},
			{"line-spacing", "line-spacing: 1em", "folio.manuscript.line-spacing"},
			{"paragraph-length", "paragraph-spacing: huge", "folio.manuscript.paragraph-spacing"},
			{"quote-indent", "quote-block-indent: inward", "folio.manuscript.quote-block-indent"},
		} {
			t.Run(format+"/"+tc.name, func(t *testing.T) {
				t.Setenv("HOME", t.TempDir())
				dir := t.TempDir()
				src := proseSource(t, dir, "source.md", "---\nfolio:\n  manuscript:\n    "+tc.yaml+"\n---\nBODY\n")
				target := proseSource(t, dir, "output."+format, "PRIOR")
				status, _, stderr := runApp(t, "manuscript", src, target)
				if status == 0 || !strings.Contains(stderr, tc.path) {
					t.Errorf("invalid configuration must name %s: status=%d stderr=%s", tc.path, status, stderr)
				}
				raw, err := os.ReadFile(target)
				if err != nil || string(raw) != "PRIOR" {
					t.Errorf("invalid configuration replaced destination: %v", err)
				}
			})
		}
	}
}

func TestRT063_5_OfficeChapterNumberReset(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		for _, reset := range []string{"never", "per-part"} {
			t.Run(format+"/"+reset, func(t *testing.T) {
				t.Setenv("HOME", t.TempDir())
				dir := t.TempDir()
				cfg := "folio:\n  manuscript:\n    title-page:\n      enabled: false\n    toc:\n      enabled: false\n    part:\n      show-number: true\n      prefix: 'Part '\n      separator: ': '\n      number-format: I\n    chapter:\n      show-number: true\n      prefix: 'Chapter '\n      separator: ': '\n      number-format: I.1\n      number-reset: " + reset + "\n"
				proseSource(t, dir, "script.yaml", cfg)
				src := proseSource(t, dir, "source.md", "# First\n\n## Opening\n\nFirst body.\n\n## Middle\n\nMiddle body.\n\n# Second\n\n## Ending\n\nLast body.\n")
				pkg := proseConvert(t, src, filepath.Join(dir, "output."+format))
				last := 3
				if reset == "per-part" {
					last = 1
				}
				for _, heading := range []string{"Part I: First", "Chapter I.1: Opening", "Chapter I.2: Middle", "Part II: Second", fmt.Sprintf("Chapter II.%d: Ending", last)} {
					p := officeParagraph(t, officeBody(t, pkg, format), heading)
					if strings.TrimSpace(p.text()) != heading {
						t.Errorf("composed heading=%q; want %q", p.text(), heading)
					}
				}
			})
		}
	}
}
