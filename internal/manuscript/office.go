// ABOUTME: Adapts canonical manuscript content directly into structured office output.
// ABOUTME: Reuses manuscript source/configuration and heading semantics before any Typst rendering.
package manuscript

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/tigger-developer/first-folio/internal/office"
)

func runOffice(inputs InputSet, opts Options, stdout, stderr io.Writer) error {
	raws, err := office.ReadSources(inputs.Paths)
	if err != nil {
		return err
	}
	var text string
	var source map[string]any
	if inputs.Format == "markdown" {
		byPath := map[string][]byte{}
		for i, path := range inputs.Paths {
			byPath[path] = raws[i]
		}
		text, source, err = readConfiguredMarkdown(inputs, func(path string) ([]byte, error) { return byPath[path], nil })
	} else {
		var b strings.Builder
		for i, raw := range raws {
			if i > 0 {
				b.WriteString("\n\n")
			}
			b.Write(raw)
			b.WriteString("\n")
		}
		text = b.String()
	}
	if err != nil {
		return err
	}
	cfg, err := loadConfigWithSource(filepath.Dir(inputs.Paths[0]), opts, source)
	if err != nil {
		return err
	}
	for _, warning := range cfg.Warnings {
		if _, err = fmt.Fprintln(stderr, "Warning:", warning); err != nil {
			return err
		}
	}
	doc, meta, err := parseOffice(inputs.Format, text)
	if err != nil {
		return err
	}
	applyMetadataOverrides(&meta, opts, cfg)
	format := strings.ToLower(strings.TrimPrefix(filepath.Ext(opts.Output), "."))
	if opts.DryRun {
		printDryRun(stdout, inputs, opts, cfg)
		_, err = fmt.Fprintln(stdout, "output format:", format)
		return err
	}
	if err = office.CheckTarget(inputs.Paths, opts.Output); err != nil {
		return err
	}
	sidecar := ""
	var svg []byte
	c := cfg.Folio.Manuscript.Copyright
	if c.Enabled && c.ISBN != "" && (c.ISBNBarcode == "file" || c.ISBNBarcode == "render-and-file") {
		sidecar = strings.TrimSuffix(opts.Output, filepath.Ext(opts.Output)) + ".barcode.svg"
		if err = office.CheckTarget(inputs.Paths, sidecar); err != nil {
			return err
		}
		value, e := renderEAN13SVG(c.ISBN)
		if e != nil {
			return e
		}
		svg = []byte(value)
	}
	plan, err := officePlan(doc, meta, cfg)
	if err != nil {
		return err
	}
	raw, warnings, err := plan.Export(format)
	if err != nil {
		return err
	}
	for _, warning := range warnings {
		if _, err = fmt.Fprintln(stderr, "Warning:", warning); err != nil {
			return err
		}
	}
	return office.PublishManuscript(opts.Output, raw, sidecar, svg)
}

func parseOffice(format, text string) (office.Document, Metadata, error) {
	var meta Metadata
	var err error
	if format == "org" {
		meta = parseOrgMetadataBlock(text)
		text = replaceOrgSectionBreaksWithSentinel(removeOrgPrivateSections(text))
		raw, _, e := office.Pandoc([]byte(text), "--from=org", "--to=gfm")
		if e != nil {
			return office.Document{}, meta, e
		}
		text = normalizePandocMarkdownEscapes(restoreSectionBreakSentinel(stripRawOrgBlocks(string(raw))))
	}
	text, err = parseMarkdownFrontmatter(&meta, text)
	if err != nil {
		return office.Document{}, meta, err
	}
	if err = rejectUnsupportedMarkdown(text); err != nil {
		return office.Document{}, meta, err
	}
	text = promoteStandaloneMarkdownCodeSpans(removeMarkdownPrivateSections(text))
	raw, _, err := office.Pandoc([]byte(text), "--from=markdown-raw_html", "--to=json")
	if err != nil {
		return office.Document{}, meta, err
	}
	doc, err := office.Decode(raw)
	return doc, meta, err
}

