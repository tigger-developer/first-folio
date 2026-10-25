// ABOUTME: Exercises the make release command boundary without network publication.
// ABOUTME: Checks the test gate, selected tag and configurable tap destination.
package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestReleaseGateAndTap(t *testing.T) {
	makefile, err := filepath.Abs("../../Makefile")
	if err != nil {
		t.Fatal(err)
	}
	makePath, err := exec.LookPath("make")
	if err != nil {
		t.Fatal(err)
	}
	for _, fail := range []bool{false, true} {
		t.Run(map[bool]string{false: "publish", true: "failed-tests"}[fail], func(t *testing.T) {
			dir := t.TempDir()
			log := filepath.Join(dir, "calls")
			// POSIX fixture executables only record argv; no real Git or Go runs.
			for _, name := range []string{"go", "git", "typst", "pandoc", "govulncheck", "soffice", "pdftotext", "pdftohtml", "pdftoppm"} {
				script := "#!/bin/sh\nset -eu\nprintf '%s' \"${0##*/}\" >> \"$RELEASE_TEST_LOG\"\nprintf ' <%s>' \"$@\" >> \"$RELEASE_TEST_LOG\"\nprintf '\\n' >> \"$RELEASE_TEST_LOG\"\nif [ \"${0##*/}\" = go ] && [ \"${1:-}\" = test ] && [ \"$RELEASE_TEST_FAIL\" = yes ]; then exit 1; fi\n"
				if err := os.WriteFile(filepath.Join(dir, name), []byte(script), 0700); err != nil {
					t.Fatal(err)
				}
			}
			expr, err := exec.LookPath("expr")
			if err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(expr, filepath.Join(dir, "expr")); err != nil {
				t.Fatal(err)
			}
			tap := filepath.Join(dir, "tap checkout")
			command := exec.Command(makePath, "-f", makefile, "release", "CURRENT_VERSION=v0.4.15", "HOMEBREW_TAP_DIR="+tap, "SKIP_TESTS=0", "MAKE="+makePath+" -f "+makefile)
			command.Dir = dir
			command.Env = append(os.Environ(), "PATH="+dir, "RELEASE_TEST_LOG="+log, "RELEASE_TEST_FAIL="+map[bool]string{false: "no", true: "yes"}[fail])
			output, runErr := command.CombinedOutput()
			raw, err := os.ReadFile(log)
			if err != nil {
				t.Fatal(err)
			}
			calls := string(raw)
			if fail {
				if runErr == nil || strings.Contains(calls, "git <tag>") || strings.Contains(calls, "--publish-tap") {
					t.Fatalf("failed tests allowed publication: %v\n%s\n%s", runErr, output, calls)
				}
				return
			}
			if runErr != nil {
				t.Fatalf("release: %v\n%s", runErr, output)
			}
			for _, want := range []string{"go <test> <-timeout=20m> <./...>", "git <tag> <-a> <v0.4.16>", "git <push> <origin> <v0.4.16>", "<--publish-tap> <" + tap + "> <0.4.16>"} {
				if !strings.Contains(calls, want) {
					t.Errorf("missing %q in calls:\n%s", want, calls)
				}
			}
		})
	}
}
