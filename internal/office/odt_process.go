// ABOUTME: Runs and validates the configured ODT writer shared by all modes.
// ABOUTME: Bounds subprocess output, diagnostics and package expansion.
package office

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

const odtMIME = "application/vnd.oasis.opendocument.text"

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

// RunODT invokes the bounded Pandoc ODT process.
func RunODT(ctx context.Context, input []byte, args ...string) ([]byte, string, error) {
	raw, diagnostic, err := runProcess(ctx, "pandoc", args, input, processLimits{docxXMLLimit, 64 << 10})
	if err != nil {
		return nil, "", fmt.Errorf("ODT pandoc failed (requires a compatible Pandoc 3.x writer version): %w: %s", err, diagnostic)
	}
	return raw, diagnostic, nil
}

// ValidateODT checks package structure and bounded XML members.
func ValidateODT(raw []byte) error {
	archive, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		return fmt.Errorf("ODT pandoc produced an invalid ZIP (requires a compatible Pandoc 3.x writer version): %w", err)
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
