// ABOUTME: Exercises canonical Pandoc data limits before expensive prose rendering.
// ABOUTME: Separately verifies semantic node counts, nesting, encoding and prohibited source constructs.
package office

import (
	"encoding/json"
	"strings"
	"testing"
)

func astDocument(t *testing.T, blocks []any) []byte {
	t.Helper()
	raw, err := json.Marshal(map[string]any{"pandoc-api-version": []int{1, 23, 1, 1}, "meta": map[string]any{}, "blocks": blocks})
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func astNode(kind string, contents any) map[string]any {
	if contents == nil {
		return map[string]any{"t": kind}
	}
	return map[string]any{"t": kind, "c": contents}
}

func testASTLimits() astLimits {
	return astLimits{bytes: 128 << 20, blocks: 50000, inlines: 100000, depth: 128}
}

func TestRT063_9_ASTResourceThresholds(t *testing.T) {
	paragraph := astNode("Para", []any{astNode("Str", "One"), astNode("Space", nil), astNode("Str", "two")})
	raw := astDocument(t, []any{paragraph, paragraph})
	for _, bound := range []string{"bytes", "blocks", "inlines"} {
		for _, exceeds := range []bool{false, true} {
			name := bound + "/at-limit"
			if exceeds {
				name = bound + "/over-limit"
			}
			t.Run(name, func(t *testing.T) {
				limits := testASTLimits()
				delta := 0
				if exceeds {
					delta = 1
				}
				switch bound {
				case "bytes":
					limits.bytes = len(raw) - delta
				case "blocks":
					limits.blocks = 2 - delta
				case "inlines":
					limits.inlines = 6 - delta
				}
				err := validateAST(raw, limits)
				if (err != nil) != exceeds {
					t.Errorf("exceeds=%v: validation error=%v", exceeds, err)
				}
			})
		}
	}
}

func TestRT063_9_ASTNesting(t *testing.T) {
	for _, depth := range []int{128, 129} {
		// Semantic depth starts with a top-level block at 1; containers and
		// JSON object/array scaffolding do not count as additional nodes.
		var inline any = astNode("Str", "leaf")
		for i := 0; i < depth-2; i++ {
			inline = astNode("Emph", []any{inline})
		}
		raw := astDocument(t, []any{astNode("Para", []any{inline})})
		err := validateAST(raw, testASTLimits())
		if (err != nil) != (depth > 128) {
			t.Errorf("semantic depth %d: error=%v", depth, err)
		}
	}
}

func TestRT063_3_ASTRejectsUnsupportedContent(t *testing.T) {
	for _, tc := range []struct {
		name  string
		block any
	}{
		{"raw-block", astNode("RawBlock", []any{"html", "<script>not executed</script>"})},
		{"raw-inline", astNode("Para", []any{astNode("RawInline", []any{"html", "<script>not executed</script>"})})},
		{"source-image", astNode("Para", []any{astNode("Image", []any{[]any{"", []any{}, []any{}}, []any{astNode("Str", "alt")}, []any{"https://example.invalid/image.png", ""}})})},
		{"unknown-node", astNode("UnexpectedFutureNode", []any{})},
		{"xml-control", astNode("Para", []any{astNode("Str", "bad\x01text")})},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := validateAST(astDocument(t, []any{tc.block}), testASTLimits()); err == nil {
				t.Fatal("unsupported source content accepted")
			}
		})
	}
}

func TestRT063_9_ASTMalformedInput(t *testing.T) {
	valid := astDocument(t, []any{astNode("Para", []any{astNode("Str", "text")})})
	for _, raw := range [][]byte{
		[]byte("{"), []byte("null"), []byte(`{"blocks":[],"meta":{}}`),
		[]byte(strings.Replace(string(valid), "[1,23,1,1]", "[999,0]", 1)),
		[]byte(strings.Replace(string(valid), "text", "\xff", 1)),
		append(append([]byte{}, valid...), valid...),
	} {
		if err := validateAST(raw, testASTLimits()); err == nil {
			t.Errorf("invalid canonical JSON accepted: %q", raw)
		}
	}
}

func TestRT063_9_ASTRejectsMalformedAdapterNodes(t *testing.T) {
	attr := []any{"", []any{}, []any{}}
	for _, node := range []any{
		astNode("Header", []any{2, attr, "not inlines"}),
		astNode("Header", []any{"two", attr, []any{}}),
		astNode("Para", []any{astNode("Str", 17)}),
		astNode("Para", []any{astNode("Code", []any{attr})}),
		astNode("Para", []any{astNode("Span", []any{attr, "not inlines"})}),
		astNode("Para", []any{astNode("Link", []any{attr, []any{}, "not target"})}),
		astNode("Para", []any{astNode("Strong", "not inlines")}),
	} {
		if _, err := Decode(astDocument(t, []any{node})); err == nil {
			t.Errorf("malformed adapter input accepted: %#v", node)
		}
	}
}
