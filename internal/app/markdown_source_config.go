// ABOUTME: Extracts folio and render mappings from bounded Markdown script frontmatter.
// ABOUTME: Projects them into the shared source configuration layer without emitting metadata as body text.
package app

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

const markdownConfigurationPrefixLimit = 1 << 20

func markdownScriptSourceConfig(path string) (map[string]any, error) {
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("file not found: %s", path)
	}
	if err != nil {
		return nil, fmt.Errorf("cannot open %s: %w", path, err)
	}
	defer func() {
		_ = file.Close()
	}()
	raw, err := io.ReadAll(io.LimitReader(file, markdownConfigurationPrefixLimit+1))
	if err != nil {
		return nil, fmt.Errorf("reading Markdown frontmatter from %s: %w", path, err)
	}
	normalized := strings.ReplaceAll(string(raw), "\r\n", "\n")
	if !strings.HasPrefix(normalized, "---\n") {
		return nil, nil
	}
	rest := normalized[len("---\n"):]
	end := strings.Index(rest, "\n---\n")
	if end < 0 && strings.HasSuffix(rest, "\n---") {
		end = len(rest) - len("\n---")
	}
	if end < 0 {
		if len(raw) > markdownConfigurationPrefixLimit {
			return nil, fmt.Errorf("%s: Markdown configuration frontmatter exceeds %d bytes", path, markdownConfigurationPrefixLimit)
		}
		return nil, fmt.Errorf("%s: Markdown frontmatter has no closing --- delimiter", path)
	}
	content := rest[:end]
	values := map[string]any{}
	if err := yaml.Unmarshal([]byte(content), &values); err != nil {
		return nil, fmt.Errorf("%s: parsing Markdown frontmatter: %w", path, err)
	}
	source := map[string]any{}
	for _, key := range []string{"folio", "render"} {
		value, exists := values[key]
		if !exists {
			continue
		}
		mapping, ok := value.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("%s: frontmatter %s must be a mapping", path, key)
		}
		source[key] = mapping
	}
	return source, nil
}
