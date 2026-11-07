// ABOUTME: Owns script cover-image source limits and preparation for PDF rendering.
// ABOUTME: Starts as execution scaffolding for the W073 test-code phase.
package app

import "fmt"

const (
	coverSourceByteLimit  int64  = 100 * 1024 * 1024
	coverSourceAxisLimit  uint32 = 32768
	coverSourcePixelLimit uint64 = 100_000_000
)

func validateCoverSourceLimits(_ int64, _, _ uint32) error {
	return fmt.Errorf("cover-image source-limit validation is not implemented")
}
