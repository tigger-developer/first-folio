// ABOUTME: Typst rendering for manuscript documents.
// ABOUTME: Uses a file-backed template and generated block markup.
package manuscript

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"

	folio "github.com/tigger-developer/first-folio"
)

type templateData struct {
	Config           Config
	Meta             Metadata
	Header           string
	Footer           string
	Body             string
	IsUS             bool
	Leading          string
	Spacing          string
	PartVertical     string
	ChapterPosition  string
	SceneBreakMarker string
	HasContact       bool

	// Page dimensions: either a named preset or a custom W x H (both non-empty means custom).
	PageSpec PageSpec

	// Binding gutter: when GutterActive, the running-page margin uses inside/outside idiom.
	Gutter       string
	GutterActive bool

	// Header/footer alignment expressions (rendered directly into `align(...)` in the template).
	HeaderAlignExpr    string
	FooterAlignExpr    string
	HeaderAlignIsPair  bool
	FooterAlignIsPair  bool
	PageFooterEnabled  bool

	// Title-page per-item alignment expressions (empty = use legacy group fallback).
	TitleAlignExpr     string
	SubtitleAlignExpr  string
	AuthorAlignExpr    string
	DateAlignExpr      string
	WordCountAlignExpr string
	VersionAlignExpr   string
	ContactAlignExpr   string

	// Legacy title-block-align, used for the title/subtitle/author group when no per-item align is set.
	TitleBlockAlignExpr string
	FooterGroupAlignExpr string
}

func RenderTypst(doc Document, cfg Config) (string, error) {
	pageSpec, err := ParsePageSpec(cfg.Folio.Manuscript.Page)
	if err != nil {
		return "", err
	}
	headerAlign, err := ParseHeaderFooterAlign(cfg.Folio.Manuscript.PageHeader.Align)
	if err != nil {
		return "", err
	}
	footerAlign, err := ParseHeaderFooterAlign(cfg.Folio.Manuscript.PageFooter.Align)
	if err != nil {
		return "", err
	}
	titleBlockExpr, err := TitleItemAlign(cfg.Folio.Manuscript.TitlePage.TitleBlockAlign)
	if err != nil {
		return "", err
	}
	footerGroupExpr, err := TitleItemAlign(cfg.Folio.Manuscript.TitlePage.FooterAlign)
	if err != nil {
		return "", err
	}
	itemExpr := func(value string) (string, error) { return TitleItemAlign(value) }
	titleExpr, err := itemExpr(cfg.Folio.Manuscript.TitlePage.Title.Align)
	if err != nil {
		return "", err
	}
	subtitleExpr, err := itemExpr(cfg.Folio.Manuscript.TitlePage.Subtitle.Align)
	if err != nil {
		return "", err
	}
	authorExpr, err := itemExpr(cfg.Folio.Manuscript.TitlePage.Author.Align)
	if err != nil {
		return "", err
	}
	dateExpr, err := itemExpr(cfg.Folio.Manuscript.TitlePage.Date.Align)
	if err != nil {
		return "", err
	}
	wordCountExpr, err := itemExpr(cfg.Folio.Manuscript.TitlePage.WordCount.Align)
	if err != nil {
		return "", err
	}
	versionExpr, err := itemExpr(cfg.Folio.Manuscript.TitlePage.Version.Align)
	if err != nil {
		return "", err
	}
	contactExpr, err := itemExpr(cfg.Folio.Manuscript.TitlePage.Contact.Align)
	if err != nil {
		return "", err
	}

	body, err := renderBlocks(doc.Blocks, cfg)
	if err != nil {
		return "", err
	}
	safeMeta := escapedMetadata(doc.Metadata)
	safeMeta.Date = escapeTypst(renderDate(doc.Metadata.Date, cfg.Folio.Manuscript.DateFormat))
	leading := lineSpacingLeading(cfg.Folio.Manuscript.LineSpacing)
	pageFooterEnabled := cfg.Folio.Manuscript.PageFooter.Enabled != nil && *cfg.Folio.Manuscript.PageFooter.Enabled
	data := templateData{
		Config:              cfg,
		Meta:                safeMeta,
		Header:              renderHeader(doc.Metadata, cfg),
		Footer:              renderFooter(doc.Metadata, cfg),
		Body:                body,
		IsUS:                cfg.Folio.Manuscript.Style == "us",
		Leading:             leading,
		Spacing:             paragraphSpacing(cfg.Folio.Manuscript.ParagraphSpacing, leading),
		PartVertical:        typstVerticalAlign(cfg.Folio.Manuscript.Part.VerticalAlign),
		ChapterPosition:     chapterPosition(cfg.Folio.Manuscript.Chapter.Position),
		SceneBreakMarker:    escapeTypst(cfg.Folio.Manuscript.SceneBreak.Marker),
		HasContact:          hasContactBlock(doc.Metadata, cfg),
		PageSpec:            pageSpec,
		Gutter:              cfg.Folio.Manuscript.Gutter,
		GutterActive:        isGutterActive(cfg.Folio.Manuscript.Gutter),
		HeaderAlignExpr:     headerAlign.TypstAlignExpression(),
		FooterAlignExpr:     footerAlign.TypstAlignExpression(),
		HeaderAlignIsPair:   headerAlign.IsPair,
		FooterAlignIsPair:   footerAlign.IsPair,
		PageFooterEnabled:   pageFooterEnabled,
		TitleAlignExpr:      titleExpr,
		SubtitleAlignExpr:   subtitleExpr,
		AuthorAlignExpr:     authorExpr,
		DateAlignExpr:       dateExpr,
		WordCountAlignExpr:  wordCountExpr,
		VersionAlignExpr:    versionExpr,
		ContactAlignExpr:    contactExpr,
		TitleBlockAlignExpr: titleBlockExpr,
		FooterGroupAlignExpr: footerGroupExpr,
	}
	raw, err := folio.Assets.ReadFile("templates/manuscript.typ")
	if err != nil {
		return "", fmt.Errorf("loading Typst template: %w", err)
	}
	tmpl, err := template.New("manuscript.typ").Parse(string(raw))
	if err != nil {
		return "", fmt.Errorf("parsing Typst template: %w", err)
	}
	var out bytes.Buffer
	if err := tmpl.Execute(&out, data); err != nil {
		return "", fmt.Errorf("executing Typst template: %w", err)
	}
	return out.String(), nil
}

