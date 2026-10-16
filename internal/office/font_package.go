// ABOUTME: Embeds full font faces into native DOCX and ODT package structures.
// ABOUTME: Preserves existing XML namespaces and bounds the repackaged document.
package office

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/xml"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"sort"
	"strings"
)

func EmbedFonts(raw []byte, format string) ([]byte, []string, error) {
	z, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		return nil, nil, err
	}
	files := map[string][]byte{}
	families := map[string]bool{}
	remaining := 256 << 20
	for _, member := range z.File {
		if member.UncompressedSize64 > uint64(remaining) || member.UncompressedSize64 > docxXMLLimit {
			return nil, nil, fmt.Errorf("office package exceeds expanded limit")
		}
		r, err := member.Open()
		if err != nil {
			return nil, nil, err
		}
		data, err := io.ReadAll(io.LimitReader(r, int64(remaining)+1))
		err = errors.Join(err, r.Close())
		if err != nil {
			return nil, nil, err
		}
		remaining -= len(data)
		if remaining < 0 {
			return nil, nil, fmt.Errorf("office package exceeds expanded limit")
		}
		files[member.Name] = data
		if strings.HasSuffix(member.Name, ".xml") && member.Name != "[Content_Types].xml" {
			if err := fontFaceFamilies(data, families); err != nil {
				return nil, nil, err
			}
		}
	}
	var names []string
	for name := range families {
		names = append(names, name)
	}
	fonts, warnings := ResolveFonts(names, fontFaceDirs())
	if len(fonts) == 0 {
		return raw, warnings, nil
	}
	if format == "docx" {
		err = embedDOCXFonts(files, fonts)
	} else {
		err = embedODTFonts(files, fonts)
	}
	if err != nil {
		return nil, nil, err
	}
	if len(files) > 256 {
		return nil, nil, fmt.Errorf("office package exceeds 256 entries")
	}
	var paths []string
	for path := range files {
		if path != "mimetype" {
			paths = append(paths, path)
		}
	}
	sort.Strings(paths)
	if format == "odt" {
		paths = append([]string{"mimetype"}, paths...)
	}
	output := nativeBuffer{limit: docxXMLLimit}
	w := zip.NewWriter(&output)
	remaining = 256 << 20
	for _, path := range paths {
		data := files[path]
		remaining -= len(data)
		if remaining < 0 || len(data) > docxXMLLimit {
			return nil, nil, fmt.Errorf("office embedded package exceeds expanded limit")
		}
		method := uint16(zip.Deflate)
		if path == "mimetype" {
			method = zip.Store
		}
		header := &zip.FileHeader{Name: path, Method: method}
		var member io.Writer
		var err error
		if path == "mimetype" {
			// LibreOffice identifies ODF from the first local header, before
			// reading the central directory. A data descriptor hides its size.
			header.CRC32 = crc32.ChecksumIEEE(data)
			header.CompressedSize64 = uint64(len(data))
			header.UncompressedSize64 = uint64(len(data))
			member, err = w.CreateRaw(header)
		} else {
			member, err = w.CreateHeader(header)
		}
		if err != nil {
			return nil, nil, err
		}
		if _, err = member.Write(data); err != nil {
			return nil, nil, err
		}
	}
	if err = w.Close(); err != nil {
		return nil, nil, err
	}
	return output.Bytes(), warnings, nil
}

func fontFaceFamilies(raw []byte, families map[string]bool) error {
	d := xml.NewDecoder(bytes.NewReader(raw))
	for {
		token, err := d.Token()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if n, ok := token.(xml.StartElement); ok {
			for _, a := range n.Attr {
				if (n.Name.Space == docxWordNS && n.Name.Local == "rFonts" && (a.Name.Local == "ascii" || a.Name.Local == "hAnsi" || a.Name.Local == "eastAsia" || a.Name.Local == "cs")) || a.Name.Local == "font-family" {
					name := strings.Trim(a.Value, "'\"")
					if name != "" {
						families[name] = true
					}
				}
			}
		}
	}
}

