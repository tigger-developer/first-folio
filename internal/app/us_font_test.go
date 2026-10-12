// ABOUTME: Verifies bundled monospace fonts and US cue geometry in real PDFs.
// ABOUTME: Exercises a copied CLI with system fonts disabled and isolated configuration.
package app

import (
	"encoding/xml"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tigger-developer/first-folio/internal/config"
)

func TestRT047_1BundledUSFontsAndCueAlignment(t *testing.T) {
	for _, tool := range []string{"typst", "pdffonts", "pdftotext"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("requires %s", tool)
		}
	}
	cfg, err := config.Load(config.Options{Mode: config.ModeScript, Home: t.TempDir(), LocalDir: t.TempDir(), CLI: map[string]any{"style": "us"}})
	if err != nil {
		t.Fatal(err)
	}
	font, err := cfg.Font("folio.positioning.speech.speaker.font")
	if err != nil {
		t.Fatal(err)
	}
	if font.Weight != "regular" {
		t.Errorf("speaker weight = %s, want regular", font.Weight)
	}
	binary := officeBuild(t)
	for _, spelling := range []string{"equals", "separate", "override", "british", "manuscript"} {
		t.Run(spelling, func(t *testing.T) {
			dir, scratch := t.TempDir(), t.TempDir()
			env := officeEnv(t, dir, map[string]string{"TYPST_IGNORE_SYSTEM_FONTS": "true", "TYPST_FONT_PATHS": dir, "TMPDIR": scratch})
			source, pdf := filepath.Join(dir, "play.org"), filepath.Join(dir, "play.pdf")
			writeAppFile(t, source, "* Act One\n** Scene One\n*** Directionstart "+strings.Repeat("directionword ", 35)+"directionend.\n**** CAIT\nDialoguestart "+strings.Repeat("dialogueword ", 35)+"dialogueend.\n**** ALEXANDER\nHello there.\n")
			args := []string{"convert", source, pdf, "--style=us"}
			switch spelling {
			case "separate":
				args = []string{"convert", source, pdf, "--style", "us"}
			case "override":
				writeAppFile(t, filepath.Join(dir, "script.yaml"), "folio:\n  positioning:\n    speech:\n      dialogue:\n        font:\n          family: Libertinus Serif\n")
			case "british":
				args = []string{"convert", source, pdf, "--style=british"}
			case "manuscript":
				source = filepath.Join(dir, "chapter.md")
				writeAppFile(t, source, "# Chapter One\n\nA `monospaced` word.\n\n```\nblock code\n```\n")
				args = []string{"manuscript", source, pdf}
			}
			if _, stderr, err := officeRun(t, binary, dir, env, args...); err != nil {
				t.Fatalf("render: %v: %s", err, stderr)
			}
			fonts, err := exec.Command("pdffonts", pdf).CombinedOutput()
			if err != nil {
				t.Fatalf("fonts: %v: %s", err, fonts)
			}
			if spelling == "british" {
				if !strings.Contains(string(fonts), "LibertinusSerif") || strings.Contains(string(fonts), "Iosevka-Publish") {
					t.Errorf("British fonts changed: %s", fonts)
				}
			} else {
				if !strings.Contains(string(fonts), "Iosevka-Publish") {
					t.Errorf("bundled font absent: %s", fonts)
				}
				if spelling == "equals" || spelling == "separate" {
					for _, face := range []string{"Iosevka-Publish ", "Iosevka-Publish-Bold", "Iosevka-Publish-Italic"} {
						if !strings.Contains(string(fonts), face) {
							t.Errorf("missing %s: %s", face, fonts)
						}
					}
				}
				if spelling == "override" && !strings.Contains(string(fonts), "LibertinusSerif") {
					t.Errorf("explicit font override lost: %s", fonts)
				}
			}
			if spelling != "manuscript" {
				raw, err := exec.Command("pdftotext", "-bbox", pdf, "-").CombinedOutput()
				if err != nil {
					t.Fatalf("bbox: %v: %s", err, raw)
				}
				var doc struct {
					Pages []struct {
						Width float64 `xml:"width,attr"`
						Words []struct {
							Text  string  `xml:",chardata"`
							Left  float64 `xml:"xMin,attr"`
							Right float64 `xml:"xMax,attr"`
							Top   float64 `xml:"yMin,attr"`
						} `xml:"word"`
					} `xml:"body>doc>page"`
				}
				if err := xml.Unmarshal(raw, &doc); err != nil {
					t.Fatal(err)
				}
				found := false
				for _, page := range doc.Pages {
					previousY := -1.0
					for _, word := range page.Words {
						firstOnLine := math.Abs(word.Top-previousY) > 1
						previousY = word.Top
						if spelling != "british" {
							const margin = 25 * 72 / 25.4
							if word.Text == "Dialoguestart" || word.Text == "dialogueword" {
								// Check the first word on every dialogue line, including wrapped lines.
								if word.Text == "Dialoguestart" || firstOnLine {
									if math.Abs(word.Left-margin) > 1 {
										t.Errorf("dialogue left = %.2f", word.Left)
									}
								}
							}
							if strings.Contains(word.Text, "Directionstart") {
								if word.Text != "(Directionstart" {
									t.Errorf("missing opening parenthesis: %s", word.Text)
								}
								want := margin + 0.4*(page.Width-2*margin)
								if math.Abs(word.Left-want) > 1 {
									t.Errorf("direction left = %.2f, want %.2f", word.Left, want)
								}
							}
							if strings.Contains(word.Text, "directionend.") && word.Text != "directionend.)" {
								t.Errorf("missing closing parenthesis: %s", word.Text)
							}
						}

						if name := strings.TrimSuffix(word.Text, ":"); name == "CAIT" || name == "ALEXANDER" {
							found = true
							centred := math.Abs(word.Left-page.Width/2) < 1
							if centred != (spelling != "british") {
								t.Errorf("cue left %.2f, page centre %.2f", word.Left, page.Width/2)
							}
						}
					}
				}
				if !found {
					t.Fatal("no character cue in PDF")
				}
			}
			entries, err := os.ReadDir(scratch)
			if err != nil {
				t.Fatal(err)
			}
			if len(entries) != 0 {
				t.Errorf("temporary compilation files remain: %v", entries)
			}
		})
	}
}
