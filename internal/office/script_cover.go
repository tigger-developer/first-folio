// ABOUTME: Declares the bounded native-office projection of prepared script cover assets.
// ABOUTME: Keeps source resolution and image processing in the application layer.
package office

import (
	"fmt"
	"math"
	"strings"

	"github.com/tigger-developer/first-folio/internal/config"
)

// ScriptCoverAsset is one conversion-scoped, already validated image payload.
type ScriptCoverAsset struct {
	Data        []byte
	Format      string
	PixelWidth  uint32
	PixelHeight uint32
}

// ScriptCover projects the resolved cover contract into the native office writer.
type ScriptCover struct {
	Config           config.CoverImage
	Sharp            ScriptCoverAsset
	Blurred          ScriptCoverAsset
	Backdrop         ScriptCoverAsset
	RenderedWidthMM  float64
	RenderedHeightMM float64
}

func validateScriptCover(cover *ScriptCover) error {
	if cover == nil {
		return nil
	}
	if len(cover.Sharp.Data) == 0 {
		return fmt.Errorf("folio.cover-image.path has no prepared sharp asset")
	}
	if err := validateScriptCoverAsset("sharp", cover.Sharp); err != nil {
		return err
	}
	if cover.Config.Mode == "inset" {
		if cover.RenderedWidthMM <= 0 || cover.RenderedHeightMM <= 0 || math.IsNaN(cover.RenderedWidthMM) || math.IsNaN(cover.RenderedHeightMM) || math.IsInf(cover.RenderedWidthMM, 0) || math.IsInf(cover.RenderedHeightMM, 0) {
			return fmt.Errorf("folio.cover-image inset geometry must have finite positive dimensions")
		}
	}
	if len(cover.Blurred.Data) > 0 {
		if err := validateScriptCoverAsset("blurred", cover.Blurred); err != nil {
			return err
		}
	}
	if cover.Config.BackdropMode == "image" {
		if len(cover.Backdrop.Data) == 0 {
			return fmt.Errorf("folio.cover-image.backdrop image has no prepared asset")
		}
		if err := validateScriptCoverAsset("backdrop", cover.Backdrop); err != nil {
			return err
		}
	}
	return nil
}

func validateScriptCoverAsset(role string, asset ScriptCoverAsset) error {
	format := strings.ToLower(asset.Format)
	if format != "png" && format != "jpeg" && format != "jpg" {
		return fmt.Errorf("folio.cover-image %s asset format %q is unsupported", role, asset.Format)
	}
	if asset.PixelWidth == 0 || asset.PixelHeight == 0 {
		return fmt.Errorf("folio.cover-image %s asset dimension must be positive", role)
	}
	return nil
}
