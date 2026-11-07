// ABOUTME: Defines the typed renderer-neutral cover-image configuration contract.
// ABOUTME: Begins as execution scaffolding for W073 test-first delivery.
package config

import "fmt"

// CoverImage is the complete effective cover-image value after cascade merge.
type CoverImage struct {
	Path                 string
	Mode                 string
	Placement            string
	Size                 string
	ImageTreatment       string
	ImageOpacity         float64
	Superscale           string
	OverlayBlur          string
	OverlayWashTreatment string
	OverlayWashOpacity   float64
	OverlayPadding       string
}

// CoverImage returns the typed effective cover-image configuration when enabled.
func (c Config) CoverImage() (CoverImage, bool, error) {
	if _, ok := c.Get("folio.cover-image.path"); !ok {
		return CoverImage{}, false, nil
	}
	return CoverImage{}, true, fmt.Errorf("folio.cover-image typed resolution is not implemented")
}
