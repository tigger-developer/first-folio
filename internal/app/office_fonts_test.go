// ABOUTME: Checks editable office font packages through the public conversion command.
// ABOUTME: Verifies full font bytes, face relationships and missing-font warnings.
package app

import (
	"archive/zip"
	"bytes"
	endian "encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	folio "github.com/tigger-developer/first-folio"
	"github.com/tigger-developer/first-folio/internal/office"
)

func officeFontMembers(t *testing.T, raw []byte) map[string][]byte {
	t.Helper()
	z, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		t.Fatal(err)
	}
	files := map[string][]byte{}
	for _, f := range z.File {
		r, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		b, err := io.ReadAll(r)
		closeErr := r.Close()
		if err != nil {
			t.Fatal(err)
		}
		if closeErr != nil {
			t.Fatal(closeErr)
		}
		files[f.Name] = b
	}
	return files
}

func TestOfficeEmbeddedFonts(t *testing.T) {
	binary := officeBuild(t)
	for _, format := range []string{"docx", "odt"} {
		t.Run(format, func(t *testing.T) {
			dir := t.TempDir()
			src := filepath.Join(dir, "play.org")
			writeAppFile(t, src, officeSources[".org"])
			raw, stderr, err := officeRun(t, binary, dir, officeEnv(t, t.TempDir(), map[string]string{"TYPST_FONT_PATHS": ""}), "convert", src, "--style=us", "--to", format)
			if err != nil {
				t.Fatalf("%v: %s", err, stderr)
			}
			files := officeFontMembers(t, raw)
			pkg := officePackage(t, raw)
			if format == "odt" {
				// ODF consumers identify the format from this local ZIP header.
				// Its stored mimetype must have known sizes, with no descriptor.
				if endian.LittleEndian.Uint16(raw[6:8])&8 != 0 || endian.LittleEndian.Uint16(raw[8:10]) != zip.Store || endian.LittleEndian.Uint32(raw[18:22]) != uint32(len(odtMIME)) || string(raw[30:38]) != "mimetype" {
					t.Fatal("ODT mimetype local header cannot be read without a data descriptor")
				}
			}
			for _, face := range []struct{ file, tag string }{{"Regular", "embedRegular"}, {"Bold", "embedBold"}, {"Italic", "embedItalic"}, {"BoldItalic", "embedBoldItalic"}} {
				want, err := folio.Assets.ReadFile("fonts/iosevka-publish/IosevkaPublish-" + face.file + ".ttf")
				if err != nil {
					t.Fatal(err)
				}
				found := false
				if format == "docx" {
					if pkg["word/fontTable.xml"] == nil {
						t.Fatal("missing embedded font table")
					}
					for _, font := range pkg["word/fontTable.xml"].all("font") {
						if font.attr("name") != "Iosevka Publish" {
							continue
						}
						for _, ref := range font.all(face.tag) {
							key, err := hex.DecodeString(strings.NewReplacer("{", "", "}", "", "-", "").Replace(ref.attr("fontKey")))
							if err != nil || len(key) != 16 {
								t.Fatal("invalid font obfuscation key")
							}
							if ref.attr("subsetted") != "false" {
								t.Fatal("font is not declared full")
							}
							for _, rel := range pkg["word/_rels/fontTable.xml.rels"].all("Relationship") {
								if rel.attr("Id") != ref.attr("id") {
									continue
								}
								got := bytes.Clone(files["word/"+rel.attr("Target")])
								if len(got) < 32 {
									t.Fatal("missing font relationship target")
								}
								for i := 0; i < 32; i++ {
									got[i] ^= key[15-i%16]
								}
								found = bytes.Equal(got, want)
							}
						}
					}
				} else {
					for _, file := range []string{"styles.xml", "content.xml"} {
						for _, ref := range pkg[file].all("font-face-uri") {
							name := ref.attr("href")
							if bytes.Equal(files[name], want) {
								officeAssertAttr(t, pkg["META-INF/manifest.xml"], "file-entry", "full-path", name)
								found = true
							}
						}
					}
				}
				if !found {
					t.Errorf("missing complete %s face", face.file)
				}
			}
			if format == "docx" {
				officeAssertAttr(t, pkg["word/_rels/document.xml.rels"], "Relationship", "Target", "fontTable.xml")
				officeAssertAttr(t, pkg["[Content_Types].xml"], "Override", "ContentType", "application/vnd.openxmlformats-officedocument.wordprocessingml.fontTable+xml")
				officeAssertAttr(t, pkg["word/settings.xml"], "saveSubsetFonts", "val", "false")
			}
			writeAppFile(t, filepath.Join(dir, "script.yaml"), "folio:\n  positioning:\n    speech:\n      dialogue:\n        font:\n          family: Folio Missing Font 92831\n")
			_, stderr, err = officeRun(t, binary, dir, officeEnv(t, t.TempDir(), nil), "convert", src, "--style=us", "--to", format)
			if err != nil {
				t.Fatalf("missing font prevents conversion: %v: %s", err, stderr)
			}
			if !strings.Contains(stderr, "Folio Missing Font 92831") || !strings.Contains(stderr, "not embedded") {
				t.Fatalf("missing font warning absent: %s", stderr)
			}
		})
	}
}