var gutterZeroRE = regexp.MustCompile(`^\s*0(?:\.0+)?(?:mm|in|pt|cm|em)?\s*$`)

func isGutterActive(gutter string) bool {
	return !gutterZeroRE.MatchString(gutter)
}

func renderDate(value string, layout string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	parsed, err := time.Parse("2006-01-02", trimmed)
	if err != nil {
		return value
	}
	if strings.TrimSpace(layout) == "" {
		layout = "2 January 2006"
	}
	return parsed.Format(layout)
}

func lineSpacingLeading(lineSpacing string) string {
	multiplier, err := strconv.ParseFloat(strings.TrimSpace(lineSpacing), 64)
	if err != nil {
		return lineSpacing + "em"
	}
	if multiplier < 1 {
		multiplier = 1
	}
	return strconv.FormatFloat(multiplier-1, 'f', -1, 64) + "em"
}

func paragraphSpacing(spacing string, leading string) string {
	trimmed := strings.TrimSpace(spacing)
	if trimmed == "" || trimmed == "0" || trimmed == "0pt" {
		return leading
	}
	if leading == "0em" {
		return trimmed
	}
	return leading + " + " + trimmed
}

func renderBlocks(blocks []Block, cfg Config) (string, error) {
	partBlankBefore := cfg.Folio.Manuscript.Part.BlankPageBefore
	partBlankAfter := cfg.Folio.Manuscript.Part.BlankPageAfter
	chapterBlankBefore := cfg.Folio.Manuscript.Chapter.BlankPageBefore
	chapterBlankAfter := cfg.Folio.Manuscript.Chapter.BlankPageAfter

	var lines []string
	firstPageBlock := true
	for _, block := range blocks {
		switch block.Kind {
		case "part":
			if partBlankBefore {
				lines = append(lines, "#folio-blank-page()")
			}
			lines = append(lines, fmt.Sprintf("#folio-part(first: %t)[%s]",
				firstPageBlock,
				caseTransform(block.Text, cfg.Folio.Manuscript.Part.CaseTransform)))
			if partBlankAfter {
				lines = append(lines, "#folio-blank-page()")
			}
			firstPageBlock = false
		case "chapter":
			if chapterBlankBefore {
				lines = append(lines, "#folio-blank-page()")
			}
			lines = append(lines, fmt.Sprintf("#folio-chapter(first: %t)[%s]",
				firstPageBlock,
				caseTransform(block.Text, cfg.Folio.Manuscript.Chapter.CaseTransform)))
			if chapterBlankAfter {
				lines = append(lines, "#folio-blank-page()")
			}
			firstPageBlock = false
		case "section":
			lines = append(lines, "#folio-section["+typstInline(block.Text, cfg)+"]")
		case "paragraph":
			lines = append(lines, typstInline(block.Text, cfg))
		case "scene-break":
			lines = append(lines, "#folio-scene-break()")
		case "code":
			lines = append(lines, "#folio-code["+escapeTypst(block.Text)+"]")
		case "raw-typst":
			lines = append(lines, block.Text)
		case "footnote":
			continue
		default:
			return "", fmt.Errorf("unknown manuscript block kind: %s", block.Kind)
		}
		lines = append(lines, "")
	}
	return strings.Join(lines, "\n"), nil
}

