// ABOUTME: Extends the existing native DOCX and configured ODT exporters to prose.
// ABOUTME: Applies mode-specific layout without invoking a Pandoc DOCX writer.
package office

import (
	"encoding/json"
	"fmt"
)

// Export builds an editable package using the same serializers as script export.
func (p *Prose) Export(format string) ([]byte, []string, error) {
	if format != "docx" && format != "odt" {
		return nil, nil, fmt.Errorf("unsupported office format %q", format)
	}
	p.resolveBaselines()
	for name, band := range map[string]Band{"FolioBandheader": p.Page.Header, "FolioBandfooter": p.Page.Footer} {
		if band.Font.Size > 0 {
			p.Styles[name] = band.Font
		}
	}
	document := p.Document
	document.Meta = map[string]any{}
	document.Blocks = p.generatedImages(p.flowBoundaries(document.Blocks))
	if document.Blocks == nil {
		document.Blocks = []Node{}
	}
	encoded, err := json.Marshal(document)
	if err != nil {
		return nil, nil, err
	}
	if len(encoded) > DataLimit {
		return nil, nil, fmt.Errorf("office JSON exceeds 128 MiB")
	}
	var raw []byte
	var warnings []string
	if format == "docx" {
		// Normalize adapter-created and reader-created nodes to the same typed shape.
		var normalized Document
		if err = json.Unmarshal(encoded, &normalized); err != nil {
			return nil, nil, err
		}
		raw, err = p.nativeDOCX(normalized.Blocks)
	} else {
		// The native reference packager is shared with script ODT; no default
		// reference is obtained from Pandoc and no DOCX intermediate is involved.
		root := odtRoot("office:document-styles", ox("office:font-face-decls"), ox("office:styles", ox("style:style").attr("style:name", "Standard").attr("style:family", "paragraph")), ox("office:automatic-styles"), ox("office:master-styles"))
		var reference []byte
		reference, err = odtReferencePackage(root)
		if err != nil {
			return nil, nil, err
		}
		files, e := readPackage(reference, defaultPackageLimits())
		if e != nil {
			return nil, nil, e
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
		var diagnostic string
		raw, diagnostic, err = writeODT(encoded, reference)
		if diagnostic != "" {
			warnings = append(warnings, diagnostic)
		}
	}
	if err != nil {
		return nil, nil, err
	}
	if err := validatePackage(raw, format, defaultPackageLimits()); err != nil {
		return nil, nil, err
	}
	files, err := readPackage(raw, defaultPackageLimits())
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
	if err = p.alignListMarkers(files, format); err != nil {
		return nil, nil, err
	}
	raw, err = writePackage(files, format)
	if err != nil {
		return nil, nil, err
	}
	// Check font budgets before embedding; script retains its existing warning policy.
	families := map[string]bool{}
	for name, data := range files {
		if len(name) > 4 && name[len(name)-4:] == ".xml" {
			if err = officeFontFamilies(data, families); err != nil {
				return nil, nil, err
			}
		}
	}
	names := []string{}
	for name := range families {
		names = append(names, name)
	}
	faces, fontWarnings := ResolveFonts(names, officeFontDirs())
	if err = selectedFontLimits(names, faces, fontWarnings); err != nil {
		return nil, nil, err
	}
	raw, fontWarnings, err = embedOfficeFonts(raw, format)
	warnings = append(warnings, fontWarnings...)
	if err == nil {
		err = validatePackage(raw, format, defaultPackageLimits())
	}
	return raw, warnings, err
}
