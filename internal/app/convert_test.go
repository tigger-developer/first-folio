// ABOUTME: Characterizes text conversion through the public Go application boundary.
// ABOUTME: Covers files, stdout, config metadata, Unicode, warnings, and invalid input.
package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConvertTextFormats(t *testing.T) {
	dir := t.TempDir()
	home := t.TempDir()
	t.Setenv("HOME", home)
	orgPath := filepath.Join(dir, "play.org")
	mdPath := filepath.Join(dir, "play.md")
	writeAppFile(t, orgPath, `#+TITLE: Samhain
#+AUTHOR: Taḋg Paul
* ACT ONE
** Scene One
**** CÁIT quietly
Hello.
`)

	status, stdout, stderr := runApp(t, "convert", orgPath, mdPath)
	if status != 0 {
		t.Fatalf("status %d\nstdout:%s\nstderr:%s", status, stdout, stderr)
	}
	markdown := readAppFile(t, mdPath)
	for _, fragment := range []string{"# Samhain", "## ACT ONE", "**CÁIT:** *(quietly)*", "Hello."} {
		if !strings.Contains(markdown, fragment) {
			t.Errorf("Markdown missing %q:\n%s", fragment, markdown)
		}
	}

	status, stdout, stderr = runApp(t, "convert", mdPath, "--to", "org")
	if status != 0 || stderr != "" {
		t.Fatalf("status %d stderr %q", status, stderr)
	}
	for _, fragment := range []string{"#+TITLE: Samhain", "* ACT ONE", "**** CÁIT quietly"} {
		if !strings.Contains(stdout, fragment) {
			t.Errorf("Org stdout missing %q:\n%s", fragment, stdout)
		}
	}
}

func TestConvertConfigMetadataOverride(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	writeAppFile(t, filepath.Join(dir, "play.org"), "#+TITLE: Source Title\n* ACT ONE\n**** CÁIT\nHello.\n")
	writeAppFile(t, filepath.Join(dir, "script.yaml"), "title: Config Title\nfolio:\n  default-format: md\n")

	status, stdout, stderr := runApp(t, "convert", filepath.Join(dir, "play.org"))
	if status != 0 || stderr != "" {
		t.Fatalf("status %d stderr %q", status, stderr)
	}
	if !strings.Contains(stdout, "# Config Title") || strings.Contains(stdout, "# Source Title") {
		t.Fatalf("config metadata not applied:\n%s", stdout)
	}
}

func TestConvertInvalidInputs(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	binary := filepath.Join(dir, "binary.org")
	writeAppFile(t, binary, "text\x00binary")

	tests := []struct {
		name string
		args []string
		want string
	}{
		{"missing", []string{"convert", filepath.Join(dir, "missing.org"), "out.md"}, "file not found"},
		{"binary", []string{"convert", binary, "out.md"}, "appears to be a binary file"},
		{"unknown target", []string{"convert", binary, "out.xyz"}, "Unrecognised file extension"},
		{"no source", []string{"convert"}, "no source file specified"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, _, stderr := runApp(t, tt.args...)
			if status == 0 || !strings.Contains(stderr, tt.want) {
				t.Fatalf("status %d, stderr %q; want %q", status, stderr, tt.want)
			}
		})
	}
}

func runApp(t *testing.T, args ...string) (int, string, string) {
	t.Helper()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	status := Run(args, strings.NewReader(""), &stdout, &stderr)
	return status, stdout.String(), stderr.String()
}

func writeAppFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readAppFile(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}
