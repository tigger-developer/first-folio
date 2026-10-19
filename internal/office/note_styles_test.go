// ABOUTME: Protects native ODT note anchor configuration across reference and final styling.
// ABOUTME: Retains a genuine note superscript independently of legacy body offsets.
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
			if got := n.child("style:text-properties").get("style:text-position"); got != "33.00000% 58%" {
				t.Errorf("note anchor must not inherit an ordinary body baseline offset: %s", got)
			}
		}
	})
	if count != 1 {
		t.Errorf("got %d anchor styles, want one", count)
	}
}
