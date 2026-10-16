// ABOUTME: Resolves manuscript YAML roles into independent native prose styles.
// ABOUTME: Keeps manuscript defaults, font-relative units and generated-matter policy at the mode boundary.
package manuscript

import (
	"strconv"
	"strings"

	"github.com/tigger-developer/first-folio/internal/office"
)

func officeLayout(cfg Config) (office.Prose, error) {
	ms := cfg.Folio.Manuscript
	p := office.Prose{Anchors: map[string]string{}, Styles: map[string]office.Style{}, Boundaries: map[string]office.Boundary{}, Placements: map[string]office.Placement{}, Images: map[string][]byte{}, InitialNumberFormat: ms.PageNumbering.FrontmatterFormat}
	p.Ends = map[string]office.Boundary{}
	body, err := office.ResolveStyle(ms.Font, 12)
	if err != nil {
		return p, err
	}
	base := body.Size
	length := func(value string, em, extent float64) float64 {
		v, e := office.Points(value, em, extent)
		if err == nil {
			err = e
		}
		return v
	}
	body.Line = (office.CapHeight(body) + 0.65*base) * parseMultiplier(ms.LineSpacing)
	body.After = length(ms.ParagraphSpacing, base, 0)
	body.First = length(ms.ParagraphIndent, base, 0)
	body.Hyphenate = ms.Hyphenation == "true" || ms.Hyphenation == "auto" && ms.Justify != nil && *ms.Justify
	body.Widow = ms.WidowOrphanControl != nil && *ms.WidowOrphanControl
	if ms.Justify != nil && *ms.Justify {
		body.Align = "justify"
	}
	p.Styles["FolioBody"] = body
	for _, name := range []string{"Normal", "BodyText", "FirstParagraph", "Compact", "ListParagraph", "Text_20_body", "Standard", "Table_20_Contents", "Table_20_Heading"} {
		p.Styles[name] = body
	}
	p.ListBefore = length(ms.List.SpaceBefore, base, 0)
	p.ListAfter = length(ms.List.SpaceAfter, base, 0)
	p.TableBefore = length(ms.Table.SpaceBefore, base, 0)
	p.TableAfter = length(ms.Table.SpaceAfter, base, 0)
	opening := body
	opening.First = 0
	p.Styles["FolioOpening"] = opening
	addFont := func(name string, font FontConfig) office.Style {
		s, e := office.ResolveStyle(font, base)
		if err == nil {
			err = e
		}
		s.Line = body.Line
		s.After = body.After
		p.Styles[name] = s
		return s
	}
	heading := addFont("FolioHeading", ms.Heading.Font)
	heading.KeepNext = true
	p.Styles["FolioHeading"] = heading
	quote := addFont("FolioQuote", ms.QuotedBlock.Font)
	quote.Left = length(ms.QuoteBlockIndent.Value, base, 0)
	quote.Before = length(ms.QuotedBlockSpacing.Value, base, 0)
	quote.After += quote.Before
	quote.Hyphenate = body.Hyphenate
	p.Styles["FolioQuote"] = quote
	code := addFont("FolioCode", ms.Mono.Font)
	code.Left = length(ms.CodeBlockIndent.Value, base, 0)
	code.Before = length(ms.CodeBlock.SpaceBefore, base, 0)
	code.After += length(ms.CodeBlock.SpaceAfter, base, 0)
	code.Hyphenate = false
	p.Styles["FolioCode"] = code
	mono := code
	mono.Character = true
	p.Styles["VerbatimChar"] = mono
	p.Styles["SourceText"] = mono
	p.Styles["Source_20_Text"] = mono
	p.Styles["FolioMono"] = mono
	p.Styles["SourceCode"] = code
	p.Styles["Preformatted_20_Text"] = code
	note := body
	note.Size *= 0.85
	note.First = 0
	note.Line = office.CapHeight(note) + 0.65*note.Size
	p.Styles["FootnoteText"] = note
	p.Styles["Footnote"] = note
	scene := body
	scene.Align = "center"
	scene.First = 0
	scene.Before = base
	scene.After += base
	p.Styles["FolioScene"] = scene
	for _, pair := range []struct {
		name string
		font FontConfig
	}{{"Title", ms.TitlePage.Title.Font}, {"Subtitle", ms.TitlePage.Subtitle.Font}, {"Author", ms.TitlePage.Author.Font}, {"Date", ms.TitlePage.Date.Font}, {"Wordcount", ms.TitlePage.WordCount.Font}, {"Version", ms.TitlePage.Version.Font}, {"Contact", ms.TitlePage.Contact.Font}, {"Copyright", ms.Copyright.Font}, {"CopyrightLabel", ms.Copyright.Label.Font}, {"Contents", ms.TOC.Font}, {"ContentsHeading", ms.TOC.Heading.Font}} {
		addFont("Folio"+pair.name, pair.font)
	}
	copyright := p.Styles["FolioCopyright"]
	copyright.Line = office.CapHeight(copyright) + (parseMultiplier(ms.Copyright.LineSpacing)-1)*copyright.Size
	copyright.After = copyright.Size*0.8 + length(ms.Copyright.BlockSpacing, copyright.Size, 0)
	copyright.Align = ms.Copyright.Align
	p.Styles["FolioCopyright"] = copyright
	label := p.Styles["FolioCopyrightLabel"]
	label.Character = true
	p.Styles["FolioCopyrightLabel"] = label
	contents := p.Styles["FolioContents"]
	contents.Line = (office.CapHeight(contents) + 0.65*contents.Size) * parseMultiplier(ms.TOC.LineSpacing)
	contents.After = 0
	p.Styles["FolioContents"] = contents
	if err != nil {
		return p, err
	}
	p.Page.Width, p.Page.Height, err = office.Paper(ms.Page)
	if err != nil {
		return p, err
	}
	margin := length(ms.Margin, base, 0)
	p.Page.Top = margin
	p.Page.Bottom = margin
	p.Page.Left = margin
	p.Page.Right = margin
	p.Page.Gutter = length(ms.Gutter, base, 0)
	bands := []struct {
		target               *office.Band
		font                 FontConfig
		enabled              bool
		format, alt          string
		front, frontAlt      *string
		align, edge, padding string
	}{
		{&p.Page.Header, ms.PageHeader.Font, ms.PageHeader.Enabled, ms.PageHeader.Format, ms.PageHeader.AltFormat, ms.PageHeader.FrontmatterFormat, ms.PageHeader.AltFrontmatterFormat, ms.PageHeader.Align, ms.PageHeader.DistanceFromEdge, ms.PageHeader.ContentPaddingAfter},
		{&p.Page.Footer, ms.PageFooter.Font, ms.PageFooter.Enabled != nil && *ms.PageFooter.Enabled, ms.PageFooter.Format, ms.PageFooter.AltFormat, ms.PageFooter.FrontmatterFormat, ms.PageFooter.AltFrontmatterFormat, ms.PageFooter.Align, ms.PageFooter.DistanceFromEdge, ms.PageFooter.ContentPaddingAfter},
	}
	for i, b := range bands {
		s, e := office.ResolveStyle(b.font, base)
		if err == nil {
			err = e
		}
		s.Line = office.CapHeight(s) + 0.65*s.Size
		b.target.Font = s
		b.target.Enabled = b.enabled
		b.target.Even = b.format
		b.target.Odd = b.alt
		if b.alt == "" {
			b.target.Odd = b.format
		}
		b.target.FrontEven = b.target.Even
		b.target.FrontOdd = b.target.Odd
		if b.front != nil {
			b.target.FrontEven = *b.front
		}
		if b.frontAlt != nil {
			b.target.FrontOdd = *b.frontAlt
		}
		b.target.Align = b.align
		b.target.Edge = length(b.edge, base, 0)
		b.target.Clearance = length(b.padding, base, 0)
		if b.enabled {
			if i == 0 {
				p.Page.Top = b.target.Edge + b.target.Clearance
			} else {
				p.Page.Bottom = b.target.Edge + b.target.Clearance
			}
		}
	}
	return p, err
}
func parseMultiplier(value string) float64 {
	v, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 1
	}
	return v
}
func officeChapterPosition(value string, page office.Page, base float64) (float64, error) {
	switch strings.ToLower(value) {
	case "top":
		value = "0pt"
	case "center", "middle", "horizon":
		value = "50%"
	case "one-third":
		value = "30%"
	}
	return office.Points(value, base, page.Height-page.Top-page.Bottom)
}
