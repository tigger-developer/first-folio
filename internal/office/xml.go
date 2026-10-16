// ABOUTME: Serializes a small, fixed subset of WordprocessingML using encoding/xml.
// ABOUTME: Bounds generated XML and rejects characters that XML cannot preserve.
package office

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"unicode/utf8"
)

const docxWordNS = "http://schemas.openxmlformats.org/wordprocessingml/2006/main"
const docxRelNS = "http://schemas.openxmlformats.org/officeDocument/2006/relationships"
const docxPackageRelNS = "http://schemas.openxmlformats.org/package/2006/relationships"
const docxContentNS = "http://schemas.openxmlformats.org/package/2006/content-types"
const docxSourceLimit = 32 << 20
const docxXMLLimit = 128 << 20

// Names and attribute keys are fixed by the writer; values are always encoded.
type docxNode struct {
	name     string
	attrs    []xml.Attr
	text     string
	children []docxNode
}

func dx(name string, children ...docxNode) docxNode {
	return docxNode{name: "w:" + name, children: children}
}
func (n docxNode) attr(key, value string) docxNode {
	n.attrs = append(n.attrs, xml.Attr{Name: xml.Name{Local: key}, Value: value})
	return n
}
func dxVal(name, value string) docxNode { return dx(name).attr("w:val", value) }
func dxRoot(name string, children ...docxNode) docxNode {
	return dx(name, children...).attr("xmlns:w", docxWordNS).attr("xmlns:r", docxRelNS)
}
func dxText(text string) docxNode {
	n := dx("t").attr("xml:space", "preserve")
	n.text = text
	return n
}

func (n docxNode) MarshalXML(e *xml.Encoder, _ xml.StartElement) error {
	if err := docxValidText(n.text); err != nil {
		return err
	}
	for _, a := range n.attrs {
		if err := docxValidText(a.Value); err != nil {
			return err
		}
	}
	start := xml.StartElement{Name: xml.Name{Local: n.name}, Attr: n.attrs}
	if err := e.EncodeToken(start); err != nil {
		return err
	}
	if n.text != "" {
		if err := e.EncodeToken(xml.CharData(n.text)); err != nil {
			return err
		}
	}
	for _, child := range n.children {
		if err := e.Encode(child); err != nil {
			return err
		}
	}
	return e.EncodeToken(start.End())
}

func docxValidText(s string) error {
	if !utf8.ValidString(s) {
		return fmt.Errorf("DOCX XML requires UTF-8 text")
	}
	for _, r := range s {
		if r == 9 || r == 10 || r == 13 || r >= 0x20 && r <= 0xD7FF || r >= 0xE000 && r <= 0xFFFD || r >= 0x10000 && r <= 0x10FFFF {
			continue
		}
		return fmt.Errorf("DOCX XML cannot represent character U+%04X", r)
	}
	return nil
}

type docxLimitedBuffer struct {
	bytes.Buffer
	limit int
}

func (b *docxLimitedBuffer) Write(p []byte) (int, error) {
	if len(p) > b.limit-b.Len() {
		return 0, fmt.Errorf("DOCX generated data exceeds %d bytes", b.limit)
	}
	return b.Buffer.Write(p)
}
func docxEncode(root docxNode, limit int) ([]byte, error) {
	var b docxLimitedBuffer
	b.limit = limit
	if _, err := io.WriteString(&b, xml.Header); err != nil {
		return nil, err
	}
	e := xml.NewEncoder(&b)
	if err := e.Encode(root); err != nil {
		return nil, err
	}
	if err := e.Close(); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func ox(name string, children ...docxNode) docxNode { return docxNode{name: name, children: children} }
func docxRelationship(id, kind, target string) docxNode {
	return ox("Relationship").attr("Id", id).attr("Type", docxRelNS+"/"+kind).attr("Target", target)
}

type nativeBuffer struct {
	buffer bytes.Buffer
	limit  int
}

func (b *nativeBuffer) Write(p []byte) (int, error) {
	if len(p) > b.limit-b.buffer.Len() {
		return 0, fmt.Errorf("native office data exceeds %d bytes", b.limit)
	}
	return b.buffer.Write(p)
}
func (b *nativeBuffer) Bytes() []byte { return b.buffer.Bytes() }
