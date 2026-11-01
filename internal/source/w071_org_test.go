// ABOUTME: Defines W071's shared Org inline semantic regression boundary.
// ABOUTME: Distinguishes prose typography from protected literal and structural content.
package source

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestRT071_1_OrgInlineSemanticsAreParsedOnce(t *testing.T) {
	input := `Prose -- em --- dots ... soft\-hyphen *strong --* "quoted" =literal--...\-= ~code---...\-~.`
	for _, tc := range []struct {
		name  string
		quote string
		parse func() ([]Node, error)
	}{
		{"prose", "“quoted”", func() ([]Node, error) { return Prose("org", input) }},
		{"script", `\"quoted\"`, func() ([]Node, error) { return ScriptInline("org", input, nil) }},
		{"letter", `\"quoted\"`, func() ([]Node, error) { return NewLetterParser().Parse(input) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			nodes, err := tc.parse()
			if err != nil {
				t.Fatal(err)
			}
			raw, err := json.Marshal(nodes)
			if err != nil {
				t.Fatal(err)
			}
			got := string(raw)
			for _, want := range []string{"Prose – em — dots … soft\u00adhyphen", "strong –", tc.quote, `literal--...\\-`, `code---...\\-`} {
				if !strings.Contains(got, want) {
					t.Errorf("semantic output %s does not contain %q", got, want)
				}
			}
			if strings.Contains(got, `=literal`) || strings.Contains(got, `~code`) {
				t.Errorf("literal delimiters leaked: %s", got)
			}
		})
	}

	blocks, err := Prose("org", "-----\n\n#+begin_src text\ninside -- --- ... \\-\n#+end_src\n")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(blocks)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `inside -- --- ... \\-`) {
		t.Fatalf("source block typography changed: %s", raw)
	}
}
