// ABOUTME: Runs a bounded Pandoc ODT writer and validates its package before publication.
// ABOUTME: Keeps external processes and generated references confined to ODT export.
package app

import (
	"fmt"
	"github.com/tigger-developer/first-folio/internal/config"
	"github.com/tigger-developer/first-folio/internal/office"
	"github.com/tigger-developer/first-folio/internal/play"
	"io"
	"os"
	"strings"
)

const odtMIME = "application/vnd.oasis.opendocument.text"

// The shared layout/publication helpers retain their existing DOCX diagnostics.
// Relabel at this boundary without changing DOCX error handling or serialization.
func odtError(err error) error {
	if err == nil {
		return nil
	}
	return odtExportError{err}
}

type odtExportError struct{ cause error }

func (e odtExportError) Error() string { return strings.ReplaceAll(e.cause.Error(), "DOCX", "ODT") }
func (e odtExportError) Unwrap() error { return e.cause }

func readODTSource(path string) ([]byte, error) {
	// Reject FIFOs before opening: an open-for-read could otherwise wait forever.
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("checking ODT source: %w", err)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("ODT source must be a regular file")
	}

	raw, err := readDOCXSource(path)
	return raw, odtError(err)
}

func exportODT(doc play.Document, cfg config.Config, sourceFormat play.Format, opts convertOptions, toStdout bool, stdout, stderr io.Writer) (err error) {
	defer func() { err = odtError(err) }()
	if toStdout && writerIsTerminal(stdout) && !opts.force {
		return fmt.Errorf("ODT is binary output. Refusing to write to terminal\nUse a target file, redirect stdout, or pass --force to override")
	}
	if !toStdout {
		if err = docxTarget(opts.source, opts.target); err != nil {
			return err
		}
	}
	raw, warnings, err := office.ExportScript(doc, cfg, sourceFormat, "odt")
	if err != nil {
		return err
	}
	for _, warning := range warnings {
		if _, err = fmt.Fprintln(stderr, "ODT limitation:", warning); err != nil {
			return err
		}
	}
	if toStdout {
		n, e := stdout.Write(raw)
		if e == nil && n != len(raw) {
			e = io.ErrShortWrite
		}
		return e
	}
	return docxPublish(opts.source, opts.target, raw)
}
