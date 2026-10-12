// ABOUTME: Embeds runtime documentation, presets, and Typst templates in the Go binary.
// ABOUTME: Keeps installed commands independent of the source checkout and environment variables.
package folio

import "embed"

// Assets contains the runtime files shipped inside the folio executable.
//
//go:embed docs/*.md presets/*.yaml templates/*.typ fonts/iosevka-publish/*
var Assets embed.FS
