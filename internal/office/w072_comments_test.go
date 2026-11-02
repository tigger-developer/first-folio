// ABOUTME: Defines W072's selective, bounded Word comment package transformation.
// ABOUTME: Covers attribution metadata, paragraph retention, preservation and failures.
package office

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"
)

func TestRT072_1_DOCXPackageConversionIsSelectiveAndComplete(t *testing.T) {
	now := time.Date(2026, 11, 1, 12, 13, 14, 0, time.UTC)
	raw := w072Package(t, false, false)
	converted, count, err := ConvertAttributedFootnotes(raw, now)
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("converted %d attributed notes, want 2", count)
	}
	files, err := readPackage(converted, defaultPackageLimits())
	if err != nil {
		t.Fatal(err)
	}
	comments := string(files["word/comments.xml"])
	for _, want := range []string{`w:id="7"`, `w:id="8"`, `w:id="9"`, `w:author="Aoife Example"`, `w:initials="AE"`, `w:date="2026-11-01T12:13:14Z"`, `w:author="Siobhan Example"`, `w:initials="SE"`, `w:date="2026-10-31T17:59:22Z"`, "First comment", "Second paragraph", "Keep me"} {
		if !strings.Contains(comments, want) {
			t.Errorf("comments missing %q: %s", want, comments)
		}
	}
	if got := strings.Count(comments, "Second paragraph"); got != 1 {
		t.Errorf("multi-paragraph comment occurrence count = %d", got)
	}
	commentRoot, err := parseXML(files["word/comments.xml"])
	if err != nil {
		t.Fatal(err)
	}
	paragraphs := 0
	commentRoot.visit(func(n *docxNode) {
		if n.name == "w:comment" && n.get("w:author") == "Aoife Example" {
			for _, child := range n.children {
				if child.name == "w:p" {
					paragraphs++
				}
			}
		}
	})
	if paragraphs != 2 {
		t.Errorf("multi-paragraph comment has %d paragraphs, want 2", paragraphs)
	}
	footnotes := string(files["word/footnotes.xml"])
	if strings.Contains(footnotes, "First comment") || strings.Contains(footnotes, "Timestamped comment") {
		t.Error("converted note remains in footnotes")
	}
	if !strings.Contains(footnotes, "Ordinary note") {
		t.Error("ordinary footnote was removed")
	}
	if string(files["custom/keep.xml"]) != `<keep>unchanged</keep>` {
		t.Error("unrelated part changed")
	}
	document := string(files["word/document.xml"])
	documentRoot, err := parseXML(files["word/document.xml"])
	if err != nil {
		t.Fatal(err)
	}
	starts, commentReferences, footnoteReferences := 0, 0, 0
	documentRoot.visit(func(n *docxNode) {
		switch n.name {
		case "w:commentRangeStart":
			starts++
		case "w:commentReference":
			commentReferences++
		case "w:footnoteReference":
			footnoteReferences++
		}
	})
	if starts != 2 || commentReferences != 2 {
		t.Errorf("comment anchors incomplete: starts=%d references=%d: %s", starts, commentReferences, document)
	}
	if footnoteReferences != 1 {
		t.Errorf("ordinary footnote references = %d, want 1: %s", footnoteReferences, document)
	}
	anchored := false
	documentRoot.visit(func(n *docxNode) {
		if n.name != "w:p" {
			return
		}
		starts, ends, references := 0, 0, 0
		for _, child := range n.children {
			switch child.name {
			case "w:commentRangeStart":
				starts++
			case "w:commentRangeEnd":
				ends++
			case "w:r":
				if child.child("w:commentReference") != nil {
					references++
				}
			}
		}
		if starts == 2 && ends == 2 && references == 2 {
			anchored = true
		}
	})
	if !anchored {
		t.Error("two comments are not anchored within their shared source paragraph")
	}
	for _, commentID := range []string{"8", "9"} {
		if got := w072CommentRangeText(documentRoot, commentID); got != "Whole sentence." {
			t.Errorf("comment %s range text = %q, want whole source sentence", commentID, got)
		}
	}
	extendedRoot, err := parseXML(files["word/commentsExtended.xml"])
	if err != nil {
		t.Fatal(err)
	}
	extendedEntries := 0
	extendedRoot.visit(func(n *docxNode) {
		if n.name == "w15:commentEx" {
			extendedEntries++
		}
	})
	if extendedEntries != 3 {
		t.Errorf("extended comment entries = %d, want 3", extendedEntries)
	}
	for name, want := range map[string]string{
		"[Content_Types].xml":          "/word/commentsExtended.xml",
		"word/_rels/document.xml.rels": "commentsExtended.xml",
	} {
		if !strings.Contains(string(files[name]), want) {
			t.Errorf("%s missing %q", name, want)
		}
	}
}

