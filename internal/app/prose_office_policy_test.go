// ABOUTME: Checks prose policy switches and heading composition against native office output.
// ABOUTME: Keeps explicit expected text and property values independent of renderer helpers.
package app

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestRT063_4_OfficeParagraphPolicy(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		for _, tc := range []struct {
			name, yaml, wordElement, wordValue, odtAttribute, odtValue string
		}{
			{"justified", "justify: true", "jc", "both", "text-align", "justify"},
			{"ragged", "justify: false", "jc", "left", "text-align", "start"},
			{"hyphenation-on", "hyphenation: true", "suppressAutoHyphens", "false", "hyphenate", "true"},
			{"hyphenation-off", "hyphenation: false", "suppressAutoHyphens", "true", "hyphenate", "false"},
			{"hyphenation-auto-justified", "hyphenation: auto\n    justify: true", "suppressAutoHyphens", "false", "hyphenate", "true"},
			{"hyphenation-auto-ragged", "hyphenation: auto\n    justify: false", "suppressAutoHyphens", "true", "hyphenate", "false"},
			{"widow-orphan-on", "widow-orphan-control: true", "widowControl", "true", "widows", "2"},
			{"widow-orphan-off", "widow-orphan-control: false", "widowControl", "false", "widows", "0"},
		} {
			t.Run(format+"/"+tc.name, func(t *testing.T) {
				t.Setenv("HOME", t.TempDir())
				dir := t.TempDir()
				proseSource(t, dir, "script.yaml", "folio:\n  manuscript:\n    title-page:\n      enabled: false\n    toc:\n      enabled: false\n    "+tc.yaml+"\n")
				src := proseSource(t, dir, "source.md", "POLICY-MARKER: editable prose with independently controlled paragraph policy.\n\n```\nCODE-EXCLUDED\n```\n")
				pkg := proseConvert(t, src, filepath.Join(dir, "output."+format))
				layers := officeEffective(t, pkg, format, officeParagraph(t, officeBody(t, pkg, format), "POLICY-MARKER"))
				if format == "docx" {
					got, exists := officeEffectiveAttr(layers, tc.wordElement, "val")
					if tc.wordValue == "true" || tc.wordValue == "false" {
						// An empty on/off element means true. Resolve each layer
						// in order so a direct empty element overrides a style.
						for _, layer := range layers {
							for _, element := range layer.all(tc.wordElement) {
								got, exists = element.attr("val"), true
								if got == "" {
									got = "true"
								}
							}
						}
					}
					if got == "0" {
						got = "false"
					} else if got == "1" {
						got = "true"
					}
					if !exists || got != tc.wordValue {
						t.Errorf("%s=%q present=%v; want %s", tc.wordElement, got, exists, tc.wordValue)
					}
				} else {
					element := "paragraph-properties"
					if tc.odtAttribute == "hyphenate" {
						element = "text-properties"
					}
					got, exists := officeEffectiveAttr(layers, element, tc.odtAttribute)
					if tc.odtAttribute == "text-align" && got == "left" {
						got = "start" // These fixtures use left-to-right paragraphs.
					}
					if !exists || got != tc.odtValue {
						t.Errorf("%s=%q present=%v; want %s", tc.odtAttribute, got, exists, tc.odtValue)
					}
				}
				if strings.HasPrefix(tc.name, "hyphenation-") {
					if format == "docx" {
						officeAssertAttr(t, pkg["word/settings.xml"], "autoHyphenation", "val", tc.odtValue)
					}
					code := officeEffective(t, pkg, format, officeParagraph(t, officeBody(t, pkg, format), "CODE-EXCLUDED"))
					if format == "odt" {
						officeAssertEffective(t, code, "text-properties", "hyphenate", "false")
					} else {
						officeAssertEffective(t, code, "suppressAutoHyphens", "val", "true")
					}
				}
			})
		}
	}
}

func TestRT063_5_OfficeHeadingPresentation(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		for _, tc := range []struct{ name, override, want string }{
			{"derived", "explicit-numbering: derived", "Chapter 1: The Watch"},
			{"source", "explicit-numbering: source", "Chapter 12: The Watch"},
			{"roman-upper", "number-format: I", "Chapter I: The Watch"},
			{"roman-lower", "number-format: i", "Chapter i: The Watch"},
			{"hide-number", "show-number: false", "Chapter The Watch"},
			{"hide-name", "show-name: false", "Chapter 1"},
			{"upper-composed", "case-transform: upper", "CHAPTER 1: THE WATCH"},
			{"upper-name", "name-case: upper", "Chapter 1: THE WATCH"},
			{"suffix", "suffix: ' END'", "Chapter 1: The Watch END"},
		} {
			t.Run(format+"/"+tc.name, func(t *testing.T) {
				t.Setenv("HOME", t.TempDir())
				dir := t.TempDir()
				base := map[string]string{"prefix": "'Chapter '", "separator": "': '", "show-name": "true", "show-number": "true"}
				key, value, _ := strings.Cut(tc.override, ": ")
				base[key] = value
				cfg := "folio:\n  manuscript:\n    title-page:\n      enabled: false\n    toc:\n      enabled: false\n    chapter:\n"
				for k, v := range base {
					cfg += "      " + k + ": " + v + "\n"
				}
				proseSource(t, dir, "script.yaml", cfg)
				src := proseSource(t, dir, "source.md", "## Chapter 12: The Watch\n\nBODY-SENTINEL\n")
				pkg := proseConvert(t, src, filepath.Join(dir, "output."+format))
				p := officeParagraph(t, officeBody(t, pkg, format), tc.want)
				if strings.TrimSpace(p.text()) != tc.want {
					t.Errorf("heading %q; want %q", p.text(), tc.want)
				}
			})
		}
	}
}
