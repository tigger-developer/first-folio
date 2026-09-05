// ABOUTME: Layered YAML configuration for manuscript rendering.
// ABOUTME: Preserves script.yaml precedence while adding manuscript-specific defaults.
package manuscript

import (
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"

	sharedconfig "github.com/tigger-developer/first-folio/internal/config"
	"gopkg.in/yaml.v3"
)

// BlankPageMode describes how the blank-page-before / blank-page-after field is resolved.
// Accepted YAML values are the bools `true`/`false` and the strings `"true"`, `"false"`,
// `"enforce-right"`, `"enforce-left"`. An empty value behaves as `"false"` (no blank page).
type BlankPageMode string

const (
	BlankPageFalse        BlankPageMode = ""
	BlankPageTrue         BlankPageMode = "true"
	BlankPageEnforceRight BlankPageMode = "enforce-right"
	BlankPageEnforceLeft  BlankPageMode = "enforce-left"
)

func (b *BlankPageMode) UnmarshalYAML(node *yaml.Node) error {
	if node.Tag == "!!bool" {
		var boolVal bool
		if err := node.Decode(&boolVal); err != nil {
			return err
		}
		if boolVal {
			*b = BlankPageTrue
		} else {
			*b = BlankPageFalse
		}
		return nil
	}
	var strVal string
	if err := node.Decode(&strVal); err != nil {
		return err
	}
	switch strVal {
	case "", "false", "no":
		*b = BlankPageFalse
	case "true", "yes":
		*b = BlankPageTrue
	case "enforce-right":
		*b = BlankPageEnforceRight
	case "enforce-left":
		*b = BlankPageEnforceLeft
	default:
		return fmt.Errorf("invalid blank-page value %q: expected true, false, enforce-right, or enforce-left", strVal)
	}
	return nil
}