// RawToken plus literal qualified names preserves prefixes and mixed content.
// Input is generated/validated XML, not arbitrary external markup.
func fontFaceXML(raw []byte, start func(xml.StartElement) []docxNode, end func(xml.EndElement) []docxNode) ([]byte, error) {
	out := nativeBuffer{limit: docxXMLLimit}
	d := xml.NewDecoder(bytes.NewReader(raw))
	e := xml.NewEncoder(&out)
	qualified := func(n xml.Name) xml.Name {
		if n.Space != "" {
			n.Local = n.Space + ":" + n.Local
			n.Space = ""
		}
		return n
	}
	encode := func(nodes []docxNode) error {
		for _, n := range nodes {
			if err := e.Encode(n); err != nil {
				return err
			}
		}
		return nil
	}
	for {
		token, err := d.RawToken()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		switch n := token.(type) {
		case xml.StartElement:
			n.Name = qualified(n.Name)
			for i := range n.Attr {
				n.Attr[i].Name = qualified(n.Attr[i].Name)
			}
			if err = e.EncodeToken(n); err != nil {
				return nil, err
			}
			if start != nil {
				if err = encode(start(n)); err != nil {
					return nil, err
				}
			}
			continue
		case xml.EndElement:
			n.Name = qualified(n.Name)
			if end != nil {
				if err = encode(end(n)); err != nil {
					return nil, err
				}
			}
			token = n
		}
		if err = e.EncodeToken(token); err != nil {
			return nil, err
		}
	}
	if err := e.Close(); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func officeAppendXML(files map[string][]byte, path, parent string, nodes ...docxNode) error {
	raw, err := fontFaceXML(files[path], nil, func(n xml.EndElement) []docxNode {
		if n.Name.Local == parent {
			return nodes
		}
		return nil
	})
	if err == nil {
		files[path] = raw
	}
	return err
}

func embedDOCXFonts(files map[string][]byte, fonts []fontFace) error {
	table := dxRoot("fonts")
	rels := ox("Relationships").attr("xmlns", docxPackageRelNS)
	for i := 0; i < len(fonts); {
		font := dx("font").attr("w:name", fonts[i].family)
		family := fonts[i].family
		for i < len(fonts) && fonts[i].family == family {
			f := fonts[i]
			id := fmt.Sprintf("folioFont%d", i+1)
			path := "fonts/" + id + ".odttf"
			hash := sha256.Sum256(f.Raw)
			key := fmt.Sprintf("{%X-%X-%X-%X-%X}", hash[:4], hash[4:6], hash[6:8], hash[8:10], hash[10:16])
			data := bytes.Clone(f.Raw)
			for j := 0; j < 32; j++ {
				data[j] ^= hash[15-j%16]
			}
			files["word/"+path] = data
			font.children = append(font.children, dx([]string{"embedRegular", "embedBold", "embedItalic", "embedBoldItalic"}[f.face]).attr("r:id", id).attr("w:fontKey", key).attr("w:subsetted", "false"))
			rels.children = append(rels.children, docxRelationship(id, "font", path))
			i++
		}
		table.children = append(table.children, font)
	}
	var err error
	files["word/fontTable.xml"], err = docxEncode(table, docxXMLLimit)
	if err != nil {
		return err
	}
	files["word/_rels/fontTable.xml.rels"], err = docxEncode(rels, docxXMLLimit)
	if err != nil {
		return err
	}
	if err = officeAppendXML(files, "[Content_Types].xml", "Types", ox("Default").attr("Extension", "odttf").attr("ContentType", "application/vnd.openxmlformats-officedocument.obfuscatedFont"), ox("Override").attr("PartName", "/word/fontTable.xml").attr("ContentType", "application/vnd.openxmlformats-officedocument.wordprocessingml.fontTable+xml")); err != nil {
		return err
	}
	if err = officeAppendXML(files, "word/_rels/document.xml.rels", "Relationships", docxRelationship("folioFontTable", "fontTable", "fontTable.xml")); err != nil {
		return err
	}
	// Settings precede updateFields in the WordprocessingML settings sequence.
	files["word/settings.xml"], err = fontFaceXML(files["word/settings.xml"], func(n xml.StartElement) []docxNode {
		if n.Name.Local == "w:settings" {
			return []docxNode{dx("embedTrueTypeFonts"), dx("embedSystemFonts"), dxVal("saveSubsetFonts", "false")}
		}
		return nil
	}, nil)
	return err
}

func embedODTFonts(files map[string][]byte, fonts []fontFace) error {
	sources := map[string]docxNode{}
	var manifest []docxNode
	for i, f := range fonts {
		path := fmt.Sprintf("Fonts/folioFont%d.%s", i+1, f.ext)
		files[path] = f.Raw
		manifest = append(manifest, ox("manifest:file-entry").attr("manifest:full-path", path).attr("manifest:media-type", "font/"+f.ext))
		src, ok := sources[f.family]
		if !ok {
			src = ox("svg:font-face-src").attr("xmlns:svg", "urn:oasis:names:tc:opendocument:xmlns:svg-compatible:1.0")
		}
		fontFormat := "truetype"
		if f.ext == "otf" {
			fontFormat = "opentype"
		}
		uri := ox("svg:font-face-uri", ox("svg:font-face-format").attr("svg:string", fontFormat)).attr("xmlns:xlink", "http://www.w3.org/1999/xlink").attr("xlink:href", path).attr("xlink:type", "simple")
		src.children = append(src.children, uri)
		sources[f.family] = src
	}
	for _, path := range []string{"styles.xml", "content.xml"} {
		seen := map[string]bool{}
		raw, err := fontFaceXML(files[path], func(n xml.StartElement) []docxNode {
			if n.Name.Local == "style:font-face" {
				for _, a := range n.Attr {
					if a.Name.Local == "svg:font-family" {
						family := strings.Trim(a.Value, "'\"")
						if src, ok := sources[family]; ok {
							seen[family] = true
							return []docxNode{src}
						}
					}
				}
			}
			return nil
		}, func(n xml.EndElement) []docxNode {
			if n.Name.Local != "office:font-face-decls" {
				return nil
			}
			var nodes []docxNode
			for _, f := range fonts {
				if !seen[f.family] {
					nodes = append(nodes, ox("style:font-face", sources[f.family]).attr("style:name", f.family).attr("svg:font-family", f.family))
					seen[f.family] = true
				}
			}
			return nodes
		})
		if err != nil {
			return err
		}
		files[path] = raw
	}
	return officeAppendXML(files, "META-INF/manifest.xml", "manifest:manifest", manifest...)
}