func officeTestFont(t *testing.T, rights uint16) []byte {
	t.Helper()
	raw, err := folio.Assets.ReadFile("fonts/iosevka-publish/IosevkaPublish-Regular.ttf")
	if err != nil {
		t.Fatal(err)
	}
	// A temporary renamed derivative distinguishes local discovery from bundled
	// resolution. This metadata fixture is never installed or shipped.
	raw = bytes.ReplaceAll(raw, []byte("\x00I\x00o\x00s\x00e\x00v\x00k\x00a"), []byte("\x00F\x00i\x00x\x00t\x00u\x00r\x00e"))
	for i := 0; i < int(endian.BigEndian.Uint16(raw[4:6])); i++ {
		r := raw[12+i*16 : 28+i*16]
		if string(r[:4]) == "OS/2" {
			off := int(endian.BigEndian.Uint32(r[8:12]))
			endian.BigEndian.PutUint16(raw[off+8:off+10], rights)
		}
	}
	return raw
}

func TestOfficeFontPermissionsAndBounds(t *testing.T) {
	for _, tc := range []struct {
		name   string
		rights uint16
		want   bool
	}{{"installable", 0, true}, {"editable", 8, true}, {"full-only", 0x108, true}, {"restricted", 2, false}, {"preview-only", 4, false}, {"bitmap-only", 0x208, false}} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			raw := officeTestFont(t, tc.rights)
			if err := os.WriteFile(filepath.Join(dir, "font.ttf"), raw, 0600); err != nil {
				t.Fatal(err)
			}
			fonts, warnings := office.ResolveFonts([]string{"Fixture Publish"}, []string{dir})
			if (len(fonts) == 1) != tc.want {
				t.Fatalf("embedded=%d warnings=%v", len(fonts), warnings)
			}
			if tc.want && !bytes.Equal(fonts[0].Raw, raw) {
				t.Fatal("font was subsetted or changed")
			}
			if !tc.want && !strings.Contains(strings.Join(warnings, " "), "prohibits editable embedding") {
				t.Fatalf("missing permission warning: %v", warnings)
			}
		})
	}
	raw := officeTestFont(t, 0)
	for _, size := range []int{0, 4, 11, 12, 100, len(raw) / 2, len(raw) - 1} {
		if _, err := office.ParseFont(raw[:size]); err == nil {
			t.Errorf("accepted truncated font of %d bytes", size)
		}
	}
	broken := bytes.Clone(raw)
	endian.BigEndian.PutUint32(broken[20:24], 0xffffffff)
	if _, err := office.ParseFont(broken); err == nil {
		t.Fatal("accepted out-of-bounds table")
	}
}

func TestOfficeLocalFontOverride(t *testing.T) {
	binary := officeBuild(t)
	dir := t.TempDir()
	fontDir := t.TempDir()
	want := officeTestFont(t, 8)
	if err := os.WriteFile(filepath.Join(fontDir, "custom.ttf"), want, 0600); err != nil {
		t.Fatal(err)
	}
	src := filepath.Join(dir, "play.org")
	writeAppFile(t, src, officeSources[".org"])
	writeAppFile(t, filepath.Join(dir, "script.yaml"), "folio:\n  positioning:\n    speech:\n      dialogue:\n        font:\n          family: Fixture Publish\n")
	for _, format := range []string{"docx", "odt"} {
		raw, stderr, err := officeRun(t, binary, dir, officeEnv(t, t.TempDir(), map[string]string{"TYPST_FONT_PATHS": fontDir}), "convert", src, "--style=us", "--to", format)
		if err != nil {
			t.Fatalf("%v: %s", err, stderr)
		}
		found := false
		for name, data := range officeFontMembers(t, raw) {
			if format == "odt" && strings.HasPrefix(name, "Fonts/") && bytes.Equal(data, want) {
				found = true
			}
			// DOCX obfuscates exactly the first 32 bytes. The bundled-face test
			// independently checks keys and deobfuscation for every face.
			if format == "docx" && strings.HasSuffix(name, ".odttf") && len(data) == len(want) && bytes.Equal(data[32:], want[32:]) {
				found = true
			}
		}
		if !found {
			t.Fatalf("%s omitted local configured font: %s", format, stderr)
		}
	}
}

func TestOfficeFontDiscoverySkipsFontBodies(t *testing.T) {
	dir := t.TempDir()
	for i := 0; i < 17; i++ {
		f, err := os.Create(filepath.Join(dir, fmt.Sprintf("a%02d.ttf", i)))
		if err != nil {
			t.Fatal(err)
		}
		err = f.Truncate(32 << 20)
		closeErr := f.Close()
		if err != nil {
			t.Fatal(err)
		}
		if closeErr != nil {
			t.Fatal(closeErr)
		}
	}
	want := officeTestFont(t, 0)
	if err := os.WriteFile(filepath.Join(dir, "z-selected.ttf"), want, 0600); err != nil {
		t.Fatal(err)
	}
	fonts, warnings := office.ResolveFonts([]string{"Fixture Publish"}, []string{dir})
	if len(fonts) != 1 || !bytes.Equal(fonts[0].Raw, want) {
		t.Fatalf("unrelated font bodies hid configured font: %v", warnings)
	}
}
