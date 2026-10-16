// ABOUTME: Checks isolated running bands do not inherit body variable text.
// ABOUTME: Protects intentional blank pages and title-page numbering.
package office

import "testing"

func TestRT063_6_ODTBlankAndTitleBandsDoNotReadBodyText(t *testing.T) {
	p := &Prose{}
	for _, tc := range []struct {
		name, text string
		state      Boundary
	}{
		{"blank", "", Boundary{Blank: true}},
		{"title", "[page]", Boundary{Title: true, TitleNumber: true}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			styles := ox("office:document-styles")
			band := p.odtBand(tc.text, Band{}, "footer", "", tc.state, &styles)
			band.visit(func(n *docxNode) {
				if n.name == "text:variable-get" {
					t.Error("isolated band reads changing body text")
				}
			})
		})
	}
}
