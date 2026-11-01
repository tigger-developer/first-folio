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
			defaultTarget := w072AcceptanceOutput(t, binary, command, "docx", false, false)
			defaultPackage := w07AcceptancePackage(t, defaultTarget)
			comments := defaultPackage["word/comments.xml"]
			if comments == nil || len(comments.all("comment")) != 2 {
				t.Fatalf("%s default comments = %#v", command, comments)
			}
			if !strings.Contains(comments.text(), "Second comment paragraph") {
				t.Errorf("%s lost multi-paragraph comment: %s", command, comments.text())
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
