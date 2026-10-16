// ABOUTME: Checks the untrusted ZIP and XML boundary of editable prose packages.
// ABOUTME: Exercises exact resource thresholds and malformed native relationships without an office application.
package office

import (
	"archive/zip"
	"bytes"
	"fmt"
	"hash/crc32"
	"os"
	"strings"
	"testing"
)

type packageMember struct {
	name, body string
	mode       os.FileMode
}

func packageBytes(t *testing.T, members []packageMember) []byte {
	t.Helper()
	var output bytes.Buffer
	w := zip.NewWriter(&output)
	for _, member := range members {
		header := &zip.FileHeader{
			Name: member.name, Method: zip.Store,
			CRC32:            crc32.ChecksumIEEE([]byte(member.body)),
			CompressedSize64: uint64(len(member.body)), UncompressedSize64: uint64(len(member.body)),
		}
		header.SetMode(member.mode)
		// Stored bytes and known sizes keep ODT mimetype descriptor-free.
		entry, err := w.CreateRaw(header)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := entry.Write([]byte(member.body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return output.Bytes()
}

func validPackageMembers(format string) []packageMember {
	if format == "odt" {
		return []packageMember{
			{"mimetype", "application/vnd.oasis.opendocument.text", 0600},
			{"content.xml", `<office:document-content xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0" xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0"><office:body><office:text><text:p>Body</text:p></office:text></office:body></office:document-content>`, 0600},
			{"styles.xml", `<office:document-styles xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"><office:styles/></office:document-styles>`, 0600},
			{"META-INF/manifest.xml", `<manifest:manifest xmlns:manifest="urn:oasis:names:tc:opendocument:xmlns:manifest:1.0"><manifest:file-entry manifest:full-path="/" manifest:media-type="application/vnd.oasis.opendocument.text"/><manifest:file-entry manifest:full-path="content.xml" manifest:media-type="text/xml"/><manifest:file-entry manifest:full-path="styles.xml" manifest:media-type="text/xml"/></manifest:manifest>`, 0600},
		}
	}
	return []packageMember{
		{"[Content_Types].xml", `<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/></Types>`, 0600},
		{"_rels/.rels", `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/></Relationships>`, 0600},
		{"word/document.xml", `<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body><w:p><w:r><w:t>Body</w:t></w:r></w:p><w:sectPr/></w:body></w:document>`, 0600},
	}
}

func testPackageLimits() packageLimits {
	return packageLimits{compressed: 128 << 20, expanded: 256 << 20, member: 128 << 20, entries: 256}
}

func TestRT063_9_PackageRejectsUnsafeMembers(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		for _, tc := range []struct {
			name, path string
			mode       os.FileMode
		}{
			{"parent", "../outside.xml", 0600},
			{"absolute", "/outside.xml", 0600},
			{"backslash", `word\..\outside.xml`, 0600},
			{"dot-segment", "word/../outside.xml", 0600},
			{"symlink", "linked.xml", os.ModeSymlink | 0600},
		} {
			t.Run(format+"/"+tc.name, func(t *testing.T) {
				members := append(validPackageMembers(format), packageMember{tc.path, "<x/>", tc.mode})
				if err := validatePackage(packageBytes(t, members), format, testPackageLimits()); err == nil {
					t.Fatalf("unsafe package member %q accepted", tc.path)
				}
			})
		}
	}
}

func TestRT063_9_PackageIntegrity(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		t.Run(format, func(t *testing.T) {
			base := validPackageMembers(format)
			t.Run("valid", func(t *testing.T) {
				if err := validatePackage(packageBytes(t, base), format, testPackageLimits()); err != nil {
					t.Fatalf("valid package rejected: %v", err)
				}
			})
			t.Run("not-zip", func(t *testing.T) {
				if err := validatePackage([]byte("not a package"), format, testPackageLimits()); err == nil {
					t.Fatal("non-ZIP output accepted")
				}
			})
			t.Run("duplicate", func(t *testing.T) {
				members := append(append([]packageMember{}, base...), base[0])
				if err := validatePackage(packageBytes(t, members), format, testPackageLimits()); err == nil {
					t.Fatal("duplicate package member accepted")
				}
			})
			t.Run("checksum", func(t *testing.T) {
				raw := packageBytes(t, base)
				at := bytes.Index(raw, []byte(">Body<"))
				if at < 0 {
					t.Fatal("fixture lacks stored body bytes")
				}
				raw[at+1] = 'X'
				if err := validatePackage(raw, format, testPackageLimits()); err == nil {
					t.Fatal("member checksum mismatch accepted")
				}
			})
			for i, member := range base {
				t.Run("missing-"+member.name, func(t *testing.T) {
					members := append(append([]packageMember{}, base[:i]...), base[i+1:]...)
					if err := validatePackage(packageBytes(t, members), format, testPackageLimits()); err == nil {
						t.Fatalf("missing required member %s accepted", member.name)
					}
				})
			}
		})
	}
}

func TestRT063_9_PackageRejectsActivePayloads(t *testing.T) {
	for _, tc := range []struct{ format, path string }{
		{"docx", "word/vbaProject.bin"},
		{"docx", "word/embeddings/object.bin"},
		{"odt", "Basic/Standard/Module1.xml"},
		{"odt", "Scripts/python/script.py"},
	} {
		t.Run(tc.format+"/"+tc.path, func(t *testing.T) {
			members := append(validPackageMembers(tc.format), packageMember{tc.path, "<payload/>", 0600})
			if err := validatePackage(packageBytes(t, members), tc.format, testPackageLimits()); err == nil {
				t.Fatal("active or arbitrary embedded payload accepted")
			}
		})
	}
}

func TestRT063_9_PackageAllowsInertExternalHyperlink(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		t.Run(format, func(t *testing.T) {
			members := validPackageMembers(format)
			if format == "docx" {
				members[2].body = strings.Replace(members[2].body, "<w:p>", `<w:p xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><w:hyperlink r:id="rLink"><w:r><w:t>Link</w:t></w:r></w:hyperlink>`, 1)
				members = append(members, packageMember{"word/_rels/document.xml.rels", `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rLink" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/hyperlink" Target="https://example.invalid/not-fetched" TargetMode="External"/></Relationships>`, 0600})
			} else {
				members[1].body = strings.Replace(members[1].body, "Body", `<text:a xmlns:xlink="http://www.w3.org/1999/xlink" xlink:href="https://example.invalid/not-fetched">Link</text:a>`, 1)
			}
			if err := validatePackage(packageBytes(t, members), format, testPackageLimits()); err != nil {
				t.Fatalf("inert external link rejected: %v", err)
			}
		})
	}
}

