// ABOUTME: Exercises office-only recipient path preflight and partial-batch publication.
// ABOUTME: Injects a filesystem change through the public output stream after the first successful file.
package app

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRT063_9_OfficeLetterRejectsUnsafePrefix(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		for _, prefix := range []string{"../escape", "nested/name", `nested\name`, "/absolute"} {
			t.Run(format+"/"+prefix, func(t *testing.T) {
				t.Setenv("HOME", t.TempDir())
				dir := t.TempDir()
				src := proseSource(t, dir, "letters.org", "* Letters :letter:\n** Subject :subject:\nBody.\n*** Recipient :to:\n**** Theatre :org:\n")
				output := t.TempDir()
				status, stdout, stderr := runApp(t, "letter", src, "--format", format, "--dir", output, "--prefix", prefix)
				if status == 0 || !strings.Contains(strings.ToLower(stderr), "prefix") {
					t.Errorf("unsafe prefix must be diagnosed: status=%d stderr=%s", status, stderr)
				}
				entries, err := os.ReadDir(output)
				if err != nil || len(entries) != 0 || strings.Contains(stdout, "Generated:") {
					t.Errorf("unsafe prefix published files: entries=%v stdout=%q error=%v", entries, stdout, err)
				}
			})
		}
	}
}

type obstructSecondLetter struct {
	buffer        bytes.Buffer
	first, second string
	obstructed    bool
}

func (w *obstructSecondLetter) Write(raw []byte) (int, error) {
	n, err := w.buffer.Write(raw)
	if err != nil {
		return n, err
	}
	if !w.obstructed && strings.Contains(w.buffer.String(), "Generated: "+w.first) {
		if _, err := os.Stat(w.first); err != nil {
			return n, fmt.Errorf("first success announced before publication: %w", err)
		}
		if err := os.Mkdir(w.second, 0700); err != nil {
			return n, fmt.Errorf("injecting second-target obstruction: %w", err)
		}
		w.obstructed = true
	}
	return n, nil
}

func TestRT063_9_OfficeLetterRetainsEarlierSuccess(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		t.Run(format, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			dir := t.TempDir()
			raw, err := os.ReadFile("../../testdata/prose-office/letters.org")
			if err != nil {
				t.Fatal(err)
			}
			src := proseSource(t, dir, "letters.org", string(raw))
			output := t.TempDir()
			writer := &obstructSecondLetter{
				first:  filepath.Join(output, "letter-first-theatre."+format),
				second: filepath.Join(output, "letter-second-theatre."+format),
			}
			var stderr bytes.Buffer
			status := Run([]string{"letter", src, "--format", format, "--dir", output}, strings.NewReader(""), writer, &stderr)
			if !writer.obstructed {
				t.Fatalf("first recipient never reached publication: status=%d stderr=%s", status, stderr.String())
			}
			if status == 0 || !strings.Contains(stderr.String(), writer.second) {
				t.Errorf("second failure must name affected path: status=%d stderr=%s", status, stderr.String())
			}
			pkg := prosePackage(t, writer.first, format)
			if !strings.Contains(officeBody(t, pkg, format).text(), "Dear Alex") {
				t.Error("earlier successful recipient was altered or removed")
			}
			if strings.Count(writer.buffer.String(), "Generated:") != 1 || strings.Contains(writer.buffer.String(), "Generated: "+writer.second) {
				t.Errorf("batch report does not distinguish published and failed files: %s", writer.buffer.String())
			}
			entries, err := os.ReadDir(output)
			if err != nil || len(entries) != 2 {
				t.Errorf("batch failure leaked staging files: %v %v", entries, err)
			}
		})
	}
}

func TestRT063_9_OfficePublicSourceLimit(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		t.Run(format, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			dir := t.TempDir()
			// Whitespace avoids an unrelated unsupported-content rejection.
			// Together these sources exceed 32 MiB by exactly one byte.
			first := proseSource(t, dir, "01-first.md", strings.Repeat(" ", 16<<20))
			second := proseSource(t, dir, "02-second.md", strings.Repeat(" ", (16<<20)+1))
			target := proseSource(t, dir, "output."+format, "PRIOR-DESTINATION")
			status, _, stderr := runApp(t, "manuscript", first, second, target)
			if status == 0 || !strings.Contains(stderr, "32") || !strings.Contains(strings.ToLower(stderr), "input") {
				t.Errorf("aggregate input limit diagnostic missing: status=%d stderr=%s", status, stderr)
			}
			got, err := os.ReadFile(target)
			if err != nil || string(got) != "PRIOR-DESTINATION" {
				t.Errorf("oversized input replaced target: %v", err)
			}
		})
	}
}

func TestRT063_9_OfficeBarcodeSidecarPreflight(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		for _, kind := range []string{"directory", "symlink", "hardlink-source"} {
			t.Run(format+"/"+kind, func(t *testing.T) {
				t.Setenv("HOME", t.TempDir())
				dir := t.TempDir()
				proseSource(t, dir, "script.yaml", "folio:\n  manuscript:\n    copyright:\n      enabled: true\n      isbn: '978-0-000000-00-2'\n      isbn-barcode: file\n")
				source := proseSource(t, dir, "source.md", "BODY-SENTINEL\n")
				target := proseSource(t, dir, "output."+format, "PRIOR-DOCUMENT")
				sidecar := filepath.Join(dir, "output.barcode.svg")
				var err error
				switch kind {
				case "directory":
					err = os.Mkdir(sidecar, 0700)
				case "symlink":
					err = os.Symlink(source, sidecar)
				case "hardlink-source":
					err = os.Link(source, sidecar)
				}
				if err != nil {
					t.Fatal(err)
				}
				status, stdout, stderr := runApp(t, "manuscript", source, target)
				if status == 0 || !strings.Contains(stderr, sidecar) || strings.Contains(stdout, "Generated:") {
					t.Errorf("sidecar preflight must name unsafe target before publication: %d %q %q", status, stdout, stderr)
				}
				for path, want := range map[string]string{source: "BODY-SENTINEL\n", target: "PRIOR-DOCUMENT"} {
					raw, err := os.ReadFile(path)
					if err != nil || string(raw) != want {
						t.Errorf("preflight altered %s: %v", path, err)
					}
				}
			})
		}
	}
}
