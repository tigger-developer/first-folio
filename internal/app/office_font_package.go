// ABOUTME: Keeps script callers on the shared native office font packaging boundary.
// ABOUTME: Script layout and serialization remain owned by the existing script writers.
package app

import "github.com/tigger-developer/first-folio/internal/office"

func embedOfficeFonts(raw []byte, format string) ([]byte, []string, error) {
	return office.EmbedFonts(raw, format)
}
