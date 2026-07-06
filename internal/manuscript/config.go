// ABOUTME: Layered YAML configuration for manuscript rendering.
// ABOUTME: Preserves script.yaml precedence while adding manuscript-specific defaults.
package manuscript

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Title     string `yaml:"title"`
	Subtitle  string `yaml:"subtitle"`
	Author    string `yaml:"author"`
	Date      string `yaml:"date"`
	Version   string `yaml:"version"`
	WordCount string `yaml:"wordcount"`
	Folio     Folio  `yaml:"folio"`
}

type Folio struct {
	Style           string           `yaml:"style"`
	Font            string           `yaml:"font"`
	FontSize        string           `yaml:"font-size"`
	FontWeight      string           `yaml:"font-weight"`
	HeadingFont     string           `yaml:"heading-font"`
	HeadingFontSize string           `yaml:"heading-font-size"`
	Page            string           `yaml:"page"`
	Margin          string           `yaml:"margin"`
	Manuscript      ManuscriptConfig `yaml:"manuscript"`
}

type ManuscriptConfig struct {
	Style             string           `yaml:"style"`
	Page              string           `yaml:"page"`
	Margin            string           `yaml:"margin"`
	Font              string           `yaml:"font"`
	FontSize          string           `yaml:"font-size"`
	FontWeight        string           `yaml:"font-weight"`
	HeadingFont       string           `yaml:"heading-font"`
	HeadingFontSize   string           `yaml:"heading-font-size"`
	HeadingFontWeight string           `yaml:"heading-font-weight"`
	MonoFont          string           `yaml:"mono-font"`
	MonoFontSize      string           `yaml:"mono-font-size"`
	MonoFontWeight    string           `yaml:"mono-font-weight"`
	TitleFont         string           `yaml:"title-font"`
	TitleFontSize     string           `yaml:"title-font-size"`
	TitleFontWeight   string           `yaml:"title-font-weight"`
	SubtitleFont      string           `yaml:"subtitle-font"`
	SubtitleFontSize  string           `yaml:"subtitle-font-size"`
	SubtitleFontStyle string           `yaml:"subtitle-font-style"`
	AuthorFont        string           `yaml:"author-font"`
	AuthorFontSize    string           `yaml:"author-font-size"`
	AuthorAttribution string           `yaml:"author-attribution"`
	DateFont          string           `yaml:"date-font"`
	DateFontSize      string           `yaml:"date-font-size"`
	VersionFont       string           `yaml:"version-font"`
	VersionFontSize   string           `yaml:"version-font-size"`
	WordCountFont     string           `yaml:"wordcount-font"`
	WordCountFontSize string           `yaml:"wordcount-font-size"`
	ContactFont       string           `yaml:"contact-font"`
	ContactFontSize   string           `yaml:"contact-font-size"`
	LineSpacing       string           `yaml:"line-spacing"`
	ParagraphIndent   string           `yaml:"paragraph-indent"`
	ParagraphSpacing  string           `yaml:"paragraph-spacing"`
	PageHeader        PageHeaderConfig `yaml:"page-header"`
	TOC               TOCConfig        `yaml:"toc"`
	Part              HeadingConfig    `yaml:"part"`
	Chapter           HeadingConfig    `yaml:"chapter"`
}

type PageHeaderConfig struct {
	Enabled             bool   `yaml:"enabled"`
	Font                string `yaml:"font"`
	FontSize            string `yaml:"font-size"`
	FontWeight          string `yaml:"font-weight"`
	Format              string `yaml:"format"`
	Align               string `yaml:"align"`
	ContentPaddingAfter string `yaml:"content-padding-after"`
}

type TOCConfig struct {
	Enabled           bool   `yaml:"enabled"`
	Title             string `yaml:"title"`
	Font              string `yaml:"font"`
	FontSize          string `yaml:"font-size"`
	FontWeight        string `yaml:"font-weight"`
	HeadingFont       string `yaml:"heading-font"`
	HeadingFontSize   string `yaml:"heading-font-size"`
	HeadingFontWeight string `yaml:"heading-font-weight"`
	IncludeParts      bool   `yaml:"include-parts"`
	IncludeChapters   bool   `yaml:"include-chapters"`
}

type HeadingConfig struct {
	Position      string `yaml:"position"`
	Align         string `yaml:"align"`
	CaseTransform string `yaml:"case-transform"`
	SpaceAfter    string `yaml:"space-after"`
}

func LoadConfig(sourceDir string, opts Options) (Config, error) {
	root, err := projectRoot()
	if err != nil {
		return Config{}, err
	}

	global := filepath.Join(os.Getenv("HOME"), ".config", "first-folio", "script.yaml")
	local := filepath.Join(sourceDir, "script.yaml")
	globalMap, err := readYAMLIfExists(global)
	if err != nil {
		return Config{}, err
	}
	localMap, err := readYAMLIfExists(local)
	if err != nil {
		return Config{}, err
	}

	style := selectedStyle(opts, globalMap, localMap)
	base, err := readYAMLRequired(filepath.Join(root, "presets", "british-manuscript.yaml"))
	if err != nil {
		return Config{}, err
	}
	if style == "us" || style == "american" {
		us, err := readYAMLRequired(filepath.Join(root, "presets", "us-overrides-manuscript.yaml"))
		if err != nil {
			return Config{}, err
		}
		deepMerge(base, us)
	}

	deepMerge(base, globalMap)
	mergeStyleConfig(base, filepath.Join(os.Getenv("HOME"), ".config", "first-folio"), style)
	deepMerge(base, localMap)
	mergeStyleConfig(base, sourceDir, style)

	var cfg Config
	raw, err := yaml.Marshal(base)
	if err != nil {
		return Config{}, fmt.Errorf("serializing merged config: %w", err)
	}
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return Config{}, fmt.Errorf("parsing merged config: %w", err)
	}
	normalizeConfig(&cfg)
	return cfg, nil
}

