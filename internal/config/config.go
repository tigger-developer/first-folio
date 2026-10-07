// ABOUTME: Loads and deep-merges shared First Folio YAML configuration.
// ABOUTME: Provides dotted, boolean, font, and typed access for every mode.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	folio "github.com/tigger-developer/first-folio"
	"gopkg.in/yaml.v3"
)

type Mode string

const (
	ModeScript     Mode = "script"
	ModeLetter     Mode = "letter"
	ModeManuscript Mode = "manuscript"
)

type Options struct {
	Mode     Mode
	Home     string
	LocalDir string
	CLI      map[string]any
	// Source is the first Markdown manuscript's configuration, below local files.
	Source map[string]any
}

type Config struct {
	data     map[string]any
	warnings []string
}

func Load(opts Options) (Config, error) {
	if opts.Mode == "" {
		opts.Mode = ModeScript
	}
	base, err := readEmbedded("presets/british.yaml")
	if err != nil {
		return Config{}, err
	}

	globalDir := filepath.Join(opts.Home, ".config", "first-folio")
	global, err := readOptional(filepath.Join(globalDir, "script.yaml"))
	if err != nil {
		return Config{}, err
	}
	// #20: walk upward from LocalDir looking for script.yaml, bounded at the
	// user's home directory. This lets a manuscript with subdirectory inputs
	// (e.g. `folio manuscript 'part?/ch??.md' out.pdf` where the first input
	// resolves to `part0/ch00.md`) still discover the project-root script.yaml.
	// The nearest hit wins -- no merging of multiple walked files.
	localScriptPath := findScriptYAML(opts.LocalDir, opts.Home)
	local, err := readOptional(localScriptPath)
	if err != nil {
		return Config{}, err
	}
	style := selectStyle(opts.Mode, global, opts.Source, local, opts.CLI)

	name := ""
	switch style {
	case "us":
		name = "presets/us.yaml"
	case "screenplay":
		if opts.Mode != ModeManuscript {
			name = "presets/screenplay.yaml"
		}
	}
	if name != "" {
		override, err := readEmbedded(name)
		if err != nil {
			return Config{}, err
		}
		deepMerge(base, override)
	}

	base = runningMatterLayer(base, opts.Mode)
	warnings := []string{}
	warningSeen := map[string]bool{}
	mergeLayer := func(layer map[string]any) {
		if opts.Mode == ModeManuscript {
			for _, warning := range legacyRunningMatterWarnings(layer) {
				if !warningSeen[warning] {
					warnings = append(warnings, warning)
					warningSeen[warning] = true
				}
			}
		}
		deepMerge(base, runningMatterLayer(layer, opts.Mode))
	}
	mergeFile := func(path string) error {
		layer, err := readOptional(path)
		if err != nil {
			return err
		}
		mergeLayer(layer)
		return nil
	}
	mergeLayer(global)
	if err := mergeFile(filepath.Join(globalDir, "script-"+styleSuffix(style)+".yaml")); err != nil {
		return Config{}, err
	}
	mergeLayer(opts.Source)
	mergeLayer(local)
	if localScriptPath != "" {
		if err := mergeFile(filepath.Join(filepath.Dir(localScriptPath), "script-"+styleSuffix(style)+".yaml")); err != nil {
			return Config{}, err
		}
	}
	mergeLayer(cliLayer(base, opts.CLI))
	setPath(base, "folio.style", style)
	if opts.Mode == ModeManuscript {
		setPath(base, "folio.manuscript.style", style)
	}
	if opts.Mode == ModeScript {
		for _, field := range []string{"enabled", "skip-header", "skip-footer"} {
			path := "folio.title-page." + field
			value, _ := (Config{data: base}).Get(path)
			if _, ok := value.(bool); !ok {
				return Config{}, fmt.Errorf("%s must be a boolean; got %#v", path, value)
			}
		}
	}
	if opts.Mode != ModeLetter {
		if err := validateRunningMatter(base); err != nil {
			return Config{}, err
		}
	}
	if err := validateFonts(base, opts.Mode); err != nil {
		return Config{}, err
	}
	if err := validateAlignments(base, opts.Mode); err != nil {
		return Config{}, err
	}
	if opts.Mode == ModeManuscript {
		value, _ := (Config{data: base}).Get("folio.manuscript.hyphenation")
		switch value {
		case true, false, "auto", "true", "false":
		default:
			return Config{}, fmt.Errorf("folio.manuscript.hyphenation must be true, false, or auto; got %#v", value)
		}
	}
	return Config{data: base, warnings: warnings}, nil
}

