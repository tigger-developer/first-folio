// ABOUTME: Verifies source-format-specific Typst presentation of underline semantics.
// ABOUTME: Keeps Markdown and Fountain underlining unchanged while Org uses highlighting.
package source

import "testing"

func TestRT082_2NonOrgUnderlineRemainsUnderline(t *testing.T) {
	for _, format := range []string{"markdown", "fountain"} {
		nodes, err := ScriptInline(format, "_marked_", nil)
		if err != nil {
			t.Fatal(err)
		}
		got, err := TypstInline(nodes, nil, format)
		if err != nil {
			t.Fatal(err)
		}
		if got != "#underline[marked]" {
			t.Fatalf("%s underline = %q; want native Typst underline", format, got)
		}
	}
}
