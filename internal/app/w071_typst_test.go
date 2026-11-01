// ABOUTME: Verifies W071 Org semantics at the script Typst projection boundary.
// ABOUTME: Protects both Org literal forms from Typst typography transforms.
package app

import (
	"strings"
	"testing"

	"github.com/tigger-developer/first-folio/internal/play"
)

func TestRT071_1_ScriptTypstUsesSharedOrgSemantics(t *testing.T) {
	got := inlineTypst(`Prose -- em --- dots ... soft\-hyphen =literal--...\-= ~code---...\-~`, nil, play.FormatOrg)
	for _, want := range []string{
		"Prose – em — dots … soft\u00adhyphen",
		`#mono-text[#raw("literal--...\\-")]`,
		`#mono-text[#raw("code---...\\-")]`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("Typst output %q does not contain %q", got, want)
		}
	}
}
