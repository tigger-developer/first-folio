//go:build folio_acceptance

// ABOUTME: Exercises W071 and W072 through built public commands and native artefacts.
// ABOUTME: Keeps ticket-scoped conversion matrices outside the persistent regression suite.
package app

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAT071_1_PublicCommandAndFormatMatrix(t *testing.T) {
	for _, tool := range []string{"typst", "pdftotext"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("%s unavailable: %v", tool, err)
		}
	}
	binary := officeBuild(t)
	for _, command := range []string{"convert", "manuscript", "letter"} {
		for _, format := range []string{"pdf", "docx", "odt"} {
			t.Run(command+"/"+format, func(t *testing.T) {
				target := w071AcceptanceOutput(t, binary, command, format)
				text := w07AcceptanceText(t, target, format)
				for _, want := range []string{"Prose – em — dots …", "literal--...", "code---..."} {
					if !strings.Contains(text, want) {
						t.Errorf("%s/%s output missing %q: %s", command, format, want, text)
					}
				}
				if command != "letter" && format == "pdf" {
					first := strings.Index(text, "First note paragraph")
					second := strings.Index(text, "Second note paragraph")
					third := strings.Index(text, "Third note paragraph")
					if first < 0 || first >= second || second >= third || strings.Count(text, "note paragraph") != 3 {
						t.Errorf("%s PDF note paragraphs are not one ordered note: %s", command, text)
					}
				}
				if command != "letter" && format != "pdf" {
					pkg := w07AcceptancePackage(t, target)
					body := officeBody(t, pkg, format)
					var note *officeXML
					if format == "docx" {
						for _, candidate := range pkg["word/footnotes.xml"].all("footnote") {
							if strings.Contains(candidate.text(), "First note") {
								note = candidate
							}
						}
					} else {
						for _, candidate := range body.all("note-body") {
							if strings.Contains(candidate.text(), "First note") {
								note = candidate
							}
						}
					}
					if note == nil || len(note.all("p")) != 3 {
						t.Fatalf("%s/%s did not retain three note paragraphs", command, format)
					}
				}
			})
		}
	}
}

func TestAT072_1_AllDOCXCommandsApplySamePolicy(t *testing.T) {
	binary := officeBuild(t)
	for _, command := range []string{"convert", "manuscript", "letter"} {
		t.Run(command, func(t *testing.T) {
			before := time.Now().UTC().Add(-time.Second)
			defaultTarget := w072AcceptanceOutput(t, binary, command, "docx", false, false)
			after := time.Now().UTC().Add(time.Second)
			defaultPackage := w07AcceptancePackage(t, defaultTarget)
			comments := defaultPackage["word/comments.xml"]
			if comments == nil || len(comments.all("comment")) != 2 {
				t.Fatalf("%s default comments = %#v", command, comments)
			}
			if !strings.Contains(comments.text(), "Second comment paragraph") {
				t.Errorf("%s lost multi-paragraph comment: %s", command, comments.text())
			}
			for _, comment := range comments.all("comment") {
				switch comment.attr("author") {
				case "Aoife Example":
					date, err := time.Parse(time.RFC3339, comment.attr("date"))
					if err != nil || date.Before(before) || date.After(after) {
						t.Errorf("%s fallback comment date = %q", command, comment.attr("date"))
					}
				case "Siobhan Example":
					if comment.attr("date") != "2026-10-31T17:59:22Z" {
						t.Errorf("%s explicit comment date = %q", command, comment.attr("date"))
					}
				}
			}
			if notes := defaultPackage["word/footnotes.xml"]; notes == nil || !strings.Contains(notes.text(), "Ordinary note") || strings.Contains(notes.text(), "Timestamped comment") {
				t.Errorf("%s selective footnotes incorrect", command)
			}

			preservedTarget := w072AcceptanceOutput(t, binary, command, "docx", true, false)
			preserved := w07AcceptancePackage(t, preservedTarget)
			if preserved["word/comments.xml"] != nil {
				t.Errorf("%s preserve mode added comments", command)
			}
			if notes := preserved["word/footnotes.xml"]; notes == nil || !strings.Contains(notes.text(), "Timestamped comment") || !strings.Contains(notes.text(), "Ordinary note") {
				t.Errorf("%s preserve mode lost footnotes", command)
			}

			odtTarget := w072AcceptanceOutput(t, binary, command, "odt", false, false)
			odt := w07AcceptancePackage(t, odtTarget)
			if notes := officeBody(t, odt, "odt").all("note"); len(notes) != 3 {
				t.Errorf("%s ODT note count = %d, want 3", command, len(notes))
			}
		})
	}
}

