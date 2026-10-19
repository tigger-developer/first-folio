// ABOUTME: Packages native Word XML for all office-export callers.
// ABOUTME: Retains the original script serializer and bounded ZIP assembly.
package office

import (
	"archive/zip"
	"fmt"
	"strings"
)

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
