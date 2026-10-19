// ABOUTME: Converts script events into native DOCX paragraphs, tables and footnotes.
// ABOUTME: Interprets source-aware inline emphasis without executable markup.
package office

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/tigger-developer/first-folio/internal/play"
)

type docxInline struct{ text, kind string }

var docxNoteRE = regexp.MustCompile(`\[(?:fn:|\^)([^\]\n]+)\]`)

func (w *docxWriter) inlines(text string, notes bool) ([]docxNode, error) {
	w.textBytes += len(text)
	if w.textBytes > docxSourceLimit {
		return nil, fmt.Errorf("DOCX expanded text exceeds 32 MiB")
	}
	segments := []docxInline{{text: text}}
	starKind := "italic"
	if w.sourceFormat == play.FormatOrg {
		starKind = "bold"
	}
	rules := []struct {
		re   *regexp.Regexp
		kind string
	}{
		{boldItalicRE, "bolditalic"}, {boldRE, "bold"}, {codeTickRE, "code"}, {codeOrgRE, "code"}, {orgItalicRE, "italic"}, {underlineRE, "underline"}, {starPairRE, starKind},
	}
	if notes {
		rules = append([]struct {
			re   *regexp.Regexp
			kind string
		}{{docxNoteRE, "note"}}, rules...)
	}
	for _, rule := range rules {
		var next []docxInline
		for _, s := range segments {
			if s.kind != "" {
				next = append(next, s)
				continue
			}
			pos := 0
			for _, m := range rule.re.FindAllStringSubmatchIndex(s.text, 100001) {
				if rule.kind == "note" {
					if _, ok := w.notes[s.text[m[2]:m[3]]]; !ok {
						continue
					}
				}
				if m[0] > pos {
					next = append(next, docxInline{text: s.text[pos:m[0]]})
				}
				next = append(next, docxInline{text: s.text[m[2]:m[3]], kind: rule.kind})
				pos = m[1]
			}
			if pos < len(s.text) {
				next = append(next, docxInline{text: s.text[pos:]})
			}
		}
		if len(next) > 100000 {
			return nil, fmt.Errorf("DOCX exceeds 100000 inline segments")
		}
		segments = next
	}
	w.inlineCount += len(segments)
	if w.inlineCount > 100000 {
		return nil, fmt.Errorf("DOCX exceeds 100000 inline segments")
	}
	var out []docxNode
	for _, s := range segments {
		if s.kind == "note" {
			node, err := w.note(s.text)
			if err != nil {
				return nil, err
			}
			out = append(out, node)
			continue
		}
		props := dx("rPr")
		switch s.kind {
		case "bold":
			props.children = append(props.children, dx("b"))
		case "bolditalic":
			props.children = append(props.children, dx("b"), dx("i"))
		case "italic":
			props.children = append(props.children, dx("i"))
		case "underline":
			props.children = append(props.children, dxVal("u", "single"))
		case "code":
			props.children = append(props.children, dx("rFonts").attr("w:ascii", "Courier New").attr("w:hAnsi", "Courier New"))
		}
		out = append(out, docxRun(s.text, props))
	}
	return out, nil
}

func docxRun(text string, props docxNode) docxNode {
	run := dx("r", props)
	start := 0
	for i, r := range text {
		if r != '\n' && r != '\t' {
			continue
		}
		if i > start {
			run.children = append(run.children, dxText(text[start:i]))
		}
		if r == '\n' {
			run.children = append(run.children, dx("br"))
		} else {
			run.children = append(run.children, dx("tab"))
		}
		start = i + 1
	}
	if start < len(text) {
		run.children = append(run.children, dxText(text[start:]))
	}
	return run
}

func (w *docxWriter) note(key string) (docxNode, error) {
	id := len(w.footnotes) + 1
	content, err := w.inlines(w.notes[key], false)
	if err != nil {
		return docxNode{}, err
	}
	style, err := w.style("Footnote")
	if err != nil {
		return docxNode{}, err
	}
	p := dx("p", dx("pPr", dxVal("pStyle", style)), dx("r", dx("rPr", dxVal("vertAlign", "superscript")), dx("footnoteRef")), docxRun(" ", dx("rPr")))
	p.children = append(p.children, content...)
	w.footnotes = append(w.footnotes, dx("footnote", p).attr("w:id", dxInt(id)))
	return dx("r", dx("rPr", dxVal("vertAlign", "superscript")), dx("footnoteReference").attr("w:id", dxInt(id))), nil
}