func (c Config) Get(path string) (any, bool) {
	var node any = c.data
	for _, part := range strings.Split(path, ".") {
		mapping, ok := node.(map[string]any)
		if !ok {
			return nil, false
		}
		node, ok = mapping[part]
		if !ok {
			return nil, false
		}
	}
	return node, true
}

func (c Config) String(path string, fallback string) string {
	value, ok := c.Get(path)
	if !ok || value == nil {
		return fallback
	}
	return fmt.Sprint(value)
}

func (c Config) Bool(path string, fallback bool) bool {
	value, ok := c.Get(path)
	if !ok {
		return fallback
	}
	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		parsed, err := strconv.ParseBool(strings.ToLower(typed))
		if err == nil {
			return parsed
		}
	}
	return fallback
}

func (c Config) Decode(target any) error {
	raw, err := yaml.Marshal(c.data)
	if err != nil {
		return fmt.Errorf("serializing merged config: %w", err)
	}
	if err := yaml.Unmarshal(raw, target); err != nil {
		return fmt.Errorf("decoding merged config: %w", err)
	}
	return nil
}

func readEmbedded(name string) (map[string]any, error) {
	raw, err := folio.Assets.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("loading embedded config %s: %w", name, err)
	}
	return parseYAML(name, raw)
}

// findScriptYAML walks upward from startDir looking for a script.yaml file.
// Returns the absolute path of the nearest match, or "" if none is found before
// crossing the home boundary (or reaching the filesystem root). startDir="" or
// home="" short-circuits to a same-directory-only lookup for the given startDir.
// The home boundary prevents unrelated reads from parent projects, /etc, or
// other users' home directories.
func findScriptYAML(startDir, home string) string {
	if startDir == "" {
		return ""
	}
	abs, err := filepath.Abs(startDir)
	if err != nil {
		return ""
	}
	homeAbs, _ := filepath.Abs(home)
	dir := abs
	for {
		candidate := filepath.Join(dir, "script.yaml")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		// Stop before crossing the home boundary (but allow the home dir itself
		// to be checked, since a manuscript project rooted at $HOME is valid).
		if homeAbs != "" && dir == homeAbs {
			return ""
		}
		dir = parent
	}
}

func readOptional(path string) (map[string]any, error) {
	if path == "" {
		return nil, nil
	}
	raw, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading config %s: %w", path, err)
	}
	return parseYAML(path, raw)
}

func parseYAML(name string, raw []byte) (map[string]any, error) {
	data := map[string]any{}
	if err := yaml.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("cannot parse YAML %s: %w", name, err)
	}
	return data, nil
}

func deepMerge(base map[string]any, overlay map[string]any) {
	for key, value := range overlay {
		if child, ok := value.(map[string]any); ok {
			if existing, ok := base[key].(map[string]any); ok {
				deepMerge(existing, child)
				continue
			}
		}
		base[key] = value
	}
}

func selectStyle(mode Mode, configs ...map[string]any) string {
	style := "british"
	for _, data := range configs[:len(configs)-1] {
		cfg := Config{data: data}
		if value := cfg.String("folio.style", ""); value != "" {
			style = normalizeStyle(value)
		}
		if mode == ModeManuscript {
			if value := cfg.String("folio.manuscript.style", ""); value != "" {
				style = normalizeStyle(value)
			}
		}
	}
	cli := configs[len(configs)-1]
	if value, ok := cli["style"]; ok && value != nil && strings.TrimSpace(fmt.Sprint(value)) != "" {
		style = normalizeStyle(fmt.Sprint(value))
	}
	return style
}

func normalizeStyle(style string) string {
	switch strings.ToLower(style) {
	case "us", "american":
		return "us"
	case "uk", "british":
		return "british"
	case "screenplay":
		return "screenplay"
	default:
		return strings.ToLower(style)
	}
}

func styleSuffix(style string) string {
	return style
}

// cliLayer preserves the existing flag contract while allowing running-matter
// aliases to be normalized at CLI precedence rather than against merged data.
func cliLayer(base map[string]any, values map[string]any) map[string]any {
	layer := map[string]any{}
	for key, value := range values {
		if key == "style" || value == nil {
			continue
		}
		path := "folio." + key
		switch key {
		case "font":
			path = "folio.font.family"
		case "font-size":
			path = "folio.font.size"
		default:
			if _, ok := (Config{data: base}).Get(path); !ok {
				continue
			}
		}
		setPath(layer, path, value)
	}
	return layer
}

func setPath(data map[string]any, path string, value any) {
	parts := strings.Split(path, ".")
	node := data
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
