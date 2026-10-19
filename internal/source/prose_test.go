// ABOUTME: Exercises native source limits and syntax at the shared parser boundary.
// ABOUTME: Protects literal and table preservation independently of office serialization.
package source

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNativeSourceBudgets(t *testing.T) {
	for _, tc := range []struct{ name, text, want string }{
		{"bytes", strings.Repeat("x", (32<<20)+1), "32 MiB"},
		{"blocks", strings.Repeat("x\n\n", 50001), "block limit"},
		{"inlines", strings.Repeat("**x** ", 50001), "inline limit"},
		{"nesting", strings.Repeat("> ", 130) + "x\n", "nesting"},
		{"unterminated-org", "#+begin_src\nx\n", "unterminated"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			format := "markdown"
			if tc.name == "unterminated-org" {
				format = "org"
			}
			_, err := Prose(format, tc.text)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("expected %s, got %v", tc.want, err)
			}
		})
	}
}

func TestNativeLiteralAndEmptyTableCells(t *testing.T) {
	for _, format := range []string{"markdown", "org"} {
		literal := "`**literal**`"
		if format == "org" {
			literal = "=*literal*="
		}
		nodes, err := Prose(format, "Before "+literal+" after.\n\n| Left | Right |\n|------+-------|\n| value ||\n")
		if err != nil {
			t.Fatal(err)
		}
		if len(nodes) != 2 || nodes[1].Type != "Table" {
			t.Fatalf("expected prose and table: %#v", nodes)
		}
		raw, err := json.Marshal(nodes)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(raw), `"t":"Code"`) {
			t.Error("literal protection lost")
		}
	}
}
