// ABOUTME: Layered YAML configuration for manuscript rendering.
// ABOUTME: Preserves script.yaml precedence while adding manuscript-specific defaults.
package manuscript

import (
	"os"

	sharedconfig "github.com/tigger-developer/first-folio/internal/config"
)

type Config struct {
	Title             string `yaml:"title"`
	Subtitle          string `yaml:"subtitle"`
	Author            string `yaml:"author"`
	Attribution       string `yaml:"attribution"`
	AuthorAttribution string `yaml:"author-attribution"`
	Date              string `yaml:"date"`
	Version           string `yaml:"version"`
	WordCount         string `yaml:"wordcount"`
	ContactName       string `yaml:"contact-name"`
	Address           string `yaml:"address"`
	Phone             string `yaml:"phone"`
	Email             string `yaml:"email"`
	Website           string `yaml:"website"`
	Folio             Folio  `yaml:"folio"`
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
	Style               string            `yaml:"style"`
	Page                string            `yaml:"page"`
	Margin              string            `yaml:"margin"`
	Font                string            `yaml:"font"`
	FontSize            string            `yaml:"font-size"`
	FontWeight          string            `yaml:"font-weight"`
	HeadingFont         string            `yaml:"heading-font"`
	HeadingFontSize     string            `yaml:"heading-font-size"`
	HeadingFontWeight   string            `yaml:"heading-font-weight"`
	MonoFont            string            `yaml:"mono-font"`
	MonoFontSize        string            `yaml:"mono-font-size"`
	MonoFontWeight      string            `yaml:"mono-font-weight"`
	TitleFont           string            `yaml:"title-font"`
	TitleFontSize       string            `yaml:"title-font-size"`
	TitleFontWeight     string            `yaml:"title-font-weight"`
	SubtitleFont        string            `yaml:"subtitle-font"`
	SubtitleFontSize    string            `yaml:"subtitle-font-size"`
	SubtitleFontWeight  string            `yaml:"subtitle-font-weight"`
	SubtitleFontStyle   string            `yaml:"subtitle-font-style"`
	AuthorFont          string            `yaml:"author-font"`
	AuthorFontSize      string            `yaml:"author-font-size"`
	AuthorFontWeight    string            `yaml:"author-font-weight"`
	Attribution         string            `yaml:"attribution"`
	AuthorAttribution   string            `yaml:"author-attribution"`
	DateFont            string            `yaml:"date-font"`
	DateFontSize        string            `yaml:"date-font-size"`
	DateFontWeight      string            `yaml:"date-font-weight"`
	DateFormat          string            `yaml:"date-format"`
	VersionFont         string            `yaml:"version-font"`
	VersionFontSize     string            `yaml:"version-font-size"`
	VersionFontWeight   string            `yaml:"version-font-weight"`
	WordCountFont       string            `yaml:"wordcount-font"`
	WordCountFontSize   string            `yaml:"wordcount-font-size"`
	WordCountFontWeight string            `yaml:"wordcount-font-weight"`
	ContactFont         string            `yaml:"contact-font"`
	ContactFontSize     string            `yaml:"contact-font-size"`
	ContactFontWeight   string            `yaml:"contact-font-weight"`
	LineSpacing         string            `yaml:"line-spacing"`
	ParagraphIndent     string            `yaml:"paragraph-indent"`
	ParagraphSpacing    string            `yaml:"paragraph-spacing"`
	PageHeader          PageHeaderConfig  `yaml:"page-header"`
	PageFooter          PageFooterConfig  `yaml:"page-footer"`
	Gutter              string            `yaml:"gutter"`
	TOC                 TOCConfig         `yaml:"toc"`
	TitlePage           TitlePageConfig   `yaml:"title-page"`
	SceneBreak          SceneBreakConfig  `yaml:"scene-break"`
	List                SpacedBlockConfig `yaml:"list"`
	Table               SpacedBlockConfig `yaml:"table"`
	CodeBlock           SpacedBlockConfig `yaml:"code-block"`
	Part                HeadingConfig     `yaml:"part"`
	Chapter             HeadingConfig     `yaml:"chapter"`
}

type TitlePageConfig struct {
	Enabled            bool                    `yaml:"enabled"`
	PageNumber         bool                    `yaml:"page-number"`
	IncludeTitle       bool                    `yaml:"include-title"`
	IncludeSubtitle    bool                    `yaml:"include-subtitle"`
	IncludeAuthor      bool                    `yaml:"include-author"`
	IncludeDate        bool                    `yaml:"include-date"`
	IncludeWordCount   bool                    `yaml:"include-wordcount"`
	IncludeContactName bool                    `yaml:"include-contact-name"`
	IncludeAddress     bool                    `yaml:"include-address"`
	IncludePhone       bool                    `yaml:"include-phone"`
	IncludeEmail       bool                    `yaml:"include-email"`
	IncludeWebsite     bool                    `yaml:"include-website"`
	IncludeVersion     bool                    `yaml:"include-version"`
	TitleBlockAlign    string                  `yaml:"title-block-align"`
	FooterAlign        string                  `yaml:"footer-align"`
	Title              TitlePageItemConfig     `yaml:"title"`
	Subtitle           TitlePageItemConfig     `yaml:"subtitle"`
	Author             TitlePageItemConfig     `yaml:"author"`
	Date               TitlePageItemConfig     `yaml:"date"`
	WordCount          TitlePageItemConfig     `yaml:"wordcount"`
	Version            TitlePageItemConfig     `yaml:"version"`
	Contact            TitlePageItemConfig     `yaml:"contact"`
}

