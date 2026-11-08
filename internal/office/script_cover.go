// ABOUTME: Declares the bounded native-office projection of prepared script cover assets.
// ABOUTME: Keeps source resolution and image processing in the application layer.
package office

import (
	"fmt"

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
	return fmt.Errorf("DOCX script cover projection is not implemented")
}
