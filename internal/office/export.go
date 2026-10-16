// ABOUTME: Generates configured references and writes prose through bounded Pandoc tools.
// ABOUTME: Validates native packages before and after font embedding without invoking Typst.
package office

import (
	"archive/zip"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"sort"
)

// Export builds one editable artefact; publication remains the caller's responsibility.
func (p *Prose) Export(format string) (result []byte, warnings []string, resultErr error) {
	if format != "docx" && format != "odt" {
		return nil, nil, fmt.Errorf("unsupported office format %s", format)
	}
	p.resolveBaselines()
	writerDocument := p.Document
	writerDocument.Meta = map[string]any{}
	writerDocument.Blocks = p.generatedImages(p.flowBoundaries(writerDocument.Blocks))
	encoded, err := json.Marshal(writerDocument)
	if err != nil {
		return nil, nil, err
	}
	if len(encoded) > DataLimit {
		return nil, nil, fmt.Errorf("office JSON exceeds 128 MiB")
	}
	reference, diagnostic, err := Pandoc(nil, "--print-default-data-file=reference."+format)
	if err != nil {
		return nil, nil, err
	}
	if diagnostic != "" {
		warnings = append(warnings, diagnostic)
	}
	files, err := readPackage(reference, defaultPackageLimits())
	if err != nil {
		return nil, nil, err
	}
	if err = p.configureStyles(files, format); err != nil {
		return nil, nil, err
	}
	if err = p.configurePage(files, format); err != nil {
		return nil, nil, err
	}
	reference, err = writePackage(files, format)
	if err != nil {
		return nil, nil, err
	}
	if err = validatePackage(reference, format, defaultPackageLimits()); err != nil {
		return nil, nil, err
	}
	dir, err := os.MkdirTemp("", "folio-prose-*")
	if err != nil {
		return nil, nil, err
	}
	defer func() { resultErr = errors.Join(resultErr, os.RemoveAll(dir)) }()
	ref := filepath.Join(dir, "reference."+format)
	if err = os.WriteFile(ref, reference, 0600); err != nil {
		return nil, nil, err
	}
	raw, diagnostic, err := Pandoc(encoded, "--from=json", "--to="+format, "--standalone", "--reference-doc="+ref)
	if err != nil {
		return nil, nil, err
	}
	if diagnostic != "" {
		warnings = append(warnings, diagnostic)
	}
	if err = validatePackage(raw, format, defaultPackageLimits()); err != nil {
		return nil, nil, err
	}
	files, err = readPackage(raw, defaultPackageLimits())
	if err != nil {
		return nil, nil, err
	}
	if err = p.configureStyles(files, format); err != nil {
		return nil, nil, err
	}
	if err = p.configurePage(files, format); err != nil {
		return nil, nil, err
	}
	if format == "docx" {
		err = p.wordLayout(files)
	} else {
		err = p.odtLayout(files)
	}
	if err != nil {
		return nil, nil, err
	}
	if err = p.nativeMetadata(files, format); err != nil {
		return nil, nil, err
	}
	raw, err = writePackage(files, format)
	if err != nil {
		return nil, nil, err
	}
	raw, fontWarnings, err := embedFonts(raw, format, true)
	warnings = append(warnings, fontWarnings...)
	if err != nil {
		return nil, nil, err
	}
	if err = validatePackage(raw, format, defaultPackageLimits()); err != nil {
		return nil, nil, err
	}
	return raw, warnings, nil
}

func (p *Prose) generatedImages(blocks []Node) []Node {
	out := append([]Node(nil), blocks...)
	for i, node := range out {
		if node.Type != "Div" {
			continue
		}
		c, ok := node.Content.([]any)
		if !ok || len(c) != 2 {
			continue
		}
		attr, ok := c[0].([]any)
		if !ok || len(attr) != 3 {
			continue
		}
		pairs, ok := attr[2].([][]string)
		if !ok {
			continue
		}
		for _, pair := range pairs {
			if len(pair) != 2 || pair[0] != "custom-style" {
				continue
			}
			data, exists := p.Images[pair[1]]
			if !exists {
				continue
			}
			uri := "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString(data)
			image := N("Image", []any{Attr("", nil, []string{"width", "40mm"}), Text("ISBN barcode"), []string{uri, ""}})
			out[i].Content = []any{c[0], []Node{N("Para", []Node{image})}}
		}
	}
	return out
}
func writePackage(files map[string][]byte, format string) ([]byte, error) {
	names := make([]string, 0, len(files))
	for name := range files {
		if name != "mimetype" {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	if format == "odt" {
		names = append([]string{"mimetype"}, names...)
	}
	if len(names) > 256 {
		return nil, fmt.Errorf("office package exceeds 256 entries")
	}
	out := nativeBuffer{limit: DataLimit}
	w := zip.NewWriter(&out)
	remaining := 256 << 20
	for _, name := range names {
		data := files[name]
		if len(data) > DataLimit || len(data) > remaining {
			return nil, fmt.Errorf("office package exceeds expanded/member limit")
		}
		remaining -= len(data)
		h := &zip.FileHeader{Name: name, Method: zip.Deflate}
		var member io.Writer
		var err error
		if name == "mimetype" {
			h.Method = zip.Store
			h.CRC32 = crc32.ChecksumIEEE(data)
			h.CompressedSize64 = uint64(len(data))
			h.UncompressedSize64 = uint64(len(data))
			member, err = w.CreateRaw(h)
		} else {
			member, err = w.CreateHeader(h)
		}
		if err != nil {
			return nil, err
		}
		if _, err = member.Write(data); err != nil {
			return nil, err
		}
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
