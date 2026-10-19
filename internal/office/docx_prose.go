// ABOUTME: Adapts structured prose to the existing native DOCX paragraph and package writer.
// ABOUTME: Preserves editable lists, tables, links, notes, headings and generated images.
package office

import (
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"math"
	"strconv"
	"strings"
)

type proseDOCX struct {
	writer            *docxWriter
	plan              *Prose
	relationships     []docxNode
	noteRelationships []docxNode
	inNote            bool
	numbering         []docxNode
	assets            map[string][]byte
	bookmark          int
	listLevel         int
}

func (p *Prose) nativeDOCX(blocks []Node) ([]byte, error) {
	w := &docxWriter{width: int(p.Page.Width * 20), height: int(p.Page.Height * 20), margin: int(p.Page.Left * 20), baseSize: p.Styles["FolioBody"].Size}
	b := &proseDOCX{writer: w, plan: p, assets: map[string][]byte{}}
	body, err := b.blocks(blocks, "FolioBody", 0)
	if err != nil {
		return nil, err
	}
	parts := w.contentParts(body, p.Page.wordSection())
	if len(b.numbering) > 0 {
		parts = append(parts, docxPart{name: "word/numbering.xml", contentType: "numbering", relationship: "numbering", root: dxRoot("numbering", b.numbering...)})
	}
	raw, err := docxPackage(parts)
	if err != nil {
		return nil, err
	}
	files, err := readPackage(raw, defaultPackageLimits())
	if err != nil {
		return nil, err
	}
	for name, data := range b.assets {
		files[name] = data
	}
	if len(b.relationships) > 0 {
		if err = officeAppendXML(files, "word/_rels/document.xml.rels", "Relationships", b.relationships...); err != nil {
			return nil, err
		}
	}
	if len(b.noteRelationships) > 0 {
		files["word/_rels/footnotes.xml.rels"], err = docxEncode(ox("Relationships", b.noteRelationships...).attr("xmlns", docxPackageRelNS), DataLimit)
		if err != nil {
			return nil, err
		}
	}
	core := ox("cp:coreProperties").attr("xmlns:cp", "http://schemas.openxmlformats.org/package/2006/metadata/core-properties").attr("xmlns:dc", "http://purl.org/dc/elements/1.1/")
	files["docProps/core.xml"], err = docxEncode(core, DataLimit)
	if err != nil {
		return nil, err
	}
	if err = officeAppendXML(files, "_rels/.rels", "Relationships", ox("Relationship").attr("Id", "core").attr("Type", "http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties").attr("Target", "docProps/core.xml")); err != nil {
		return nil, err
	}
	types := []docxNode{ox("Override").attr("PartName", "/docProps/core.xml").attr("ContentType", "application/vnd.openxmlformats-package.core-properties+xml")}
	if len(b.assets) > 0 {
		types = append(types, ox("Default").attr("Extension", "svg").attr("ContentType", "image/svg+xml"))
	}
	if err = officeAppendXML(files, "[Content_Types].xml", "Types", types...); err != nil {
		return nil, err
	}
	return writePackage(files, "docx")
}