func w072CommentRangeText(root *docxNode, commentID string) string {
	var result strings.Builder
	active := false
	root.visit(func(n *docxNode) {
		if n.name != "w:p" {
			return
		}
		for i := range n.children {
			child := &n.children[i]
			switch {
			case child.name == "w:commentRangeStart" && child.get("w:id") == commentID:
				active = true
			case child.name == "w:commentRangeEnd" && child.get("w:id") == commentID:
				active = false
			case active:
				result.WriteString(wordText(child))
			}
		}
	})
	return result.String()
}

func TestRT072_1_DOCXPackageConversionRejectsInvalidCandidates(t *testing.T) {
	for _, tc := range []struct {
		name      string
		duplicate bool
		badDate   bool
		want      string
	}{
		{"invalid timestamp", false, true, "invalid UTC timestamp"},
		{"duplicate reference", true, false, "expected 1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := ConvertAttributedFootnotes(w072Package(t, tc.duplicate, tc.badDate), time.Unix(0, 0).UTC())
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want %q", err, tc.want)
			}
		})
	}
}

func TestRT072_1_DOCXPackageConversionLeavesNoCandidatesUnchanged(t *testing.T) {
	raw := w072RewritePackage(t, w072Package(t, false, false), func(files map[string][]byte) {
		files["word/footnotes.xml"] = bytes.ReplaceAll(files["word/footnotes.xml"], []byte("[Aoife Example] wrote: "), []byte("[Aoife Example] wrote "))
		files["word/footnotes.xml"] = bytes.ReplaceAll(files["word/footnotes.xml"], []byte("[Siobhan Example] wrote (2026-10-31 17:59:22Z): "), []byte("[Siobhan Example] wrote [2026-10-31 17:59:22Z]: "))
		files["word/footnotes.xml"] = bytes.ReplaceAll(files["word/footnotes.xml"], []byte("Ordinary note"), []byte("[Third Example] wrote:No separating space"))
	})
	converted, count, err := ConvertAttributedFootnotes(raw, time.Unix(0, 0).UTC())
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 || !bytes.Equal(converted, raw) {
		t.Fatalf("no-candidate conversion changed package: count=%d equal=%v", count, bytes.Equal(converted, raw))
	}
}

func TestRT072_1_DOCXPackageConversionRejectsMalformedPackage(t *testing.T) {
	if _, _, err := ConvertAttributedFootnotes([]byte("not a zip"), time.Unix(0, 0).UTC()); err == nil {
		t.Fatal("malformed package accepted")
	}
}

func TestRT072_1_DOCXPackageConversionRejectsMissingReference(t *testing.T) {
	raw := w072RewritePackage(t, w072Package(t, false, false), func(files map[string][]byte) {
		files["word/document.xml"] = bytes.Replace(files["word/document.xml"], []byte(`<w:r><w:footnoteReference w:id="2"/></w:r>`), nil, 1)
	})
	_, _, err := ConvertAttributedFootnotes(raw, time.Unix(0, 0).UTC())
	if err == nil || !strings.Contains(err.Error(), "expected 1") {
		t.Fatalf("error = %v, want missing-reference failure", err)
	}
}

func TestRT072_1_DOCXPackageConversionRejectsRelationshipConflict(t *testing.T) {
	raw := w072RewritePackage(t, w072Package(t, false, false), func(files map[string][]byte) {
		files["word/_rels/document.xml.rels"] = bytes.Replace(files["word/_rels/document.xml.rels"], []byte(`Target="comments.xml"`), []byte(`Target="wrong-comments.xml"`), 1)
	})
	_, _, err := ConvertAttributedFootnotes(raw, time.Unix(0, 0).UTC())
	if err == nil || !strings.Contains(err.Error(), "comments") {
		t.Fatalf("error = %v, want comments relationship conflict", err)
	}
}

