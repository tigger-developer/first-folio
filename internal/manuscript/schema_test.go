// ABOUTME: Checks the prose schema emitted by the existing Markdown serializer.
// ABOUTME: Covers empty metadata and preserves manuscript content on re-emission.
package manuscript

import (
	"strings"
	"testing"
)

func TestRenderMarkdownIncludesManuscriptSchema(t *testing.T) {
	for _, title := range []string{"", "A Novel"} {
		t.Run(title, func(t *testing.T) {
			doc := Document{Metadata: Metadata{Title: title}, Blocks: []Block{{Kind: "paragraph", Text: "A quiet morning."}}}
			output := RenderMarkdown(doc)
			const declaration = "schema: \"https://github.com/tigger-developer/first-folio/blob/master/schema/manuscript.md\""
			if !strings.HasPrefix(output, "---\n"+declaration+"\n") || strings.Count(output, "schema:") != 1 {
				t.Fatalf("missing manuscript schema:\n%s", output)
			}
			if !strings.Contains(output, "A quiet morning.") {
				t.Fatalf("body lost:\n%s", output)
			}
			if title != "" && !strings.Contains(output, "title: \"A Novel\"") {
				t.Fatalf("title lost:\n%s", output)
			}
		})
	}
}
