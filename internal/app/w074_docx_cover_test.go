// ABOUTME: Defines retained public DOCX cover source-cascade and package regressions.
// ABOUTME: Compares local YAML, Markdown YAML and dotted Org through native output semantics.
package app

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"fmt"
	"image/color"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

const w074CoverYAML = `folio:
  title-page:
    title:
      position: 40%
    subtitle:
      space-before: 4mm
    author:
      position: bottom-center
  cover-image:
    path: cover.png
    mode: inset
    placement: dodge
    size: 80mm
    dodge:
      padding: 5mm
    backdrop:
      mode: image
      image:
        blur: 8mm
      wash:
        treatment: dark
        opacity: 15%
`

const w074ConflictingYAML = `folio:
  title-page:
    title:
      position: 10%
    subtitle:
      space-before: 1em
    author:
      position: group
  cover-image:
    path: cover.png
    mode: inset
    placement: overlay
    size: 20mm
    backdrop:
      mode: none
`

func w074DOCXFingerprint(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	z, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		t.Fatal(err)
	}
	parts := map[string][]byte{}
	var media []string
	for _, file := range z.File {
		if file.Name != "word/document.xml" && file.Name != "word/_rels/document.xml.rels" && file.Name != "[Content_Types].xml" && !strings.HasPrefix(file.Name, "word/media/") {
			continue
		}
		reader, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		data, readErr := io.ReadAll(reader)
		closeErr := reader.Close()
		if readErr != nil || closeErr != nil {
			t.Fatalf("read %s: %v / %v", file.Name, readErr, closeErr)
		}
		if strings.HasPrefix(file.Name, "word/media/") {
			sum := sha256.Sum256(data)
			media = append(media, fmt.Sprintf("%s:%x", file.Name, sum))
		} else {
			parts[file.Name] = data
		}
	}
	sort.Strings(media)
	return string(parts["[Content_Types].xml"]) + "\n" + string(parts["word/_rels/document.xml.rels"]) + "\n" + string(parts["word/document.xml"]) + "\n" + strings.Join(media, "\n")
}

func TestRT074_3DOCXSourceConfigurationResolvesIdentically(t *testing.T) {
	type fixture struct {
		name, extension, source, local string
	}
	fixtures := []fixture{
		{
			name: "local yaml", extension: ".org", local: w074CoverYAML,
			source: "#+TITLE: W074 TITLE\n#+SUBTITLE: W074 SUBTITLE\n#+AUTHOR: W074 AUTHOR\n#+DATE: W074 DATE\n#+VERSION: W074 VERSION\n* ACT ONE\n*** BODY MARKER\n",
		},
		{
			name: "markdown frontmatter", extension: ".md", local: w074ConflictingYAML,
			source: `---
schema: https://github.com/tigger-developer/first-folio/blob/master/schema/script.md
title: W074 TITLE
subtitle: W074 SUBTITLE
author: W074 AUTHOR
date: W074 DATE
version: W074 VERSION
folio:
  title-page:
    title:
      position: 40%
    subtitle:
      space-before: 4mm
    author:
      position: bottom-center
  cover-image:
    path: cover.png
    mode: inset
    placement: dodge
    size: 80mm
    dodge:
      padding: 5mm
    backdrop:
      mode: image
      image:
        blur: 8mm
      wash:
        treatment: dark
        opacity: 15%
---
## Act One
*BODY MARKER*
`,
		},
		{
			name: "dotted org", extension: ".org", local: w074ConflictingYAML,
			source: `#+TITLE: W074 TITLE
#+SUBTITLE: W074 SUBTITLE
#+AUTHOR: W074 AUTHOR
#+DATE: W074 DATE
#+VERSION: W074 VERSION
#+FOLIO.TITLE-PAGE.TITLE.POSITION: 40%
#+FOLIO.TITLE-PAGE.SUBTITLE.SPACE-BEFORE: 4mm
#+FOLIO.TITLE-PAGE.AUTHOR.POSITION: bottom-center
#+FOLIO.COVER-IMAGE.PATH: cover.png
#+FOLIO.COVER-IMAGE.MODE: inset
#+FOLIO.COVER-IMAGE.PLACEMENT: dodge
#+FOLIO.COVER-IMAGE.SIZE: 80mm
#+FOLIO.COVER-IMAGE.DODGE.PADDING: 5mm
#+FOLIO.COVER-IMAGE.BACKDROP.MODE: image
#+FOLIO.COVER-IMAGE.BACKDROP.IMAGE.BLUR: 8mm
#+FOLIO.COVER-IMAGE.BACKDROP.WASH.TREATMENT: dark
#+FOLIO.COVER-IMAGE.BACKDROP.WASH.OPACITY: 15%
* ACT ONE
*** BODY MARKER
`,
		},
	}

	var want string
	for _, fixture := range fixtures {
		t.Run(fixture.name, func(t *testing.T) {
			dir := t.TempDir()
			source := filepath.Join(dir, "play"+fixture.extension)
			target := filepath.Join(dir, "play.docx")
			writeAppFile(t, source, fixture.source)
			writeAppFile(t, filepath.Join(dir, "script.yaml"), fixture.local)
			writeW073PNG(t, filepath.Join(dir, "cover.png"), 400, 200, color.RGBA{R: 215, G: 40, B: 120, A: 255})
			status, _, stderr := runApp(t, "convert", source, target)
			if status != 0 {
				t.Fatalf("convert status=%d: %s", status, stderr)
			}
			if strings.Contains(stderr, "unrecognized FOLIO") || strings.Contains(stderr, "not supported") {
				t.Fatalf("recognized DOCX cover configuration warned: %s", stderr)
			}
			got := w074DOCXFingerprint(t, target)
			if want == "" {
				want = got
			} else if got != want {
				t.Fatalf("DOCX title semantics differ from local YAML baseline")
			}
		})
	}
}
