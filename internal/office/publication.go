// ABOUTME: Publishes private office artefacts atomically beside their selected destinations.
// ABOUTME: Preflights source aliases and reports manuscript-first sidecar partial success.
package office

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type stagedFile interface {
	Write([]byte) (int, error)
	Close() error
	Chmod(os.FileMode) error
	Name() string
}
type publicationFS struct {
	createTemp func(string, string) (stagedFile, error)
	rename     func(string, string) error
	remove     func(string) error
}

// CheckTarget permits replacement of regular files only and rejects source aliases.
func CheckTarget(sources []string, target string) error {
	targetAbs, err := filepath.Abs(target)
	if err != nil {
		return err
	}
	info, err := os.Lstat(target)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("checking target %s: %w", target, err)
	}
	if err == nil && !info.Mode().IsRegular() {
		return fmt.Errorf("target %s must be regular, not a symlink or directory", target)
	}
	parent, err := os.Stat(filepath.Dir(target))
	if err != nil {
		return fmt.Errorf("checking target directory %s: %w", target, err)
	}
	if !parent.IsDir() {
		return fmt.Errorf("target parent is not a directory: %s", target)
	}
	for _, source := range sources {
		abs, err := filepath.Abs(source)
		if err != nil {
			return err
		}
		original, err := os.Stat(source)
		if err != nil {
			return err
		}
		if abs == targetAbs || info != nil && os.SameFile(original, info) {
			return fmt.Errorf("target %s aliases source %s", target, source)
		}
	}
	return nil
}

// Publish stages and replaces one output; callers preflight all batch targets first.
func Publish(path string, raw []byte) error {
	if err := CheckTarget(nil, path); err != nil {
		return err
	}
	fs := publicationFS{func(dir, pattern string) (stagedFile, error) { return os.CreateTemp(dir, pattern) }, os.Rename, os.Remove}
	return publishFile(path, raw, fs)
}

func publishFile(path string, raw []byte, fs publicationFS) (result error) {
	f, err := fs.createTemp(filepath.Dir(path), ".folio-office-*.tmp")
	if err != nil {
		return fmt.Errorf("staging %s: %w", path, err)
	}
	closed := false
	defer func() {
		if !closed {
			result = errors.Join(result, f.Close())
		}
		if err := fs.remove(f.Name()); err != nil && !errors.Is(err, os.ErrNotExist) {
			result = errors.Join(result, fmt.Errorf("cleaning stage for %s: %w", path, err))
		}
	}()
	if err = f.Chmod(0600); err != nil {
		return fmt.Errorf("protecting %s: %w", path, err)
	}
	n, err := f.Write(raw)
	if err == nil && n != len(raw) {
		err = io.ErrShortWrite
	}
	if err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	err = f.Close()
	closed = true
	if err != nil {
		return fmt.Errorf("closing %s: %w", path, err)
	}
	if err = fs.rename(f.Name(), path); err != nil {
		return fmt.Errorf("publishing %s: %w", path, err)
	}
	return nil
}

// PublishManuscript retains the documented manuscript-then-sidecar order.
func PublishManuscript(path string, document []byte, sidecarPath string, sidecar []byte) error {
	return publishManuscript(path, document, sidecarPath, sidecar, Publish)
}
func publishManuscript(path string, document []byte, sidecarPath string, sidecar []byte, publish func(string, []byte) error) error {
	if err := publish(path, document); err != nil {
		return fmt.Errorf("publishing manuscript %s: %w", path, err)
	}
	if sidecarPath != "" {
		if err := publish(sidecarPath, sidecar); err != nil {
			return fmt.Errorf("manuscript %s was published; publishing sidecar %s failed: %w", path, sidecarPath, err)
		}
	}
	return nil
}
