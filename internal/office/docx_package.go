// ABOUTME: Packages native Word XML for all office-export callers.
// ABOUTME: Retains the original script serializer and bounded ZIP assembly.
package office

import (
	"archive/zip"
	"fmt"
	"path/filepath"
	"strings"
)

func docxPackage(parts []docxPart) ([]byte, error) {
	content := docxNode{name: "Types"}.attr("xmlns", docxContentNS)
	content.children = append(content.children, docxNode{name: "Default"}.attr("Extension", "rels").attr("ContentType", "application/vnd.openxmlformats-package.relationships+xml"), docxNode{name: "Default"}.attr("Extension", "xml").attr("ContentType", "application/xml"))
	relationships := docxNode{name: "Relationships"}.attr("xmlns", docxPackageRelNS)
	defaults := map[string]bool{"rels": true, "xml": true}
	var overrides []docxNode
	for _, part := range parts {
		if len(part.raw) > 0 {
			ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(part.name)), ".")
			if !defaults[ext] {
				mediaType := "image/" + ext
				if ext == "jpg" {
					mediaType = "image/jpeg"
				}
				content.children = append(content.children, docxNode{name: "Default"}.attr("Extension", ext).attr("ContentType", mediaType))
				defaults[ext] = true
			}
		} else {
			overrides = append(overrides, docxNode{name: "Override"}.attr("PartName", "/"+part.name).attr("ContentType", "application/vnd.openxmlformats-officedocument.wordprocessingml."+part.contentType+"+xml"))
		}
		if part.relationship != "" {
			target := strings.TrimPrefix(part.name, "word/")
			id := part.relationshipID
			if id == "" {
				id = strings.TrimSuffix(target, ".xml")
			}
			relationships.children = append(relationships.children, docxRelationship(id, part.relationship, target))
		}
	}
	content.children = append(content.children, overrides...)
	rootRels := docxNode{name: "Relationships", children: []docxNode{docxRelationship("document", "officeDocument", "word/document.xml")}}.attr("xmlns", docxPackageRelNS)
	parts = append(parts, docxPart{name: "[Content_Types].xml", root: content}, docxPart{name: "_rels/.rels", root: rootRels}, docxPart{name: "word/_rels/document.xml.rels", root: relationships})
	var output docxLimitedBuffer
	output.limit = docxXMLLimit
	archive := zip.NewWriter(&output)
	remaining := docxXMLLimit
	for _, part := range parts {
		raw := part.raw
		if raw == nil {
			var err error
			raw, err = docxEncode(part.root, remaining)
			if err != nil {
				return nil, fmt.Errorf("encoding %s: %w", part.name, err)
			}
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
