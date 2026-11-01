// ABOUTME: Verifies W071 Org semantics at the letter Typst projection boundary.
// ABOUTME: Protects both Org literal forms from Typst typography transforms.
package letter

import (
	"strings"
	"testing"
)

func TestRT071_1_LetterTypstUsesSharedOrgSemantics(t *testing.T) {
	got := inlineOrg(`Prose -- em --- dots ... soft\-hyphen =literal--...\-= ~code---...\-~`)
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
