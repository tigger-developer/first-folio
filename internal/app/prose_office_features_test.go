// ABOUTME: Proves generated prose matter, embedded fonts, and office-only failure behaviour.
// ABOUTME: Uses exported artefacts and public errors rather than inspecting implementation source.
package app

import (
	"bytes"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	folio "github.com/tigger-developer/first-folio"
)

func TestRT063_5_OfficeMissingOptionalMatter(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		for _, content := range []string{"", "PLAIN-BODY\n", "---\ntitle: METADATA-ONLY\n---\n"} {
			t.Run(format+"/"+strings.TrimSpace(content), func(t *testing.T) {
				t.Setenv("HOME", t.TempDir())
				dir := t.TempDir()
				proseSource(t, dir, "script.yaml", "folio:\n  manuscript:\n    title-page:\n      enabled: false\n    toc:\n      enabled: false\n    copyright:\n      enabled: false\n")
				src := proseSource(t, dir, "source.md", content)
				pkg := proseConvert(t, src, filepath.Join(dir, "output."+format))
				body := officeBody(t, pkg, format)
				want := ""
				if content == "PLAIN-BODY\n" {
					want = "PLAIN-BODY"
				}
				if strings.TrimSpace(body.text()) != want {
					t.Errorf("optional matter introduced visible content: %q", body.text())
				}
				if proseNativeBreakCount(t, pkg, format) > 1 {
					t.Error("absent generated matter introduced phantom page transitions")
				}
			})
		}
	}
}

func TestRT063_4_OfficeConfiguredSceneMarker(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		t.Run(format, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			dir := t.TempDir()
			proseSource(t, dir, "script.yaml", "folio:\n  manuscript:\n    title-page:\n      enabled: false\n    toc:\n      enabled: false\n    scene-break:\n      marker: SCENE-CUSTOM\n")
			src := proseSource(t, dir, "source.md", "FIRST-SCENE\n\n---\n\nSECOND-SCENE\n")
			pkg := proseConvert(t, src, filepath.Join(dir, "output."+format))
			if strings.Count(officeBody(t, pkg, format).text(), "SCENE-CUSTOM") != 1 {
				t.Error("configured scene marker missing or duplicated")
			}
			proseAssertAlignment(t, pkg, format, "SCENE-CUSTOM", "center")
		})
	}
}

func TestRT063_5_OfficeGeneratedTitleAndContents(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		for _, enabled := range []bool{false, true} {
			label := "disabled"
			if enabled {
				label = "enabled"
			}
			t.Run(format+"/"+label, func(t *testing.T) {
				t.Setenv("HOME", t.TempDir())
				dir := t.TempDir()
				value := "false"
				if enabled {
					value = "true"
				}
				proseSource(t, dir, "script.yaml", "folio:\n  manuscript:\n    title-page:\n      enabled: "+value+"\n      include-title: true\n      include-author: true\n    toc:\n      enabled: "+value+"\n      title: DISTINCT-CONTENTS-TITLE\n      include-chapters: true\n      page-numbers: true\n      links: true\n")
				src := proseSource(t, dir, "source.md", "---\ntitle: DISTINCT-MANUSCRIPT-TITLE\nauthor: DISTINCT-AUTHOR\n---\n\n## First chapter\n\nFIRST-CHAPTER-BODY\n\n## Second chapter\n\nLAST-CHAPTER-BODY\n")
				pkg := proseConvert(t, src, filepath.Join(dir, "output."+format))
				body := officeBody(t, pkg, format)
				for _, text := range []string{"DISTINCT-MANUSCRIPT-TITLE", "DISTINCT-AUTHOR", "DISTINCT-CONTENTS-TITLE"} {
					if strings.Contains(body.text(), text) != enabled {
						t.Errorf("generated text %q visibility does not match enabled=%v", text, enabled)
					}
				}
				live := proseContentsHasPageFields(body, format)
				if live != enabled {
					t.Errorf("native updateable contents=%v, want %v", live, enabled)
				}
			})
		}
	}
}

func TestRT063_6_OfficeRunningMatterVariants(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		t.Run(format, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			dir := t.TempDir()
			proseSource(t, dir, "script.yaml", "folio:\n  page-header:\n    enabled: true\n    format: BODY-EVEN [page]\n    alt-format: BODY-ODD [page]\n    frontmatter-format: FRONT-EVEN [page]\n    alt-frontmatter-format: FRONT-ODD [page]\n  manuscript:\n    title-page:\n      enabled: true\n    toc:\n      enabled: true\n    chapter:\n      skip-header: false\n      skip-footer: false\n    page-numbering:\n      frontmatter-format: i\n      body-format: '1'\n      body-reset: first-part-or-chapter\n")
			src := proseSource(t, dir, "source.md", "---\ntitle: Page policy\n---\n\n## Chapter One\n\nBODY-CONTENT\n")
			pkg := proseConvert(t, src, filepath.Join(dir, "output."+format))
			var text strings.Builder
			fields := 0
			for _, member := range pkg {
				text.WriteString(member.text())
				if format == "odt" {
					fields += len(member.all("page-number")) + len(member.all("page-variable-get"))
				} else {
					for _, f := range member.all("instrText") {
						if strings.TrimSpace(f.text()) == "PAGE" {
							fields++
						}
					}
					for _, f := range member.all("fldSimple") {
						if strings.TrimSpace(f.attr("instr")) == "PAGE" {
							fields++
						}
					}
				}
			}
			for _, marker := range []string{"BODY-EVEN", "BODY-ODD", "FRONT-EVEN", "FRONT-ODD"} {
				if !strings.Contains(text.String(), marker) {
					t.Errorf("missing running-matter variant %s", marker)
				}
			}
			if fields == 0 {
				t.Error("page numbers are not native live fields")
			}
		})
	}
}