func proseNodes(value any) ([]Node, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var nodes []Node
	err = json.Unmarshal(raw, &nodes)
	return nodes, err
}
func proseAttr(value any, key string) string {
	attr, ok := value.([]any)
	if !ok || len(attr) != 3 {
		return ""
	}
	if key == "id" {
		s, _ := attr[0].(string)
		return s
	}
	pairs, _ := attr[2].([]any)
	for _, v := range pairs {
		pair, ok := v.([]any)
		if ok && len(pair) == 2 && pair[0] == key {
			s, _ := pair[1].(string)
			return s
		}
	}
	return ""
}
func (b *proseDOCX) childBlocks(value any, role string, depth int) ([]docxNode, error) {
	nodes, err := proseNodes(value)
	if err != nil {
		return nil, err
	}
	return b.blocks(nodes, role, depth+1)
}
func (b *proseDOCX) blocks(nodes []Node, role string, depth int) ([]docxNode, error) {
	if depth > 128 {
		return nil, fmt.Errorf("DOCX prose nesting exceeds limit")
	}
	var out []docxNode
	for _, node := range nodes {
		c, _ := node.Content.([]any)
		switch node.Type {
		case "Div":
			if len(c) != 2 {
				return nil, fmt.Errorf("invalid prose Div")
			}
			childRole := proseAttr(c[0], "custom-style")
			if childRole == "" {
				childRole = role
			}
			children, err := b.childBlocks(c[1], childRole, depth)
			if err != nil {
				return nil, err
			}
			out = append(out, children...)
		case "Para", "Plain", "Header":
			value := node.Content
			id := ""
			if node.Type == "Header" {
				if len(c) != 3 {
					return nil, fmt.Errorf("invalid prose heading")
				}
				value = c[2]
				id = proseAttr(c[1], "id")
			}
			inlines, err := b.inlines(value, dx("rPr"), depth+1)
			if err != nil {
				return nil, err
			}
			paragraph := dx("p", dx("pPr", dxVal("pStyle", role)))
			if node.Type == "Header" {
				level, _ := c[0].(float64)
				paragraph.children[0].children = append(paragraph.children[0].children, dxVal("outlineLvl", dxInt(int(level)-1)))
			}
			if id != "" {
				b.bookmark++
				paragraph.children = append(paragraph.children, dx("bookmarkStart").attr("w:id", dxInt(b.bookmark)).attr("w:name", id))
			}
			paragraph.children = append(paragraph.children, inlines...)
			if id != "" {
				paragraph.children = append(paragraph.children, dx("bookmarkEnd").attr("w:id", dxInt(b.bookmark)))
			}
			out = append(out, paragraph)
		case "BlockQuote":
			children, err := b.childBlocks(node.Content, "FolioQuote", depth)
			if err != nil {
				return nil, err
			}
			out = append(out, children...)
		case "CodeBlock":
			if len(c) != 2 {
				return nil, fmt.Errorf("invalid prose code block")
			}
			out = append(out, dx("p", dx("pPr", dxVal("pStyle", "FolioCode")), docxRun(fmt.Sprint(c[1]), dx("rPr"))))
		case "LineBlock":
			for _, line := range c {
				runs, err := b.inlines(line, dx("rPr"), depth+1)
				if err != nil {
					return nil, err
				}
				out = append(out, dx("p", append([]docxNode{dx("pPr", dxVal("pStyle", role))}, runs...)...))
			}
		case "BulletList", "OrderedList":
			items := c
			start := 1
			kind := "bullet"
			label := "•"
			if node.Type == "OrderedList" {
				if len(c) != 2 {
					return nil, fmt.Errorf("invalid ordered list")
				}
				items, _ = c[1].([]any)
				attrs, _ := c[0].([]any)
				if len(attrs) == 3 {
					v, _ := attrs[0].(float64)
					start = int(v)
				}
				kind = "decimal"
				label = "%1."
			}
			number := len(b.numbering)/2 + 1
			listLevel := b.listLevel
			// Word numbers nine levels per definition. Deeper source lists
			// retain their indentation in their own editable definition.
			nativeLevel := min(listLevel, 8)
			indent := dxInt((listLevel + 1) * 360)
			if kind != "bullet" {
				label = "%" + dxInt(nativeLevel+1) + "."
			}
			level := dx("lvl", dxVal("start", dxInt(start)), dxVal("numFmt", kind), dxVal("lvlText", label), dx("pPr", dx("tabs", dx("tab").attr("w:val", "num").attr("w:pos", indent)), dx("ind").attr("w:left", indent).attr("w:hanging", "360"))).attr("w:ilvl", dxInt(nativeLevel))
			b.numbering = append(b.numbering, dx("abstractNum", level).attr("w:abstractNumId", dxInt(number)), dx("num", dxVal("abstractNumId", dxInt(number))).attr("w:numId", dxInt(number)))
			for _, item := range items {
				b.listLevel++
				children, err := b.childBlocks(item, role, depth)
				b.listLevel--
				if err != nil {
					return nil, err
				}
				for i := range children {
					if children[i].name == "w:p" {
						children[i].ensure("w:pPr").children = append(children[i].ensure("w:pPr").children, dx("numPr", dxVal("ilvl", dxInt(nativeLevel)), dxVal("numId", dxInt(number))))
						break
					}
				}
				out = append(out, children...)
			}
		case "Table":
			table, err := b.table(c, role, depth)
			if err != nil {
				return nil, err
			}
			out = append(out, table)
		case "DefinitionList":
			for _, item := range c {
				pair, _ := item.([]any)
				if len(pair) != 2 {
					return nil, fmt.Errorf("invalid definition list")
				}
				runs, err := b.inlines(pair[0], dx("rPr", dx("b")), depth+1)
				if err != nil {
					return nil, err
				}
				out = append(out, dx("p", append([]docxNode{dx("pPr", dxVal("pStyle", role))}, runs...)...))
				defs, _ := pair[1].([]any)
				for _, def := range defs {
					children, err := b.childBlocks(def, role, depth)
					if err != nil {
						return nil, err
					}
					out = append(out, children...)
				}
			}
		case "HorizontalRule":
			out = append(out, dx("p", dx("pPr", dxVal("pStyle", "FolioScene")), docxRun("* * *", dx("rPr"))))
		case "Null":
		default:
			return nil, fmt.Errorf("unsupported native DOCX block %q", node.Type)
		}
	}
	return out, nil
}