// TypstDirective returns the Typst source line(s) to emit for this mode, or the empty string
// when no blank/parity break is required. For enforce-right / enforce-left, the emitted code
// records the page number of the anticipated parity blank (pg+1 when the current page's parity
// requires a skip) into folio-skip-header-pages / folio-skip-footer-pages so the running
// header/footer stays hidden on that inserted blank. The CURRENT page number is deliberately
// NOT added -- it is either a content page (mid-block) or a phantom fresh page after a
// previous section's pagebreak, and adding it would falsely suppress a real body page's
// running header/footer.
func (b BlankPageMode) TypstDirective() string {
	switch b {
	case BlankPageTrue:
		return "#folio-blank-page()"
	case BlankPageEnforceRight:
		return `#context { let pg = counter(page).at(here()).first(); if calc.odd(pg) { state("folio-skip-header-pages", ()).update(pages => pages + (pg + 1,)); state("folio-skip-footer-pages", ()).update(pages => pages + (pg + 1,)) } }
#pagebreak(weak: true, to: "odd")`
	case BlankPageEnforceLeft:
		return `#context { let pg = counter(page).at(here()).first(); if calc.even(pg) { state("folio-skip-header-pages", ()).update(pages => pages + (pg + 1,)); state("folio-skip-footer-pages", ()).update(pages => pages + (pg + 1,)) } }
#pagebreak(weak: true, to: "even")`
	default:
		return ""
	}
}

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
	Style               string              `yaml:"style"`
	Page                string              `yaml:"page"`
	Margin              string              `yaml:"margin"`
	Font                string              `yaml:"font"`
	FontSize            string              `yaml:"font-size"`
	FontWeight          string              `yaml:"font-weight"`
	HeadingFont         string              `yaml:"heading-font"`
	HeadingFontSize     string              `yaml:"heading-font-size"`
	HeadingFontWeight   string              `yaml:"heading-font-weight"`
	MonoFont            string              `yaml:"mono-font"`
	MonoFontSize        string              `yaml:"mono-font-size"`
	MonoFontWeight      string              `yaml:"mono-font-weight"`
	TitleFont           string              `yaml:"title-font"`
	TitleFontSize       string              `yaml:"title-font-size"`
	TitleFontWeight     string              `yaml:"title-font-weight"`
	SubtitleFont        string              `yaml:"subtitle-font"`
	SubtitleFontSize    string              `yaml:"subtitle-font-size"`
	SubtitleFontWeight  string              `yaml:"subtitle-font-weight"`
	SubtitleFontStyle   string              `yaml:"subtitle-font-style"`
	AuthorFont          string              `yaml:"author-font"`
	AuthorFontSize      string              `yaml:"author-font-size"`
	AuthorFontWeight    string              `yaml:"author-font-weight"`
	Attribution         string              `yaml:"attribution"`
	AuthorAttribution   string              `yaml:"author-attribution"`
	DateFont            string              `yaml:"date-font"`
	DateFontSize        string              `yaml:"date-font-size"`
	DateFontWeight      string              `yaml:"date-font-weight"`
	DateFormat          string              `yaml:"date-format"`
	VersionFont         string              `yaml:"version-font"`
	VersionFontSize     string              `yaml:"version-font-size"`
	VersionFontWeight   string              `yaml:"version-font-weight"`
	WordCountFont       string              `yaml:"wordcount-font"`
	WordCountFontSize   string              `yaml:"wordcount-font-size"`
	WordCountFontWeight string              `yaml:"wordcount-font-weight"`
	ContactFont         string              `yaml:"contact-font"`
	ContactFontSize     string              `yaml:"contact-font-size"`
	ContactFontWeight   string              `yaml:"contact-font-weight"`
	LineSpacing         string              `yaml:"line-spacing"`
	LetterSpacing       string              `yaml:"letter-spacing"`
	Justify             *bool               `yaml:"justify"`
	WidowOrphanControl  *bool               `yaml:"widow-orphan-control"`
	ParagraphIndent     string              `yaml:"paragraph-indent"`
	ParagraphSpacing    string              `yaml:"paragraph-spacing"`
	QuotedBlockSpacing  ConfigString        `yaml:"quoted-block-spacing"`
	CodeBlockSpacing    ConfigString        `yaml:"code-block-spacing"`
	QuotedBlock         QuotedBlockConfig   `yaml:"quoted-block"`
	PageHeader          PageHeaderConfig    `yaml:"page-header"`
	PageFooter          PageFooterConfig    `yaml:"page-footer"`
	Gutter              string              `yaml:"gutter"`
	TOC                 TOCConfig           `yaml:"toc"`
	TitlePage           TitlePageConfig     `yaml:"title-page"`
	SceneBreak          SceneBreakConfig    `yaml:"scene-break"`
	List                SpacedBlockConfig   `yaml:"list"`
	Table               SpacedBlockConfig   `yaml:"table"`
	CodeBlock           SpacedBlockConfig   `yaml:"code-block"`
	Part                HeadingConfig       `yaml:"part"`
	Chapter             HeadingConfig       `yaml:"chapter"`
	Copyright           CopyrightConfig     `yaml:"copyright"`
	PageNumbering       PageNumberingConfig `yaml:"page-numbering"`
}

// #16 PageNumberingConfig -- controls the page-number style on frontmatter
// vs body pages and whether the display counter restarts at the frontmatter/body
// boundary. Number-format values match the existing part/chapter number-format
// convention: "1" (arabic, default), "I" (Roman upper), "i" (Roman lower).
type PageNumberingConfig struct {
	FrontmatterFormat string `yaml:"frontmatter-format"` // "1" | "I" | "i"; default "1"
	BodyFormat        string `yaml:"body-format"`        // "1" | "I" | "i"; default "1"
	BodyReset         string `yaml:"body-reset"`         // "first-part-or-chapter" (default) | "never"
}

