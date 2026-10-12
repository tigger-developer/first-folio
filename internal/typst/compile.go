// ABOUTME: Makes bundled fonts available during PDF compilation for every document type.
// ABOUTME: Preserves caller font paths and removes private extracted files after compilation.
package typst

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	folio "github.com/tigger-developer/first-folio"
)

// Compile renders a Typst source using the fonts shipped inside the executable.
func Compile(source, output string) (err error) {
	dir, err := os.MkdirTemp("", "folio-fonts-")
	if err != nil {
		return fmt.Errorf("creating temporary font directory: %w", err)
	}
	defer func() {
		if cleanupErr := os.RemoveAll(dir); cleanupErr != nil {
			err = errors.Join(err, fmt.Errorf("removing temporary fonts: %w", cleanupErr))
		}
	}()
	const bundled = "fonts/iosevka-publish"
	entries, err := folio.Assets.ReadDir(bundled)
	if err != nil {
		return fmt.Errorf("reading bundled font directory: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".ttf" {
			continue
		}
		data, err := folio.Assets.ReadFile(bundled + "/" + entry.Name())
		if err != nil {
			return fmt.Errorf("reading bundled font: %w", err)
		}
		if err := os.WriteFile(filepath.Join(dir, entry.Name()), data, 0o600); err != nil {
			return fmt.Errorf("extracting bundled font: %w", err)
		}
	}
	paths := dir
	if existing := os.Getenv("TYPST_FONT_PATHS"); existing != "" {
		paths = existing + string(os.PathListSeparator) + dir
	}
	cmd := exec.Command("typst", "compile", "--font-path", paths, source, output)
	if result, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("typst compile failed: %w: %s", err, strings.TrimSpace(string(result)))
	}
	return nil
}
