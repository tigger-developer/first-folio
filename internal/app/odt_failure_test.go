// ABOUTME: Exercises ODT subprocess, resource and package failure boundaries.
// ABOUTME: Uses local synthetic writers only; no remote services or metered calls.
package app

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func init() {
	if os.Getenv("FOLIO_TEST_ODT_OVERSIZE") == "1" && filepath.Base(os.Args[0]) == "pandoc" {
		chunk := bytes.Repeat([]byte("x"), 64<<10)
		for i := 0; i < 2049; i++ {
			if _, err := os.Stdout.Write(chunk); err != nil {
				os.Exit(7)
			}
		}
		os.Exit(0)
	}

	if os.Getenv("FOLIO_TEST_ODT_DIAGNOSTIC") == "1" && filepath.Base(os.Args[0]) == "pandoc" {
		fmt.Fprint(os.Stderr, strings.Repeat("diagnostic ", 20000))
		os.Exit(9)
	}
}
func odtFakePandoc(t *testing.T) string {
	t.Helper()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(executable)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err = os.WriteFile(filepath.Join(dir, "pandoc"), raw, 0700); err != nil {
		t.Fatal(err)
	}
	return dir
}
func TestRT058_3ODTPandocFailures(t *testing.T) {
	binary := officeBuild(t)
	dir := t.TempDir()
	src := filepath.Join(dir, "play.org")
	target := filepath.Join(dir, "play.odt")
	writeAppFile(t, src, officeSources[".org"])
	fake := odtFakePandoc(t)
	for _, mode := range []string{"exit", "malformed", "diagnostic", "oversize"} {
		writeAppFile(t, target, "PRESERVE TARGET")
		extra := map[string]string{"PATH": fake, "FOLIO_TEST_PANDOC_FAILURE": mode}
		if mode == "diagnostic" {
			delete(extra, "FOLIO_TEST_PANDOC_FAILURE")
			extra["FOLIO_TEST_ODT_DIAGNOSTIC"] = "1"
		}
		if mode == "oversize" {
			delete(extra, "FOLIO_TEST_PANDOC_FAILURE")
			extra["FOLIO_TEST_ODT_OVERSIZE"] = "1"
		}
		env := officeEnv(t, t.TempDir(), extra)
		raw, stderr, err := officeRun(t, binary, dir, env, "convert", src, target)
		if err == nil || len(raw) != 0 || stderr == "" || len(stderr) > 66000 {
			t.Fatalf("failure %s: %v, stdout %d, stderr %d", mode, err, len(raw), len(stderr))
		}
		if readAppFile(t, target) != "PRESERVE TARGET" {
			t.Fatal("target changed")
		}
		for _, e := range env {
			if temp, ok := strings.CutPrefix(e, "TMPDIR="); ok {
				entries, err := os.ReadDir(temp)
				if err != nil {
					t.Fatal(err)
				}
				if len(entries) != 0 {
					t.Fatal("private temporary reference leaked")
				}
			}
		}
	}
	t.Setenv("PATH", fake)
	t.Setenv("FOLIO_TEST_PANDOC_FAILURE", "timeout")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	start := time.Now()
	raw, _, err := odtPandoc(ctx, nil, "--from=json", "--to=odt")
	if err == nil || !strings.Contains(err.Error(), "deadline") || len(raw) != 0 || time.Since(start) > 3*time.Second {
		t.Fatalf("deadline: %v", err)
	}
}

