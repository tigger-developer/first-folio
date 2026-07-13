// ABOUTME: Exercises the built Go executable from outside the source checkout.
// ABOUTME: Verifies embedded assets, public dispatch, and installation-shaped operation.
package app

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuiltBinaryOutsideCheckout(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	binary := filepath.Join(dir, "folio")
	build := exec.Command("go", "build", "-o", binary, "./cmd/folio")
	build.Dir = root
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("building folio: %v\n%s", err, output)
	}
	help := exec.Command(binary, "--help")
	help.Dir = dir
	help.Env = []string{"HOME=" + filepath.Join(dir, "home"), "PATH="}
	output, err := help.CombinedOutput()
	if err != nil || !strings.Contains(string(output), "manuscript") {
		t.Fatalf("installed-shaped help: %v\n%s", err, output)
	}
	source := filepath.Join(dir, "play.org")
	if err := os.WriteFile(source, []byte("#+TITLE: Installed\n* ACT ONE\n**** CÁIT\nHello.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	convert := exec.Command(binary, "convert", source, filepath.Join(dir, "play.md"))
	convert.Dir = dir
	convert.Env = help.Env
	if output, err := convert.CombinedOutput(); err != nil {
		t.Fatalf("installed-shaped convert: %v\n%s", err, output)
	}
	markdown, err := os.ReadFile(filepath.Join(dir, "play.md"))
	if err != nil || !strings.Contains(string(markdown), "# Installed") {
		t.Fatalf("installed output: %v\n%s", err, markdown)
	}
}