// #21 CopyrightConfig -- frontmatter copyright page rendered from declarative
// config. Every field is optional. Enabled defaults to false for backwards
// compatibility. When enabled, the page renders on the frontmatter (verso, page
// ii by default) between the title page and the TOC.
type CopyrightConfig struct {
	Enabled         bool          `yaml:"enabled"`
	Position        string        `yaml:"position"`              // "after-title" (default), "after-toc", "after-frontmatter"
	SkipHeader      *bool         `yaml:"skip-header,omitempty"` // default true
	SkipFooter      *bool         `yaml:"skip-footer,omitempty"` // default false
	BlankPageBefore BlankPageMode `yaml:"blank-page-before"`
	BlankPageAfter  BlankPageMode `yaml:"blank-page-after"`
	Align           string        `yaml:"align"`
	// Credits is a list of free-text credit lines rendered as centred paragraphs at
	// the top of the copyright page (before the body). Users write the full copyright
	// text they want -- including the © character, year, and holder name(s). Markdown-
	// mini formatting (**bold**, *italic*, en-dash, em-dash) applies. When unset, a
	// single default line is generated from folio.author and the folio.date year:
	// `Copyright © YEAR Author Name.` To omit the credit block entirely, set credits
	// to an empty list AND unset folio.author.
	Credits              []string `yaml:"credits"`
	Body                 []string `yaml:"body"`
	Separator            string   `yaml:"separator"`
	SeparatorSpaceBefore string   `yaml:"separator-space-before"`
	SeparatorSpaceAfter  string   `yaml:"separator-space-after"`
	Publication          []string `yaml:"publication"`
	Publisher            string   `yaml:"publisher"`
	PublisherPreposition string   `yaml:"publisher-preposition"`
	ISBN                 string   `yaml:"isbn"`
	ISBNLabel            string   `yaml:"isbn-label"`
	ISBNBarcode          string   `yaml:"isbn-barcode"` // "none" (default), "render", "file", "render-and-file"
	Font                 string   `yaml:"font"`
	FontSize             string   `yaml:"font-size"`
	HeadingFontWeight    string   `yaml:"heading-font-weight"`
	LineSpacing          string   `yaml:"line-spacing"`
	BlockSpacing         string   `yaml:"block-spacing"`
}

type TitlePageConfig struct {
	Enabled            bool                `yaml:"enabled"`
	PageNumber         bool                `yaml:"page-number"`
	IncludeTitle       bool                `yaml:"include-title"`
	IncludeSubtitle    bool                `yaml:"include-subtitle"`
	IncludeAuthor      bool                `yaml:"include-author"`
	IncludeDate        bool                `yaml:"include-date"`
	IncludeWordCount   bool                `yaml:"include-wordcount"`
	IncludeContactName bool                `yaml:"include-contact-name"`
	IncludeAddress     bool                `yaml:"include-address"`
	IncludePhone       bool                `yaml:"include-phone"`
	IncludeEmail       bool                `yaml:"include-email"`
	IncludeWebsite     bool                `yaml:"include-website"`
	IncludeVersion     bool                `yaml:"include-version"`
	TitleBlockAlign    string              `yaml:"title-block-align"`
	FooterAlign        string              `yaml:"footer-align"`
	Title              TitlePageItemConfig `yaml:"title"`
	Subtitle           TitlePageItemConfig `yaml:"subtitle"`
	Author             TitlePageItemConfig `yaml:"author"`
	Date               TitlePageItemConfig `yaml:"date"`
	WordCount          TitlePageItemConfig `yaml:"wordcount"`
	Version            TitlePageItemConfig `yaml:"version"`
	Contact            TitlePageItemConfig `yaml:"contact"`
}

type TitlePageItemConfig struct {
	Align string `yaml:"align"`
}

type PageHeaderConfig struct {
	Enabled       bool   `yaml:"enabled"`
	Font          string `yaml:"font"`
	FontSize      string `yaml:"font-size"`
	FontWeight    string `yaml:"font-weight"`
	FontStyle     string `yaml:"font-style"`
	LetterSpacing string `yaml:"letter-spacing"`
	Format        string `yaml:"format"`
	// #24: frontmatter-format / alt-frontmatter-format apply on frontmatter pages
	// (title, copyright, TOC, and any page before the first part/chapter). nil ->
	// fall back to Format / AltFormat. Non-nil empty string -> blank header on
	// frontmatter. Non-nil non-empty -> use this format on frontmatter.
	FrontmatterFormat    *string `yaml:"frontmatter-format,omitempty"`
	AltFrontmatterFormat *string `yaml:"alt-frontmatter-format,omitempty"`
	// AC18.6: when set, AltFormat renders on right-hand (recto, odd) pages while Format
	// continues to render on left-hand (verso, even) pages. When unset, Format renders
	// on every page (unchanged from AC15.1).
	AltFormat           string `yaml:"alt-format"`
	Align               string `yaml:"align"`
	DistanceFromEdge    string `yaml:"distance-from-edge"`
	ContentPaddingAfter string `yaml:"content-padding-after"`
}

