// ABOUTME: Validates untrusted native office ZIP/XML packages and internal references.
// ABOUTME: Enforces byte/member budgets and rejects active content before publication.
package office

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"strings"
)

type packageLimits struct {
	compressed, expanded, member uint64
	entries                      int
}

func defaultPackageLimits() packageLimits { return packageLimits{DataLimit, 256 << 20, DataLimit, 256} }
func validatePackage(raw []byte, format string, limits packageLimits) error {
	files, err := readPackage(raw, limits)
	if err != nil {
		return err
	}
	return validateParts(files, format)
}

func readPackage(raw []byte, limits packageLimits) (map[string][]byte, error) {
	if uint64(len(raw)) > limits.compressed {
		return nil, fmt.Errorf("office compressed package exceeds limit")
	}
	z, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		return nil, fmt.Errorf("invalid office ZIP: %w", err)
	}
	if len(z.File) > limits.entries {
		return nil, fmt.Errorf("office package entry limit exceeded")
	}
	files := map[string][]byte{}
	remaining := limits.expanded
	for _, f := range z.File {
		if f.Name == "" || path.Clean(f.Name) != f.Name || strings.HasPrefix(f.Name, "/") || strings.Contains(f.Name, "\\") || strings.Contains(f.Name, ":") || f.Name == ".." || strings.HasPrefix(f.Name, "../") || f.Mode()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("unsafe office member %q", f.Name)
		}
		if _, ok := files[f.Name]; ok {
			return nil, fmt.Errorf("duplicate office member %s", f.Name)
		}
		if activePart(f.Name) {
			return nil, fmt.Errorf("active office payload prohibited: %s", f.Name)
		}
		if f.UncompressedSize64 > limits.member || f.UncompressedSize64 > remaining {
			return nil, fmt.Errorf("office expanded/member byte limit exceeded")
		}
		r, err := f.Open()
		if err != nil {
			return nil, err
		}
		data, err := io.ReadAll(io.LimitReader(r, int64(f.UncompressedSize64)+1))
		err = errors.Join(err, r.Close())
		if err != nil {
			return nil, fmt.Errorf("reading office member %s: %w", f.Name, err)
		}
		if uint64(len(data)) != f.UncompressedSize64 {
			return nil, fmt.Errorf("office member size mismatch: %s", f.Name)
		}
		remaining -= uint64(len(data))
		files[f.Name] = data
	}
	return files, nil
}
func activePart(name string) bool {
	lower := strings.ToLower(name)
	return strings.Contains(lower, "vbaproject") || strings.HasPrefix(lower, "basic/") || strings.HasPrefix(lower, "scripts/") || strings.Contains(lower, "/embeddings/")
}

func validateParts(files map[string][]byte, format string) error {
	required := []string{"[Content_Types].xml", "_rels/.rels", "word/document.xml"}
	if format == "odt" {
		required = []string{"mimetype", "content.xml", "styles.xml", "META-INF/manifest.xml"}
	} else if format != "docx" {
		return fmt.Errorf("unsupported office package format %s", format)
	}
	for _, name := range required {
		if _, ok := files[name]; !ok {
			return fmt.Errorf("missing office member %s", name)
		}
	}
	if format == "odt" && string(files["mimetype"]) != "application/vnd.oasis.opendocument.text" {
		return fmt.Errorf("invalid ODT MIME type")
	}
	for name, raw := range files {
		if strings.HasSuffix(name, ".xml") || strings.HasSuffix(name, ".rels") {
			if err := validatePartXML(name, raw, files, format); err != nil {
				return fmt.Errorf("invalid office XML %s: %w", name, err)
			}
		}
	}
	return nil
}
func validatePartXML(name string, raw []byte, files map[string][]byte, format string) error {
	if err := validText(string(raw)); err != nil {
		return err
	}
	d := xml.NewDecoder(bytes.NewReader(raw))
	depth, roots := 0, 0
	for {
		token, err := d.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		switch n := token.(type) {
		case xml.StartElement:
			if depth == 0 {
				roots++
			}
			depth++
			if depth > 512 {
				return fmt.Errorf("XML nesting limit exceeded")
			}
			if n.Name.Local == "script" || n.Name.Local == "event-listener" {
				return fmt.Errorf("active XML content prohibited")
			}
			if err := validateReference(name, n, files, format); err != nil {
				return err
			}
		case xml.EndElement:
			depth--
		case xml.Directive:
			return fmt.Errorf("XML declarations/entities prohibited")
		case xml.CharData:
			if depth == 0 && strings.TrimSpace(string(n)) != "" {
				return fmt.Errorf("text outside XML root")
			}
		}
	}
	if roots != 1 || depth != 0 {
		return fmt.Errorf("XML must have exactly one complete root")
	}
	return nil
}
func xmlAttr(n xml.StartElement, key string) string {
	for _, a := range n.Attr {
		if a.Name.Local == key {
			return a.Value
		}
	}
	return ""
}
func validateReference(member string, n xml.StartElement, files map[string][]byte, format string) error {
	ref, base := "", ""
	if format == "docx" && n.Name.Local == "Relationship" {
		if xmlAttr(n, "TargetMode") == "External" {
			if !strings.HasSuffix(xmlAttr(n, "Type"), "/hyperlink") {
				return fmt.Errorf("external non-hyperlink relationship prohibited")
			}
			return nil
		}
		ref = xmlAttr(n, "Target")
		base = path.Dir(path.Dir(member))
		if member == "_rels/.rels" {
			base = ""
		}
	} else if format == "odt" {
		switch n.Name.Local {
		case "image", "font-face-uri":
			ref = xmlAttr(n, "href")
		case "file-entry":
			ref = xmlAttr(n, "full-path")
			if ref == "/" || strings.HasSuffix(ref, "/") {
				return nil
			}
		}
	}
	if ref == "" {
		return nil
	}
	u, err := url.Parse(ref)
	if err != nil {
		return err
	}
	if u.IsAbs() || u.Host != "" {
		return fmt.Errorf("external package asset prohibited: %s", ref)
	}
	target := path.Clean(path.Join(base, u.Path))
	if strings.HasPrefix(u.Path, "/") {
		target = strings.TrimPrefix(path.Clean(u.Path), "/")
	}
	if _, ok := files[target]; !ok {
		return fmt.Errorf("broken internal reference %s -> %s", member, target)
	}
	return nil
}
