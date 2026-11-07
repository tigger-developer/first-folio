// ABOUTME: Owns script cover-image source limits and preparation for PDF rendering.
// ABOUTME: Starts as execution scaffolding for the W073 test-code phase.
package app

import (
	"fmt"

	"github.com/tigger-developer/first-folio/internal/config"
)

const (
	coverSourceByteLimit  int64  = 100 * 1024 * 1024
	coverSourceAxisLimit  uint32 = 32768
	coverSourcePixelLimit uint64 = 100_000_000
)

func validateCoverSourceLimits(_ int64, _, _ uint32) error {
	return fmt.Errorf("cover-image source-limit validation is not implemented")
}

type coverImagePlan struct {
	Config                            config.CoverImage
	SourcePath                        string
	PixelWidth, PixelHeight           uint32
	RenderedWidthMM, RenderedHeightMM float64
}

func planCoverImage(_ config.Config, _ string) (*coverImagePlan, error) {
	return nil, fmt.Errorf("folio.cover-image planning is not implemented")
}

func validateDodgeGroupFit(_ string, _, _ float64) error {
	return fmt.Errorf("folio.cover-image dodge fit validation is not implemented")
}
