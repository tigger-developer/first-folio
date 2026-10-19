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
	"regexp"
	"strings"
	"testing"
	"time"
)

func proseSource(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	writeAppFile(t, path, content)
	return path
}

func TestRT063_7_OfficeLetterFallbacksAndEscaping(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		t.Run(format, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			dir := t.TempDir()
			src := proseSource(t, dir, "letters.org", "#+AUTHOR: Writer & Partner\n* Letters :letter:\n** SUBJECT <A&B> :subject:\nFirst paragraph for [moniker] at [org].\n\nSecond paragraph keeps [unknown] and <literal> & text.\n*** Recipient / Address Line / Cork :to:\n**** Theatre & Company :org:\n**** Friend :moniker:\n")
			before := time.Now().Format("2 January 2006")
			status, _, stderr := runApp(t, "letter", src, "--format", format)
			after := time.Now().Format("2 January 2006")
			if status != 0 {
				t.Fatalf("letter export: %s", stderr)
			}
			pkg := prosePackage(t, filepath.Join(dir, "letter-theatre-company."+format), format)
			body := officeBody(t, pkg, format)
			text := body.text()
			for _, want := range []string{"SUBJECT <A&B>", "First paragraph for Friend at Theatre & Company.", "Second paragraph keeps [unknown] and <literal> & text.", "Yours sincerely,"} {
				if strings.Count(text, want) != 1 {
					t.Errorf("expected one occurrence of %q in letter", want)
				}
			}
			if strings.Count(text, "Writer & Partner") != 2 {
				t.Error("author must supply both fallback sender and fallback signoff")
			}
			if !strings.Contains(text, before) && !strings.Contains(text, after) {
				t.Error("omitted date must resolve to the current local date")
			}
			// The two address separators become native line breaks, not literal slashes.
			breaks := "br"
			if format == "odt" {
				breaks = "line-break"
			}
			address := officeParagraph(t, body, "Address Line")
			if len(address.all(breaks)) != 2 || strings.Contains(address.text(), " / ") {
				t.Error("recipient address did not retain its three editable lines")
			}
		})
	}
}

