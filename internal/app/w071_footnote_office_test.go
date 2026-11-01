// ABOUTME: Verifies W071 structured Org footnotes at native DOCX and ODT boundaries.
// ABOUTME: Detects displaced continuation paragraphs independently of editor rendering.
package app

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestRT071_2_StructuredFootnotesSurviveNativeSerialization(t *testing.T) {
	binary := officeBuild(t)
	source := "#+TITLE: Notes\n* ACT ONE\n** SCENE ONE\n**** CÁIT\nLine.[fn:note]\n\n[fn:note] First *bold* paragraph.\n\nSecond /italic/ paragraph.\n\nThird paragraph.\n"
	for _, format := range []string{"docx", "odt"} {
		t.Run(format, func(t *testing.T) {
			dir := t.TempDir()
			src := filepath.Join(dir, "notes.org")
			writeAppFile(t, src, source)
			raw, diagnostic, err := officeRun(t, binary, dir, officeEnv(t, t.TempDir(), nil), "convert", src, "--to", format)
			if err != nil {
				t.Fatalf("convert: %v: %s", err, diagnostic)
			}
			pkg := officePackage(t, raw)
			body := officeBody(t, pkg, format)
			if got := strings.Join(w071BodyParagraphs(body, false), " | "); strings.Contains(got, "Second /italic/") || strings.Contains(got, "Third paragraph") {
				t.Errorf("note continuation leaked into body: %s", got)
			}

			var note *officeXML
			if format == "docx" {
				for _, candidate := range pkg["word/footnotes.xml"].all("footnote") {
					if strings.Contains(candidate.text(), "First") {
						note = candidate
					}
				}
			} else {
				for _, candidate := range body.all("note-body") {
					if strings.Contains(candidate.text(), "First") {
						note = candidate
					}
				}
			}
			if note == nil {
				t.Fatal("structured footnote missing")
			}
			paragraphs := note.all("p")
			if len(paragraphs) != 3 {
				t.Fatalf("note paragraph count = %d, want 3: %s", len(paragraphs), note.text())
			}
			for i, want := range []string{"First bold paragraph.", "Second italic paragraph.", "Third paragraph."} {
				if got := strings.Join(strings.Fields(paragraphs[i].text()), " "); got != want {
					t.Errorf("paragraph %d = %q, want %q", i, got, want)
				}
			}
			officeAssertEmphasis(t, pkg, format, note, "bold", true, false)
			officeAssertEmphasis(t, pkg, format, note, "italic", false, true)
		})
	}
}

func w071BodyParagraphs(node *officeXML, inNote bool) []string {
	if node.Name.Local == "note-body" {
		inNote = true
	}
	var out []string
	if node.Name.Local == "p" && !inNote {
		out = append(out, w071TextOutsideNotes(node))
		return out
	}
	for _, child := range node.Children {
		out = append(out, w071BodyParagraphs(child, inNote)...)
	}
	return out
}

func w071TextOutsideNotes(node *officeXML) string {
	if node.Name.Local == "note-body" {
		return ""
	}
	var out strings.Builder
	out.WriteString(node.Text)
	for _, child := range node.Children {
		out.WriteString(w071TextOutsideNotes(child))
	}
	return out.String()
}