// PageFooterConfig mirrors PageHeaderConfig. Fields left empty inherit from PageHeaderConfig
// during normalization; PageHeader values in turn inherit from root folio settings.
// Enabled is a *bool so normalizeConfig can distinguish "unset" (default true) from
// an explicit `enabled: false`.
type PageFooterConfig struct {
	Enabled       *bool  `yaml:"enabled,omitempty"`
	Font          string `yaml:"font"`
	FontSize      string `yaml:"font-size"`
	FontWeight    string `yaml:"font-weight"`
	FontStyle     string `yaml:"font-style"`
	LetterSpacing string `yaml:"letter-spacing"`
	Format        string `yaml:"format"`
	// #24: frontmatter-format / alt-frontmatter-format apply on frontmatter pages.
	// nil -> fall back to Format / AltFormat. Non-nil empty string -> blank footer
	// on frontmatter. Non-nil non-empty -> use this format on frontmatter.
	FrontmatterFormat    *string `yaml:"frontmatter-format,omitempty"`
	AltFrontmatterFormat *string `yaml:"alt-frontmatter-format,omitempty"`
	// AC18.6: verso uses Format, recto uses AltFormat when set (see PageHeaderConfig).
	AltFormat           string `yaml:"alt-format"`
	Align               string `yaml:"align"`
	DistanceFromEdge    string `yaml:"distance-from-edge"`
	ContentPaddingAfter string `yaml:"content-padding-after"`
}

type TOCConfig struct {
	Enabled           bool          `yaml:"enabled"`
	Links             bool          `yaml:"links"`
	Title             string        `yaml:"title"`
	Font              string        `yaml:"font"`
	FontSize          string        `yaml:"font-size"`
	FontWeight        string        `yaml:"font-weight"`
	HeadingFont       string        `yaml:"heading-font"`
	HeadingFontSize   string        `yaml:"heading-font-size"`
	HeadingFontWeight string        `yaml:"heading-font-weight"`
	IncludeParts      bool          `yaml:"include-parts"`
	IncludeChapters   bool          `yaml:"include-chapters"`
	IncludeSections   bool          `yaml:"include-sections"`
	DotLeaders        bool          `yaml:"dot-leaders"`
	PageNumbers       bool          `yaml:"page-numbers"`
	PageBreakBefore   bool          `yaml:"page-break-before"`
	BlankPageBefore   BlankPageMode `yaml:"blank-page-before"`
	BlankPageAfter    BlankPageMode `yaml:"blank-page-after"`
	LineSpacing       string        `yaml:"line-spacing"`
	PartGapBefore     string        `yaml:"part-gap-before"`
	ContinuationPad   string        `yaml:"continuation-padding-before"`
	PartBold          bool          `yaml:"part-bold"`
}

type SceneBreakConfig struct {
	Marker string `yaml:"marker"`
}

type SpacedBlockConfig struct {
	SpaceBefore string `yaml:"space-before"`
	SpaceAfter  string `yaml:"space-after"`
}

type QuotedBlockConfig struct {
	Font FontConfig `yaml:"font"`
}

// ConfigString retains whether a YAML property was omitted so normalization
// can distinguish inheritance from an explicitly empty or null value.
type ConfigString struct {
	Value string
	Set   bool
}

func (s *ConfigString) UnmarshalYAML(node *yaml.Node) error {
	s.Set = true
	if node.Tag == "!!null" {
		s.Value = ""
		return nil
	}
	if node.Kind != yaml.ScalarNode {
		return nil
	}
	return node.Decode(&s.Value)
}

// FontConfig is the shared six-property font block introduced for quoted
// manuscript blocks and intended for reuse by the unified font migration.
type FontConfig struct {
	Family        string `yaml:"family"`
	Size          string `yaml:"size"`
	Weight        string `yaml:"weight"`
	Stretch       string `yaml:"stretch"`
	Style         string `yaml:"style"`
	LetterSpacing string `yaml:"letter-spacing"`
	set           map[string]bool
}