type TitlePageItemConfig struct {
	Align string `yaml:"align"`
}

type PageHeaderConfig struct {
	Enabled             bool   `yaml:"enabled"`
	Font                string `yaml:"font"`
	FontSize            string `yaml:"font-size"`
	FontWeight          string `yaml:"font-weight"`
	Format              string `yaml:"format"`
	Align               string `yaml:"align"`
	DistanceFromEdge    string `yaml:"distance-from-edge"`
	ContentPaddingAfter string `yaml:"content-padding-after"`
}

// PageFooterConfig mirrors PageHeaderConfig. Fields left empty inherit from PageHeaderConfig
// during normalization; PageHeader values in turn inherit from root folio settings.
// Enabled is a *bool so normalizeConfig can distinguish "unset" (default true) from
// an explicit `enabled: false`.
type PageFooterConfig struct {
	Enabled             *bool  `yaml:"enabled,omitempty"`
	Font                string `yaml:"font"`
	FontSize            string `yaml:"font-size"`
	FontWeight          string `yaml:"font-weight"`
	Format              string `yaml:"format"`
	Align               string `yaml:"align"`
	DistanceFromEdge    string `yaml:"distance-from-edge"`
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
	IncludeSections   bool   `yaml:"include-sections"`
	DotLeaders        bool   `yaml:"dot-leaders"`
	PageNumbers       bool   `yaml:"page-numbers"`
	PageBreakBefore   bool   `yaml:"page-break-before"`
	LineSpacing       string `yaml:"line-spacing"`
	PartGapBefore     string `yaml:"part-gap-before"`
	PartBold          bool   `yaml:"part-bold"`
}

type SceneBreakConfig struct {
	Marker string `yaml:"marker"`
}

type SpacedBlockConfig struct {
	SpaceBefore string `yaml:"space-before"`
	SpaceAfter  string `yaml:"space-after"`
}

type HeadingConfig struct {
	PageBreakBefore bool   `yaml:"page-break-before"`
	BlankPageBefore bool   `yaml:"blank-page-before"`
	BlankPageAfter  bool   `yaml:"blank-page-after"`
	VerticalAlign   string `yaml:"vertical-align"`
	Position        string `yaml:"position"`
	Align           string `yaml:"align"`
	CaseTransform   string `yaml:"case-transform"`
	SpaceAfter      string `yaml:"space-after"`
}