func TestRT063_8_OfficeProseEmbedsFullFonts(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		for _, mode := range []string{"manuscript", "letter"} {
			t.Run(format+"/"+mode, func(t *testing.T) {
				t.Setenv("HOME", t.TempDir())
				dir := t.TempDir()
				proseSource(t, dir, "script.yaml", "folio:\n  "+mode+":\n    font:\n      family: Iosevka Publish\n")
				target := filepath.Join(dir, "output."+format)
				var pkg map[string]*officeXML
				if mode == "manuscript" {
					src := proseSource(t, dir, "source.md", "## Chapter One\n\nFONT-CONTENT\n")
					pkg = proseConvert(t, src, target)
				} else {
					raw, err := os.ReadFile("../../testdata/prose-office/letters.org")
					if err != nil {
						t.Fatal(err)
					}
					src := proseSource(t, dir, "letters.org", string(raw))
					status, _, stderr := runApp(t, "letter", src, "--format", format, "--to", "First")
					if status != 0 {
						t.Fatalf("letter export: %s", stderr)
					}
					target = filepath.Join(dir, "letter-first-theatre."+format)
					pkg = prosePackage(t, target, format)
				}
				raw, err := os.ReadFile(target)
				if err != nil {
					t.Fatal(err)
				}
				members := officeFontMembers(t, raw)
				for _, faceSpec := range []struct{ file, tag string }{{"Regular", "embedRegular"}, {"Bold", "embedBold"}, {"Italic", "embedItalic"}, {"BoldItalic", "embedBoldItalic"}} {
					want, err := folio.Assets.ReadFile("fonts/iosevka-publish/IosevkaPublish-" + faceSpec.file + ".ttf")
					if err != nil {
						t.Fatal(err)
					}
					found := false
					if format == "odt" {
						for _, file := range []string{"styles.xml", "content.xml"} {
							for _, ref := range pkg[file].all("font-face-uri") {
								if bytes.Equal(members[ref.attr("href")], want) {
									found = true
									manifest := pkg["META-INF/manifest.xml"]
									if manifest == nil {
										t.Fatal("missing font manifest")
									}
									officeAssertAttr(t, manifest, "file-entry", "full-path", ref.attr("href"))
								}
							}
						}
					} else {
						fonts, rels := pkg["word/fontTable.xml"], pkg["word/_rels/fontTable.xml.rels"]
						if fonts == nil || rels == nil {
							t.Fatal("font table or relationships absent")
						}
						for _, font := range fonts.all("font") {
							if font.attr("name") != "Iosevka Publish" {
								continue
							}
							for _, ref := range font.all(faceSpec.tag) {
								key, err := hex.DecodeString(strings.NewReplacer("{", "", "}", "", "-", "").Replace(ref.attr("fontKey")))
								if err != nil || len(key) != 16 {
									t.Fatal("invalid embedded font key")
								}
								for _, rel := range rels.all("Relationship") {
									if rel.attr("Id") != ref.attr("id") {
										continue
									}
									face := bytes.Clone(members["word/"+rel.attr("Target")])
									if len(face) < 32 {
										t.Fatal("missing embedded font payload")
									}
									for i := range 32 {
										face[i] ^= key[15-i%16]
									}
									found = bytes.Equal(face, want)
								}
							}
						}
					}
					if !found {
						t.Errorf("configured complete %s font face absent or modified", faceSpec.file)
					}
				}
			})
		}
	}
}

func TestRT063_9_OfficeLetterCollisionPreflight(t *testing.T) {
	raw, err := os.ReadFile("../../testdata/prose-office/letters.org")
	if err != nil {
		t.Fatal(err)
	}
	for _, format := range []string{"docx", "odt"} {
		t.Run(format, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			dir := t.TempDir()
			src := proseSource(t, dir, "letters.org", strings.ReplaceAll(string(raw), "Second Theatre", "First Theatre"))
			out := t.TempDir()
			status, _, stderr := runApp(t, "letter", src, "--format", format, "--dir", out)
			if status == 0 || !strings.Contains(strings.ToLower(stderr), "collision") {
				t.Errorf("collision must be diagnosed before publication: status=%d stderr=%s", status, stderr)
			}
			files, err := os.ReadDir(out)
			if err != nil || len(files) != 0 {
				t.Errorf("collision left outputs: %v, %v", files, err)
			}
		})
	}
}

