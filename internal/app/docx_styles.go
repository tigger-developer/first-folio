// ABOUTME: Maps resolved script font roles and geometry to editable DOCX styles.
// ABOUTME: Validates native limits and confines approximations to documented spike policy.
package app

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/tigger-developer/first-folio/internal/config"
	"github.com/tigger-developer/first-folio/internal/play"
)

type docxWriter struct {
	cfg                                                                config.Config
	doc                                                                play.Document
	width, height, margin, top, bottom, headerDistance, footerDistance int
	baseSize                                                           float64
	sameLine                                                           bool
	styles                                                             []docxNode
	styleIDs                                                           map[string]bool
	notes                                                              map[string]string
	footnotes                                                          []docxNode
	warnings                                                           []string
	textBytes, inlineCount                                             int
}

func newDOCXWriter(doc play.Document, cfg config.Config) (*docxWriter, error) {
	w := &docxWriter{cfg: cfg, doc: doc, styleIDs: map[string]bool{}, notes: map[string]string{}}
	if len(doc.Events) > 50000 {
		return nil, fmt.Errorf("DOCX exceeds 50000 script events")
	}
	paper := strings.ToLower(cfg.String("folio.page", "a4"))
	papers := map[string][2]float64{"a3": {297, 420}, "a4": {210, 297}, "a5": {148, 210}, "a6": {105, 148}, "us-letter": {215.9, 279.4}, "letter": {215.9, 279.4}, "us-legal": {215.9, 355.6}, "legal": {215.9, 355.6}}
	size, ok := papers[paper]
	if !ok {
		return nil, fmt.Errorf("DOCX folio.page %q is unsupported; use a3, a4, a5, a6, letter or legal", paper)
	}
	w.width = int(math.Round(size[0] * 1440 / 25.4))
	w.height = int(math.Round(size[1] * 1440 / 25.4))
	font, err := cfg.Font("folio.font")
	if err != nil {
		return nil, err
	}
	w.baseSize, err = docxPoints(font.Size, 12, 0)
	if err != nil {
		return nil, fmt.Errorf("DOCX folio.font.size: %w", err)
	}
	w.margin, err = w.length("folio.margin", "25mm", w.baseSize, w.width)
	if err != nil {
		return nil, err
	}
	if w.margin < 0 {
		return nil, fmt.Errorf("DOCX folio.margin must not be negative")
	}
	w.top, w.bottom = w.margin, w.margin
	for _, role := range []string{"header", "footer"} {
		path := "folio.page-" + role
		if !cfg.Bool(path+".enabled", role == "footer") {
			continue
		}
		distance, e := w.length(path+".distance-from-edge", "20mm", w.baseSize, w.height)
		if e != nil {
			return nil, e
		}
		padding, e := w.length(path+".content-padding-after", "10mm", w.baseSize, w.height)
		if e != nil {
			return nil, e
		}
		if distance < 0 || padding < 0 {
			return nil, fmt.Errorf("DOCX %s distance and clearance must not be negative", path)
		}
		if role == "header" {
			w.top = distance + padding
			w.headerDistance = distance
		} else {
			w.bottom = distance + padding
			w.footerDistance = distance
		}
	}
	if w.width-2*w.margin < 720 || w.height-w.top-w.bottom < 720 {
		return nil, fmt.Errorf("DOCX folio.margin and running matter leave less than 36pt of usable page")
	}
	w.sameLine = cfg.String("folio.positioning.speech.dialogue.placement", "same-line") == "same-line"
	for _, event := range doc.Events {
		if event.Kind == play.EventFootnote {
			w.notes[event.Name] = event.Text
		}
	}
	return w, nil
}

