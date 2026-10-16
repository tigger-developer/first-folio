// ABOUTME: Exercises manuscript and letter office exports through the public command boundary.
// ABOUTME: Checks native packages, semantic content, recipient selection, and publication safety.
package app

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func proseSource(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	writeAppFile(t, path, content)
	return path
}

func prosePackage(t *testing.T, path, format string) map[string]*officeXML {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	z, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		t.Fatalf("invalid %s package: %v", format, err)
	}
	seen := map[string]bool{}
	for _, member := range z.File {
		if seen[member.Name] {
			t.Errorf("duplicate package member %q", member.Name)
		}
		seen[member.Name] = true
		if format == "odt" && member.Name == "mimetype" {
			r, err := member.Open()
			if err != nil {
				t.Fatal(err)
			}
			mime, readErr := io.ReadAll(r)
			closeErr := r.Close()
			if readErr != nil || closeErr != nil || string(mime) != "application/vnd.oasis.opendocument.text" || member.Method != zip.Store {
				t.Fatalf("invalid ODT mimetype: %q, method %d, errors %v/%v", mime, member.Method, readErr, closeErr)
			}
		}
	}
	if format == "odt" && !seen["mimetype"] {
		t.Fatal("ODT mimetype absent")
	}
	pkg := officePackage(t, raw)
	if format == "docx" {
		ct := pkg["[Content_Types].xml"]
		if ct == nil {
			t.Fatal("DOCX content types absent")
		}
		officeAssertAttr(t, ct, "Override", "ContentType", "application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml")
	}
	return pkg
}

func proseConvert(t *testing.T, source, target string) map[string]*officeXML {
	t.Helper()
	status, stdout, stderr := runApp(t, "manuscript", source, target)
	if status != 0 {
		t.Fatalf("manuscript export status %d\nstdout: %s\nstderr: %s", status, stdout, stderr)
	}
	return prosePackage(t, target, strings.ToLower(strings.TrimPrefix(filepath.Ext(target), ".")))
}

func TestRT063_1_ManuscriptOfficeSelection(t *testing.T) {
	for _, format := range []string{"docx", "odt", "DOCX", "ODT"} {
		for _, source := range []struct{ suffix, content string }{
			{"md", "---\ntitle: Selection manuscript\n---\n\n## Chapter One\n\nVISIBLE-SELECTION.\n"},
			{"org", "#+TITLE: Selection manuscript\n** Chapter One\nVISIBLE-SELECTION.\n"},
		} {
			t.Run(format+"/"+source.suffix, func(t *testing.T) {
				t.Setenv("HOME", t.TempDir())
				dir := t.TempDir()
				src := proseSource(t, dir, "source with spaces."+source.suffix, source.content)
				target := filepath.Join(dir, "output with spaces."+format)
				pkg := proseConvert(t, src, target)
				body := officeBody(t, pkg, strings.ToLower(format))
				if strings.Count(body.text(), "VISIBLE-SELECTION") != 1 {
					t.Fatalf("source marker missing or duplicated: %s", body.text())
				}
			})
		}
	}
}

func TestRT063_1_ManuscriptOfficeDryRun(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		t.Run(format, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			dir := t.TempDir()
			src := proseSource(t, dir, "source.md", "## Chapter One\n\nDRY-RUN-CONTENT.\n")
			target := filepath.Join(dir, "output."+format)
			status, stdout, stderr := runApp(t, "manuscript", "--dry-run", src, target)
			if status != 0 || !strings.Contains(strings.ToLower(stdout), "output format: "+format+"\n") {
				t.Fatalf("dry run must identify %s: status=%d stdout=%s stderr=%s", format, status, stdout, stderr)
			}
			entries, err := os.ReadDir(dir)
			if err != nil || len(entries) != 1 {
				t.Fatalf("dry run created files: %v, %v", entries, err)
			}
		})
	}
}