func (b *proseDOCX) inlines(value any, props docxNode, depth int) ([]docxNode, error) {
	if depth > 128 {
		return nil, fmt.Errorf("DOCX inline nesting exceeds limit")
	}
	nodes, err := proseNodes(value)
	if err != nil {
		return nil, err
	}
	var out []docxNode
	for _, node := range nodes {
		c, _ := node.Content.([]any)
		childProps := dx("rPr", append([]docxNode(nil), props.children...)...)
		var children []docxNode
		switch node.Type {
		case "Str":
			text, ok := node.Content.(string)
			if !ok {
				return nil, fmt.Errorf("invalid prose text")
			}
			children = []docxNode{docxRun(text, props)}
		case "Space", "SoftBreak":
			children = []docxNode{docxRun(" ", props)}
		case "LineBreak":
			children = []docxNode{docxRun("\n", props)}
		case "Strong", "Emph", "Underline", "Strikeout", "Superscript", "Subscript", "SmallCaps":
			property := map[string]docxNode{"Strong": dx("b"), "Emph": dx("i"), "Underline": dxVal("u", "single"), "Strikeout": dx("strike"), "Superscript": dxVal("vertAlign", "superscript"), "Subscript": dxVal("vertAlign", "subscript"), "SmallCaps": dx("smallCaps")}[node.Type]
			childProps.children = append(childProps.children, property)
			children, err = b.inlines(node.Content, childProps, depth+1)
		case "Span":
			if len(c) != 2 {
				return nil, fmt.Errorf("invalid prose span")
			}
			if style := proseAttr(c[0], "custom-style"); style != "" {
				childProps.children = append(childProps.children, dxVal("rStyle", style))
			}
			children, err = b.inlines(c[1], childProps, depth+1)
		case "Code":
			if len(c) != 2 {
				return nil, fmt.Errorf("invalid inline code")
			}
			childProps.children = append(childProps.children, dxVal("rStyle", "FolioMono"))
			children = []docxNode{docxRun(fmt.Sprint(c[1]), childProps)}
		case "Link":
			if len(c) != 3 {
				return nil, fmt.Errorf("invalid prose link")
			}
			target, _ := c[2].([]any)
			if len(target) != 2 {
				return nil, fmt.Errorf("invalid link target")
			}
			runs, e := b.inlines(c[1], props, depth+1)
			if e != nil {
				return nil, e
			}
			address, _ := target[0].(string)
			link := dx("hyperlink", runs...)
			if strings.HasPrefix(address, "#") {
				link = link.attr("w:anchor", strings.TrimPrefix(address, "#"))
			} else {
				id := "proseLink" + dxInt(len(b.relationships)+len(b.noteRelationships)+1)
				rel := docxRelationship(id, "hyperlink", address).attr("TargetMode", "External")
				if b.inNote {
					b.noteRelationships = append(b.noteRelationships, rel)
				} else {
					b.relationships = append(b.relationships, rel)
				}
				link = link.attr("r:id", id)
			}
			children = []docxNode{link}
		case "Note":
			id := len(b.writer.footnotes) + 1
			previousNote := b.inNote
			b.inNote = true
			paragraphs, e := b.childBlocks(node.Content, "FootnoteText", depth)
			b.inNote = previousNote
			if e != nil {
				return nil, e
			}
			if len(paragraphs) == 0 {
				paragraphs = []docxNode{dx("p")}
			}
			paragraphs[0].children = append(paragraphs[0].children[:1], append([]docxNode{dx("r", dx("footnoteRef")), docxRun(" ", dx("rPr"))}, paragraphs[0].children[1:]...)...)
			b.writer.footnotes = append(b.writer.footnotes, dx("footnote", paragraphs...).attr("w:id", dxInt(id)))
			children = []docxNode{dx("r", dx("rPr", dxVal("vertAlign", "superscript")), dx("footnoteReference").attr("w:id", dxInt(id)))}
		case "Quoted":
			if len(c) != 2 {
				return nil, fmt.Errorf("invalid quoted text")
			}
			runs, e := b.inlines(c[1], props, depth+1)
			if e != nil {
				return nil, e
			}
			children = append([]docxNode{docxRun("“", props)}, runs...)
			children = append(children, docxRun("”", props))
		case "Image":
			children, err = b.image(c)
		default:
			return nil, fmt.Errorf("unsupported native DOCX inline %q", node.Type)
		}
		if err != nil {
			return nil, err
		}
		out = append(out, children...)
	}
	return out, nil
}