func docxPoints(raw string, em, extent float64) (float64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "0" {
		return 0, nil
	}
	for _, unit := range []struct {
		suffix string
		scale  float64
	}{{"pt", 1}, {"mm", 72 / 25.4}, {"cm", 720 / 25.4}, {"in", 72}, {"em", em}, {"%", extent / 100}} {
		if !strings.HasSuffix(raw, unit.suffix) {
			continue
		}
		n, err := strconv.ParseFloat(strings.TrimSuffix(raw, unit.suffix), 64)
		if err != nil || math.IsNaN(n) || math.IsInf(n, 0) || unit.suffix == "%" && extent == 0 {
			break
		}
		value := n * unit.scale
		if math.IsInf(value, 0) || math.Abs(value) > 1584 {
			break
		}
		return value, nil
	}
	return 0, fmt.Errorf("invalid or out-of-range length %q", raw)
}
func (w *docxWriter) length(path, fallback string, em float64, extent int) (int, error) {
	n, err := docxPoints(w.cfg.String(path, fallback), em, float64(extent)/20)
	if err != nil {
		return 0, fmt.Errorf("DOCX %s: %w", path, err)
	}
	return int(math.Round(n * 20)), nil
}
func dxInt(n int) string { return strconv.Itoa(n) }

func (w *docxWriter) font(path string) (docxNode, float64, error) {
	f, err := w.cfg.Font(path)
	if err != nil {
		return docxNode{}, 0, err
	}
	size, err := docxPoints(f.Size, w.baseSize, 0)
	if path == "folio.font" {
		size = w.baseSize
	}
	if err != nil || size < 0.5 {
		return docxNode{}, 0, fmt.Errorf("DOCX %s.size %q is outside supported range", path, f.Size)
	}
	bold := false
	switch f.Weight {
	case "regular", "400":
	case "bold", "700":
		bold = true
	default:
		return docxNode{}, 0, fmt.Errorf("DOCX %s.weight %q is unsupported; use regular or bold", path, f.Weight)
	}
	scale, err := strconv.ParseFloat(strings.TrimSuffix(f.Stretch, "%"), 64)
	if err != nil || scale < 1 || scale > 600 || scale != math.Trunc(scale) {
		return docxNode{}, 0, fmt.Errorf("DOCX %s.stretch %q requires an integer from 1%% to 600%%", path, f.Stretch)
	}
	tracking, err := docxPoints(f.LetterSpacing, size, 0)
	if err != nil {
		return docxNode{}, 0, fmt.Errorf("DOCX %s.letter-spacing: %w", path, err)
	}
	props := dx("rPr", dx("rFonts").attr("w:ascii", f.Family).attr("w:hAnsi", f.Family).attr("w:eastAsia", f.Family).attr("w:cs", f.Family), dxVal("b", strconv.FormatBool(bold)), dxVal("i", strconv.FormatBool(f.Style != "regular")), dxVal("spacing", dxInt(int(math.Round(tracking*20)))), dxVal("w", dxInt(int(scale))), dxVal("sz", dxInt(int(math.Round(size*2)))), dxVal("szCs", dxInt(int(math.Round(size*2)))))
	if f.Style == "oblique" {
		w.warnings = append(w.warnings, path+".style oblique uses native italic")
	}
	return props, size, nil
}

func docxRolePath(role string) string {
	switch role {
	case "Speaker":
		return "folio.positioning.speech.speaker"
	case "Instruction":
		return "folio.positioning.speech.speech-instruction"
	case "Dialogue":
		return "folio.positioning.speech.dialogue"
	case "Direction":
		return "folio.positioning.stage-direction"
	case "Transition":
		return "folio.positioning.transition"
	case "Act":
		return "folio.positioning.act-header"
	case "Scene":
		return "folio.positioning.scene-header"
	case "Intro":
		return "folio.positioning.frontmatter.header"
	case "Header":
		return "folio.page-header"
	case "Footer":
		return "folio.page-footer"
	case "Title", "Subtitle", "Author", "Date", "Version":
		return "folio.title-page." + strings.ToLower(role)
	default:
		return "folio"
	}
}

