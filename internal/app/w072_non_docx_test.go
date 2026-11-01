// ABOUTME: Protects PDF and ODT from W072's DOCX-only comment transformation.
// ABOUTME: Confirms all public commands retain footnotes outside the DOCX boundary.
package app

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRT072_2_ODTAndPDFRemainOutsideCommentConversion(t *testing.T) {
	for _, tool := range []string{"typst", "pdftotext"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("%s unavailable: %v", tool, err)
		}
	}
	binary := officeBuild(t)
	for _, command := range []string{"convert", "manuscript", "letter"} {
		for _, format := range []string{"odt", "pdf"} {
			t.Run(command+"/"+format, func(t *testing.T) {
				dir := t.TempDir()
				src := filepath.Join(dir, "source.org")
				writeAppFile(t, src, w072NonDOCXSource(command))
				target := filepath.Join(dir, "target."+format)
				args := []string{command, src, target, "--preserve-footnotes"}
				if command == "letter" {
					target = filepath.Join(dir, "boundary-example-theatre."+format)
					args = []string{"letter", src, "--format", format, "--dir", dir, "--prefix", "boundary", "--preserve-footnotes"}
				}
				_, diagnostic, err := officeRun(t, binary, dir, officeEnv(t, t.TempDir(), nil), args...)
				if err != nil {
					t.Fatalf("%s/%s conversion: %v: %s", command, format, err, diagnostic)
				}
				if format == "odt" {
					raw, err := os.ReadFile(target)
					if err != nil {
						t.Fatal(err)
					}
					pkg := officePackage(t, raw)
					if got := officeBody(t, pkg, format).all("note"); len(got) != 1 || !strings.Contains(got[0].text(), "Aoife Example") {
						t.Fatalf("ODT footnote not retained: %#v", got)
					}
					return
				}
				raw, err := exec.Command("pdftotext", target, "-").CombinedOutput()
				if err != nil {
					t.Fatalf("pdftotext: %v: %s", err, raw)
				}
				if !strings.Contains(string(raw), "Keep as a non-DOCX footnote") {
					t.Fatalf("PDF footnote changed or missing: %s", raw)
				}
			})
		}
	}
}

func w072NonDOCXSource(command string) string {
	body := "Line.[fn:note]\n\n[fn:note] [Aoife Example] wrote: Keep as a non-DOCX footnote.\n"
	switch command {
	case "convert":
		return "#+TITLE: Boundary\n* ACT ONE\n** SCENE ONE\n**** CÁIT\n" + body
	case "manuscript":
		return "#+TITLE: Boundary\n* Part One\n** Chapter One\n" + body
	default:
		return "#+TITLE: Boundary\n#+AUTHOR: Example Author\n* Letters :letter:\n** Sender / Dublin :sender:\n** Subject :subject:\n" + body + "*** Yours sincerely :closing:\n*** [2026-11-01]\n*** Recipient / Cork :to:\n**** Example Theatre :org:\n**** Friend :moniker:\n"
	}
}
