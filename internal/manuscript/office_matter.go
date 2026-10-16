// ABOUTME: Composes editable manuscript title, contents and copyright matter.
// ABOUTME: Reuses existing metadata and barcode rules while recording native field/placement plans.
package manuscript

import (
	"fmt"
	"strings"

	"github.com/tigger-developer/first-folio/internal/office"
)

func officeGeneratedMatter(p *office.Prose, meta Metadata, cfg Config) error {
	var title, contents, legal []office.Node
	ms := cfg.Folio.Manuscript
	tp := ms.TitlePage
	if tp.Enabled {
		contact := []string{}
		for _, item := range []struct {
			include bool
			text    string
		}{{tp.IncludeContactName, meta.ContactName}, {tp.IncludeAddress, meta.Address}, {tp.IncludePhone, meta.Phone}, {tp.IncludeEmail, meta.Email}, {tp.IncludeWebsite, meta.Website}} {
			if item.include && item.text != "" {
				contact = append(contact, item.text)
			}
		}
		author := meta.Author
		if author != "" && meta.AuthorAttribution != "" {
			author = meta.AuthorAttribution + " " + author
		}
		for _, item := range []struct {
			role, text, align string
			include           bool
		}{{"Contact", strings.Join(contact, "\n"), tp.Contact.Align, len(contact) > 0}, {"Title", meta.Title, tp.Title.Align, tp.IncludeTitle}, {"Subtitle", meta.Subtitle, tp.Subtitle.Align, tp.IncludeSubtitle}, {"Author", author, tp.Author.Align, tp.IncludeAuthor}, {"Date", renderDate(meta.Date, ms.DateFormat), tp.Date.Align, tp.IncludeDate}, {"Wordcount", meta.WordCount, tp.WordCount.Align, tp.IncludeWordCount}, {"Version", meta.Version, tp.Version.Align, tp.IncludeVersion}} {
			if !item.include || item.text == "" {
				continue
			}
			role := "Folio" + item.role
			placement := officeTitlePlacement(item.role, item.align, tp)
			p.Placements[role] = placement
			style := p.Styles[role]
			style.Align = placement.Horizontal
			if item.align == "group" {
				if item.role == "Subtitle" {
					style.Before = p.Styles["FolioBody"].Size
				}
				if item.role == "Author" {
					style.Before = 2 * p.Styles["FolioBody"].Size
				}
			}
			style.Line = office.CapHeight(style) + 0.65*style.Size
			p.Styles[role] = style
			title = append(title, office.Paragraph(role, officeTextLines(item.text)))
		}
		if len(title) > 0 {
			first := title[0].Content.([]any)[0].([]any)[2].([][]string)[0][1]
			p.Boundaries[first] = office.Boundary{Title: true, TitleNumber: tp.PageNumber, SkipHeader: true, SkipFooter: !tp.PageNumber, NumberFormat: ms.PageNumbering.FrontmatterFormat}
			p.Ends[officeBlockRole(title[len(title)-1])] = office.Boundary{Break: ms.TOC.PageBreakBefore}
		}
	}
	if ms.TOC.Enabled {
		contents = append(contents, office.Paragraph("FolioContentsHeading", office.Text(ms.TOC.Title)))
		p.Boundaries["FolioContentsHeading"] = office.Boundary{Break: ms.TOC.PageBreakBefore, Before: string(ms.TOC.BlankPageBefore), After: string(ms.TOC.BlankPageAfter), NumberFormat: ms.PageNumbering.FrontmatterFormat}
		padding, err := office.Points(ms.TOC.ContinuationPad, p.Styles["FolioContents"].Size, 0)
		if err != nil {
			return err
		}
		boundary := p.Boundaries["FolioContentsHeading"]
		boundary.ContinuationPadding = padding
		p.Boundaries["FolioContentsHeading"] = boundary
		for i, node := range p.Document.Blocks {
			c, ok := node.Content.([]any)
			if node.Type != "Div" || !ok || len(c) != 2 {
				continue
			}
			children, ok := c[1].([]office.Node)
			if !ok || len(children) != 1 || children[0].Type != "Header" {
				continue
			}
			h := children[0].Content.([]any)
			level := int(h[0].(float64))
			if level == 1 && !ms.TOC.IncludeParts || level == 2 && !ms.TOC.IncludeChapters || level > 2 && !ms.TOC.IncludeSections {
				continue
			}
			role := fmt.Sprintf("FolioContentsEntry%d", i)
			style := p.Styles["FolioContents"]
			if level == 1 {
				style.Bold = ms.TOC.PartBold
				gap, err := office.Points(ms.TOC.PartGapBefore, style.Size, 0)
				if err != nil {
					return err
				}
				style.Before = gap
			}
			p.Styles[role] = style
			attr := h[1].([]any)
			anchor := attr[0].(string)
			p.Contents = append(p.Contents, office.ContentsEntry{Style: role, Anchor: anchor, Links: ms.TOC.Links, Leaders: ms.TOC.DotLeaders, Pages: ms.TOC.PageNumbers})
			contents = append(contents, office.Paragraph(role, office.Text(officeInlineText(h[2]))))
		}
	}
	if ms.Copyright.Enabled {
		var err error
		legal, err = officeCopyright(p, meta, cfg)
		if err != nil {
			return err
		}
	}
	if len(contents) > 0 {
		p.Ends[officeBlockRole(contents[len(contents)-1])] = office.Boundary{Break: true, Before: string(ms.TOC.BlankPageAfter)}
	}
	if len(legal) > 0 {
		p.Ends[officeBlockRole(legal[len(legal)-1])] = office.Boundary{Break: true, Before: string(ms.Copyright.BlankPageAfter)}
	}
	blocks := append([]office.Node{}, title...)
	if ms.Copyright.Position == "after-title" {
		blocks = append(blocks, legal...)
	}
	blocks = append(blocks, contents...)
	if ms.Copyright.Position == "after-toc" || ms.Copyright.Position == "after-frontmatter" {
		blocks = append(blocks, legal...)
	}
	p.Document.Blocks = append(blocks, p.Document.Blocks...)
	return nil
}

