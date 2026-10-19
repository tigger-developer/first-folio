// ABOUTME: Publishes bounded, native DOCX packages through the conversion command.
// ABOUTME: Keeps output atomic and independent of the existing PDF and text routes.
package app

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/tigger-developer/first-folio/internal/config"
	"github.com/tigger-developer/first-folio/internal/office"
	"github.com/tigger-developer/first-folio/internal/play"
)

func readDOCXSource(path string) (raw []byte, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening DOCX source %q: %w", path, err)
	}
	defer func() { err = errors.Join(err, f.Close()) }()
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("DOCX source must be a regular file")
	}
	raw, err = io.ReadAll(io.LimitReader(f, office.SourceLimit+1))
	if err != nil {
		return nil, fmt.Errorf("reading DOCX source: %w", err)
	}
	if len(raw) > office.SourceLimit {
		return nil, fmt.Errorf("DOCX source exceeds 32 MiB")
	}
	if err := office.ValidText(string(raw)); err != nil {
		return nil, err
	}
	return raw, nil
}

func exportDOCX(doc play.Document, cfg config.Config, sourceFormat play.Format, opts convertOptions, toStdout bool, stdout, stderr io.Writer) error {
	if toStdout && writerIsTerminal(stdout) && !opts.force {
		return fmt.Errorf("DOCX is binary output. Refusing to write to terminal\nUse a target file, redirect stdout, or pass --force to override")
	}
	if !toStdout {
		if err := docxTarget(opts.source, opts.target); err != nil {
			return err
		}
	}
	raw, warnings, err := office.ExportScript(doc, cfg, sourceFormat, "docx")
	if err != nil {
		return err
	}
	for _, warning := range warnings {
		if _, err := fmt.Fprintln(stderr, "DOCX limitation:", warning); err != nil {
			return err
		}
	}
	if toStdout {
		n, err := stdout.Write(raw)
		if err == nil && n != len(raw) {
			err = io.ErrShortWrite
		}
		return err
	}
	return docxPublish(opts.source, opts.target, raw)
}

func docxTarget(source, target string) error {
	sourceInfo, err := os.Stat(source)
	if err != nil {
		return fmt.Errorf("checking DOCX source: %w", err)
	}
	info, err := os.Lstat(target)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("checking DOCX destination: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("DOCX destination %q is a symlink", target)
	}
	if os.SameFile(sourceInfo, info) {
		return fmt.Errorf("DOCX source and destination identify the same file")
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("DOCX destination must be a regular file")
	}
	return nil
}

func docxPublish(source, target string, raw []byte) (err error) {
	staging, err := os.CreateTemp(filepath.Dir(target), ".folio-docx-*")
	if err != nil {
		return fmt.Errorf("staging DOCX: %w", err)
	}
	name := staging.Name()
	defer func() {
		if removeErr := os.Remove(name); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			err = errors.Join(err, fmt.Errorf("removing DOCX staging file: %w", removeErr))
		}
	}()
	_, writeErr := staging.Write(raw)
	closeErr := staging.Close()
	if err = errors.Join(writeErr, closeErr); err != nil {
		return fmt.Errorf("writing DOCX staging file: %w", err)
	}
	// Recheck after generation: never knowingly replace a newly introduced link.
	if err = docxTarget(source, target); err != nil {
		return err
	}
	if err = os.Rename(name, target); err != nil {
		return fmt.Errorf("publishing DOCX %q: %w", target, err)
	}
	return nil
}