func prosePackage(t *testing.T, path, format string) map[string]*officeXML {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm() != 0600 {
		t.Fatalf("office output must have private mode 0600: %v, %v", info, err)
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
			src := proseSource(t, dir, "source.md", "---\nfolio:\n  manuscript:\n    copyright:\n      enabled: true\n      isbn: '978-0-000000-00-2'\n      isbn-barcode: render-and-file\n---\n\n## Chapter One\n\nDRY-RUN-CONTENT.\n")
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
	for _, selection := range []string{"docx", "odt", "DOCX", "ODT"} {
		format := strings.ToLower(selection)
		for _, filter := range []string{"", "fIrSt tHeAtRe", "RECIPIENT-TWO"} {
			t.Run(selection+"/"+filter, func(t *testing.T) {
				t.Setenv("HOME", t.TempDir())
				dir := t.TempDir()
				src := proseSource(t, dir, "letters.org", string(fixture))
				out := t.TempDir()
				args := []string{"letter", src, "--format", selection, "--dir", out, "--prefix", "submission"}
				if format == "odt" {
					args = []string{"letter", src, "--format=" + selection, "--dir", out, "--prefix", "submission"}
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

func TestRT063_2_LetterFormatErrors(t *testing.T) {
	fixture, err := os.ReadFile("../../testdata/prose-office/letters.org")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		args []string
		want []string
	}{
		{"unknown-format", []string{"--format", "html"}, []string{"pdf", "docx", "odt"}},
		{"missing-value", []string{"--format"}, []string{"format", "value"}},
		{"invalid-regex", []string{"--format", "docx", "--to", "["}, []string{"recipient", "filter"}},
		{"no-matches", []string{"--format", "odt", "--to", "DOES-NOT-EXIST"}, []string{"no recipients"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			dir, out := t.TempDir(), t.TempDir()
			src := proseSource(t, dir, "letters.org", string(fixture))
			args := append([]string{"letter", src, "--dir", out}, tc.args...)
			status, _, stderr := runApp(t, args...)
			if status == 0 {
				t.Fatal("invalid invocation succeeded")
			}
			for _, want := range tc.want {
				if !strings.Contains(strings.ToLower(stderr), want) {
					t.Errorf("diagnostic missing %q: %s", want, stderr)
				}
			}
			entries, err := os.ReadDir(out)
			if err != nil || len(entries) != 0 {
				t.Fatalf("invalid invocation published files: %v, %v", entries, err)
			}
		})
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
			literal := proseInlineLayers(t, pkg, format, "https://example.invalid/submissions/letter")
			if format == "docx" {
				officeAssertEffective(t, literal, "rFonts", "ascii", "Libertinus Mono")
			} else {
				name, _ := officeEffectiveAttr(literal, "text-properties", "font-name")
				family := ""
				for _, file := range []string{"styles.xml", "content.xml"} {
					for _, face := range pkg[file].all("font-face") {
						if face.attr("name") == name {
							family = strings.Trim(face.attr("font-family"), "'\"")
						}
					}
				}
				if family != "Libertinus Mono" {
					t.Errorf("letter literal family=%q; want existing fixed Libertinus Mono contract", family)
				}
			}
			for _, emphasis := range []struct{ marker, word, attribute, value string }{
				{"BOLD-SAMPLE", "b", "font-weight", "bold"},
				{"ITALIC-SAMPLE", "i", "font-style", "italic"},
				{"UNDERLINE-SAMPLE", "u", "text-underline-style", "solid"},
			} {
				layers := proseInlineLayers(t, pkg, format, emphasis.marker)
				if format == "odt" {
					officeAssertEffective(t, layers, "text-properties", emphasis.attribute, emphasis.value)
				} else {
					active := false
					for _, layer := range layers {
						for _, flag := range layer.all(emphasis.word) {
							value := flag.attr("val")
							active = value != "false" && value != "0" && value != "none"
						}
					}
					if !active {
						t.Errorf("%s lost native emphasis %s", emphasis.marker, emphasis.word)
					}
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
				inlineCode := proseInlineLayers(t, pkg, format, "INLINE-CODE-SAMPLE")
				if format == "docx" {
					officeAssertEffective(t, inlineCode, "rFonts", "ascii", "Iosevka Publish")
					officeAssertEffective(t, inlineCode, "sz", "val", "18")
				} else {
					officeAssertEffective(t, inlineCode, "text-properties", "font-size", "9pt")
				}
				last := -1
				for _, marker := range []string{"HEADING-SAMPLE", "OPENING-SAMPLE", "BODY-SAMPLE", "BOLD-SAMPLE", "ITALIC-SAMPLE", "STRIKE-SAMPLE", "INLINE-CODE-SAMPLE", "LINK-SAMPLE", "QUOTE-SAMPLE", "QUOTE-SECOND-PARAGRAPH", "CODE-SAMPLE:", "STANDALONE-CODE-SAMPLE", "LIST-FIRST", "LIST-NESTED", "LIST-LAST", "ORDERED-FIRST", "ORDERED-LAST", "TABLE-HEADER-LEFT", "TABLE-CELL-LEFT", "NOTE-ANCHOR", "FINAL-SENTINEL"} {
					// Match marker boundaries: the fenced-code sentence is also a suffix
					// of the inline-code specimen and must not count as that block.
					if marker == "CODE-SAMPLE:" {
						marker = "CODE-SAMPLE: 9pt Iosevka Publish."
					}
					pattern := `(` + regexp.QuoteMeta(marker) + `)`
					if strings.HasPrefix(marker, "CODE-SAMPLE:") {
						pattern = `(?:^|[^A-Z-])` + pattern
					}
					matches := regexp.MustCompile(pattern).FindAllStringSubmatchIndex(text, -1)
					at := -1
					if len(matches) > 0 {
						at = matches[0][2]
					}
					if at < last || at < 0 || len(matches) != 1 {
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
					heading := officeParagraph(t, body, "HEADING-SAMPLE")
					officeAssertAttr(t, heading, "outlineLvl", "val", "1")
					nested := officeParagraph(t, body, "LIST-NESTED")
					officeAssertAttr(t, nested, "ilvl", "val", "1")
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

func TestRT063_9_OfficeAtomicReplacement(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		t.Run(format, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			dir := t.TempDir()
			src := proseSource(t, dir, "source.md", "## Chapter One\n\nFIRST-VERSION\n")
			target := proseSource(t, dir, "output."+format, "OLD-DESTINATION")
			proseConvert(t, src, target)
			if !strings.Contains(readAppFile(t, src), "FIRST-VERSION") {
				t.Fatal("export changed source")
			}
			proseSource(t, dir, "source.md", "## Chapter One\n\nSECOND-VERSION\n")
			pkg := proseConvert(t, src, target)
			text := officeBody(t, pkg, format).text()
			if !strings.Contains(text, "SECOND-VERSION") || strings.Contains(text, "FIRST-VERSION") {
				t.Error("repeat export did not replace the selected document")
			}
			before, err := os.ReadFile(target)
			if err != nil {
				t.Fatal(err)
			}
			proseSource(t, dir, "script.yaml", "folio:\n  manuscript:\n    font:\n      size: -5pt\n")
			status, _, stderr := runApp(t, "manuscript", src, target)
			if status == 0 || !strings.Contains(stderr, "font.size") {
				t.Errorf("invalid config was not diagnosed: status=%d stderr=%s", status, stderr)
			}
			after, err := os.ReadFile(target)
			if err != nil || !bytes.Equal(before, after) {
				t.Fatalf("failed export changed destination: %v", err)
			}
			entries, err := os.ReadDir(dir)
			if err != nil || len(entries) != 3 {
				t.Fatalf("staging files leaked beside output: %v, %v", entries, err)
			}
		})
	}
}

func TestRT063_9_OfficeInvalidTextPreservesDestination(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		for _, text := range []struct{ name, input string }{{"utf8", "bad \xff text"}, {"xml-control", "bad \x01 text"}} {
			t.Run(format+"/"+text.name, func(t *testing.T) {
				t.Setenv("HOME", t.TempDir())
				dir := t.TempDir()
				src := proseSource(t, dir, "source.md", text.input)
				target := proseSource(t, dir, "output."+format, "KEEP-DESTINATION")
				status, _, stderr := runApp(t, "manuscript", src, target)
				if status == 0 || stderr == "" || len(stderr) > 64<<10 {
					t.Errorf("invalid source must produce a bounded failure: status=%d bytes=%d", status, len(stderr))
				}
				if readAppFile(t, target) != "KEEP-DESTINATION" {
					t.Fatal("invalid source changed destination")
				}
			})
		}
	}
}

func TestRT063_3_DOCXFootnoteLinksRemainEditable(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	proseSource(t, dir, "script.yaml", "folio:\n  manuscript:\n    title-page: {enabled: false}\n    toc: {enabled: false}\n")
	source := proseSource(t, dir, "source.md", "Body[^note].\n\n[^note]: Note with [NOTE-LINK](https://example.invalid/note).\n")
	pkg := proseConvert(t, source, filepath.Join(dir, "output.docx"))
	note := officeParagraph(t, pkg["word/footnotes.xml"], "NOTE-LINK")
	links := note.all("hyperlink")
	if len(links) != 1 {
		t.Fatalf("note has %d native links", len(links))
	}
	rels := pkg["word/_rels/footnotes.xml.rels"]
	if rels == nil {
		t.Fatal("footnote hyperlink relationship part missing")
	}
	found := false
	for _, rel := range rels.all("Relationship") {
		if rel.attr("Id") == links[0].attr("id") && rel.attr("Target") == "https://example.invalid/note" && rel.attr("TargetMode") == "External" {
			found = true
		}
	}
	if !found {
		t.Error("footnote link does not resolve from its owning part")
	}
}
