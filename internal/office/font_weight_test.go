// ABOUTME: Verifies configured font weights survive native prose serialization.
// ABOUTME: Distinguishes numeric weights from semantic bold overrides.
package office

import (
	"testing"

	"github.com/tigger-developer/first-folio/internal/config"
)

func TestRT063_4_ODTFontWeightValues(t *testing.T) {
	for _, tc := range []struct{ input, want string }{
		{"regular", "normal"}, {"bold", "bold"}, {"light", "300"},
		{"medium", "500"}, {"semibold", "600"}, {"extrabold", "800"},
		{"600", "600"}, {"900", "900"},
	} {
		t.Run(tc.input, func(t *testing.T) {
			s, err := ResolveStyle(config.Font{Size: "12pt", Weight: tc.input, Style: "regular"}, 12)
			if err != nil {
				t.Fatal(err)
			}
			n := s.odtFont()
			if got := n.get("fo:font-weight"); got != tc.want {
				t.Errorf("weight = %q, want %q", got, tc.want)
			}
		})
	}
	for _, tc := range []struct {
		weight string
		bold   bool
		want   string
	}{{"regular", true, "bold"}, {"bold", false, "normal"}, {"600", false, "normal"}} {
		s, err := ResolveStyle(config.Font{Size: "12pt", Weight: tc.weight, Style: "regular"}, 12)
		if err != nil {
			t.Fatal(err)
		}
		s.Bold = tc.bold
		n := s.odtFont()
		if got := n.get("fo:font-weight"); got != tc.want {
			t.Errorf("override %s/%v = %s, want %s", tc.weight, tc.bold, got, tc.want)
		}
	}
}