func officeBlockRole(node office.Node) string {
	return node.Content.([]any)[0].([]any)[2].([][]string)[0][1]
}
func officePlacement(value string) office.Placement {
	parts := strings.Split(value, "-")
	if len(parts) == 2 {
		return office.Placement{Vertical: parts[0], Horizontal: parts[1]}
	}
	return office.Placement{Vertical: "center", Horizontal: value}
}
func officeTextLines(value string) []office.Node {
	var nodes []office.Node
	for i, line := range strings.Split(value, "\n") {
		if i > 0 {
			nodes = append(nodes, office.N("LineBreak", nil))
		}
		nodes = append(nodes, office.N("Str", line))
	}
	return nodes
}
func officeCopyright(p *office.Prose, meta Metadata, cfg Config) ([]office.Node, error) {
	c := cfg.Folio.Manuscript.Copyright
	var nodes []office.Node
	add := func(text string, bottom bool) error {
		raw, _, err := office.Pandoc([]byte(text), "--from=markdown", "--to=json")
		if err != nil {
			return err
		}
		doc, err := office.Decode(raw)
		if err != nil {
			return err
		}
		role := fmt.Sprintf("FolioLegal%d", len(nodes))
		p.Styles[role] = p.Styles["FolioCopyright"]
		if bottom {
			p.Placements[role] = office.Placement{Vertical: "bottom", Horizontal: c.Align, Group: "copyright-bottom"}
		}
		nodes = append(nodes, office.N("Div", []any{office.Attr("", nil, []string{"custom-style", role}), doc.Blocks}))
		return nil
	}
	for _, text := range append(effectiveCredits(meta, c), c.Body...) {
		if strings.TrimSpace(text) != "" {
			if err := add(text, false); err != nil {
				return nil, err
			}
		}
	}
	if c.Separator != "" {
		role := "FolioLegalSeparator"
		style := p.Styles["FolioCopyright"]
		var err error
		style.Before, err = office.Points(c.SeparatorSpaceBefore, style.Size, 0)
		if err != nil {
			return nil, err
		}
		style.After, err = office.Points(c.SeparatorSpaceAfter, style.Size, 0)
		if err != nil {
			return nil, err
		}
		style.Align = "center"
		p.Styles[role] = style
		nodes = append(nodes, office.Paragraph(role, office.Text(c.Separator)))
	}
	if len(c.Publication) > 0 {
		if err := add(strings.Join(c.Publication, "  \n"), true); err != nil {
			return nil, err
		}
	}
	if c.Publisher != "" {
		officeCopyrightLabel(p, &nodes, c.PublisherPreposition+" ", c.Publisher, "", c.Align)
	}
	if c.ISBN != "" {
		officeCopyrightLabel(p, &nodes, "", c.ISBNLabel, ": "+c.ISBN, c.Align)
	}
	if c.ISBNBarcode == "render" || c.ISBNBarcode == "render-and-file" {
		svg, err := renderEAN13SVG(c.ISBN)
		if err != nil {
			return nil, err
		}
		p.Images["FolioBarcode"] = []byte(svg)
		nodes = append(nodes, office.Paragraph("FolioBarcode", nil))
		p.Styles["FolioBarcode"] = p.Styles["FolioCopyright"]
	}
	if len(nodes) > 0 {
		role := nodes[0].Content.([]any)[0].([]any)[2].([][]string)[0][1]
		p.Boundaries[role] = office.Boundary{Before: string(c.BlankPageBefore), After: string(c.BlankPageAfter), SkipHeader: c.SkipHeader != nil && *c.SkipHeader, SkipFooter: c.SkipFooter != nil && *c.SkipFooter, NumberFormat: cfg.Folio.Manuscript.PageNumbering.FrontmatterFormat}
	}
	return nodes, nil
}

