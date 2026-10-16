// ABOUTME: Checks atomic private publication and cleanup after real filesystem operations fail.
// ABOUTME: Fault wrappers preserve real temporary files while forcing write, close and rename failures.
package office

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func realPublicationFS() publicationFS {
	return publicationFS{
		createTemp: func(dir, pattern string) (stagedFile, error) { return os.CreateTemp(dir, pattern) },
		rename:     os.Rename,
		remove:     os.Remove,
	}
}

type failingStage struct {
	*os.File
	operation string
	failure   error
}

func (f *failingStage) Write(raw []byte) (int, error) {
	if f.operation == "write" {
		n, err := f.File.Write(raw[:len(raw)/2])
		return n, errors.Join(f.failure, err)
	}
	if f.operation == "short-write" {
		return f.File.Write(raw[:len(raw)/2])
	}
	return f.File.Write(raw)
}

func (f *failingStage) Close() error {
	err := f.File.Close()
	if f.operation == "close" {
		return errors.Join(f.failure, err)
	}
	return err
}

func (f *failingStage) Chmod(mode os.FileMode) error {
	if f.operation == "chmod" {
		return f.failure
	}
	return f.File.Chmod(mode)
}

func TestRT063_9_AtomicPublicationAndPermissions(t *testing.T) {
	dir := t.TempDir()
	path := sourceFile(t, dir, "output.docx", []byte("PRIOR"))
	if err := os.Chmod(path, 0644); err != nil {
		t.Fatal(err)
	}
	for _, content := range [][]byte{[]byte("FIRST-PUBLICATION"), []byte("SECOND-PUBLICATION")} {
		if err := publishFile(path, content, realPublicationFS()); err != nil {
			t.Fatal(err)
		}
		got, err := os.ReadFile(path)
		if err != nil || !bytes.Equal(got, content) {
			t.Fatalf("published bytes differ: %q, %v", got, err)
		}
		info, err := os.Stat(path)
		if err != nil || info.Mode().Perm() != 0600 {
			t.Fatalf("published file is not private: %v", err)
		}
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 1 {
		t.Errorf("success leaked staging files: %v %v", entries, err)
	}
}

func TestRT063_9_AtomicPublicationFailures(t *testing.T) {
	for _, operation := range []string{"create", "write", "short-write", "chmod", "close", "rename"} {
		t.Run(operation, func(t *testing.T) {
			dir := t.TempDir()
			path := sourceFile(t, dir, "output.odt", []byte("PRIOR"))
			failure := errors.New("injected filesystem failure")
			fs := realPublicationFS()
			fs.createTemp = func(parent, pattern string) (stagedFile, error) {
				if parent != dir {
					t.Errorf("staging is not beside destination: %s", parent)
				}
				if operation == "create" {
					return nil, failure
				}
				file, err := os.CreateTemp(parent, pattern)
				if err != nil {
					return nil, err
				}
				return &failingStage{File: file, operation: operation, failure: failure}, nil
			}
			if operation == "rename" {
				fs.rename = func(stage, target string) error {
					if filepath.Dir(stage) != dir || target != path {
						t.Error("publication used unexpected paths")
					}
					got, err := os.ReadFile(stage)
					if err != nil || string(got) != "REPLACEMENT" {
						t.Error("rename occurred before complete staging")
					}
					return failure
				}
			}
			err := publishFile(path, []byte("REPLACEMENT"), fs)
			want := failure
			if operation == "short-write" {
				want = io.ErrShortWrite
			}
			if !errors.Is(err, want) {
				t.Errorf("filesystem cause lost: %v", err)
			}
			got, readErr := os.ReadFile(path)
			if readErr != nil || string(got) != "PRIOR" {
				t.Errorf("failure altered prior destination: %q %v", got, readErr)
			}
			entries, readErr := os.ReadDir(dir)
			if readErr != nil || len(entries) != 1 {
				t.Errorf("failure leaked staging files: %v %v", entries, readErr)
			}
		})
	}
}
