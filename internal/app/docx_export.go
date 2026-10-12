// ABOUTME: Publishes bounded, native DOCX packages through the conversion command.
// ABOUTME: Keeps output atomic and independent of the existing PDF and text routes.
package app

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/tigger-developer/first-folio/internal/config"
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
	raw, err = io.ReadAll(io.LimitReader(f, docxSourceLimit+1))
	if err != nil {
		return nil, fmt.Errorf("reading DOCX source: %w", err)
	}
	if len(raw) > docxSourceLimit {
		return nil, fmt.Errorf("DOCX source exceeds 32 MiB")
	}
	if err := docxValidText(string(raw)); err != nil {
		return nil, err
	}
	return raw, nil
}

func exportDOCX(doc play.Document, cfg config.Config, opts convertOptions, toStdout bool, stdout, stderr io.Writer) error {
	if toStdout && writerIsTerminal(stdout) && !opts.force {
		return fmt.Errorf("DOCX is binary output. Refusing to write to terminal\nUse a target file, redirect stdout, or pass --force to override")
	}
	if !toStdout {
		if err := docxTarget(opts.source, opts.target); err != nil {
			return err
		}
	}
	writer, err := newDOCXWriter(doc, cfg)
	if err != nil {
		return err
	}
	parts, err := writer.parts()
	if err != nil {
		return err
	}
	raw, err := docxPackage(parts)
	if err != nil {
		return err
	}
	raw, fontWarnings, err := embedOfficeFonts(raw, "docx")
	if err != nil {
		return err
	}
	writer.warnings = append(writer.warnings, fontWarnings...)
	for _, warning := range writer.warnings {
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

func docxPackage(parts []docxPart) ([]byte, error) {
	content := docxNode{name: "Types"}.attr("xmlns", docxContentNS)
	content.children = append(content.children, docxNode{name: "Default"}.attr("Extension", "rels").attr("ContentType", "application/vnd.openxmlformats-package.relationships+xml"), docxNode{name: "Default"}.attr("Extension", "xml").attr("ContentType", "application/xml"))
	relationships := docxNode{name: "Relationships"}.attr("xmlns", docxPackageRelNS)
	for _, part := range parts {
		content.children = append(content.children, docxNode{name: "Override"}.attr("PartName", "/"+part.name).attr("ContentType", "application/vnd.openxmlformats-officedocument.wordprocessingml."+part.contentType+"+xml"))
		if part.relationship != "" {
			target := strings.TrimPrefix(part.name, "word/")
			relationships.children = append(relationships.children, docxRelationship(strings.TrimSuffix(target, ".xml"), part.relationship, target))
		}
	}
	rootRels := docxNode{name: "Relationships", children: []docxNode{docxRelationship("document", "officeDocument", "word/document.xml")}}.attr("xmlns", docxPackageRelNS)
	parts = append(parts, docxPart{name: "[Content_Types].xml", root: content}, docxPart{name: "_rels/.rels", root: rootRels}, docxPart{name: "word/_rels/document.xml.rels", root: relationships})
	var output docxLimitedBuffer
	output.limit = docxXMLLimit
	archive := zip.NewWriter(&output)
	remaining := docxXMLLimit
	for _, part := range parts {
		raw, err := docxEncode(part.root, remaining)
		if err != nil {
			return nil, fmt.Errorf("encoding %s: %w", part.name, err)
		}
		remaining -= len(raw)
		file, err := archive.Create(part.name)
		if err != nil {
			return nil, err
		}
		if _, err := file.Write(raw); err != nil {
			return nil, err
		}
	}
	if err := archive.Close(); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

func docxRelationship(id, kind, target string) docxNode {
	return docxNode{name: "Relationship"}.attr("Id", id).attr("Type", docxRelNS+"/"+kind).attr("Target", target)
}
