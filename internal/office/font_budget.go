// ABOUTME: Enforces full-face and aggregate selected-font byte budgets.
// ABOUTME: Checks subtraction before addition to avoid overflow from untrusted lengths.
package office

import "fmt"

func checkFontBudget(sizes []uint64) error {
	remaining := uint64(64 << 20)
	for _, size := range sizes {
		if size > 32<<20 {
			return fmt.Errorf("selected font exceeds 32 MiB per-face limit")
		}
		if size > remaining {
			return fmt.Errorf("selected fonts exceed 64 MiB aggregate limit")
		}
		remaining -= size
	}
	return nil
}