func TestRT063_5_OfficeTitleFieldSwitches(t *testing.T) {
	metadata := "---\ntitle: TITLE-MARKER\nsubtitle: SUBTITLE-MARKER\nauthor: AUTHOR-MARKER\ndate: DATE-MARKER\nwordcount: WORDCOUNT-MARKER\ncontact-name: CONTACT-MARKER\naddress: ADDRESS-MARKER\nphone: PHONE-MARKER\nemail: EMAIL-MARKER\nwebsite: WEBSITE-MARKER\nversion: VERSION-MARKER\n---\n\nBODY-MARKER\n"
	for _, format := range []string{"docx", "odt"} {
		for _, tc := range []struct{ key, marker string }{
			{"title", "TITLE-MARKER"}, {"subtitle", "SUBTITLE-MARKER"}, {"author", "AUTHOR-MARKER"},
			{"date", "DATE-MARKER"}, {"wordcount", "WORDCOUNT-MARKER"}, {"contact-name", "CONTACT-MARKER"},
			{"address", "ADDRESS-MARKER"}, {"phone", "PHONE-MARKER"}, {"email", "EMAIL-MARKER"},
			{"website", "WEBSITE-MARKER"}, {"version", "VERSION-MARKER"},
		} {
			t.Run(format+"/include-"+tc.key, func(t *testing.T) {
				t.Setenv("HOME", t.TempDir())
				dir := t.TempDir()
				src := proseSource(t, dir, "source.md", metadata)
				for _, value := range []string{"true", "false"} {
					cfg := "folio:\n  page-header:\n    enabled: false\n  page-footer:\n    enabled: false\n  manuscript:\n    toc:\n      enabled: false\n    title-page:\n      enabled: true\n      include-" + tc.key + ": " + value + "\n"
					proseSource(t, dir, "script.yaml", cfg)
					pkg := proseConvert(t, src, filepath.Join(dir, "output."+format))
					// Match whole marker tokens: TITLE-MARKER is a suffix of SUBTITLE-MARKER.
					present := false
					for _, member := range officeBody(t, pkg, format).all("p") {
						for _, word := range strings.Fields(member.text()) {
							if word == tc.marker {
								present = true
							}
						}
					}
					if present != (value == "true") {
						t.Errorf("include-%s=%s: %s present=%v", tc.key, value, tc.marker, present)
					}
				}
			})
		}
	}
}

func TestRT063_5_OfficeCopyrightBarcodeModes(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		for _, mode := range []string{"none", "render", "file", "render-and-file"} {
			t.Run(format+"/"+mode, func(t *testing.T) {
				t.Setenv("HOME", t.TempDir())
				dir := t.TempDir()
				cfg := "folio:\n  manuscript:\n    title-page:\n      enabled: false\n    toc:\n      enabled: false\n    copyright:\n      enabled: true\n      credits: ['COPYRIGHT-CREDIT']\n      body: ['COPYRIGHT-BODY with **bold credit** and *italic credit*.']\n      publication: ['PUBLICATION-MARKER']\n      publisher: PUBLISHER-MARKER\n      isbn: '978-0-000000-00-2'\n      isbn-barcode: " + mode + "\n"
				proseSource(t, dir, "script.yaml", cfg)
				src := proseSource(t, dir, "source.org", "#+TITLE: Barcode book\n** Chapter One\nBODY-SENTINEL\n")
				target := filepath.Join(dir, "output."+format)
				pkg := proseConvert(t, src, target)
				body := officeBody(t, pkg, format)
				for _, text := range []string{"COPYRIGHT-CREDIT", "COPYRIGHT-BODY", "PUBLICATION-MARKER", "PUBLISHER-MARKER", "978-0-000000-00-2"} {
					if !strings.Contains(body.text(), text) {
						t.Errorf("missing copyright content %q", text)
					}
				}
				if strings.Contains(body.text(), "**bold credit**") || strings.Contains(body.text(), "*italic credit*") {
					t.Error("configuration mini-Markdown was not interpreted independently of Org source")
				}
				image := "drawing"
				if format == "odt" {
					image = "image"
				}
				if (len(body.all(image)) > 0) != (mode == "render" || mode == "render-and-file") {
					t.Errorf("embedded barcode does not match mode %s", mode)
				}
				data, err := os.ReadFile(filepath.Join(dir, "output.barcode.svg"))
				if mode == "file" || mode == "render-and-file" {
					if err != nil {
						t.Fatal(err)
					}
					svg := officeParseXML(t, data)
					info, err := os.Stat(filepath.Join(dir, "output.barcode.svg"))
					if err != nil || info.Mode().Perm() != 0600 {
						t.Errorf("barcode sidecar is not private: %v", err)
					}
					if len(svg.all("rect")) < 20 || !strings.Contains(strings.Join(strings.Fields(svg.text()), ""), "9780000000002") {
						t.Error("sidecar lacks barcode bars or expected ISBN digits")
					}
				} else if !os.IsNotExist(err) {
					t.Errorf("unexpected barcode sidecar for %s: %v", mode, err)
				}
			})
		}
	}
}
