// ABOUTME: Runs a bounded Pandoc ODT writer and validates its package before publication.
// ABOUTME: Keeps external processes and generated references confined to ODT export.
package app

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/tigger-developer/first-folio/internal/config"
	"github.com/tigger-developer/first-folio/internal/play"
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
	layout, err := newDOCXWriter(doc, cfg, sourceFormat)
	if err != nil {
		return err
	}
	body, err := layout.body()
	if err != nil {
		return err
	}
	p := &odtProjection{layout: layout, variants: map[string]string{}}
	ast, err := p.json(body)
	if err != nil {
		return err
	}
	reference, err := p.reference()
	if err != nil {
		return err
	}
	dir, err := os.MkdirTemp("", "folio-odt-*")
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, os.RemoveAll(dir)) }()
	path := filepath.Join(dir, "reference.odt")
	if err = os.WriteFile(path, reference, 0600); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	raw, diagnostic, err := odtPandoc(ctx, ast, "--from=json", "--to=odt", "--standalone", "--reference-doc="+path)
	if err != nil {
		return err
	}
	if err = odtValidate(raw); err != nil {
		return err
	}
	raw, fontWarnings, err := embedOfficeFonts(raw, "odt")
	if err != nil {
		return err
	}
	if err = odtValidate(raw); err != nil {
		return err
	}
	layout.warnings = append(layout.warnings, fontWarnings...)
	for _, warning := range layout.warnings {
		if _, err = fmt.Fprintln(stderr, "ODT limitation:", warning); err != nil {
			return err
		}
	}
	if diagnostic != "" {
		if _, err = fmt.Fprintln(stderr, "Pandoc:", diagnostic); err != nil {
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

// Do not embed bytes.Buffer: its promoted ReadFrom would bypass Write limits
// when os/exec copies a subprocess pipe through io.Copy.
type odtBuffer struct {
	buffer   bytes.Buffer
	limit    int
	truncate bool
}

func (b *odtBuffer) Write(raw []byte) (int, error) {
	n := len(raw)
	left := b.limit - b.buffer.Len()
	if n > left {
		if !b.truncate {
			return 0, fmt.Errorf("ODT generated data exceeds %d bytes", b.limit)
		}
		raw = raw[:left]
	}
	_, err := b.buffer.Write(raw)
	return n, err
}
func (b *odtBuffer) Bytes() []byte  { return b.buffer.Bytes() }
func (b *odtBuffer) String() string { return b.buffer.String() }

func odtPandoc(ctx context.Context, input []byte, args ...string) ([]byte, string, error) {
	cmd := exec.CommandContext(ctx, "pandoc", args...)
	cmd.Stdin = bytes.NewReader(input)
	output := odtBuffer{limit: docxXMLLimit}
	diagnostic := odtBuffer{limit: 64 << 10, truncate: true}
	cmd.Stdout = &output
	cmd.Stderr = &diagnostic
	cmd.WaitDelay = time.Second
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return nil, "", fmt.Errorf("ODT pandoc deadline exceeded: %w", ctx.Err())
		}
		return nil, "", fmt.Errorf("ODT pandoc failed (requires a Pandoc 3.x JSON writer): %w: %s", err, diagnostic.String())
	}
	return output.Bytes(), diagnostic.String(), nil
}

func odtValidate(raw []byte) error {
	archive, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		return fmt.Errorf("ODT pandoc produced an invalid ZIP: %w", err)
	}
	if len(archive.File) > 256 {
		return fmt.Errorf("ODT package exceeds 256 entries")
	}
	total := uint64(0)
	seen := map[string]bool{}
	for _, f := range archive.File {
		if seen[f.Name] {
			return fmt.Errorf("ODT duplicate package member %q", f.Name)
		}
		seen[f.Name] = true
		if f.Mode()&os.ModeSymlink != 0 || strings.Contains(f.Name, "..") || strings.HasPrefix(f.Name, "/") {
			return fmt.Errorf("ODT unsafe package member %q", f.Name)
		}
		if f.UncompressedSize64 > 128<<20 || total > 256<<20-f.UncompressedSize64 {
			return fmt.Errorf("ODT expanded package exceeds limit")
		}
		total += f.UncompressedSize64
		member, err := f.Open()
		if err != nil {
			return err
		}
		if strings.HasSuffix(f.Name, ".xml") {
			err = odtValidateXML(member)
		} else if f.Name == "mimetype" {
			data, e := io.ReadAll(io.LimitReader(member, 128))
			err = e
			if string(data) != odtMIME && err == nil {
				err = fmt.Errorf("incorrect ODT mimetype")
			}
		} else {
			_, err = io.Copy(io.Discard, io.LimitReader(member, (128<<20)+1))
		}
		err = errors.Join(err, member.Close())
		if err != nil {
			return fmt.Errorf("ODT invalid %s: %w", f.Name, err)
		}
	}
	for _, name := range []string{"mimetype", "content.xml", "styles.xml", "META-INF/manifest.xml"} {
		if !seen[name] {
			return fmt.Errorf("ODT missing package member %s", name)
		}
	}
	return nil
}

// Validate XML incrementally so large package members are never expanded in memory.
func odtValidateXML(member io.Reader) error {
	decoder := xml.NewDecoder(io.LimitReader(member, (128<<20)+1))
	var err error
	depth := 0
	root := false
	for {
		token, e := decoder.Token()
		if e == io.EOF {
			break
		}
		if e != nil {
			err = e
			break
		}
		switch v := token.(type) {
		case xml.StartElement:
			depth++
			if depth == 1 {
				if root {
					err = fmt.Errorf("multiple XML roots")
				}
				root = true
			}
			if depth > 256 {
				err = fmt.Errorf("XML nesting exceeds limit")
			}
		case xml.EndElement:
			depth--
		case xml.Directive:
			err = fmt.Errorf("XML directives are not allowed")
		case xml.CharData:
			if depth == 0 && len(bytes.TrimSpace(v)) != 0 {
				err = fmt.Errorf("text outside XML root")
			}
		}
		if err != nil {
			break
		}
	}
	if !root && err == nil {
		err = fmt.Errorf("missing XML root")
	}

	return err
}
