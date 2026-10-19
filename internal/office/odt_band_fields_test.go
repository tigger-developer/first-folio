// ABOUTME: Checks isolated running bands do not inherit body variable text.
// ABOUTME: Protects intentional blank pages and title-page numbering.
package office

import (
	"strings"
	"testing"
)

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

func TestRT063_6_ODTBoundaryFieldSequences(t *testing.T) {
	for _, tc := range []struct {
		name, front, body string
		enabled, change   bool
	}{
		{"wording", "Introduction [page]", "[chapter] [page]", true, false},
		{"add page", "Introduction", "[page]", true, true},
		{"remove page", "[page]", "Chapter", true, true},
		{"field type", "[page]", "[total-pages]", true, true},
		{"field order", "[page] of [total-pages]", "[total-pages] / [page]", true, true},
		{"disabled", "Introduction", "[page]", false, false},
	} {
		for _, kind := range []string{"header odd", "header even", "footer odd", "footer even"} {
			t.Run(tc.name+"/"+kind, func(t *testing.T) {
				p := Prose{}
				band := &p.Page.Header
				if strings.HasPrefix(kind, "footer") {
					band = &p.Page.Footer
				}
				band.Enabled = tc.enabled
				if strings.HasSuffix(kind, "even") {
					band.FrontEven, band.Even = tc.front, tc.body
				} else {
					band.FrontOdd, band.Odd = tc.front, tc.body
				}
				if got := p.odtBodyMasterChange(Boundary{}, Boundary{Body: true}); got != tc.change {
					t.Errorf("master change=%v, want %v", got, tc.change)
				}
				if p.odtBodyMasterChange(Boundary{Body: true}, Boundary{Body: true}) {
					t.Error("later body heading acquired the frontmatter exception")
				}
			})
		}
	}
}