func TestRT063_9_PackageXMLIntegrity(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		for _, tc := range []struct{ name, xml string }{
			{"unclosed", "<root>"}, {"multiple-roots", "<a/><b/>"},
			{"outside-text", "text<root/>"}, {"empty", ""},
			{"control", "<root>\x01</root>"}, {"invalid-utf8", "<root>\xff</root>"},
			{"entity", `<!DOCTYPE root [<!ENTITY x SYSTEM "file:///not-read">]><root>&x;</root>`},
		} {
			t.Run(format+"/"+tc.name, func(t *testing.T) {
				members := validPackageMembers(format)
				index := 2
				if format == "odt" {
					index = 1
				}
				members[index].body = tc.xml
				if err := validatePackage(packageBytes(t, members), format, testPackageLimits()); err == nil {
					t.Fatal("invalid XML accepted")
				}
			})
		}
	}
}

func TestRT063_9_PackageResourceThresholds(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		members := validPackageMembers(format)
		index := 2
		if format == "odt" {
			index = 1
		}
		members[index].body = strings.Replace(members[index].body, "Body", strings.Repeat("x", 4096), 1)
		raw := packageBytes(t, members)
		var expanded, largest uint64
		for _, member := range members {
			size := uint64(len(member.body))
			expanded += size
			if size > largest {
				largest = size
			}
		}
		// Small injected limits exercise the same owning boundary without
		// allocating hundreds of MiB for every boundary permutation.
		for _, bound := range []string{"compressed", "expanded", "member", "entries"} {
			for _, exceeds := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/%s/exceeds=%v", format, bound, exceeds), func(t *testing.T) {
					limits := testPackageLimits()
					delta := uint64(0)
					if exceeds {
						delta = 1
					}
					switch bound {
					case "compressed":
						limits.compressed = uint64(len(raw)) - delta
					case "expanded":
						limits.expanded = expanded - delta
					case "member":
						limits.member = largest - delta
					case "entries":
						limits.entries = len(members) - int(delta)
					}
					err := validatePackage(raw, format, limits)
					if (err != nil) != exceeds {
						t.Errorf("exceeds=%v: validation error=%v", exceeds, err)
					}
				})
			}
		}
	}
}

func TestRT063_9_PackageInternalReferences(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		t.Run(format, func(t *testing.T) {
			members := validPackageMembers(format)
			if format == "docx" {
				members = append(members, packageMember{"word/_rels/document.xml.rels", `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rIdImage" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/image" Target="media/missing.png"/></Relationships>`, 0600})
			} else {
				members[1].body = strings.Replace(members[1].body, "<text:p>Body</text:p>", `<text:p><draw:frame xmlns:draw="urn:oasis:names:tc:opendocument:xmlns:drawing:1.0"><draw:image xmlns:xlink="http://www.w3.org/1999/xlink" xlink:href="Pictures/missing.png"/></draw:frame></text:p>`, 1)
			}
			if err := validatePackage(packageBytes(t, members), format, testPackageLimits()); err == nil {
				t.Fatal("dangling internal image reference accepted")
			}
		})
	}
}