func TestRT063_2_OfficeLetterRecipientBatches(t *testing.T) {
	fixture, err := os.ReadFile("../../testdata/prose-office/letters.org")
	if err != nil {
		t.Fatal(err)
	}
	for _, format := range []string{"docx", "odt"} {
		for _, filter := range []string{"", "fIrSt tHeAtRe", "RECIPIENT-TWO"} {
			t.Run(format+"/"+filter, func(t *testing.T) {
				t.Setenv("HOME", t.TempDir())
				dir := t.TempDir()
				src := proseSource(t, dir, "letters.org", string(fixture))
				out := t.TempDir()
				args := []string{"letter", src, "--format", format, "--dir", out, "--prefix", "submission"}
				if format == "odt" {
					args = []string{"letter", src, "--format=odt", "--dir", out, "--prefix", "submission"}
				}
				if filter != "" {
					args = append(args, "--to", filter)
				}
				status, stdout, stderr := runApp(t, args...)
				if status != 0 {
					t.Fatalf("letter export status %d: %s", status, stderr)
				}
				for _, recipient := range []struct {
					slug, salutation, forbidden string
					selected                    bool
				}{
					{"first-theatre", "Dear Alex", "Dear Sam", filter != "RECIPIENT-TWO"},
					{"second-theatre", "Dear Sam", "Dear Alex", filter != "fIrSt tHeAtRe"},
				} {
					path := filepath.Join(out, "submission-"+recipient.slug+"."+format)
					if !recipient.selected {
						if _, err := os.Stat(path); !os.IsNotExist(err) {
							t.Errorf("unselected recipient output exists: %s", path)
						}
						continue
					}
					text := officeBody(t, prosePackage(t, path, format), format).text()
					if !strings.Contains(text, recipient.salutation) || strings.Contains(text, recipient.forbidden) {
						t.Errorf("incorrect recipient substitution: %s", text)
					}
					if !strings.Contains(stdout, "Generated: "+path) {
						t.Errorf("missing generated-file report: %s", stdout)
					}
				}
			})
		}
	}
}

func TestRT063_7_OfficeLetterLiterals(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		t.Run(format, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			dir := t.TempDir()
			for _, file := range []struct{ source, target string }{{"letters.org", "letters.org"}, {"letter.yaml", "script.yaml"}} {
				raw, err := os.ReadFile("../../testdata/prose-office/" + file.source)
				if err != nil {
					t.Fatal(err)
				}
				proseSource(t, dir, file.target, string(raw))
			}
			status, _, stderr := runApp(t, "letter", filepath.Join(dir, "letters.org"), "--format", format, "--to", "First")
			if status != 0 {
				t.Fatalf("letter export: %s", stderr)
			}
			pkg := prosePackage(t, filepath.Join(dir, "letter-first-theatre."+format), format)
			body := officeBody(t, pkg, format)
			for _, want := range []string{"https://example.invalid/submissions/letter", "[unresolved]", "16 October 2026", "BOLD-SAMPLE", "ITALIC-SAMPLE", "UNDERLINE-SAMPLE"} {
				if !strings.Contains(body.text(), want) {
					t.Errorf("missing letter text %q", want)
				}
			}
			for _, bad := range []string{"*BOLD-SAMPLE*", "/ITALIC-SAMPLE/", "_UNDERLINE-SAMPLE_", "=https://"} {
				if strings.Contains(body.text(), bad) {
					t.Errorf("literal markup leaked: %q", bad)
				}
			}
			for name, member := range pkg {
				if strings.Contains(member.text(), "LETTER-MUST-NOT-HAVE-") {
					t.Errorf("letter inherited running matter in %s", name)
				}
			}
		})
	}
}

// Resolve the style of the smallest native text span containing the marker.
// Paragraph descendants cannot supply another span's emphasis as evidence.
func proseInlineLayers(t *testing.T, pkg map[string]*officeXML, format, marker string) []*officeXML {
	t.Helper()
	kind := "r"
	if format == "odt" {
		kind = "span"
	}
	var selected *officeXML
	for _, n := range officeBody(t, pkg, format).all(kind) {
		if strings.Contains(n.text(), marker) && (selected == nil || len(n.text()) < len(selected.text())) {
			selected = n
		}
	}
	if selected == nil {
		t.Fatalf("no native %s containing %q", kind, marker)
	}
	proxy := &officeXML{Attr: selected.Attr}
	if format == "docx" {
		for _, ref := range selected.all("rStyle") {
			proxy.Children = append(proxy.Children, &officeXML{Name: xml.Name{Local: "pStyle"}, Attr: ref.Attr})
		}
	}
	return append(officeEffective(t, pkg, format, proxy), selected)
}

