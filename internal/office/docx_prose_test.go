// ABOUTME: Checks native note construction at empty and structured-content boundaries.
// ABOUTME: Protects reference ownership without requiring an external writer.
package office

import "testing"

func TestRT063_3_NativeEmptyFootnote(t *testing.T) {
	writer := &docxWriter{}
	adapter := &proseDOCX{writer: writer}
	runs, err := adapter.inlines([]Node{N("Note", []Node{})}, dx("rPr"), 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(runs) != 1 || runs[0].child("w:footnoteReference") == nil {
		t.Fatal("missing body note reference")
	}
	if len(writer.footnotes) != 1 {
		t.Fatal("missing native note")
	}
	note := writer.footnotes[0]
	if len(note.children) != 1 || note.children[0].name != "w:p" {
		t.Fatal("empty note needs an editable paragraph")
	}
	paragraph := note.children[0]
	if paragraph.children[0].name != "w:pPr" || paragraph.children[1].child("w:footnoteRef") == nil {
		t.Fatal("note reference must follow paragraph properties")
	}
}
