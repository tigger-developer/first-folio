// ABOUTME: Checks separators between speaker names and delivery instructions.
// ABOUTME: Keeps comma-separated names intact while stripping an attached instruction separator.
package play

import "testing"

func TestSpeakerInstructionSeparator(t *testing.T) {
	for _, tc := range []struct{ source, name, direction string }{
		{"Bob, cheerfully", "Bob", "cheerfully"},
		{"Bob (cheerfully)", "Bob", "cheerfully"},
		{"Bob cheerfully", "Bob", "cheerfully"},
		{"BOB, ALICE", "BOB, ALICE", ""},
		{"BOB, ALICE, together", "BOB, ALICE", "together"},
	} {
		t.Run(tc.source, func(t *testing.T) {
			name, direction := parseCharacterLine(tc.source)
			if name != tc.name || direction != tc.direction {
				t.Fatalf("got (%q, %q), want (%q, %q)", name, direction, tc.name, tc.direction)
			}
		})
	}
}
