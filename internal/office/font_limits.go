// ABOUTME: Enforces strict prose font budgets while retaining legacy script warning policy.
// ABOUTME: Reads bounded metadata to identify oversized selected faces without loading their bodies.
package office

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func selectedFontLimits(names []string, faces []FontFace, warnings []string) error {
	counts := map[string]int{}
	var sizes []uint64
	for _, face := range faces {
		counts[face.family]++
		sizes = append(sizes, uint64(len(face.Raw)))
	}
	if err := checkFontBudget(sizes); err != nil {
		return err
	}
	for _, warning := range warnings {
		if strings.Contains(warning, "64 MiB font budget exceeded") {
			return fmt.Errorf("selected fonts exceed 64 MiB aggregate limit")
		}
	}
	wanted := map[string]bool{}
	for _, name := range names {
		if counts[name] < 4 {
			wanted[name] = true
		}
	}
	if len(wanted) == 0 {
		return nil
	}
	visited := 0
	for _, dir := range officeFontDirs() {
		if dir == "" {
			continue
		}
		err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, walkErr error) error {
			visited++
			if visited > 10000 {
				return fs.SkipAll
			}
			if walkErr != nil || entry.IsDir() || entry.Type()&os.ModeSymlink != 0 {
				return nil
			}
			ext := strings.ToLower(filepath.Ext(path))
			if ext != ".ttf" && ext != ".otf" {
				return nil
			}
			info, err := entry.Info()
			if err != nil || !info.Mode().IsRegular() || info.Size() <= officeFontLimit {
				return nil
			}
			f, err := os.Open(path)
			if err != nil {
				return nil
			}
			// All metadata reads remain bounded by the supported-face limit. The
			// physical length is used only after the selected family is identified.
			face, metadataErr := readOfficeFontInfo(f, officeFontLimit)
			closeErr := f.Close()
			if closeErr != nil {
				return closeErr
			}
			if metadataErr == nil && wanted[face.family] && face.editable() {
				return fmt.Errorf("selected font %q exceeds 32 MiB per-face limit", face.family)
			}
			return nil
		})
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	return nil
}