func TestAT072_2_InvalidConversionIsAtomic(t *testing.T) {
	binary := officeBuild(t)
	for _, command := range []string{"convert", "manuscript", "letter"} {
		t.Run(command, func(t *testing.T) {
			dir := t.TempDir()
			source := filepath.Join(dir, "source.org")
			writeAppFile(t, source, w072AcceptanceSource(command, true))
			target := filepath.Join(dir, "target.docx")
			args := []string{command, source, target}
			if command == "letter" {
				target = filepath.Join(dir, "atomic-example-theatre.docx")
				args = []string{"letter", source, "--format", "docx", "--dir", dir, "--prefix", "atomic"}
			}
			const sentinel = "existing target sentinel"
			writeAppFile(t, target, sentinel)
			_, diagnostic, err := officeRun(t, binary, dir, officeEnv(t, t.TempDir(), nil), args...)
			if err == nil || !strings.Contains(diagnostic, "invalid UTC timestamp") {
				t.Fatalf("%s error = %v, diagnostic = %s", command, err, diagnostic)
			}
			raw, readErr := os.ReadFile(target)
			if readErr != nil || string(raw) != sentinel {
				t.Fatalf("%s replaced existing target: %v %q", command, readErr, raw)
			}

			absentDir := t.TempDir()
			absentSource := filepath.Join(absentDir, "source.org")
			writeAppFile(t, absentSource, w072AcceptanceSource(command, true))
			absentTarget := filepath.Join(absentDir, "absent.docx")
			absentArgs := []string{command, absentSource, absentTarget}
			if command == "letter" {
				absentTarget = filepath.Join(absentDir, "absent-example-theatre.docx")
				absentArgs = []string{"letter", absentSource, "--format", "docx", "--dir", absentDir, "--prefix", "absent"}
			}
			_, absentDiagnostic, absentErr := officeRun(t, binary, absentDir, officeEnv(t, t.TempDir(), nil), absentArgs...)
			if absentErr == nil || !strings.Contains(absentDiagnostic, "invalid UTC timestamp") {
				t.Fatalf("%s absent-target error = %v, diagnostic = %s", command, absentErr, absentDiagnostic)
			}
			if _, statErr := os.Stat(absentTarget); !os.IsNotExist(statErr) {
				t.Fatalf("%s created failed target: %v", command, statErr)
			}
		})
	}

	dir := t.TempDir()
	source := filepath.Join(dir, "source.org")
	writeAppFile(t, source, w072AcceptanceSource("convert", true))
	stdout, diagnostic, err := officeRun(t, binary, dir, officeEnv(t, t.TempDir(), nil), "convert", source, "--to", "docx")
	if err == nil || !strings.Contains(diagnostic, "invalid UTC timestamp") {
		t.Fatalf("stdout conversion error = %v, diagnostic = %s", err, diagnostic)
	}
	if len(stdout) != 0 {
		t.Fatalf("failed conversion wrote %d stdout bytes", len(stdout))
	}
}

func w071AcceptanceOutput(t *testing.T, binary, command, format string) string {
	t.Helper()
	dir := t.TempDir()
	source := filepath.Join(dir, "source.org")
	writeAppFile(t, source, w071AcceptanceSource(command))
	target := filepath.Join(dir, "target."+format)
	args := []string{command, source, target}
	if command == "letter" {
		target = filepath.Join(dir, "matrix-example-theatre."+format)
		args = []string{"letter", source, "--format", format, "--dir", dir, "--prefix", "matrix"}
	}
	_, diagnostic, err := officeRun(t, binary, dir, officeEnv(t, t.TempDir(), nil), args...)
	if err != nil {
		t.Fatalf("%s/%s: %v: %s", command, format, err, diagnostic)
	}
	return target
}

