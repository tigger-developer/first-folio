// ABOUTME: Packages generated prose assets with the shared bounded office archive format.
// ABOUTME: Keeps barcode images local and native package members deterministic.
package office

import (
	"archive/zip"
	"encoding/base64"
	"fmt"
	"hash/crc32"
	"io"
	"sort"
)

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
