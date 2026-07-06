// ABOUTME: Input path resolution for manuscript mode.
// ABOUTME: Handles explicit paths, quoted globs, deterministic ordering, and format validation.
package manuscript

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func ResolveInputs(patterns []string) (InputSet, error) {
	seen := map[string]bool{}
	var paths []string

	for _, pattern := range patterns {
		matches, err := expandInput(pattern)
		if err != nil {
			return InputSet{}, err
		}
		for _, match := range matches {
			cleaned := filepath.Clean(match)
			if !seen[cleaned] {
				seen[cleaned] = true
				paths = append(paths, cleaned)
			}
		}
	}

	if len(paths) == 0 {
		return InputSet{}, fmt.Errorf("no manuscript inputs matched")
	}
	sort.Strings(paths)

	format := ""
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			return InputSet{}, fmt.Errorf("checking input %s: %w", path, err)
		}
		if info.IsDir() {
			return InputSet{}, fmt.Errorf("input is a directory: %s", path)
		}
		thisFormat, err := manuscriptFormat(path)
		if err != nil {
			return InputSet{}, err
		}
		if format == "" {
			format = thisFormat
			continue
		}
		if thisFormat != format {
			return InputSet{}, fmt.Errorf("manuscript inputs must not mix Markdown and org-mode files")
		}
	}
	return InputSet{Format: format, Paths: paths}, nil
}

func expandInput(pattern string) ([]string, error) {
	pattern = expandHome(pattern)
	if hasGlob(pattern) {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid glob %s: %w", pattern, err)
		}
		if len(matches) == 0 {
			return nil, fmt.Errorf("no files match glob: %s", pattern)
		}
		return matches, nil
	}
	if _, err := os.Stat(pattern); err != nil {
		return nil, fmt.Errorf("input not found: %s", pattern)
	}
	return []string{pattern}, nil
}

func expandHome(path string) string {
	if path == "~" {
		if home, err := os.UserHomeDir(); err == nil {
			return home
		}
	}
	if strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, strings.TrimPrefix(path, "~/"))
		}
	}
	return path
}

func hasGlob(path string) bool {
	return strings.ContainsAny(path, "*?[")
}

func manuscriptFormat(path string) (string, error) {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".md":
		return "markdown", nil
	case ".org":
		return "org", nil
	case ".fountain", ".ftn":
		return "", fmt.Errorf("manuscript mode accepts only Markdown or org-mode input, not Fountain")
	default:
		return "", fmt.Errorf("unsupported manuscript input format: %s", path)
	}
}

func ReadJoined(paths []string) (string, error) {
	var builder strings.Builder
	for index, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("reading %s: %w", path, err)
		}
		if index > 0 {
			builder.WriteString("\n\n")
		}
		builder.Write(data)
		builder.WriteString("\n")
	}
	return builder.String(), nil
}