func (w *docxWriter) style(role string) (string, error) {
	id := "Folio" + role
	if w.styleIDs[id] {
		return id, nil
	}
	path := docxRolePath(role)
	font, size, err := w.font(path + ".font")
	if err != nil {
		return "", err
	}
	props, err := w.paragraphProperties(role, path, size)
	if err != nil {
		return "", err
	}
	if role == "Prop" {
		for i, n := range font.children {
			if n.name == "w:b" || n.name == "w:i" {
				font.children[i] = dxVal(strings.TrimPrefix(n.name, "w:"), "true")
			}
		}
	}
	if w.cfg.String(path+".case-transform", "") == "small-caps" {
		// rFonts, b and i precede smallCaps in the WordprocessingML sequence.
		font.children = append(append(font.children[:3:3], dx("smallCaps")), font.children[3:]...)
	}
	style := dx("style", dxVal("name", "Folio "+role), dx("qFormat"), props, font).attr("w:type", "paragraph").attr("w:styleId", id)
	w.styles = append(w.styles, style)
	w.styleIDs[id] = true
	return id, nil
}

func (w *docxWriter) paragraphProperties(role, path string, size float64) (docxNode, error) {
	before, after, left, right := 0, 0, 0, 0
	align := "left"
	var err error
	extent := w.width - 2*w.margin
	before, err = w.length(path+".space-before", "0", size, w.height-w.top-w.bottom)
	if err != nil {
		return docxNode{}, err
	}
	after, err = w.length(path+".space-after", "0", size, w.height-w.top-w.bottom)
	if err != nil {
		return docxNode{}, err
	}
	align = w.cfg.String(path+".align", "left")
	switch role {
	case "Body", "Footnote":
		before = 0
		after = int(math.Round(size * 12))
		align = "left"
	case "Prop":
		before = int(math.Round(size * 32))
		after = int(math.Round(size * 12))
		align = "center"
	case "Speaker", "Instruction", "Dialogue":
		before, after = 0, 0
		if !w.sameLine {
			left, err = w.length("folio.positioning.speech.dialogue.indent", "0", size, extent)
			if err != nil {
				return docxNode{}, err
			}
			right, err = w.length("folio.positioning.speech.dialogue.wrap-indent", "0", size, extent)
			if err != nil {
				return docxNode{}, err
			}
			if role == "Speaker" {
				extra, e := w.length(path+".indent", "0", size, extent)
				if e != nil {
					return docxNode{}, e
				}
				left += extra
				before, err = w.length("folio.positioning.speech.space-before", "1.6em", size, w.height)
				if err != nil {
					return docxNode{}, err
				}
			}
		}
	case "Direction", "Transition":
		left, err = w.length(path+".indent", "0", size, extent)
		if err != nil {
			return docxNode{}, err
		}
		if role == "Direction" {
			after = int(math.Round(size * 12))
		}
	case "Act":
		after = int(math.Round(size * 6))
	case "Title", "Subtitle", "Author":
		align = w.cfg.String("folio.title-page.title.align", "center")
	case "Date", "Version":
		align = "left"
		if strings.HasSuffix(w.cfg.String(path+".position", "bottom-right"), "right") {
			align = "right"
		}
	case "Header", "Footer":
		align = strings.Split(align, "-")[0]
	}
	if before < 0 {
		w.warnings = append(w.warnings, path+".space-before is negative; using zero")
		before = 0
	}
	if after < 0 {
		w.warnings = append(w.warnings, path+".space-after is negative; using zero")
		after = 0
	}
	if left < 0 || right < 0 || left+right >= extent {
		return docxNode{}, fmt.Errorf("DOCX %s indentation leaves no usable text width", path)
	}
	props := dx("pPr")
	if role == "Speaker" || role == "Instruction" || role == "Act" || role == "Scene" || role == "Intro" {
		props.children = append(props.children, dx("keepNext"))
	}
	props.children = append(props.children, dx("spacing").attr("w:before", dxInt(before)).attr("w:after", dxInt(after)).attr("w:line", "288").attr("w:lineRule", "auto"), dx("ind").attr("w:left", dxInt(left)).attr("w:right", dxInt(right)), dxVal("jc", align))
	return props, nil
}