func w072AcceptanceOutput(t *testing.T, binary, command, format string, preserve, badDate bool) string {
	t.Helper()
	dir := t.TempDir()
	source := filepath.Join(dir, "source.org")
	writeAppFile(t, source, w072AcceptanceSource(command, badDate))
	target := filepath.Join(dir, "target."+format)
	args := []string{command, source, target}
	if command == "letter" {
		target = filepath.Join(dir, "comments-example-theatre."+format)
		args = []string{"letter", source, "--format", format, "--dir", dir, "--prefix", "comments"}
	}
	if preserve {
		args = append(args, "--preserve-footnotes")
	}
	_, diagnostic, err := officeRun(t, binary, dir, officeEnv(t, t.TempDir(), nil), args...)
	if err != nil {
		t.Fatalf("%s/%s preserve=%v: %v: %s", command, format, preserve, err, diagnostic)
	}
	return target
}

func w07AcceptancePackage(t *testing.T, path string) map[string]*officeXML {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return officePackage(t, raw)
}

func w07AcceptanceText(t *testing.T, target, format string) string {
	t.Helper()
	if format == "pdf" {
		raw, err := exec.Command("pdftotext", "-layout", target, "-").CombinedOutput()
		if err != nil {
			t.Fatalf("pdftotext: %v: %s", err, raw)
		}
		return strings.Join(strings.Fields(string(raw)), " ")
	}
	return strings.Join(strings.Fields(officeBody(t, w07AcceptancePackage(t, target), format).text()), " ")
}

func w071AcceptanceSource(command string) string {
	prose := `Prose -- em --- dots ... soft\-hyphen =literal--...= ~code---...~.`
	note := "[fn:note] First note paragraph.\n\nSecond note paragraph.\n\nThird note paragraph.\n"
	switch command {
	case "convert":
		return "#+TITLE: Matrix\n* ACT ONE\n** SCENE ONE\n**** CÁIT\n" + prose + "[fn:note]\n\n" + note
	case "manuscript":
		return "#+TITLE: Matrix\n* Part One\n** Chapter One\n" + prose + "[fn:note]\n\n" + note
	default:
		return "#+TITLE: Matrix\n#+AUTHOR: Example Author\n* Letters :letter:\n** Sender / Dublin :sender:\n** Subject :subject:\n" + prose + "\n*** Yours sincerely :closing:\n*** [2026-11-01]\n*** Recipient / Cork :to:\n**** Example Theatre :org:\n**** Friend :moniker:\n"
	}
}

func w072AcceptanceSource(command string, badDate bool) string {
	date := "2026-10-31 17:59:22Z"
	if badDate {
		date = "31 October"
	}
	body := "Text one.[fn:a] Text two.[fn:b] Text three.[fn:c]\n\n[fn:a] [Aoife Example] wrote: First comment\n\n[fn:b] [Siobhan Example] wrote (" + date + "): Timestamped comment\n\nSecond comment paragraph\n\nThird comment paragraph\n\n[fn:c] Ordinary note\n"
	switch command {
	case "convert":
		return "#+TITLE: Comments\n* ACT ONE\n** SCENE ONE\n**** CÁIT\n" + body
	case "manuscript":
		return "#+TITLE: Comments\n* Part One\n** Chapter One\n" + body
	default:
		return "#+TITLE: Comments\n#+AUTHOR: Example Author\n* Letters :letter:\n** Sender / Dublin :sender:\n** Subject :subject:\n" + body + "*** Yours sincerely :closing:\n*** [2026-11-01]\n*** Recipient / Cork :to:\n**** Example Theatre :org:\n**** Friend :moniker:\n"
	}
}
