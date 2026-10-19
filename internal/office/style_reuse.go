// ABOUTME: Separates final named formatting styles from temporary layout-instance identities.
// ABOUTME: Reuses equivalent role styles while retaining bookmarks, fields and page overrides.
package office

import (
	"encoding/xml"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var instanceStyle = regexp.MustCompile(`^((?:FolioFrame)?(?:FolioHeading|FolioContentsEntry)Level[1-6])Instance[0-9]+$|^(Folio(?:LetterBody|Legal|Flow|Blank|Band))[0-9]+(?:_[0-9]+)?$`)

// compactProseStyles runs after instance-specific layout has been applied. Only
// exact formatting matches within one role family share a named style. Layout
// and anchor identities have already been written on the individual paragraphs.
func compactProseStyles(files map[string][]byte, format string) error {
	path, key := "word/styles.xml", "w:styleId"
	if format == "odt" {
		path, key = "styles.xml", "style:name"
	}
	root, err := parseXML(files[path])
	if err != nil {
		return err
	}
	styles := root
	if format == "odt" {
		styles = root.ensure("office:styles")
	}
	aliases := map[string]string{}
	fingerprints, variants := map[string]string{}, map[string]int{}
	var retained, automatic []docxNode
	for _, style := range styles.children {
		name := style.get(key)
		if format == "odt" && strings.HasPrefix(name, "FolioPageParagraph") {
			automatic = append(automatic, style)
			continue
		}
		match := instanceStyle.FindStringSubmatch(name)
		if match == nil {
			retained = append(retained, style)
			continue
		}
		family := match[1]
		if family == "" {
			family = match[2]
		}
		encoded, e := docxEncode(namedProseStyle(style, key, format, ""), DataLimit)
		if e != nil {
			return e
		}
		fingerprint := family + "\x00" + string(encoded)
		canonical := fingerprints[fingerprint]
		if canonical == "" {
			variants[family]++
			canonical = family + "Variant" + strconv.Itoa(variants[family])
			fingerprints[fingerprint] = canonical
			style = namedProseStyle(style, key, format, canonical)
			retained = append(retained, style)
		}
		aliases[name] = canonical
	}
	styles.children = retained
	files[path], err = docxEncode(*root, DataLimit)
	if err != nil {
		return err
	}
	return rewriteProseStyleReferences(files, format, aliases, automatic)
}

// Copy before renaming: signatures and aliases must never mutate retained nodes.
func namedProseStyle(style docxNode, key, format, name string) docxNode {
	style.attrs = append([]xml.Attr(nil), style.attrs...)
	style.children = append([]docxNode(nil), style.children...)
	style.set(key, name)
	if format == "docx" {
		for i := range style.children {
			if style.children[i].name == "w:name" {
				style.children[i] = dxVal("name", name)
			}
		}
	} else {
		style.set("style:display-name", name)
	}
	return style
}

func rewriteProseStyleReferences(files map[string][]byte, format string, aliases map[string]string, automatic []docxNode) error {
	paths := []string{}
	for path := range files {
		if strings.HasSuffix(path, ".xml") {
			paths = append(paths, path)
		}
	}
	sort.Strings(paths)
	for _, path := range paths {
		root, err := parseXML(files[path])
		if err != nil {
			return err
		}
		if format == "odt" && path == "content.xml" {
			root.ensure("office:automatic-styles").children = append(root.ensure("office:automatic-styles").children, automatic...)
		}
		root.visit(func(n *docxNode) {
			for i, attr := range n.attrs {
				reference := attr.Name.Local == "text:style-name" || attr.Name.Local == "style:parent-style-name" || attr.Name.Local == "style:next-style-name" || attr.Name.Local == "draw:style-name"
				if format == "docx" {
					reference = attr.Name.Local == "w:val" && (n.name == "w:pStyle" || n.name == "w:rStyle" || n.name == "w:basedOn" || n.name == "w:next" || n.name == "w:link")
				}
				if replacement := aliases[attr.Value]; reference && replacement != "" {
					n.attrs[i].Value = replacement
				}
			}
		})
		files[path], err = docxEncode(*root, DataLimit)
		if err != nil {
			return err
		}
	}
	return nil
}

func withoutStyleAttribute(attrs []xml.Attr, name string) []xml.Attr {
	out := make([]xml.Attr, 0, len(attrs))
	for _, attr := range attrs {
		if attr.Name.Local != name {
			out = append(out, attr)
		}
	}
	return out
}