func TestRT072_1_DOCXPackageConversionRejectsContentTypeConflict(t *testing.T) {
	raw := w072RewritePackage(t, w072Package(t, false, false), func(files map[string][]byte) {
		files["[Content_Types].xml"] = bytes.Replace(files["[Content_Types].xml"], []byte(`application/vnd.openxmlformats-officedocument.wordprocessingml.comments+xml`), []byte(`application/x-wrong-comments`), 1)
	})
	_, _, err := ConvertAttributedFootnotes(raw, time.Unix(0, 0).UTC())
	if err == nil || !strings.Contains(err.Error(), "content type") {
		t.Fatalf("error = %v, want comments content-type conflict", err)
	}
}

func w072Package(t *testing.T, duplicate, badDate bool) []byte {
	t.Helper()
	date := "2026-10-31 17:59:22Z"
	if badDate {
		date = "31 October"
	}
	ref2 := ""
	if duplicate {
		ref2 = `<w:r><w:footnoteReference w:id="1"/></w:r>`
	}
	parts := map[string]string{
		"[Content_Types].xml":          `<?xml version="1.0"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/><Override PartName="/word/footnotes.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.footnotes+xml"/><Override PartName="/word/comments.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.comments+xml"/><Override PartName="/custom/keep.xml" ContentType="application/xml"/></Types>`,
		"_rels/.rels":                  `<?xml version="1.0"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="document" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/></Relationships>`,
		"word/_rels/document.xml.rels": `<?xml version="1.0"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="footnotes" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/footnotes" Target="footnotes.xml"/><Relationship Id="comments" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/comments" Target="comments.xml"/></Relationships>`,
		"word/document.xml":            fmt.Sprintf(`<?xml version="1.0"?><w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body><w:p><w:r><w:t>Whole </w:t></w:r><w:r><w:t>sentence.</w:t></w:r><w:r><w:footnoteReference w:id="1"/></w:r>%s<w:r><w:footnoteReference w:id="2"/></w:r></w:p><w:p><w:r><w:t>Ordinary</w:t></w:r><w:r><w:footnoteReference w:id="3"/></w:r></w:p></w:body></w:document>`, ref2),
		"word/footnotes.xml":           fmt.Sprintf(`<?xml version="1.0"?><w:footnotes xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:footnote w:id="-1"><w:p/></w:footnote><w:footnote w:id="0"><w:p/></w:footnote><w:footnote w:id="1"><w:p><w:r><w:footnoteRef/></w:r><w:r><w:t>[Aoife Example] wrote: First comment</w:t></w:r></w:p><w:p><w:r><w:t>Second paragraph</w:t></w:r></w:p></w:footnote><w:footnote w:id="2"><w:p><w:r><w:footnoteRef/></w:r><w:r><w:t>[Siobhan Example] wrote (%s): Timestamped comment</w:t></w:r></w:p></w:footnote><w:footnote w:id="3"><w:p><w:r><w:footnoteRef/></w:r><w:r><w:t>Ordinary note</w:t></w:r></w:p></w:footnote></w:footnotes>`, date),
		"word/comments.xml":            `<?xml version="1.0"?><w:comments xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:comment w:id="7" w:author="Existing"><w:p><w:r><w:t>Keep me</w:t></w:r></w:p></w:comment></w:comments>`,
		"custom/keep.xml":              `<keep>unchanged</keep>`,
	}
	var out bytes.Buffer
	zw := zip.NewWriter(&out)
	for _, name := range []string{"[Content_Types].xml", "_rels/.rels", "word/_rels/document.xml.rels", "word/document.xml", "word/footnotes.xml", "word/comments.xml", "custom/keep.xml"} {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = io.WriteString(w, parts[name]); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return out.Bytes()
}

func w072RewritePackage(t *testing.T, raw []byte, update func(map[string][]byte)) []byte {
	t.Helper()
	files, err := readPackage(raw, defaultPackageLimits())
	if err != nil {
		t.Fatal(err)
	}
	update(files)
	var out bytes.Buffer
	zw := zip.NewWriter(&out)
	for name, body := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = w.Write(body); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return out.Bytes()
}