func TestRT058_3ODTBrokenOutputAndInputLimits(t *testing.T) {
	// Build before isolating application configuration: Go derives its default
	// module cache from HOME and creates read-only directories there.
	binary := officeBuild(t)
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	src := filepath.Join(dir, "play.org")
	writeAppFile(t, src, officeSources[".org"])
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	if err = reader.Close(); err != nil {
		t.Fatal(err)
	}
	defer writer.Close()
	var diagnostic bytes.Buffer
	if status := Run([]string{"convert", src, "--to", "odt"}, strings.NewReader(""), writer, &diagnostic); status == 0 || !strings.Contains(diagnostic.String(), "broken pipe") {
		t.Fatalf("broken output: %s", diagnostic.String())
	}
	env := officeEnv(t, t.TempDir(), nil)
	target := filepath.Join(dir, "play.odt")
	for _, value := range []string{strings.Repeat("x", (32<<20)+1), "* ACT ONE\n**** CÁIT\ninvalid\x01XML", "* ACT ONE\n**** CÁIT\n" + strings.Repeat("**x** ", 100001)} {
		writeAppFile(t, src, value)
		writeAppFile(t, target, "PRESERVE")
		raw, stderr, err := officeRun(t, binary, dir, env, "convert", src, target)
		if err == nil || len(raw) != 0 || !strings.Contains(stderr, "ODT") || readAppFile(t, target) != "PRESERVE" {
			t.Fatalf("input limit: %v %.200s", err, stderr)
		}
	}
	for _, value := range []string{"", "**** CÁIT\nHello.\n", "#+TITLE: Table after title\n**** CÁIT\nHello.\n"} {
		writeAppFile(t, src, value)
		raw, stderr, err := officeRun(t, binary, dir, env, "convert", src, "--to", "odt")
		if err != nil {
			t.Fatalf("empty/headingless: %v %s", err, stderr)
		}
		officePackage(t, raw)
	}
}

func TestRT058_3ODTRejectsInvalidPackages(t *testing.T) {
	for _, mode := range []string{"missing", "duplicate", "oversize", "entries", "xml", "directive", "mimetype", "path"} {
		t.Run(mode, func(t *testing.T) {
			var b bytes.Buffer
			z := zip.NewWriter(&b)
			write := func(name, text string) {
				w, err := z.Create(name)
				if err != nil {
					t.Fatal(err)
				}
				if _, err = io.WriteString(w, text); err != nil {
					t.Fatal(err)
				}
			}
			write("mimetype", odtMIME)
			switch mode {
			case "duplicate":
				write("mimetype", odtMIME)
			case "oversize":
				_, err := z.CreateRaw(&zip.FileHeader{Name: "large.xml", UncompressedSize64: (128 << 20) + 1})
				if err != nil {
					t.Fatal(err)
				}
			case "entries":
				for i := 0; i < 256; i++ {
					write(fmt.Sprintf("file%d", i), "")
				}
			case "xml":
				write("content.xml", "<a>")
			case "directive":
				write("content.xml", "<!DOCTYPE a><a/>")
			case "mimetype":
				b.Reset()
				z = zip.NewWriter(&b)
				write("mimetype", "incorrect")
			case "path":
				write("../escape", "")
			}
			if err := z.Close(); err != nil {
				t.Fatal(err)
			}
			if err := odtValidate(b.Bytes()); err == nil {
				t.Fatalf("accepted %s", mode)
			}
		})
	}
}

func TestRT058_3ODTTerminalGuard(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("uses macOS PTY command")
	}
	binary := officeBuild(t)
	dir := t.TempDir()
	src := filepath.Join(dir, "play.org")
	writeAppFile(t, src, officeSources[".org"])
	for _, force := range []bool{false, true} {
		args := []string{"-q", "/dev/null", binary, "convert", src, "--to", "odt"}
		if force {
			args = append(args, "--force")
		}
		command := exec.Command("/usr/bin/script", args...)
		command.Env = officeEnv(t, t.TempDir(), nil)
		raw, err := command.CombinedOutput()
		if err != nil {
			exit, ok := err.(*exec.ExitError)
			if force || !ok || exit.ExitCode() != 1 {
				t.Fatalf("PTY: %v %.200s", err, raw)
			}
		}
		if force {
			if !bytes.Contains(raw, []byte("PK")) {
				t.Fatal("forced binary absent")
			}
		} else if !bytes.Contains(raw, []byte("Refusing to write to terminal")) {
			t.Fatalf("terminal guard absent: %.200s", raw)
		}
	}
}