func (f *FontConfig) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return fmt.Errorf("folio.manuscript.quoted-block.font must be a mapping")
	}
	f.set = make(map[string]bool, len(node.Content)/2)
	for index := 0; index < len(node.Content); index += 2 {
		key := node.Content[index].Value
		valueNode := node.Content[index+1]
		f.set[key] = true
		var target *string
		switch key {
		case "family":
			target = &f.Family
		case "size":
			target = &f.Size
		case "weight":
			target = &f.Weight
		case "stretch":
			target = &f.Stretch
		case "style":
			target = &f.Style
		case "letter-spacing":
			target = &f.LetterSpacing
		default:
			return fmt.Errorf("folio.manuscript.quoted-block.font.%s is not supported", key)
		}
		if valueNode.Tag == "!!null" {
			*target = ""
			continue
		}
		if valueNode.Kind != yaml.ScalarNode {
			return fmt.Errorf("folio.manuscript.quoted-block.font.%s must be a scalar value", key)
		}
		if err := valueNode.Decode(target); err != nil {
			return fmt.Errorf("decoding folio.manuscript.quoted-block.font.%s: %w", key, err)
		}
	}
	return nil
}

func (f *FontConfig) propertyWasSet(name string) bool {
	return f.set != nil && f.set[name]
}

var typstLengthRE = regexp.MustCompile(`^([+-]?(?:[0-9]+(?:\.[0-9]+)?|\.[0-9]+))(pt|mm|cm|in|em)$`)

var typstFontWeights = map[string]bool{
	"thin": true, "extralight": true, "light": true, "regular": true,
	"medium": true, "semibold": true, "bold": true, "extrabold": true, "black": true,
}

type HeadingConfig struct {
	PageBreakBefore bool          `yaml:"page-break-before"`
	BlankPageBefore BlankPageMode `yaml:"blank-page-before"`
	BlankPageAfter  BlankPageMode `yaml:"blank-page-after"`
	SkipHeader      bool          `yaml:"skip-header"`
	SkipFooter      bool          `yaml:"skip-footer"`
	VerticalAlign   string        `yaml:"vertical-align"`
	Position        string        `yaml:"position"`
	Align           string        `yaml:"align"`
	CaseTransform   string        `yaml:"case-transform"`
	SpaceAfter      string        `yaml:"space-after"`
	// AC18.3: presentation-side fields that compose the rendered heading string
	// as Prefix + FormatNumber(Number, NumberFormat) + Separator + Name + Suffix.
	Prefix       string `yaml:"prefix"`
	NumberFormat string `yaml:"number-format"` // part: 1/I/i; chapter also accepts a part.chapter pair
	Separator    string `yaml:"separator"`
	Suffix       string `yaml:"suffix"`
	ShowName     *bool  `yaml:"show-name,omitempty"`   // nil = default true
	ShowNumber   *bool  `yaml:"show-number,omitempty"` // nil = default false
	// AC18.5: name-case applies to the name segment only, taking precedence over
	// case-transform for that segment. Values: "" (as-written), "upper", "lower", "title".
	NameCase string `yaml:"name-case"`
	// AC18.2: choose between derived (source-order) and source (parsed from source) numbering.
	ExplicitNumbering string `yaml:"explicit-numbering"` // "" or "derived" (default), "source"
	// #16 AC16.4: controls whether the counter resets. "never" (default) numbers
	// continuously; "per-part" restarts chapters at each part.
	// Currently applied only to chapter (part counter has no reset boundary above it).
	NumberReset string `yaml:"number-reset"` // "never" (default for chapter) | "per-part"
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
	if err := validateCopyright(&ms.Copyright); err != nil {
		return err
	}
	if err := validatePageNumbering(&ms.PageNumbering); err != nil {
		return err
	}
	if err := validateNumberReset("chapter", ms.Chapter.NumberReset); err != nil {
		return err
	}
	if err := validateQuotedBlockConfig(ms); err != nil {
		return err
	}
	if err := validateHeadingNumberFormat("part", ms.Part.NumberFormat, false); err != nil {
		return err
	}
	if err := validateHeadingNumberFormat("chapter", ms.Chapter.NumberFormat, true); err != nil {
		return err
	}
	return nil
}