func (b *proseDOCX) table(c []any, role string, depth int) (docxNode, error) {
	if len(c) != 6 {
		return docxNode{}, fmt.Errorf("invalid prose table")
	}
	columns, _ := c[2].([]any)
	if len(columns) == 0 {
		return docxNode{}, fmt.Errorf("table has no columns")
	}
	width := int((b.plan.Page.Width - b.plan.Page.Left - b.plan.Page.Right - b.plan.Page.Gutter) * 20)
	table := b.writer.table(width, width/len(columns))
	grid := dx("tblGrid")
	for range columns {
		grid.children = append(grid.children, dx("gridCol").attr("w:w", dxInt(width/len(columns))))
	}
	table.children[1] = grid
	var rows []any
	head, _ := c[3].([]any)
	if len(head) == 2 {
		r, _ := head[1].([]any)
		rows = append(rows, r...)
	}
	bodies, _ := c[4].([]any)
	for _, body := range bodies {
		v, _ := body.([]any)
		if len(v) == 4 {
			r, _ := v[2].([]any)
			rows = append(rows, r...)
			r, _ = v[3].([]any)
			rows = append(rows, r...)
		}
	}
	foot, _ := c[5].([]any)
	if len(foot) == 2 {
		r, _ := foot[1].([]any)
		rows = append(rows, r...)
	}
	for _, value := range rows {
		row, _ := value.([]any)
		if len(row) != 2 {
			return docxNode{}, fmt.Errorf("invalid table row")
		}
		cells, _ := row[1].([]any)
		tr := dx("tr")
		for _, value := range cells {
			cell, _ := value.([]any)
			if len(cell) != 5 {
				return docxNode{}, fmt.Errorf("invalid table cell")
			}
			span, _ := cell[3].(float64)
			if span < 1 {
				span = 1
			}
			props := dx("tcPr", dx("tcW").attr("w:w", dxInt(width*int(span)/len(columns))).attr("w:type", "dxa"))
			if span > 1 {
				props.children = append(props.children, dxVal("gridSpan", strconv.Itoa(int(span))))
			}
			paragraphs, err := b.childBlocks(cell[4], role, depth)
			if err != nil {
				return docxNode{}, err
			}
			if len(paragraphs) == 0 {
				paragraphs = []docxNode{dx("p")}
			}
			tr.children = append(tr.children, dx("tc", append([]docxNode{props}, paragraphs...)...))
		}
		table.children = append(table.children, tr)
	}
	return table, nil
}

func (b *proseDOCX) image(c []any) ([]docxNode, error) {
	if len(c) != 3 {
		return nil, fmt.Errorf("invalid generated image")
	}
	target, _ := c[2].([]any)
	if len(target) != 2 {
		return nil, fmt.Errorf("invalid generated image target")
	}
	uri, _ := target[0].(string)
	prefix := "data:image/svg+xml;base64,"
	if !strings.HasPrefix(uri, prefix) {
		return nil, fmt.Errorf("only generated local barcode images are supported")
	}
	data, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(uri, prefix))
	if err != nil {
		return nil, err
	}
	var svg struct {
		ViewBox string `xml:"viewBox,attr"`
	}
	if err := xml.Unmarshal(data, &svg); err != nil {
		return nil, fmt.Errorf("reading generated SVG dimensions: %w", err)
	}
	var x, y, width, height float64
	if n, err := fmt.Sscanf(svg.ViewBox, "%f %f %f %f", &x, &y, &width, &height); err != nil || n != 4 || width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid generated SVG viewBox %q", svg.ViewBox)
	}
	imageHeight := strconv.FormatInt(int64(math.Round(1440000*height/width)), 10)
	id := "proseImage" + dxInt(len(b.assets)+1)
	name := "media/" + id + ".svg"
	b.assets["word/"+name] = data
	b.relationships = append(b.relationships, docxRelationship(id, "image", name))
	// DrawingML remains a packaged image, separate from editable body text.
	pic := ox("pic:pic", ox("pic:nvPicPr", ox("pic:cNvPr").attr("id", "0").attr("name", "ISBN barcode"), ox("pic:cNvPicPr")), ox("pic:blipFill", ox("a:blip").attr("r:embed", id), ox("a:stretch", ox("a:fillRect"))), ox("pic:spPr", ox("a:xfrm", ox("a:off").attr("x", "0").attr("y", "0"), ox("a:ext").attr("cx", "1440000").attr("cy", imageHeight)), ox("a:prstGeom", ox("a:avLst")).attr("prst", "rect"))).attr("xmlns:pic", "http://schemas.openxmlformats.org/drawingml/2006/picture")
	drawing := dx("drawing", ox("wp:inline", ox("wp:extent").attr("cx", "1440000").attr("cy", imageHeight), ox("wp:docPr").attr("id", dxInt(len(b.assets))).attr("name", "ISBN barcode"), ox("a:graphic", ox("a:graphicData", pic).attr("uri", "http://schemas.openxmlformats.org/drawingml/2006/picture")).attr("xmlns:a", "http://schemas.openxmlformats.org/drawingml/2006/main")).attr("xmlns:wp", "http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing"))
	return []docxNode{dx("r", drawing)}, nil
}