func officeCopyrightLabel(p *office.Prose, nodes *[]office.Node, prefix, label, value, align string) {
	role := fmt.Sprintf("FolioLegal%d", len(*nodes))
	p.Styles[role] = p.Styles["FolioCopyright"]
	p.Placements[role] = office.Placement{Vertical: "bottom", Horizontal: align, Group: "copyright-bottom"}
	inlines := office.Text(prefix)
	inlines = append(inlines, office.Span("FolioCopyrightLabel", office.Text(label)))
	inlines = append(inlines, office.Text(value)...)
	*nodes = append(*nodes, office.Paragraph(role, inlines))
}

// Group footer fields occupy the bottom band: version left, count in the
// configured middle alignment, date right. Explicit item placements stay independent.
func officeTitlePlacement(role, align string, title TitlePageConfig) office.Placement {
	group := "title"
	if role == "Date" || role == "Wordcount" || role == "Version" {
		group = "footer"
	}
	if align == "group" {
		align = title.TitleBlockAlign
		if role == "Contact" {
			align = "top-left"
		}
		if group == "footer" {
			align = "bottom-" + officePlacement(title.FooterAlign).Horizontal
			if role == "Date" {
				align = "bottom-right"
			}
			if role == "Version" {
				align = "bottom-left"
			}
		}
	}
	placement := officePlacement(align)
	placement.Group = group
	return placement
}

// Vertical alignment occupies a page only when a break follows the part.
// In continuous flow, the PDF renderer keeps it before the following content;
// an independent page-centred frame would instead overlap later paragraphs.
func officePartOwnPage(blocks []office.Node, index int, cfg ManuscriptConfig) bool {
	if cfg.Part.BlankPageAfter != "" && cfg.Part.BlankPageAfter != "false" || index+1 == len(blocks) {
		return true
	}
	next := blocks[index+1]
	if next.Type != "Header" {
		return false
	}
	content, ok := next.Content.([]any)
	if !ok || len(content) != 3 {
		return false
	}
	level, ok := content[0].(float64)
	if !ok || level > 2 {
		return false
	}
	heading := cfg.Chapter
	if level == 1 {
		heading = cfg.Part
	}
	return heading.PageBreakBefore || heading.BlankPageBefore != "" && heading.BlankPageBefore != "false"
}
