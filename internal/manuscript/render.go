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
	HeaderAlt        string
	HasHeaderAlt     bool
	Footer           string
	FooterAlt        string
	HasFooterAlt     bool
	Body             string

	// #18: seed the first part / first chapter semantic-authoring state at document top so
	// the FIRST body page's header context (which evaluates at page-top, before the first
	// folio-part / folio-chapter macro body has run its state.update) still sees the
	// correct values. Multi-part / multi-chapter manuscripts update the state again in
	// their block emissions; the seed is a starting value only.
	FirstPartName     string
	FirstPartNumber   string
	FirstPartPrefix   string
	FirstPartFull     string
	FirstChapterName   string
	FirstChapterNumber string
	FirstChapterPrefix string
	FirstChapterFull   string
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


	// Title-page per-item alignment expressions (empty = use legacy group fallback) and a
	// paired Floatable boolean per item so the template can emit `float: true` only when the
	// vertical axis is top or bottom (Typst's floating placement forbids center).
	TitleAlignExpr        string
	TitleFloatable        bool
	SubtitleAlignExpr     string
	SubtitleFloatable     bool
	AuthorAlignExpr       string
	AuthorFloatable       bool
	DateAlignExpr         string
	DateFloatable         bool
	WordCountAlignExpr    string
	WordCountFloatable    bool
	VersionAlignExpr      string
	VersionFloatable      bool
	ContactAlignExpr      string
	ContactFloatable      bool

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
	firstPart, firstChapter := firstHeadingSeeds(doc.Blocks, cfg)
	safeMeta := escapedMetadata(doc.Metadata)
	safeMeta.Date = escapeTypst(renderDate(doc.Metadata.Date, cfg.Folio.Manuscript.DateFormat))
	leading := lineSpacingLeading(cfg.Folio.Manuscript.LineSpacing)
	pageFooterEnabled := cfg.Folio.Manuscript.PageFooter.Enabled != nil && *cfg.Folio.Manuscript.PageFooter.Enabled
	data := templateData{
		Config:              cfg,
		Meta:                safeMeta,
		Header:              renderHeader(doc.Metadata, cfg),
		HeaderAlt:           renderHeaderAlt(doc.Metadata, cfg),
		HasHeaderAlt:        cfg.Folio.Manuscript.PageHeader.AltFormat != "",
		Footer:              renderFooter(doc.Metadata, cfg),
		FooterAlt:           renderFooterAlt(doc.Metadata, cfg),
		HasFooterAlt:        cfg.Folio.Manuscript.PageFooter.AltFormat != "",
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
		TitleFloatable:      TitleItemFloatable(cfg.Folio.Manuscript.TitlePage.Title.Align),
		SubtitleAlignExpr:   subtitleExpr,
		SubtitleFloatable:   TitleItemFloatable(cfg.Folio.Manuscript.TitlePage.Subtitle.Align),
		AuthorAlignExpr:     authorExpr,
		AuthorFloatable:     TitleItemFloatable(cfg.Folio.Manuscript.TitlePage.Author.Align),
		DateAlignExpr:       dateExpr,
		DateFloatable:       TitleItemFloatable(cfg.Folio.Manuscript.TitlePage.Date.Align),
		WordCountAlignExpr:  wordCountExpr,
		WordCountFloatable:  TitleItemFloatable(cfg.Folio.Manuscript.TitlePage.WordCount.Align),
		VersionAlignExpr:    versionExpr,
		VersionFloatable:    TitleItemFloatable(cfg.Folio.Manuscript.TitlePage.Version.Align),
		ContactAlignExpr:    contactExpr,
		ContactFloatable:    TitleItemFloatable(cfg.Folio.Manuscript.TitlePage.Contact.Align),
		TitleBlockAlignExpr: titleBlockExpr,
		FooterGroupAlignExpr: footerGroupExpr,
		FirstPartName:       firstPart.Name,
		FirstPartNumber:     firstPart.Number,
		FirstPartPrefix:     firstPart.Prefix,
		FirstPartFull:       firstPart.Full,
		FirstChapterName:    firstChapter.Name,
		FirstChapterNumber:  firstChapter.Number,
		FirstChapterPrefix:  firstChapter.Prefix,
		FirstChapterFull:    firstChapter.Full,
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

// firstHeadingSeeds returns the composed heading pieces for the FIRST heading block only,
// used to seed the semantic-authoring state at document top so the first body page's
// header context (which evaluates before any block's state.update runs) sees the correct
// initial values. If the first heading is a part, the chapter seed stays empty (so [chapter]
// renders empty on the part page, as expected). If the first heading is a chapter, the part
// seed stays empty. Subsequent block emissions overwrite these seeds as parts/chapters change.
func firstHeadingSeeds(blocks []Block, cfg Config) (part, chapter HeadingParts) {
	for _, block := range blocks {
		if block.Kind == "part" {
			part = composeHeadingParts(block, cfg.Folio.Manuscript.Part)
			return
		}
		if block.Kind == "chapter" {
			chapter = composeHeadingParts(block, cfg.Folio.Manuscript.Chapter)
			return
		}
	}
	return
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
	partSkipHeader := cfg.Folio.Manuscript.Part.SkipHeader
	partSkipFooter := cfg.Folio.Manuscript.Part.SkipFooter
	chapterSkipHeader := cfg.Folio.Manuscript.Chapter.SkipHeader
	chapterSkipFooter := cfg.Folio.Manuscript.Chapter.SkipFooter

	emitDirective := func(lines []string, directive string) []string {
		if directive == "" {
			return lines
		}
		return append(lines, directive)
	}

	// Skip flags are propagated to folio-part / folio-chapter which record the resulting
	// page number in `folio-skip-header-pages` / `folio-skip-footer-pages` at context time.
	// The header/footer context reads that list via `.final()` and hides the running band only
	// on those pages, leaving subsequent body pages of a multi-page block unaffected.
	// emitPartState / emitChapterState pre-populate the semantic-authoring state at a source
	// position before the folio-part / folio-chapter call. This matters for the first body
	// page: its header context evaluates at page-top, and state updates inside the folio-part
	// macro body happen at the call position (which is after page-top for a first block).
	// Pre-emitting means header sees the correct [part] / [chapter] values on the very first
	// body page.
	emitPartState := func(lines []string, cp HeadingParts) []string {
		return append(lines,
			fmt.Sprintf(`#state("folio-current-part-name").update(%q)`, cp.Name),
			fmt.Sprintf(`#state("folio-current-part-number").update(%q)`, cp.Number),
			fmt.Sprintf(`#state("folio-current-part-prefix").update(%q)`, cp.Prefix),
			fmt.Sprintf(`#state("folio-current-part-full").update(%q)`, cp.Full),
			`#state("folio-current-chapter-name").update("")`,
			`#state("folio-current-chapter-number").update("")`,
			`#state("folio-current-chapter-prefix").update("")`,
			`#state("folio-current-chapter-full").update("")`,
		)
	}
	emitChapterState := func(lines []string, cp HeadingParts) []string {
		return append(lines,
			fmt.Sprintf(`#state("folio-current-chapter-name").update(%q)`, cp.Name),
			fmt.Sprintf(`#state("folio-current-chapter-number").update(%q)`, cp.Number),
			fmt.Sprintf(`#state("folio-current-chapter-prefix").update(%q)`, cp.Prefix),
			fmt.Sprintf(`#state("folio-current-chapter-full").update(%q)`, cp.Full),
		)
	}

	// For the FIRST block, the "current page" at the point of an enforce-right context is a
	// phantom empty page created by the TOC's pagebreak (or the fresh body page after any
	// front matter). Add that page number to the skip lists so its header/footer stay hidden.
	// Subsequent blocks' "current page" is a real content page, so we don't skip it there.
	seedFirstBlockPhantomSkip := func(lines []string, mode BlankPageMode) []string {
		if mode == BlankPageEnforceRight || mode == BlankPageEnforceLeft {
			return append(lines,
				`#context { let pg = counter(page).at(here()).first(); state("folio-skip-header-pages", ()).update(pages => pages + (pg,)); state("folio-skip-footer-pages", ()).update(pages => pages + (pg,)) }`,
			)
		}
		return lines
	}

	var lines []string
	firstPageBlock := true
	for _, block := range blocks {
		switch block.Kind {
		case "part":
			hc := cfg.Folio.Manuscript.Part
			if firstPageBlock {
				lines = seedFirstBlockPhantomSkip(lines, hc.BlankPageBefore)
			}
			lines = emitDirective(lines, hc.BlankPageBefore.TypstDirective())
			composed := composeHeadingParts(block, hc)
			lines = emitPartState(lines, composed)
			lines = append(lines, fmt.Sprintf(
				"#folio-part(first: %t, skip-header: %t, skip-footer: %t, name: %q, number: %q, prefix: %q, full: %q)[%s]",
				firstPageBlock,
				partSkipHeader,
				partSkipFooter,
				composed.Name,
				composed.Number,
				composed.Prefix,
				composed.Full,
				caseTransform(composed.Full, hc.CaseTransform),
			))
			lines = emitDirective(lines, hc.BlankPageAfter.TypstDirective())
			firstPageBlock = false
		case "chapter":
			hc := cfg.Folio.Manuscript.Chapter
			if firstPageBlock {
				lines = seedFirstBlockPhantomSkip(lines, hc.BlankPageBefore)
			}
			lines = emitDirective(lines, hc.BlankPageBefore.TypstDirective())
			composed := composeHeadingParts(block, hc)
			lines = emitChapterState(lines, composed)
			lines = append(lines, fmt.Sprintf(
				"#folio-chapter(first: %t, skip-header: %t, skip-footer: %t, name: %q, number: %q, prefix: %q, full: %q)[%s]",
				firstPageBlock,
				chapterSkipHeader,
				chapterSkipFooter,
				composed.Name,
				composed.Number,
				composed.Prefix,
				composed.Full,
				caseTransform(composed.Full, hc.CaseTransform),
			))
			lines = emitDirective(lines, hc.BlankPageAfter.TypstDirective())
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

// HeadingParts is the decomposed form of a rendered part or chapter heading.
// Full is the composed string that renders in the body; Name is the semantic
// name only (used by the [chapter] / [part] placeholders); Number is the
// formatted number (arabic, roman upper, or roman lower per NumberFormat);
// Prefix is the configured prefix as-is.
type HeadingParts struct {
	Name   string
	Number string
	Prefix string
	Full   string
}

func composeHeadingParts(block Block, hc HeadingConfig) HeadingParts {
	showName := true
	if hc.ShowName != nil {
		showName = *hc.ShowName
	}
	showNumber := false
	if hc.ShowNumber != nil {
		showNumber = *hc.ShowNumber
	}

	name := block.Name
	if name == "" {
		name = block.Text
	}
	name = applyNameCase(name, hc.NameCase)

	number := block.Number
	if hc.ExplicitNumbering == "source" && block.SourceNumber != "" {
		if parsed, err := strconv.Atoi(block.SourceNumber); err == nil {
			number = parsed
		}
	}
	numberStr := formatHeadingNumber(number, hc.NumberFormat)

	prefix := hc.Prefix

	var full strings.Builder
	if showNumber || showName {
		full.WriteString(prefix)
	}
	if showNumber {
		full.WriteString(numberStr)
	}
	if showNumber && showName && name != "" {
		full.WriteString(hc.Separator)
	}
	if showName {
		full.WriteString(name)
	}
	full.WriteString(hc.Suffix)

	return HeadingParts{
		Name:   name,
		Number: numberStr,
		Prefix: prefix,
		Full:   full.String(),
	}
}

// formatHeadingNumber renders an integer per the configured number-format:
//   - "1" or empty (default): arabic ("1", "2", "10")
//   - "I": roman upper ("I", "II", "X")
//   - "i": roman lower ("i", "ii", "x")
func formatHeadingNumber(n int, format string) string {
	switch format {
	case "I":
		return toRomanUpper(n)
	case "i":
		return strings.ToLower(toRomanUpper(n))
	default:
		return strconv.Itoa(n)
	}
}

// toRomanUpper converts a positive integer to upper-case Roman numerals.
// Returns the arabic form for zero or negative inputs (defensive fallback).
func toRomanUpper(n int) string {
	if n <= 0 {
		return strconv.Itoa(n)
	}
	values := []struct {
		v int
		s string
	}{
		{1000, "M"}, {900, "CM"}, {500, "D"}, {400, "CD"},
		{100, "C"}, {90, "XC"}, {50, "L"}, {40, "XL"},
		{10, "X"}, {9, "IX"}, {5, "V"}, {4, "IV"}, {1, "I"},
	}
	var b strings.Builder
	for _, val := range values {
		for n >= val.v {
			b.WriteString(val.s)
			n -= val.v
		}
	}
	return b.String()
}

// applyNameCase applies the configured name-case to the semantic name segment.
// Values: "" (as-written), "upper", "lower", "title".
func applyNameCase(name string, nameCase string) string {
	switch nameCase {
	case "upper":
		return strings.ToUpper(name)
	case "lower":
		return strings.ToLower(name)
	case "title":
		return simpleTitleCase(name)
	default:
		return name
	}
}

// simpleTitleCase capitalises the first rune of each whitespace-separated word and
// lower-cases the rest. Adequate for AC18.5's name-case: title requirement without
// pulling in golang.org/x/text/cases.
func simpleTitleCase(s string) string {
	var b strings.Builder
	upcomingWord := true
	for _, r := range s {
		if r == ' ' || r == '\t' || r == '\n' {
			b.WriteRune(r)
			upcomingWord = true
			continue
		}
		if upcomingWord {
			b.WriteRune(unicodeToUpper(r))
			upcomingWord = false
		} else {
			b.WriteRune(unicodeToLower(r))
		}
	}
	return b.String()
}

func unicodeToUpper(r rune) rune {
	if r >= 'a' && r <= 'z' {
		return r - 32
	}
	return r
}

func unicodeToLower(r rune) rune {
	if r >= 'A' && r <= 'Z' {
		return r + 32
	}
	return r
}

var placeholderRE = regexp.MustCompile(`\[([a-z-]+)\]`)

func renderHeader(meta Metadata, cfg Config) string {
	return substitutePlaceholders(cfg.Folio.Manuscript.PageHeader.Format, meta)
}

func renderHeaderAlt(meta Metadata, cfg Config) string {
	if cfg.Folio.Manuscript.PageHeader.AltFormat == "" {
		return ""
	}
	return substitutePlaceholders(cfg.Folio.Manuscript.PageHeader.AltFormat, meta)
}

func renderFooter(meta Metadata, cfg Config) string {
	return substitutePlaceholders(cfg.Folio.Manuscript.PageFooter.Format, meta)
}

func renderFooterAlt(meta Metadata, cfg Config) string {
	if cfg.Folio.Manuscript.PageFooter.AltFormat == "" {
		return ""
	}
	return substitutePlaceholders(cfg.Folio.Manuscript.PageFooter.AltFormat, meta)
}

// substitutePlaceholders resolves [author], [title], [page], [part], [chapter] placeholders in
// a header/footer format string. Recognized tokens are replaced with escaped metadata or the
// corresponding Typst state/counter reads; interstitial text is escapeTypst'd so unknown
// tokens like [unknown] survive as literal brackets in the rendered Typst source.
//
// When the format has a consistent separator between placeholders (e.g. " * " between three
// placeholders) and at least one dynamic placeholder ([part] or [chapter] whose value may be
// empty on some pages), the emission uses a Typst filter-join block so an empty state
// placeholder drops its surrounding separator too -- avoiding the "TITLE * * AUTHOR" quirk on
// pages preceding the first chapter. Formats with irregular separators fall back to naive
// substitution.
func substitutePlaceholders(format string, meta Metadata) string {
	if joined, ok := tryFilterJoin(format, meta); ok {
		return joined
	}
	return naiveSubstitute(format, meta)
}

// tryFilterJoin returns a Typst filter-join emission if the format's placeholders are
// separated by a single consistent literal separator (leading and trailing literals may vary).
// Returns ok=false if the format has irregular separators, no placeholders, or only literal
// text; the caller falls back to naiveSubstitute.
func tryFilterJoin(format string, meta Metadata) (string, bool) {
	matches := placeholderRE.FindAllStringSubmatchIndex(format, -1)
	if len(matches) < 2 {
		return "", false
	}
	sep := format[matches[0][1]:matches[1][0]]
	for i := 1; i < len(matches)-1; i++ {
		if format[matches[i][1]:matches[i+1][0]] != sep {
			return "", false
		}
	}
	if sep == "" {
		return "", false
	}
	leading := format[:matches[0][0]]
	trailing := format[matches[len(matches)-1][1]:]
	items := make([]string, 0, len(matches))
	for _, m := range matches {
		name := format[m[2]:m[3]]
		items = append(items, placeholderItemExpr(name, meta, format[m[0]:m[1]]))
	}
	var out strings.Builder
	out.WriteString(escapeTypst(leading))
	out.WriteString("#{ (")
	for i, item := range items {
		if i > 0 {
			out.WriteString(", ")
		}
		out.WriteString(item)
	}
	out.WriteString(`).filter(x => x != none and x != "" and x != []).join(`)
	out.WriteString("[" + escapeTypst(sep) + "]")
	out.WriteString(") }")
	out.WriteString(escapeTypst(trailing))
	return out.String(), true
}

// placeholderItemExpr returns the Typst expression for a placeholder in an array-item context.
// Unknown placeholders return their literal bracketed text (escaped) so the fallback naive
// substitute path is symmetric.
func placeholderItemExpr(name string, meta Metadata, literal string) string {
	switch name {
	case "author":
		if meta.Author == "" {
			return "none"
		}
		return "[" + escapeTypst(meta.Author) + "]"
	case "title":
		if meta.Title == "" {
			return "none"
		}
		return "[" + escapeTypst(meta.Title) + "]"
	case "page":
		return "folio-display-page()"
	// AC18.4: [part] returns the semantic name only (folio-current-part-name state).
	case "part":
		return placeholderStateGuard("folio-current-part-name")
	case "part-number":
		return placeholderStateGuard("folio-current-part-number")
	case "part-prefix":
		return placeholderStateGuard("folio-current-part-prefix")
	case "part-full":
		return placeholderStateGuard("folio-current-part-full")
	case "chapter":
		return placeholderStateGuard("folio-current-chapter-name")
	case "chapter-number":
		return placeholderStateGuard("folio-current-chapter-number")
	case "chapter-prefix":
		return placeholderStateGuard("folio-current-chapter-prefix")
	case "chapter-full":
		return placeholderStateGuard("folio-current-chapter-full")
	default:
		return "[" + escapeTypst(literal) + "]"
	}
}

// placeholderStateGuard returns a Typst expression that reads a state value and
// evaluates to `none` when the state is empty (unset, empty string, or empty content).
func placeholderStateGuard(stateName string) string {
	return `{ let v = state(` + strconv.Quote(stateName) + `).get(); if v == none or v == "" or v == [] { none } else { v } }`
}

// naiveSubstitute is the pre-smart-join substitution kept as the fallback for formats with
// irregular separators or a single placeholder.
func naiveSubstitute(format string, meta Metadata) string {
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
			out.WriteString("#folio-display-page()")
		// AC18.4 extended placeholders.
		case "part":
			out.WriteString(`#context state("folio-current-part-name").get()`)
		case "part-number":
			out.WriteString(`#context state("folio-current-part-number").get()`)
		case "part-prefix":
			out.WriteString(`#context state("folio-current-part-prefix").get()`)
		case "part-full":
			out.WriteString(`#context state("folio-current-part-full").get()`)
		case "chapter":
			out.WriteString(`#context state("folio-current-chapter-name").get()`)
		case "chapter-number":
			out.WriteString(`#context state("folio-current-chapter-number").get()`)
		case "chapter-prefix":
			out.WriteString(`#context state("folio-current-chapter-prefix").get()`)
		case "chapter-full":
			out.WriteString(`#context state("folio-current-chapter-full").get()`)
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
