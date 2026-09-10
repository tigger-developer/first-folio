// ABOUTME: Verifies destination schema declarations through the conversion command.
// ABOUTME: Protects round trips and keeps schema metadata out of other output formats.
package app

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestConvertDestinationSchema(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	sources := map[string]string{
		"org":      "#+SCHEMA: https://example.invalid/old.org\n#+TITLE: Sample\n* ACT ONE\n** Scene One\n**** ALEX\nHello.\n",
		"md":       "# Sample\n\n## ACT ONE\n\n### Scene One\n\n**ALEX:**\nHello.\n",
		"fountain": "Title: Sample\n\n.SCENE ONE\n\nALEX\nHello.\n",
	}
	for ext, text := range sources {
		source := filepath.Join(dir, "source."+ext)
		writeAppFile(t, source, text)
		for _, style := range []string{"british", "us", "screenplay"} {
			for _, target := range []string{"org", "md", "markdown"} {
				t.Run(ext+"/"+style+"/"+target, func(t *testing.T) {
					status, output, stderr := runApp(t, "convert", source, "--to", target, "--style", style)
					if status != 0 || stderr != "" {
						t.Fatalf("status %d stderr %q", status, stderr)
					}
					assertScriptSchema(t, output, target)
					if !strings.Contains(output, "Hello.") || strings.Contains(output, "example.invalid") {
						t.Fatalf("body lost or source schema copied:\n%s", output)
					}
					path := filepath.Join(t.TempDir(), "converted."+target)
					status, _, stderr = runApp(t, "convert", source, path, "--style", style)
					if status != 0 || stderr != "" || readAppFile(t, path) != output {
						t.Fatalf("file output differs from stdout: status %d stderr %q", status, stderr)
					}
				})
			}
		}
	}
}

func TestConvertSchemaRoundTrip(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	source := filepath.Join(dir, "source.org")
	writeAppFile(t, source, "#+TITLE: Sample\n* ACT ONE\n**** ALEX\nHello.\n")
	for _, ext := range []string{"md", "org", "md", "md"} {
		target := filepath.Join(t.TempDir(), "roundtrip."+ext)
		status, _, stderr := runApp(t, "convert", source, target)
		if status != 0 || stderr != "" {
			t.Fatalf("status %d stderr %q", status, stderr)
		}
		output := readAppFile(t, target)
		assertScriptSchema(t, output, ext)
		if strings.Count(output, "Hello.") != 1 || strings.Count(output, "schema:")+strings.Count(output, "#+SCHEMA:") != 1 {
			t.Fatalf("round trip duplicated content or schema:\n%s", output)
		}
		source = target
	}
}

func TestConvertSchemaDoesNotEnterOtherFormats(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	const body = "# Sample\n\n## ACT ONE\n\n**ALEX:**\nHello.\n"
	plain := filepath.Join(dir, "plain.md")
	declared := filepath.Join(dir, "declared.md")
	writeAppFile(t, plain, body)
	writeAppFile(t, declared, "---\nschema: https://example.invalid/schema.md\n---\n\n"+body)
	for _, ext := range []string{"fountain", "typ"} {
		t.Run(ext, func(t *testing.T) {
			outputs := make([]string, 0, 2)
			for _, source := range []string{plain, declared} {
				target := filepath.Join(t.TempDir(), "result."+ext)
				status, _, stderr := runApp(t, "convert", source, target)
				if status != 0 || stderr != "" {
					t.Fatalf("status %d stderr %q", status, stderr)
				}
				outputs = append(outputs, readAppFile(t, target))
			}
			if outputs[0] != outputs[1] || strings.Contains(outputs[1], "schema:") {
				t.Fatalf("schema changed %s output", ext)
			}
		})
	}
}

func TestConvertSchemaWithEmptySource(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	source := filepath.Join(t.TempDir(), "empty.org")
	writeAppFile(t, source, "")
	for _, target := range []string{"org", "md"} {
		status, output, stderr := runApp(t, "convert", source, "--to", target)
		if status != 0 || stderr != "" {
			t.Fatalf("status %d stderr %q", status, stderr)
		}
		assertScriptSchema(t, output, target)
	}
}

func TestConvertRejectsMalformedSchemaFrontmatter(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	for _, text := range []string{"---\nschema: missing closing delimiter\n", "---\nschema: [unclosed\n---\n", "---\nschema: [one, two]\n---\n"} {
		source := filepath.Join(t.TempDir(), "malformed.md")
		writeAppFile(t, source, text)
		status, output, stderr := runApp(t, "convert", source, "--to", "org")
		if status == 0 || output != "" || !strings.Contains(stderr, "frontmatter") {
			t.Fatalf("malformed frontmatter accepted: status %d output %q stderr %q", status, output, stderr)
		}
	}
}

func assertScriptSchema(t *testing.T, output, format string) {
	t.Helper()
	url := "https://github.com/tigger-developer/first-folio/blob/master/schema/script."
	prefix := "#+SCHEMA: " + url + "org\n"
	if format != "org" {
		prefix = "---\nschema: " + url + "md\n---\n"
	}
	if !strings.HasPrefix(output, prefix) {
		t.Fatalf("missing destination schema prefix %q:\n%s", prefix, output)
	}
}