func selectedStyle(opts Options, configs ...map[string]any) string {
	style := "british"
	for _, cfg := range configs {
		if value := stringAt(cfg, "folio", "style"); value != "" {
			style = strings.ToLower(value)
		}
		if value := stringAt(cfg, "folio", "manuscript", "style"); value != "" {
			style = strings.ToLower(value)
		}
	}
	if opts.Style != "" {
		style = strings.ToLower(opts.Style)
	}
	if style == "american" {
		return "us"
	}
	return style
}

func normalizeConfig(cfg *Config) {
	folio := &cfg.Folio
	ms := &folio.Manuscript
	fill(&folio.Page, "a4")
	fill(&folio.Margin, "25mm")
	fill(&folio.Font, "Libertinus Serif")
	fill(&folio.FontSize, "12pt")
	fill(&folio.HeadingFont, folio.Font)
	fill(&folio.HeadingFontSize, folio.FontSize)
	fill(&ms.Page, folio.Page)
	fill(&ms.Margin, folio.Margin)
	fill(&ms.Font, folio.Font)
	fill(&ms.FontSize, folio.FontSize)
	fill(&ms.HeadingFont, folio.HeadingFont)
	fill(&ms.HeadingFontSize, folio.HeadingFontSize)
	fill(&ms.MonoFont, "Libertinus Mono")
	fill(&ms.MonoFontSize, ms.FontSize)
	fill(&ms.TitleFont, ms.HeadingFont)
	fill(&ms.TitleFontSize, "20pt")
	fill(&ms.SubtitleFont, ms.HeadingFont)
	fill(&ms.SubtitleFontSize, "14pt")
	fill(&ms.AuthorFont, ms.HeadingFont)
	fill(&ms.AuthorFontSize, ms.FontSize)
	fill(&ms.AuthorAttribution, "by")
	fill(&ms.DateFont, ms.HeadingFont)
	fill(&ms.DateFontSize, "10pt")
	fill(&ms.VersionFont, ms.HeadingFont)
	fill(&ms.VersionFontSize, "10pt")
	fill(&ms.WordCountFont, ms.HeadingFont)
	fill(&ms.WordCountFontSize, "10pt")
	fill(&ms.ContactFont, ms.HeadingFont)
	fill(&ms.ContactFontSize, "10pt")
	fill(&ms.LineSpacing, "1.5")
	fill(&ms.ParagraphIndent, "10mm")
	fill(&ms.ParagraphSpacing, "0")
	if ms.ParagraphSpacing == "0" {
		ms.ParagraphSpacing = "0pt"
	}
	fill(&ms.PageHeader.Font, ms.HeadingFont)
	fill(&ms.PageHeader.FontSize, "10pt")
	fill(&ms.PageHeader.FontWeight, "regular")
	fill(&ms.PageHeader.Format, "[author] / [title] / [page]")
	fill(&ms.PageHeader.Align, "right")
	fill(&ms.PageHeader.ContentPaddingAfter, "10mm")
	fill(&ms.TOC.Title, "Contents")
	fill(&ms.TOC.Font, ms.HeadingFont)
	fill(&ms.TOC.FontSize, "11pt")
	fill(&ms.TOC.HeadingFont, ms.HeadingFont)
	fill(&ms.TOC.HeadingFontSize, "16pt")
	fill(&ms.Part.Align, "center")
	fill(&ms.Part.CaseTransform, "as-written")
	fill(&ms.Chapter.Align, "center")
	fill(&ms.Chapter.Position, "one-third")
	fill(&ms.Chapter.CaseTransform, "as-written")
	fill(&ms.Chapter.SpaceAfter, "2em")
}

func fill(target *string, value string) {
	if *target == "" {
		*target = value
	}
}

func mergeStyleConfig(base map[string]any, dir string, style string) {
	if dir == "" {
		return
	}
	suffix := "british"
	if style == "us" || style == "american" {
		suffix = "us"
	}
	path := filepath.Join(dir, "script-"+suffix+".yaml")
	if cfg, err := readYAMLIfExists(path); err == nil {
		deepMerge(base, cfg)
	}
}

func readYAMLRequired(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	return parseYAML(path, data)
}

func readYAMLIfExists(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]any{}, nil
		}
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	return parseYAML(path, data)
}

func parseYAML(path string, data []byte) (map[string]any, error) {
	out := map[string]any{}
	if err := yaml.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	return out, nil
}

func deepMerge(base map[string]any, overlay map[string]any) {
	for key, value := range overlay {
		baseChild, baseOK := base[key].(map[string]any)
		overlayChild, overlayOK := value.(map[string]any)
		if baseOK && overlayOK {
			deepMerge(baseChild, overlayChild)
			continue
		}
		base[key] = value
	}
}

func stringAt(node map[string]any, parts ...string) string {
	var current any = node
	for _, part := range parts {
		asMap, ok := current.(map[string]any)
		if !ok {
			return ""
		}
		current = asMap[part]
	}
	value, ok := current.(string)
	if !ok {
		return ""
	}
	return value
}

func projectRoot() (string, error) {
	if root := os.Getenv("FIRST_FOLIO_ROOT"); root != "" {
		return root, nil
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("finding working directory: %w", err)
	}
	for dir := wd; ; dir = filepath.Dir(dir) {
		if _, err := os.Stat(filepath.Join(dir, "presets", "british-script.yaml")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
	}
	return "", fmt.Errorf("cannot find First Folio project root")
}