func (w *docxWriter) paragraph(role, text string, properties ...docxNode) (docxNode, error) {
	id, err := w.style(role)
	if err != nil {
		return docxNode{}, err
	}
	runs, err := w.inlines(text, true)
	if err != nil {
		return docxNode{}, err
	}
	props := dx("pPr", dxVal("pStyle", id))
	props.children = append(props.children, properties...)
	p := dx("p", props)
	p.children = append(p.children, runs...)
	return p, nil
}

func (w *docxWriter) literalParagraph(role, text string) (docxNode, error) {
	id, err := w.style(role)
	if err != nil {
		return docxNode{}, err
	}
	w.textBytes += len(text)
	if w.textBytes > docxSourceLimit {
		return docxNode{}, fmt.Errorf("DOCX expanded text exceeds 32 MiB")
	}
	return dx("p", dx("pPr", dxVal("pStyle", id)), docxRun(text, dx("rPr"))), nil
}

func (w *docxWriter) body() ([]docxNode, error) {
	out, err := w.title()
	if err != nil {
		return nil, err
	}
	titleCount := len(out)
	firstAct := true
	sceneTable := -1
	appendScene := func(table docxNode) {
		if sceneTable < 0 {
			sceneTable = len(out)
			out = append(out, table)
		} else {
			out[sceneTable].children = append(out[sceneTable].children, table.children[2:]...)
		}
	}
	for i := 0; i < len(w.doc.Events); i++ {
		event := w.doc.Events[i]
		if event.Kind == play.EventCharacter {
			var dialogue []string
			for i+1 < len(w.doc.Events) && w.doc.Events[i+1].Kind == play.EventDialogue {
				i++
				dialogue = append(dialogue, w.doc.Events[i].Text)
			}
			speech, e := w.speech(event, strings.Join(dialogue, "\n"))
			if e != nil {
				return nil, e
			}
			if w.sameLine {
				appendScene(speech[0])
			} else {
				out = append(out, speech...)
			}
			continue
		}
		if event.Kind == play.EventCharacterTableStart {
			sceneTable = -1
			heading, e := w.paragraph("Intro", defaultString(event.Text, "Characters"))
			if e != nil {
				return nil, e
			}
			out = append(out, heading)
			table := w.table(w.width-2*w.margin, (w.width-2*w.margin)*3/10)
			for i+1 < len(w.doc.Events) && w.doc.Events[i+1].Kind == play.EventCharacterTableRow {
				i++
				row := w.doc.Events[i]
				primary, _ := play.ParseCastName(row.Name)
				left, e := w.paragraph("Cast", primary)
				if e != nil {
					return nil, e
				}
				right, e := w.paragraph("Cast", row.Text)
				if e != nil {
					return nil, e
				}
				table.children = append(table.children, w.tableRow((w.width-2*w.margin)*3/10, []docxNode{left}, []docxNode{right}))
			}
			if len(table.children) > 2 {
				out = append(out, table)
			}
			continue
		}
		role := ""
		var properties []docxNode
		switch event.Kind {
		case play.EventActHeader:
			role = "Act"
			if w.cfg.Bool("folio.positioning.act-header.page-break-before", true) && (!firstAct || w.cfg.Bool("folio.title-page.enabled", true) && len(out) > 0) {
				properties = append(properties, dx("pageBreakBefore"))
			}
			firstAct = false
		case play.EventSceneHeader:
			role = "Scene"
		case play.EventIntroHeader:
			role = "Intro"
		case play.EventStageDirection:
			role = "Direction"
		case play.EventTransition:
			role = "Transition"
		case play.EventPropText:
			role = "Prop"
		case play.EventIntroText:
			role = "FrontmatterParagraph"
		case play.EventDialogue:
			role = "Body"
		default:
			continue
		}
		text := event.Text
		if role == "Act" || role == "Scene" || role == "Transition" {
			text = docxCase(text, w.cfg.String(docxRolePath(role)+".case-transform", "as-written"))
		}
		p, e := w.paragraph(role, text, properties...)
		if e != nil {
			return nil, e
		}
		if w.sameLine && (role == "Direction" || role == "Transition" || role == "Prop" || event.Kind == play.EventDialogue) {
			table, _, e := w.sceneTable()
			if e != nil {
				return nil, e
			}
			cell := dx("tc", dx("tcPr", dx("tcW").attr("w:w", dxInt(w.width-2*w.margin)).attr("w:type", "dxa"), dxVal("gridSpan", "2")), p)
			table.children = append(table.children, dx("tr", cell))
			appendScene(table)
		} else {
			sceneTable = -1
			out = append(out, p)
		}
	}
	if titleCount > 0 && len(out) > titleCount && w.cfg.Bool("folio.title-page.enabled", true) {
		if out[titleCount].name == "w:tbl" {
			// Keep section pagination outside table properties when no heading precedes it.
			separator := dx("p", dx("pPr", dx("spacing").attr("w:line", "1").attr("w:lineRule", "exact")))
			out = append(out[:titleCount], append([]docxNode{separator}, out[titleCount:]...)...)
		}
		// A break on the first body paragraph avoids a trailing empty title page.
		props := &out[titleCount].children[0]
		var before, after []docxNode
		for _, property := range props.children {
			switch property.name {
			case "w:pStyle", "w:keepNext":
				before = append(before, property)
			case "w:pageBreakBefore": // Replace an existing act break, never duplicate it.
			default:
				after = append(after, property)
			}
		}
		props.children = append(append(before, dx("pageBreakBefore")), after...)
	}
	if len(out) == 0 {
		p, e := w.paragraph("Body", "")
		if e != nil {
			return nil, e
		}
		out = append(out, p)
	}
	return out, nil
}