func validateQuotedBlockConfig(ms *ManuscriptConfig) error {
	if err := validateTypstLength("folio.manuscript.quoted-block-spacing", ms.QuotedBlockSpacing.Value, true, false); err != nil {
		return err
	}
	if err := validateTypstLength("folio.manuscript.code-block-spacing", ms.CodeBlockSpacing.Value, true, false); err != nil {
		return err
	}
	font := &ms.QuotedBlock.Font
	if strings.TrimSpace(font.Family) == "" {
		return fmt.Errorf("folio.manuscript.quoted-block.font.family must not be empty")
	}
	if err := validateTypstLength("folio.manuscript.quoted-block.font.size", font.Size, false, false); err != nil {
		return err
	}
	if err := validateFontWeight("folio.manuscript.quoted-block.font.weight", font.Weight); err != nil {
		return err
	}
	if err := validateFontStretch("folio.manuscript.quoted-block.font.stretch", font.Stretch); err != nil {
		return err
	}
	if err := validateFontStyle("folio.manuscript.quoted-block.font.style", font.Style); err != nil {
		return err
	}
	return validateTypstLength("folio.manuscript.quoted-block.font.letter-spacing", font.LetterSpacing, true, true)
}

func validateTypstLength(path string, value string, allowZero bool, allowNegative bool) error {
	match := typstLengthRE.FindStringSubmatch(strings.TrimSpace(value))
	if match == nil {
		return fmt.Errorf("%s %q must be a length in pt, mm, cm, in, or em", path, value)
	}
	number, err := strconv.ParseFloat(match[1], 64)
	if err != nil {
		return fmt.Errorf("%s %q has an invalid numeric value: %w", path, value, err)
	}
	if number < 0 && !allowNegative {
		return fmt.Errorf("%s %q must not be negative", path, value)
	}
	if number == 0 && !allowZero {
		return fmt.Errorf("%s %q must be greater than zero", path, value)
	}
	return nil
}

func validateFontWeight(path string, value string) error {
	value = strings.ToLower(strings.TrimSpace(value))
	if typstFontWeights[value] {
		return nil
	}
	weight, err := strconv.Atoi(value)
	if err == nil && weight >= 100 && weight <= 900 {
		return nil
	}
	return fmt.Errorf("%s %q must be a Typst weight name or a number from 100 to 900", path, value)
}

func validateFontStretch(path string, value string) error {
	value = strings.TrimSuffix(strings.TrimSpace(value), "%")
	stretch, err := strconv.ParseFloat(value, 64)
	if err != nil || math.IsNaN(stretch) || math.IsInf(stretch, 0) || stretch <= 0 {
		return fmt.Errorf("%s %q must be a positive percentage or plain number", path, value)
	}
	return nil
}

func validateFontStyle(path string, value string) error {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "regular", "italic", "oblique":
		return nil
	default:
		return fmt.Errorf("%s %q must be regular, italic, or oblique", path, value)
	}
}

// validatePageNumbering enforces the "1" / "I" / "i" enum for both format fields
// and the enum for body-reset.
func validatePageNumbering(p *PageNumberingConfig) error {
	switch p.FrontmatterFormat {
	case "1", "I", "i":
	default:
		return fmt.Errorf("page-numbering.frontmatter-format %q is not one of 1, I, i", p.FrontmatterFormat)
	}
	switch p.BodyFormat {
	case "1", "I", "i":
	default:
		return fmt.Errorf("page-numbering.body-format %q is not one of 1, I, i", p.BodyFormat)
	}
	switch p.BodyReset {
	case "first-part-or-chapter", "never":
	default:
		return fmt.Errorf("page-numbering.body-reset %q is not one of first-part-or-chapter, never", p.BodyReset)
	}
	return nil
}

func validateNumberReset(field, value string) error {
	switch value {
	case "per-part", "never", "":
		return nil
	default:
		return fmt.Errorf("%s.number-reset %q is not one of per-part, never", field, value)
	}
}

