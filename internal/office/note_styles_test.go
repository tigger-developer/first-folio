// ABOUTME: Protects native ODT note anchor configuration across reference and final styling.
// ABOUTME: Checks that the two style passes retain a single editable anchor definition.
package office

import "testing"

func TestRT063_3_ODTNoteAnchorStyleIsStable(t *testing.T) {
	p := &Prose{Styles: map[string]Style{"FolioBody": {Size: 12, FontBox: 15, BaselineShift: 6}}}
	files := map[string][]byte{}
	var err error
	files["styles.xml"], err = docxEncode(odtRoot("office:document-styles", ox("office:styles")), DataLimit)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if err = p.configureStyles(files, "odt"); err != nil {
			t.Fatal(err)
		}
	}
	root, err := parseXML(files["styles.xml"])
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	root.visit(func(n *docxNode) {
		if n.name == "style:style" && n.get("style:name") == "FolioNoteAnchor" {
			count++
			if got := n.child("style:text-properties").get("style:text-position"); got != "73.00000% 58%" {
				t.Errorf("anchor does not incorporate the resolved body baseline: %s", got)
			}
		}
	})
	if count != 1 {
		t.Errorf("got %d anchor styles, want one", count)
	}
}
