// ABOUTME: Adapts parsed plays to the shared native DOCX and configured ODT writers.
// ABOUTME: Returns validated packages without selecting or publishing destinations.
package office

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/tigger-developer/first-folio/internal/config"
	"github.com/tigger-developer/first-folio/internal/play"
)

// ValidText rejects text that cannot be preserved in office XML.
func ValidText(text string) error { return docxValidText(text) }

// ExportScript uses the established script layout and shared office serializers.
func ExportScript(doc play.Document, cfg config.Config, sourceFormat play.Format, format string) ([]byte, []string, error) {
	layout, err := newDOCXWriter(doc, cfg, sourceFormat)
	if err != nil {
		return nil, nil, err
	}
	var raw []byte
	switch format {
	case "docx":
		parts, e := layout.parts()
		if e != nil {
			return nil, nil, e
		}
		raw, err = docxPackage(parts)
	case "odt":
		body, e := layout.body()
		if e != nil {
			return nil, nil, e
		}
		projection := &odtProjection{layout: layout, variants: map[string]string{}}
		ast, e := projection.json(body)
		if e != nil {
			return nil, nil, e
		}
		reference, e := projection.reference()
		if e != nil {
			return nil, nil, e
		}
		var diagnostic string
		raw, diagnostic, err = writeODT(ast, reference)
		if diagnostic != "" {
			layout.warnings = append(layout.warnings, "Pandoc: "+diagnostic)
		}
	default:
		return nil, nil, fmt.Errorf("unsupported office format %q", format)
	}
	if err != nil {
		return nil, nil, err
	}
	raw, warnings, err := embedOfficeFonts(raw, format)
	if err == nil && format == "odt" {
		err = ValidateODT(raw)
	}
	return raw, append(layout.warnings, warnings...), err
}

// writeODT supplies the native configured reference to the bounded ODT writer.
func writeODT(ast, reference []byte) (raw []byte, diagnostic string, err error) {
	dir, err := os.MkdirTemp("", "folio-odt-*")
	if err != nil {
		return nil, "", err
	}
	defer func() { err = errors.Join(err, os.RemoveAll(dir)) }()
	path := filepath.Join(dir, "reference.odt")
	if err = os.WriteFile(path, reference, 0600); err != nil {
		return nil, "", err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	raw, diagnostic, err = RunODT(ctx, ast, "--from=json", "--to=odt", "--standalone", "--reference-doc="+path)
	if err == nil {
		err = ValidateODT(raw)
	}
	return
}
