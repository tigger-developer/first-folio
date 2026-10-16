// ABOUTME: Checks the full-face and aggregate byte limits for selected office fonts.
// ABOUTME: Uses declared sizes so exact production thresholds need no giant synthetic font bodies.
package office

import "testing"

func TestRT063_8_SelectedFontBudgets(t *testing.T) {
	for _, tc := range []struct {
		name      string
		sizes     []uint64
		wantError bool
	}{
		{"empty", nil, false},
		{"face-at-limit", []uint64{32 << 20}, false},
		{"face-over-limit", []uint64{(32 << 20) + 1}, true},
		{"aggregate-at-limit", []uint64{32 << 20, 32 << 20}, false},
		{"aggregate-over-limit", []uint64{32 << 20, 32 << 20, 1}, true},
		{"integer-overflow", []uint64{1, ^uint64(0)}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := checkFontBudget(tc.sizes)
			if (err != nil) != tc.wantError {
				t.Errorf("font sizes %v: error=%v; rejection required=%v", tc.sizes, err, tc.wantError)
			}
		})
	}
}