func TestRT063_3_OfficeManuscriptStructuredContent(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		for _, suffix := range []string{"md", "org"} {
			t.Run(format+"/"+suffix, func(t *testing.T) {
				t.Setenv("HOME", t.TempDir())
				dir := t.TempDir()
				raw, err := os.ReadFile("../../testdata/prose-office/typography." + suffix)
				if err != nil {
					t.Fatal(err)
				}
				src := proseSource(t, dir, "source."+suffix, string(raw))
				if suffix == "org" {
					cfg, err := os.ReadFile("../../testdata/prose-office/typography.yaml")
					if err != nil {
						t.Fatal(err)
					}
					proseSource(t, dir, "script.yaml", string(cfg))
				}
				pkg := proseConvert(t, src, filepath.Join(dir, "output."+format))
				body := officeBody(t, pkg, format)
				text := body.text()
				last := -1
				for _, marker := range []string{"HEADING-SAMPLE", "OPENING-SAMPLE", "BODY-SAMPLE", "BOLD-SAMPLE", "ITALIC-SAMPLE", "STRIKE-SAMPLE", "INLINE-CODE-SAMPLE", "LINK-SAMPLE", "QUOTE-SAMPLE", "QUOTE-SECOND-PARAGRAPH", "CODE-SAMPLE:", "STANDALONE-CODE-SAMPLE", "LIST-FIRST", "LIST-NESTED", "LIST-LAST", "ORDERED-FIRST", "ORDERED-LAST", "TABLE-HEADER-LEFT", "TABLE-CELL-LEFT", "NOTE-ANCHOR", "FINAL-SENTINEL"} {
					// CODE-SAMPLE also appears in INLINE-CODE-SAMPLE; use its full sentence.
					if marker == "CODE-SAMPLE:" {
						marker = "CODE-SAMPLE: 9pt Iosevka Publish."
					}
					at := strings.Index(text, marker)
					if at < last || at < 0 || strings.Count(text, marker) != 1 {
						t.Errorf("missing, duplicated, or reordered marker %q at %d after %d", marker, at, last)
					}
					last = at
				}
				if strings.Contains(text, "PRIVATE-SENTINEL") || !strings.Contains(text, "CÁIT") {
					t.Errorf("private/Unicode content incorrect: %s", text)
				}
				table, list, note, link := "tbl", "numPr", "footnoteReference", "hyperlink"
				if format == "odt" {
					table, list, note, link = "table", "list", "note", "a"
				}
				for _, element := range []string{table, list, note, link} {
					if len(body.all(element)) == 0 {
						t.Errorf("missing editable native %s", element)
					}
				}
				if format == "odt" {
					officeAssertAttr(t, body, "a", "href", "https://example.invalid/manuscript/path")
					officeAssertEffective(t, proseInlineLayers(t, pkg, format, "BOLD-SAMPLE"), "text-properties", "font-weight", "bold")
					officeAssertEffective(t, proseInlineLayers(t, pkg, format, "ITALIC-SAMPLE"), "text-properties", "font-style", "italic")
				} else {
					rels := pkg["word/_rels/document.xml.rels"]
					if rels == nil {
						t.Fatal("DOCX relationships absent")
					}
					officeAssertAttr(t, rels, "Relationship", "Target", "https://example.invalid/manuscript/path")
					for _, inline := range []struct{ marker, property string }{{"BOLD-SAMPLE", "b"}, {"ITALIC-SAMPLE", "i"}} {
						present := false
						for _, layer := range proseInlineLayers(t, pkg, format, inline.marker) {
							for _, property := range layer.all(inline.property) {
								present = property.attr("val") != "false" && property.attr("val") != "0"
							}
						}
						if !present {
							t.Errorf("%s lacks native %s", inline.marker, inline.property)
						}
					}
				}
			})
		}
	}
}

func TestRT063_9_OfficeManuscriptRejectsUnsafeDestinations(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		for _, kind := range []string{"symlink", "hardlink", "directory"} {
			t.Run(format+"/"+kind, func(t *testing.T) {
				t.Setenv("HOME", t.TempDir())
				dir := t.TempDir()
				original := "## Chapter One\n\nSOURCE-MUST-SURVIVE.\n"
				src := proseSource(t, dir, "source.md", original)
				target := filepath.Join(dir, "output."+format)
				var err error
				switch kind {
				case "symlink":
					err = os.Symlink(src, target)
				case "hardlink":
					err = os.Link(src, target)
				case "directory":
					err = os.Mkdir(target, 0700)
				}
				if err != nil {
					t.Fatal(err)
				}
				status, _, stderr := runApp(t, "manuscript", src, target)
				if status == 0 || !strings.Contains(stderr, target) {
					t.Errorf("unsafe destination must fail naming path: status=%d stderr=%s", status, stderr)
				}
				if readAppFile(t, src) != original {
					t.Fatal("source changed during rejected export")
				}
			})
		}
	}
}
