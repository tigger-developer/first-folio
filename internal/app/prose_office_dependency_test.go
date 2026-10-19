// ABOUTME: Tests prose dependencies in child processes with a controlled executable search path.
// ABOUTME: Leaves the parent agent and test process environment unchanged.
package app

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func init() {
	mode := os.Getenv("FOLIO_TEST_PROSE_TOOL")
	if mode == "" {
		return
	}
	switch filepath.Base(os.Args[0]) {
	case "typst":
		if err := os.WriteFile(os.Getenv("FOLIO_TEST_PROSE_TYPST_MARKER"), []byte("invoked"), 0600); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		fmt.Fprintln(os.Stderr, "FORBIDDEN-TYPST-INVOCATION")
		os.Exit(73)
	case "pandoc":
		if mode == "reader-only" {
			allowed := false
			for _, arg := range os.Args[1:] {
				if arg == "--to=json" || arg == "--to=markdown" {
					allowed = true
				}
				if strings.Contains(arg, "docx") || strings.Contains(arg, "reference") {
					allowed = false
					break
				}
			}
			if !allowed {
				fmt.Fprintln(os.Stderr, "FORBIDDEN-PANDOC-WRITER", os.Args[1:])
				os.Exit(74)
			}
			command := exec.Command(os.Getenv("FOLIO_TEST_REAL_PANDOC"), os.Args[1:]...)
			command.Stdin, command.Stdout, command.Stderr = os.Stdin, os.Stdout, os.Stderr
			if err := command.Run(); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(75)
			}
			os.Exit(0)
		}

		if mode != "incompatible" {
			return
		}
		if len(os.Args) > 1 && os.Args[1] == "--version" {
			fmt.Println("pandoc 3.9.0.2")
		} else {
			fmt.Print(`{"pandoc-api-version":[999],"meta":{},"blocks":[]}`)
		}
		os.Exit(0)
	}
}

func TestProseDependencyChild(t *testing.T) {
	if os.Getenv("FOLIO_TEST_PROSE_CHILD") != "1" {
		return
	}
	for i, arg := range os.Args {
		if arg == "--" {
			os.Exit(Run(os.Args[i+1:], strings.NewReader(""), os.Stdout, os.Stderr))
		}
	}
	t.Fatal("child arguments absent")
}

func TestRT063_1_LetterDOCXWithoutPandoc(t *testing.T) {
	proseDependencies(t, []string{"docx"}, []string{"letter"}, []string{"missing", "incompatible"})
}
func TestRT063_1_ManuscriptDOCXReaderOnly(t *testing.T) {
	proseDependencies(t, []string{"docx"}, []string{"manuscript"}, []string{"reader-only"})
}
func TestRT063_1_OfficeDependencyIsolation(t *testing.T) {
	proseDependencies(t, []string{"docx", "odt"}, []string{"manuscript", "letter"}, []string{"available", "missing", "incompatible"})
}
func proseDependencies(t *testing.T, formats, documents, modes []string) {
	t.Helper()

	pandoc, err := exec.LookPath("pandoc")
	if err != nil {
		t.Fatalf("real Pandoc required for success case: %v", err)
	}
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, format := range formats {
		for _, document := range documents {
			for _, mode := range modes {
				t.Run(format+"/"+document+"/"+mode, func(t *testing.T) {
					dir, bin := t.TempDir(), t.TempDir()
					if err := os.Symlink(self, filepath.Join(bin, "typst")); err != nil {
						t.Fatal(err)
					}
					if mode != "missing" {
						executable := pandoc
						if mode == "incompatible" || mode == "reader-only" {
							executable = self
						}
						if err := os.Symlink(executable, filepath.Join(bin, "pandoc")); err != nil {
							t.Fatal(err)
						}
					}
					src := proseSource(t, dir, "source.md", "DEPENDENCY-BODY\n")
					target := proseSource(t, dir, "output."+format, "PRIOR")
					args := []string{"manuscript", src, target}
					if document == "letter" {
						src = proseSource(t, dir, "letters.org", "* Letters :letter:\n** Subject :subject:\nDEPENDENCY-BODY\n*** Recipient :to:\n**** Theatre :org:\n")
						target = proseSource(t, dir, "letter-theatre."+format, "PRIOR")
						args = []string{"letter", src, "--format", format, "--dir", dir}
					}
					ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
					defer cancel()
					cmd := exec.CommandContext(ctx, self, append([]string{"-test.run=^TestProseDependencyChild$", "--"}, args...)...)
					marker := filepath.Join(dir, "typst-invoked")
					cmd.Env = officeEnv(t, t.TempDir(), map[string]string{"PATH": bin, "FOLIO_TEST_REAL_PANDOC": pandoc, "FOLIO_TEST_PROSE_CHILD": "1", "FOLIO_TEST_PROSE_TOOL": mode, "FOLIO_TEST_PROSE_TYPST_MARKER": marker})
					stdout, stderr := proseDiagnostic{cancel: cancel}, proseDiagnostic{cancel: cancel}
					cmd.Stdout, cmd.Stderr = &stdout, &stderr
					err := cmd.Run()
					if _, err := os.Stat(marker); !os.IsNotExist(err) {
						t.Errorf("office export invoked Typst (marker stat: %v)", err)
					}
					if strings.Contains(stderr.String(), "FORBIDDEN-TYPST") {
						t.Fatalf("office route invoked Typst: %s", stderr.String())
					}
					if mode == "available" || mode == "reader-only" || document == "letter" && format == "docx" {
						if err != nil {
							t.Fatalf("export with only Pandoc: %v: %s", err, stderr.String())
						}
						pkg := prosePackage(t, target, format)
						if !strings.Contains(officeBody(t, pkg, format).text(), "DEPENDENCY-BODY") {
							t.Error("export lost body")
						}
						return
					}
					if err == nil || !strings.Contains(strings.ToLower(stderr.String()), "pandoc") {
						t.Errorf("dependency failure must identify Pandoc: %v %s", err, stderr.String())
					}
					if mode == "incompatible" && !strings.Contains(strings.ToLower(stderr.String()), "version") {
						t.Errorf("incompatible API version not diagnosed: %s", stderr.String())
					}
					proseAssertFile(t, target, "PRIOR")
				})
			}
		}
	}
}