func validateHeadingNumberFormat(field, value string, allowComposite bool) error {
	if value == "" {
		return nil
	}
	segments := strings.Split(value, ".")
	if len(segments) > 1 && !allowComposite {
		return fmt.Errorf("%s.number-format %q must be one of 1, I, i", field, value)
	}
	if len(segments) > 2 {
		return fmt.Errorf("%s.number-format %q must be one or two segments chosen from 1, I, i", field, value)
	}
	for _, segment := range segments {
		switch segment {
		case "1", "I", "i":
		default:
			return fmt.Errorf("%s.number-format %q must be one or two segments chosen from 1, I, i", field, value)
		}
	}
	return nil
}

// validateCopyright enforces enum and format constraints on the copyright block.
// Only runs when Enabled == true, so an unconfigured manuscript still loads even
// if defaults would trip validation (they don't, but the guard is defensive).
func validateCopyright(c *CopyrightConfig) error {
	if !c.Enabled {
		return nil
	}
	switch c.Position {
	case "after-title", "after-toc", "after-frontmatter":
	default:
		return fmt.Errorf("copyright.position %q is not one of after-title, after-toc, after-frontmatter", c.Position)
	}
	switch c.ISBNBarcode {
	case "none", "render", "file", "render-and-file":
	default:
		return fmt.Errorf("copyright.isbn-barcode %q is not one of none, render, file, render-and-file", c.ISBNBarcode)
	}
	if c.ISBN != "" {
		if err := validateISBN13(c.ISBN); err != nil {
			return err
		}
	}
	return nil
}

