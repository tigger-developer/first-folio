// ABOUTME: Checks native Word page fields retain the configured number format.
// ABOUTME: Prevents readers from defaulting title-page fields to decimal.
package office

import (
	"strings"
	"testing"
)

func TestRT063_6_WordPageFieldNumberFormat(t *testing.T) {
	for _, tc := range []struct{ format, suffix string }{{"1", "ARABIC"}, {"i", "roman"}, {"I", "ROMAN"}} {
		t.Run(tc.format, func(t *testing.T) {
			p := &Prose{}
			band := p.wordBand("[page]", Band{}, "first", Boundary{Title: true, TitleNumber: true, NumberFormat: tc.format})
			found := false
			band.visit(func(n *docxNode) {
				if n.name == "w:fldSimple" && strings.TrimSpace(n.get("w:instr")) == `PAGE \* `+tc.suffix {
					found = true
				}
			})
			if !found {
				t.Errorf("page field has no explicit %s numeric format", tc.format)
			}
		})
	}
}