var placeholderRE = regexp.MustCompile(`\[([a-z-]+)\]`)

func renderHeader(meta Metadata, cfg Config) string {
	return substitutePlaceholders(cfg.Folio.Manuscript.PageHeader.Format, meta)
}

func renderFooter(meta Metadata, cfg Config) string {
	return substitutePlaceholders(cfg.Folio.Manuscript.PageFooter.Format, meta)
}

// substitutePlaceholders resolves [author], [title], [page], [part], [chapter] placeholders in
// a header/footer format string. Recognized tokens are replaced with escaped metadata or the
// corresponding Typst state/counter reads; interstitial text is escapeTypst'd so unknown
// tokens like [unknown] survive as literal brackets in the rendered Typst source.
func substitutePlaceholders(format string, meta Metadata) string {
	var out strings.Builder
	last := 0
	for _, m := range placeholderRE.FindAllStringSubmatchIndex(format, -1) {
		out.WriteString(escapeTypst(format[last:m[0]]))
		name := format[m[2]:m[3]]
		switch name {
		case "author":
			out.WriteString(escapeTypst(meta.Author))
		case "title":
			out.WriteString(escapeTypst(meta.Title))
		case "page":
			out.WriteString("#context counter(page).display()")
		case "part":
			out.WriteString(`#context state("folio-current-part").get()`)
		case "chapter":
			out.WriteString(`#context state("folio-current-chapter").get()`)
		default:
			out.WriteString(escapeTypst(format[m[0]:m[1]]))
		}
		last = m[1]
	}
	out.WriteString(escapeTypst(format[last:]))
	return out.String()
}

func hasContactBlock(meta Metadata, cfg Config) bool {
	titlePage := cfg.Folio.Manuscript.TitlePage
	return titlePage.IncludeContactName && meta.ContactName != "" ||
		titlePage.IncludeAddress && meta.Address != "" ||
		titlePage.IncludePhone && meta.Phone != "" ||
		titlePage.IncludeEmail && meta.Email != "" ||
		titlePage.IncludeWebsite && meta.Website != ""
}

func typstInline(text string, cfg Config) string {
	return renderInlineMarkup(text, cfg.Folio.Manuscript.MonoFont, cfg.Folio.Manuscript.MonoFontSize, cfg.Folio.Manuscript.MonoFontWeight)
}

