// ABOUTME: Projects dotted Org script frontmatter into the shared YAML configuration tree.
// ABOUTME: Warns about unknown FOLIO fields while keeping configuration out of play content.
package app

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/tigger-developer/first-folio/internal/config"
	nativesource "github.com/tigger-developer/first-folio/internal/source"
	"gopkg.in/yaml.v3"
)

const orgConfigurationPrefixLimit = 1 << 20

func readOrgConfigurationPrefix(path string) (string, error) {
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("file not found: %s", path)
	}
	if err != nil {
		return "", fmt.Errorf("cannot open %s: %w", path, err)
	}
	defer func() {
		// The file is read-only; a close failure cannot invalidate consumed bytes.
		_ = file.Close()
	}()
	raw, err := io.ReadAll(io.LimitReader(file, orgConfigurationPrefixLimit+1))
	if err != nil {
		return "", fmt.Errorf("reading Org frontmatter from %s: %w", path, err)
	}
	if len(raw) > orgConfigurationPrefixLimit {
		prefix := string(raw[:orgConfigurationPrefixLimit])
		if !orgConfigurationEnded(prefix) {
			return "", fmt.Errorf("%s: Org configuration frontmatter exceeds %d bytes", path, orgConfigurationPrefixLimit)
		}
		return prefix, nil
	}
	return string(raw), nil
}

func orgConfigurationEnded(raw string) bool {
	for _, line := range nativesource.Lines(raw) {
		token := nativesource.Recognize("org", line)
		if token.Kind != "keyword" && token.Kind != "blank" {
			return true
		}
	}
	return false
}

func orgScriptSourceConfig(raw string, path string) (map[string]any, []string, error) {
	known, err := config.ScriptSourcePaths()
	if err != nil {
		return nil, nil, err
	}
	result := map[string]any{}
	warnings := []string{}
	seen := map[string]int{}
	for index, line := range nativesource.Lines(raw) {
		token := nativesource.Recognize("org", line)
		if token.Kind != "keyword" && token.Kind != "blank" {
			break
		}
		if token.Kind != "keyword" {
			continue
		}
		field := strings.ToLower(token.Name)
		isFolio := folioFrontmatterPrefix(field)
		isRender := strings.HasPrefix(field, "render.")
		if !isFolio && !isRender {
			continue
		}
		lineNumber := index + 1
		if previous, duplicate := seen[field]; duplicate {
			return nil, nil, fmt.Errorf("%s:%d: duplicate Org configuration field %q; first declared on line %d", path, lineNumber, field, previous)
		}
		seen[field] = lineNumber
		if _, ok := known[field]; !ok {
			if isFolio {
				warnings = append(warnings, fmt.Sprintf("%s:%d: unrecognized FOLIO frontmatter field %q", path, lineNumber, field))
			}
			continue
		}
		var value any
		if err := yaml.Unmarshal([]byte(token.Text), &value); err != nil {
			return nil, nil, fmt.Errorf("%s:%d: parsing Org configuration field %q: %w", path, lineNumber, field, err)
		}
		switch value.(type) {
		case map[string]any, []any:
			return nil, nil, fmt.Errorf("%s:%d: Org configuration field %q must contain one YAML scalar", path, lineNumber, field)
		}
		setOrgSourcePath(result, field, value)
	}
	return result, warnings, nil
}

func folioFrontmatterPrefix(field string) bool {
	return field == "folio" || strings.HasPrefix(field, "folio.") || strings.HasPrefix(field, "folio-") || strings.HasPrefix(field, "folio_")
}

func setOrgSourcePath(root map[string]any, path string, value any) {
	parts := strings.Split(path, ".")
	node := root
	for _, part := range parts[:len(parts)-1] {
		child, ok := node[part].(map[string]any)
		if !ok {
			child = map[string]any{}
			node[part] = child
		}
		node = child
	}
	node[parts[len(parts)-1]] = value
}