func officePlan(doc office.Document, meta Metadata, cfg Config) (office.Prose, error) {
	plan, err := officeLayout(cfg)
	if err != nil {
		return plan, err
	}
	plan.Document = doc
	plan.Metadata = map[string]string{"title": meta.Title, "author": meta.Author}
	plan.Document.Meta = map[string]any{"title": office.N("MetaString", meta.Title), "author": office.N("MetaString", meta.Author)}
	var blocks []office.Node
	part, chapter := 0, 0
	previousChapter := false
	headingFields := map[string]string{}
	seeded := map[string]bool{}
	for i, node := range doc.Blocks {
		role := "FolioBody"
		switch node.Type {
		case "Header":
			c, ok := node.Content.([]any)
			if !ok || len(c) != 3 {
				return plan, fmt.Errorf("invalid Pandoc heading")
			}
			level, ok := c[0].(float64)
			if !ok {
				return plan, fmt.Errorf("invalid Pandoc heading level")
			}
			name := officeInlineText(c[2])
			heading := Block{Text: name, Name: name, Level: int(level)}
			hc := cfg.Folio.Manuscript.Chapter
			if level == 1 {
				part++
				if cfg.Folio.Manuscript.Chapter.NumberReset == "per-part" {
					chapter = 0
				}
				heading.Kind = "part"
				heading.Number = part
				heading.Name, heading.SourceNumber, heading.SourceSeparator = splitHeadingPrefix(name, partPrefixRE)
				hc = cfg.Folio.Manuscript.Part
			}
			if level == 2 {
				chapter++
				heading.Kind = "chapter"
				heading.Number = chapter
				heading.Name, heading.SourceNumber, heading.SourceSeparator = splitHeadingPrefix(name, chapterPrefixRE)
			}
			if level <= 2 {
				parts := composeHeadingParts(heading, hc)
				if level == 2 {
					parts = composeChapterHeadingParts(heading, hc, part)
				}
				c[2] = officeHeadingInlines(c[2], heading, parts, hc)
				kind := "chapter"
				if level == 1 {
					kind = "part"
					for _, suffix := range []string{"", "-number", "-prefix", "-full"} {
						headingFields["chapter"+suffix] = ""
					}
				}
				values := map[string]string{kind: parts.Name, kind + "-number": parts.Number, kind + "-prefix": parts.Prefix, kind + "-full": parts.Full}
				for key, value := range values {
					headingFields[key] = value
					if !seeded[kind] {
						plan.Metadata[key] = value
					}
				}
				seeded[kind] = true
			}
			role = fmt.Sprintf("FolioHeading%d", i)
			style := plan.Styles["FolioHeading"]
			style.Align = hc.Align
			boundary := office.Boundary{Before: string(hc.BlankPageBefore), After: string(hc.BlankPageAfter), Break: hc.PageBreakBefore, Body: level <= 2, NumberFormat: cfg.Folio.Manuscript.PageNumbering.BodyFormat, Reset: (part+chapter == 1 && cfg.Folio.Manuscript.PageNumbering.BodyReset != "never")}
			boundary.SkipHeader = hc.SkipHeader
			boundary.SkipFooter = hc.SkipFooter
			if level == 1 {
				boundary.Vertical = hc.VerticalAlign
				if hc.VerticalAlign != "" && hc.VerticalAlign != "top" {
					plan.Placements[role] = office.Placement{Vertical: hc.VerticalAlign, Horizontal: hc.Align, Group: role}
				}
			}
			if level == 2 {
				style.Before, err = officeChapterPosition(hc.Position, plan.Page, plan.Styles["FolioBody"].Size)
				if err != nil {
					return plan, err
				}
				style.After, err = office.Points(hc.SpaceAfter, style.Size, 0)
				if err != nil {
					return plan, err
				}
			}
			plan.Anchors[fmt.Sprintf("folio-heading-%d", i)] = role
			plan.Styles[role] = style
			if level <= 2 {
				boundary.Fields = map[string]string{}
				for key, value := range headingFields {
					boundary.Fields[key] = value
				}
				plan.Boundaries[role] = boundary
				plan.Ends[role] = office.Boundary{Before: boundary.After}
			}
			c[1] = office.Attr(fmt.Sprintf("folio-heading-%d", i), nil, []string{"custom-style", role})
			node.Content = c
			// Pandoc custom paragraph styles are carried by a Div around the heading.
			blocks = append(blocks, office.N("Div", []any{office.Attr("", nil, []string{"custom-style", role}), []office.Node{node}}))
			previousChapter = level == 2
			continue
		case "HorizontalRule":
			blocks = append(blocks, office.Paragraph("FolioScene", office.Text(cfg.Folio.Manuscript.SceneBreak.Marker)))
			previousChapter = false
			continue
		case "Para", "Plain":
			if previousChapter {
				role = "FolioOpening"
			}
		case "CodeBlock":
			role = "FolioCode"
			c, ok := node.Content.([]any)
			if !ok || len(c) != 2 {
				return plan, fmt.Errorf("invalid Pandoc code block")
			}
			text, ok := c[1].(string)
			if !ok {
				return plan, fmt.Errorf("invalid Pandoc code text")
			}
			blocks = append(blocks, office.Paragraph(role, officeTextLines(text)))
			previousChapter = false
			continue
		case "BlockQuote":
			role = "FolioQuote"
			blocks = append(blocks, office.N("Div", []any{office.Attr("", nil, []string{"custom-style", role}), node.Content}))
			previousChapter = false
			continue
		}
		previousChapter = false
		blocks = append(blocks, office.N("Div", []any{office.Attr("", nil, []string{"custom-style", role}), []office.Node{node}}))
	}
	plan.Document.Blocks = blocks
	if err = officeGeneratedMatter(&plan, meta, cfg); err != nil {
		return plan, err
	}
	return plan, nil
}
func officeInlineText(value any) string {
	raw, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	var v any
	if json.Unmarshal(raw, &v) != nil {
		return ""
	}
	var b strings.Builder
	var visit func(any)
	visit = func(v any) {
		switch n := v.(type) {
		case []any:
			for _, x := range n {
				visit(x)
			}
		case map[string]any:
			kind, _ := n["t"].(string)
			switch kind {
			case "Str":
				if s, ok := n["c"].(string); ok {
					b.WriteString(s)
				}
			case "Space", "SoftBreak", "LineBreak":
				b.WriteByte(' ')
			case "Code":
				if c, ok := n["c"].([]any); ok && len(c) == 2 {
					b.WriteString(fmt.Sprint(c[1]))
				}
			default:
				visit(n["c"])
			}
		}
	}
	visit(v)
	return b.String()
}