func (w *docxWriter) speech(event play.Event, text string) ([]docxNode, error) {
	path := "folio.positioning.speech."
	name := w.cfg.String(path+"speaker.prefix", "") + event.Name + w.cfg.String(path+"speaker.suffix", ":")
	name = docxCase(name, w.cfg.String(path+"speaker.case-transform", "upper"))
	speaker, err := w.literalParagraph("Speaker", name)
	if err != nil {
		return nil, err
	}
	var content []docxNode
	if event.Direction != "" {
		instruction := w.cfg.String(path+"speech-instruction.prefix", "(") + event.Direction + w.cfg.String(path+"speech-instruction.suffix", ")")
		p, e := w.literalParagraph("Instruction", instruction)
		if e != nil {
			return nil, e
		}
		content = append(content, p)
	}
	dialogue, err := w.paragraph("Dialogue", text)
	if err != nil {
		return nil, err
	}
	content = append(content, dialogue)
	if !w.sameLine {
		return append([]docxNode{speaker}, content...), nil
	}
	table, cue, err := w.sceneTable()
	if err != nil {
		return nil, err
	}
	space, err := w.length(path+"space-before", "1.6em", w.baseSize, w.height)
	if err != nil {
		return nil, err
	}
	if space < 0 {
		return nil, fmt.Errorf("DOCX %sspace-before must not be negative", path)
	}
	// Space belongs inside the row so no outside paragraph splits an editable scene.
	for _, p := range []*docxNode{&speaker, &content[0]} {
		p.children[0].children = append(p.children[0].children, dx("spacing").attr("w:before", dxInt(space/2)))
	}
	table.children = append(table.children, w.tableRow(cue, []docxNode{speaker}, content))
	return []docxNode{table}, nil
}

func (w *docxWriter) sceneTable() (docxNode, int, error) {
	width := w.width - 2*w.margin
	cue, err := w.length("folio.positioning.speech.dialogue.wrap-indent", "7em", w.baseSize, width)
	if err != nil {
		return docxNode{}, 0, err
	}
	if cue < 1 || cue >= width {
		return docxNode{}, 0, fmt.Errorf("DOCX folio.positioning.speech.dialogue.wrap-indent leaves no usable speech column")
	}
	return w.table(width, cue), cue, nil
}

func (w *docxWriter) table(width, left int) docxNode {
	borders := dx("tblBorders")
	for _, edge := range []string{"top", "left", "bottom", "right", "insideH", "insideV"} {
		borders.children = append(borders.children, dxVal(edge, "nil"))
	}
	margins := dx("tblCellMar")
	for _, edge := range []string{"top", "left", "bottom", "right"} {
		margins.children = append(margins.children, dx(edge).attr("w:w", "0").attr("w:type", "dxa"))
	}
	return dx("tbl", dx("tblPr", dx("tblW").attr("w:w", dxInt(width)).attr("w:type", "dxa"), borders, dx("tblLayout").attr("w:type", "fixed"), margins), dx("tblGrid", dx("gridCol").attr("w:w", dxInt(left)), dx("gridCol").attr("w:w", dxInt(width-left))))
}
func (w *docxWriter) tableRow(left int, a, b []docxNode) docxNode {
	cells := []docxNode{}
	for i, paragraphs := range [][]docxNode{a, b} {
		width := left
		if i == 1 {
			width = w.width - 2*w.margin - left
		}
		cell := dx("tc", dx("tcPr", dx("tcW").attr("w:w", dxInt(width)).attr("w:type", "dxa"), dxVal("vAlign", "top")))
		cell.children = append(cell.children, paragraphs...)
		cells = append(cells, cell)
	}
	return dx("tr", cells...)
}
func docxCase(text, transform string) string {
	switch transform {
	case "upper":
		return strings.ToUpper(text)
	case "lower":
		return strings.ToLower(text)
	default:
		return text
	}
}