// validateISBN13 checks that an ISBN string has 13 digits (ignoring hyphens) and
// a valid EAN-13 check digit. Non-numeric, wrong length, or wrong check digit
// input is rejected with a diagnostic naming the offending value.
func validateISBN13(isbn string) error {
	digits := strings.ReplaceAll(isbn, "-", "")
	digits = strings.ReplaceAll(digits, " ", "")
	if len(digits) != 13 {
		return fmt.Errorf("copyright.isbn %q must be 13 digits (ignoring hyphens); got %d", isbn, len(digits))
	}
	sum := 0
	for i, r := range digits {
		if r < '0' || r > '9' {
			return fmt.Errorf("copyright.isbn %q contains non-digit character %q", isbn, r)
		}
		d := int(r - '0')
		if i%2 == 0 {
			sum += d
		} else {
			sum += d * 3
		}
	}
	if sum%10 != 0 {
		return fmt.Errorf("copyright.isbn %q has invalid EAN-13 check digit", isbn)
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
	fill(&ms.LetterSpacing, "0em")
	if ms.Justify == nil {
		justify := true
		ms.Justify = &justify
	}
	if ms.WidowOrphanControl == nil {
		enabled := true
		ms.WidowOrphanControl = &enabled
	}
	fill(&ms.ParagraphIndent, "10mm")
	fill(&ms.ParagraphSpacing, "0")
	if ms.ParagraphSpacing == "0" {
		ms.ParagraphSpacing = "0pt"
	}
	defaultConfigString(&ms.QuotedBlockSpacing, "0.5em")
	defaultConfigString(&ms.CodeBlockSpacing, "0.5em")
	inheritFontProperty(&ms.QuotedBlock.Font, "family", &ms.QuotedBlock.Font.Family, ms.Font)
	inheritFontProperty(&ms.QuotedBlock.Font, "size", &ms.QuotedBlock.Font.Size, ms.FontSize)
	inheritFontProperty(&ms.QuotedBlock.Font, "weight", &ms.QuotedBlock.Font.Weight, ms.FontWeight)
	inheritFontProperty(&ms.QuotedBlock.Font, "stretch", &ms.QuotedBlock.Font.Stretch, "100%")
	inheritFontProperty(&ms.QuotedBlock.Font, "style", &ms.QuotedBlock.Font.Style, "regular")
	inheritFontProperty(&ms.QuotedBlock.Font, "letter-spacing", &ms.QuotedBlock.Font.LetterSpacing, ms.LetterSpacing)
	ms.QuotedBlock.Font.Weight = strings.ToLower(strings.TrimSpace(ms.QuotedBlock.Font.Weight))
	ms.QuotedBlock.Font.Stretch = normalizeFontStretch(ms.QuotedBlock.Font.Stretch)
	ms.QuotedBlock.Font.Style = strings.ToLower(strings.TrimSpace(ms.QuotedBlock.Font.Style))
	fill(&ms.PageHeader.Font, ms.HeadingFont)
	fill(&ms.PageHeader.FontSize, "10pt")
	fill(&ms.PageHeader.FontWeight, "regular")
	fill(&ms.PageHeader.LetterSpacing, ms.LetterSpacing)
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
	fill(&ms.PageFooter.FontStyle, ms.PageHeader.FontStyle)
	fill(&ms.PageFooter.LetterSpacing, ms.LetterSpacing)
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
	fill(&ms.TOC.ContinuationPad, "15mm")
	fill(&ms.SceneBreak.Marker, "#")
	fill(&ms.List.SpaceBefore, "0.5em")
	fill(&ms.List.SpaceAfter, "0.5em")
	fill(&ms.Table.SpaceBefore, "0.5em")
	fill(&ms.Table.SpaceAfter, "0.5em")
	fill(&ms.CodeBlock.SpaceBefore, ms.CodeBlockSpacing.Value)
	fill(&ms.CodeBlock.SpaceAfter, ms.CodeBlockSpacing.Value)
	fill(&ms.TitlePage.TitleBlockAlign, "center")
	fill(&ms.TitlePage.FooterAlign, "center")
	fill(&ms.Part.Align, "center")
	fill(&ms.Part.VerticalAlign, "center")
	fill(&ms.Part.CaseTransform, "as-written")
	fill(&ms.Chapter.Align, "center")
	fill(&ms.Chapter.Position, "one-third")
	fill(&ms.Chapter.CaseTransform, "as-written")
	fill(&ms.Chapter.SpaceAfter, "2em")
	// #21 Copyright page defaults.
	fill(&ms.Copyright.Position, "after-title")
	if ms.Copyright.SkipHeader == nil {
		t := true
		ms.Copyright.SkipHeader = &t
	}
	if ms.Copyright.SkipFooter == nil {
		f := false
		ms.Copyright.SkipFooter = &f
	}
	if ms.Copyright.BlankPageBefore == BlankPageMode("") {
		ms.Copyright.BlankPageBefore = BlankPageMode("enforce-left")
	}
	fill(&ms.Copyright.Align, "center")
	fill(&ms.Copyright.Separator, "———")
	fill(&ms.Copyright.SeparatorSpaceBefore, "1.5em")
	fill(&ms.Copyright.SeparatorSpaceAfter, "1.5em")
	fill(&ms.Copyright.PublisherPreposition, "by")
	fill(&ms.Copyright.ISBNLabel, "ISBN")
	fill(&ms.Copyright.ISBNBarcode, "none")
	fill(&ms.Copyright.Font, ms.Font)
	fill(&ms.Copyright.FontSize, ms.FontSize)
	fill(&ms.Copyright.HeadingFontWeight, "bold")
	// #21: copyright page uses a more generous default line-spacing than body
	// prose. British / US body may be 1.0 (single) or 2.0 (double) — neither is
	// ideal for a densely-set legal boilerplate page. 1.4 gives publisher-typical
	// breathing room. Users can override via copyright.line-spacing.
	fill(&ms.Copyright.LineSpacing, "1.4")
	fill(&ms.Copyright.BlockSpacing, "0.75em")
	// #16 page-numbering defaults.
	fill(&ms.PageNumbering.FrontmatterFormat, "1")
	fill(&ms.PageNumbering.BodyFormat, "1")
	fill(&ms.PageNumbering.BodyReset, "first-part-or-chapter")
	// #32 chapter numbering is continuous unless the user explicitly requests per-part reset.
	fill(&ms.Chapter.NumberReset, "never")
	// Parts have no scope above them; hardcode "never" for symmetry / future-proofing
	// but never actually reset the part counter.
	fill(&ms.Part.NumberReset, "never")
}

func normalizeFontStretch(value string) string {
	value = strings.TrimSpace(value)
	if value != "" && !strings.HasSuffix(value, "%") {
		return value + "%"
	}
	return value
}

func inheritFontProperty(font *FontConfig, name string, target *string, value string) {
	if !font.propertyWasSet(name) && *target == "" {
		*target = value
	}
}

func defaultConfigString(target *ConfigString, value string) {
	if !target.Set {
		target.Value = value
	}
}

func fill(target *string, value string) {
	if *target == "" {
		*target = value
	}
}
