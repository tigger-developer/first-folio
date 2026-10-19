// ABOUTME: Exercises letter-wide source budgets across individually composed paragraphs.
// ABOUTME: Prevents resetting semantic limits at recipient layout block boundaries.
package source

import (
	"strings"
	"testing"
)

func TestLetterAggregateBudgets(t *testing.T) {
	for _, tc := range []struct {
		name, text, want string
		count            int
	}{
		{"blocks", "x", "block limit", 50001},
		{"inlines", "*x* *y*", "inline limit", 20001},
		{"bytes", strings.Repeat("x", 1<<20), "32 MiB", 33},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := NewLetterParser()
			var err error
			for i := 0; i < tc.count; i++ {
				_, err = p.Parse(tc.text)
				if err != nil {
					if i != tc.count-1 {
						t.Fatalf("budget failed early at %d: %v", i, err)
					}
					break
				}
			}
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("expected letter-wide %s, got %v", tc.want, err)
			}
		})
	}
}