func renderInlineMarkup(text string, monoFont string, monoSize string, monoWeight string) string {
	var out strings.Builder
	for i := 0; i < len(text); {
		switch {
		case strings.HasPrefix(text[i:], "`"):
			if end := strings.Index(text[i+1:], "`"); end >= 0 {
				content := text[i+1 : i+1+end]
				out.WriteString(fmt.Sprintf(`#text(font: "%s", size: %s, weight: "%s")[%s]`, escapeTypst(monoFont), monoSize, escapeTypst(monoWeight), escapeTypst(content)))
				i += end + 2
				continue
			}
		case strings.HasPrefix(text[i:], "**"):
			if end := strings.Index(text[i+2:], "**"); end >= 0 {
				content := text[i+2 : i+2+end]
				out.WriteString("*")
				out.WriteString(renderInlineMarkup(content, monoFont, monoSize, monoWeight))
				out.WriteString("*")
				i += end + 4
				continue
			}
		case strings.HasPrefix(text[i:], "*"):
			if end := strings.Index(text[i+1:], "*"); end >= 0 {
				content := text[i+1 : i+1+end]
				out.WriteString("_")
				out.WriteString(renderInlineMarkup(content, monoFont, monoSize, monoWeight))
				out.WriteString("_")
				i += end + 2
				continue
			}
		case strings.HasPrefix(text[i:], "[fn:"):
			if end := strings.Index(text[i:], "]"); end >= 0 {
				out.WriteString("#footnote[")
				out.WriteString(escapeTypst(text[i+4 : i+end]))
				out.WriteString("]")
				i += end + 1
				continue
			}
		case strings.HasPrefix(text[i:], "[^"):
			if end := strings.Index(text[i:], "]"); end >= 0 {
				out.WriteString("#footnote[")
				out.WriteString(escapeTypst(text[i+2 : i+end]))
				out.WriteString("]")
				i += end + 1
				continue
			}
		case strings.HasPrefix(text[i:], "---"):
			out.WriteString("—")
			i += 3
			continue
		case strings.HasPrefix(text[i:], "--"):
			out.WriteString("–")
			i += 2
			continue
		}

		next := nextInlineMarker(text[i+1:])
		if next < 0 {
			out.WriteString(escapeTypst(applyMarkdownDashes(text[i:])))
			break
		}
		next += i + 1
		out.WriteString(escapeTypst(applyMarkdownDashes(text[i:next])))
		i = next
	}
	return out.String()
}

func nextInlineMarker(text string) int {
	index := -1
	for _, marker := range []string{"`", "**", "*", "[fn:", "[^", "---", "--"} {
		if found := strings.Index(text, marker); found >= 0 && (index < 0 || found < index) {
			index = found
		}
	}
	return index
}

func applyMarkdownDashes(text string) string {
	text = strings.ReplaceAll(text, "---", "—")
	return strings.ReplaceAll(text, "--", "–")
}

func typstVerticalAlign(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "center", "middle", "horizon", "":
		return "horizon"
	case "top":
		return "top"
	case "bottom":
		return "bottom"
	default:
		return value
	}
}

func chapterPosition(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "one-third", "third", "":
		return "30%"
	case "center", "middle":
		return "50%"
	case "top":
		return "0%"
	default:
		return value
	}
}

func escapeTypst(text string) string {
	replacer := strings.NewReplacer(
		`\`, `\\`,
		`[`, `\[`,
		`]`, `\]`,
		`#`, `\#`,
		`$`, `\$`,
		`@`, `\@`,
		`_`, `\_`,
		`*`, `\*`,
		`/`, `\/`,
	)
	return replacer.Replace(text)
}

func caseTransform(text string, mode string) string {
	if mode == "upper" {
		return escapeTypst(strings.ToUpper(text))
	}
	return typstInline(text, Config{})
}

func escapedMetadata(meta Metadata) Metadata {
	return Metadata{
		Title:             escapeTypst(meta.Title),
		Subtitle:          escapeTypst(meta.Subtitle),
		Author:            escapeTypst(meta.Author),
		AuthorAttribution: escapeTypst(meta.AuthorAttribution),
		Date:              escapeTypst(meta.Date),
		Version:           escapeTypst(meta.Version),
		WordCount:         escapeTypst(meta.WordCount),
		ContactName:       escapeTypst(meta.ContactName),
		Address:           escapeTypst(meta.Address),
		Phone:             escapeTypst(meta.Phone),
		Email:             escapeTypst(meta.Email),
		Website:           escapeTypst(meta.Website),
	}
}

func templateExistsForTests() bool {
	_, err := folio.Assets.ReadFile("templates/manuscript.typ")
	return err == nil
}
