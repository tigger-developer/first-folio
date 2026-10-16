// ABOUTME: Tests aggregate source limits and lossless UTF-8/XML-safe text acceptance.
// ABOUTME: Owns all temporary inputs and verifies that rejected files remain unchanged.
package office

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func sourceFile(t *testing.T, dir, name string, raw []byte) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestRT063_9_SourceAggregateThreshold(t *testing.T) {
	dir := t.TempDir()
	first := []byte("First café\n")
	second := []byte("Second & <literal>\t\r\n")
	paths := []string{sourceFile(t, dir, "first.md", first), sourceFile(t, dir, "second.md", second)}
	limit := int64(len(first) + len(second))
	t.Run("at-limit", func(t *testing.T) {
		got, err := readSources(paths, limit)
		if err != nil || len(got) != 2 || !bytes.Equal(got[0], first) || !bytes.Equal(got[1], second) {
			t.Fatalf("exact aggregate limit must return original per-file bytes: files=%d error=%v", len(got), err)
		}
	})
	t.Run("over-limit", func(t *testing.T) {
		if _, err := readSources(paths, limit-1); err == nil {
			t.Error("aggregate limit exceeded by one byte was accepted")
		}
	})
}

func TestRT063_9_SourceInvalidEncodingAndType(t *testing.T) {
	for _, tc := range []struct {
		name string
		raw  []byte
	}{
		{"utf8", []byte{'a', 0xff, 'b'}},
		{"nul", []byte{'a', 0, 'b'}},
		{"xml-control", []byte{'a', 1, 'b'}},
		{"xml-noncharacter", []byte("a\ufffeb")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := sourceFile(t, t.TempDir(), "source.md", tc.raw)
			if _, err := readSources([]string{path}, 1024); err == nil {
				t.Fatal("invalid source text accepted")
			}
			got, err := os.ReadFile(path)
			if err != nil || !bytes.Equal(got, tc.raw) {
				t.Errorf("rejected input changed: %v", err)
			}
		})
	}
	t.Run("directory", func(t *testing.T) {
		if _, err := readSources([]string{t.TempDir()}, 1024); err == nil {
			t.Error("nonregular input accepted")
		}
	})
	t.Run("missing", func(t *testing.T) {
		if _, err := readSources([]string{filepath.Join(t.TempDir(), "absent.md")}, 1024); err == nil {
			t.Error("missing input accepted")
		}
	})
}