func LoadConfig(sourceDir string, opts Options) (Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Config{}, err
	}
	merged, err := sharedconfig.Load(sharedconfig.Options{
		Mode:     sharedconfig.ModeManuscript,
		Home:     home,
		LocalDir: sourceDir,
		CLI:      map[string]any{"style": opts.Style},
	})
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := merged.Decode(&cfg); err != nil {
		return Config{}, err
	}
	normalizeConfig(&cfg)
	if err := validateConfig(&cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func validateConfig(cfg *Config) error {
	ms := &cfg.Folio.Manuscript
	if _, err := ParsePageSpec(ms.Page); err != nil {
		return err
	}
	if _, err := ParseHeaderFooterAlign(ms.PageHeader.Align); err != nil {
		return err
	}
	if _, err := ParseHeaderFooterAlign(ms.PageFooter.Align); err != nil {
		return err
	}
	if _, err := TitleItemAlign(ms.TitlePage.TitleBlockAlign); err != nil {
		return err
	}
	if _, err := TitleItemAlign(ms.TitlePage.FooterAlign); err != nil {
		return err
	}
	for _, item := range []struct {
		name  string
		value string
	}{
		{"title", ms.TitlePage.Title.Align},
		{"subtitle", ms.TitlePage.Subtitle.Align},
		{"author", ms.TitlePage.Author.Align},
		{"date", ms.TitlePage.Date.Align},
		{"wordcount", ms.TitlePage.WordCount.Align},
		{"version", ms.TitlePage.Version.Align},
		{"contact", ms.TitlePage.Contact.Align},
	} {
		if _, err := TitleItemAlign(item.value); err != nil {
			return err
		}
		_ = item.name
	}
	return nil
}

func normalizeConfig(cfg *Config) {
	folio := &cfg.Folio
	ms := &folio.Manuscript
	fill(&folio.Page, "a4")
	fill(&folio.Margin, "25mm")
	fill(&folio.Font, "Libertinus Serif")
	fill(&folio.FontSize, "12pt")
	fill(&folio.FontWeight, "regular")
	fill(&folio.HeadingFont, folio.Font)
	fill(&folio.HeadingFontSize, folio.FontSize)
	fill(&ms.Page, folio.Page)
	fill(&ms.Margin, folio.Margin)
	fill(&ms.Font, folio.Font)
	fill(&ms.FontSize, folio.FontSize)
	fill(&ms.FontWeight, folio.FontWeight)
	fill(&ms.HeadingFont, folio.HeadingFont)
	fill(&ms.HeadingFontSize, folio.HeadingFontSize)
	fill(&ms.HeadingFontWeight, "regular")
	fill(&ms.MonoFont, "Libertinus Mono")
	fill(&ms.MonoFontSize, ms.FontSize)
	fill(&ms.MonoFontWeight, "regular")
	fill(&ms.TitleFont, ms.HeadingFont)
	fill(&ms.TitleFontSize, "20pt")
	fill(&ms.TitleFontWeight, "bold")
	fill(&ms.SubtitleFont, ms.HeadingFont)
	fill(&ms.SubtitleFontSize, "14pt")
	fill(&ms.SubtitleFontWeight, "regular")
	fill(&ms.AuthorFont, ms.HeadingFont)
	fill(&ms.AuthorFontSize, ms.FontSize)
	fill(&ms.AuthorFontWeight, "regular")
	fill(&ms.DateFont, ms.HeadingFont)
	fill(&ms.DateFontSize, "10pt")
	fill(&ms.DateFontWeight, "regular")
	fill(&ms.DateFormat, "2 January 2006")
	fill(&ms.VersionFont, ms.HeadingFont)
	fill(&ms.VersionFontSize, "10pt")
	fill(&ms.VersionFontWeight, "regular")
	fill(&ms.WordCountFont, ms.HeadingFont)
	fill(&ms.WordCountFontSize, "10pt")
	fill(&ms.WordCountFontWeight, "regular")
	fill(&ms.ContactFont, ms.HeadingFont)
	fill(&ms.ContactFontSize, "10pt")
	fill(&ms.ContactFontWeight, "regular")
	fill(&ms.LineSpacing, "1.5")
	fill(&ms.ParagraphIndent, "10mm")
	fill(&ms.ParagraphSpacing, "0")
	if ms.ParagraphSpacing == "0" {
		ms.ParagraphSpacing = "0pt"
	}
	fill(&ms.PageHeader.Font, ms.HeadingFont)
	fill(&ms.PageHeader.FontSize, "10pt")
	fill(&ms.PageHeader.FontWeight, "regular")
	fill(&ms.PageHeader.Format, "[title] • [chapter] • [author]")
	fill(&ms.PageHeader.Align, "left-right")
	fill(&ms.PageHeader.DistanceFromEdge, ms.Margin)
	fill(&ms.PageHeader.ContentPaddingAfter, "10mm")
	if ms.PageFooter.Enabled == nil {
		t := true
		ms.PageFooter.Enabled = &t
	}
	fill(&ms.PageFooter.Font, ms.PageHeader.Font)
	fill(&ms.PageFooter.FontSize, ms.PageHeader.FontSize)
	fill(&ms.PageFooter.FontWeight, ms.PageHeader.FontWeight)
	fill(&ms.PageFooter.Format, "[page]")
	fill(&ms.PageFooter.Align, "center")
	fill(&ms.PageFooter.DistanceFromEdge, ms.Margin)
	fill(&ms.PageFooter.ContentPaddingAfter, "10mm")
	fill(&ms.Gutter, "0mm")
	fill(&ms.TitlePage.Contact.Align, "top-left")
	fill(&ms.TOC.Title, "Contents")
	fill(&ms.TOC.Font, ms.HeadingFont)
	fill(&ms.TOC.FontSize, "11pt")
	fill(&ms.TOC.FontWeight, "regular")
	fill(&ms.TOC.HeadingFont, ms.HeadingFont)
	fill(&ms.TOC.HeadingFontSize, "16pt")
	fill(&ms.TOC.HeadingFontWeight, "bold")
	fill(&ms.TOC.LineSpacing, "1.15em")
	fill(&ms.TOC.PartGapBefore, "0.5em")
	fill(&ms.SceneBreak.Marker, "#")
	fill(&ms.List.SpaceBefore, "0.5em")
	fill(&ms.List.SpaceAfter, "0.5em")
	fill(&ms.Table.SpaceBefore, "0.5em")
	fill(&ms.Table.SpaceAfter, "0.5em")
	fill(&ms.CodeBlock.SpaceBefore, "0.5em")
	fill(&ms.CodeBlock.SpaceAfter, "0.5em")
	fill(&ms.TitlePage.TitleBlockAlign, "center")
	fill(&ms.TitlePage.FooterAlign, "center")
	fill(&ms.Part.Align, "center")
	fill(&ms.Part.VerticalAlign, "center")
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
