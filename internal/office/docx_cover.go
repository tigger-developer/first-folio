// ABOUTME: Projects prepared script-cover assets into native page-anchored DOCX drawings.
// ABOUTME: Keeps title metadata editable while preserving shared cover geometry and treatments.
package office

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

const (
	docxEMUPerMM = 36000
	docxEMUPerPT = 12700
)

func coverXML(name string, children ...docxNode) docxNode {
	return docxNode{name: name, children: children}
}

func (w *docxWriter) coverMediaParts() []docxPart {
	if w.cover == nil {
		return nil
	}
	parts := []docxPart{coverMediaPart("folio-cover-sharp", "coverSharp", w.cover.Sharp)}
	if len(w.cover.Blurred.Data) > 0 && !sameCoverAsset(w.cover.Sharp, w.cover.Blurred) {
		parts = append(parts, coverMediaPart("folio-cover-blurred", "coverBlurred", w.cover.Blurred))
	}
	if len(w.cover.Backdrop.Data) > 0 {
		parts = append(parts, coverMediaPart("folio-cover-backdrop", "coverBackdrop", w.cover.Backdrop))
	}
	return parts
}

func coverMediaPart(base, relationshipID string, asset ScriptCoverAsset) docxPart {
	ext := strings.ToLower(asset.Format)
	if ext == "jpeg" {
		ext = "jpg"
	}
	return docxPart{name: "word/media/" + base + "." + ext, relationship: "image", relationshipID: relationshipID, raw: append([]byte(nil), asset.Data...)}
}

func sameCoverAsset(left, right ScriptCoverAsset) bool {
	if len(left.Data) != len(right.Data) || left.Format != right.Format || left.PixelWidth != right.PixelWidth || left.PixelHeight != right.PixelHeight {
		return false
	}
	for i := range left.Data {
		if left.Data[i] != right.Data[i] {
			return false
		}
	}
	return true
}

func (w *docxWriter) coverTitle() ([]docxNode, error) {
	pageW := math.Round(float64(w.width)*25.4/1440*10) / 10
	pageH := math.Round(float64(w.height)*25.4/1440*10) / 10
	cover := w.cover
	var drawings []docxNode
	if cover.Config.BackdropMode == "image" {
		drawings = append(drawings, w.coverPicture("Folio backdrop image", "coverBackdrop", 0, 0, pageW, pageH, cover.Backdrop, true, false))
	} else if cover.Config.BackdropMode == "solid" {
		drawings = append(drawings, w.coverShape("Folio backdrop solid", 0, 0, pageW, pageH, coverHex(cover.Config.BackdropSolidColor, "FFFFFF"), 1, true))
	}
	if cover.Config.BackdropMode != "" && cover.Config.BackdropWashOpacity > 0 {
		drawings = append(drawings, w.coverShape("Folio backdrop wash", 0, 0, pageW, pageH, treatmentHex(cover.Config.BackdropWashTreatment), cover.Config.BackdropWashOpacity, true))
	}

	imageW, imageH := pageW, pageH
	imageX, imageY := 0.0, 0.0
	name := "Folio cover background"
	if cover.Config.Mode == "inset" {
		imageW, imageH = cover.RenderedWidthMM, cover.RenderedHeightMM
		imageX, imageY = (pageW-imageW)/2, (pageH-imageH)/2
		name = "Folio cover inset"
	}
	drawings = append(drawings, w.coverPicture(name, "coverSharp", imageX, imageY, imageW, imageH, cover.Sharp, true, cover.Config.Mode == "background"))

	textColour := "000000"
	if cover.Config.OverlayWashTreatment == "dark" || cover.Config.BackdropMode != "" && cover.Config.BackdropWashTreatment == "dark" {
		textColour = "FFFFFF"
	}
	if cover.Config.Placement == "dodge" && cover.Config.Mode == "inset" {
		upper, bottom, err := w.coverDodgeText(imageY, imageH, pageW, pageH, textColour)
		if err != nil {
			return nil, err
		}
		drawings = append(drawings, upper, bottom)
	} else {
		overlay, err := w.coverOverlayText(pageW, pageH, textColour)
		if err != nil {
			return nil, err
		}
		drawings = append(drawings, overlay...)
	}
	paragraph := dx("p", dx("pPr", dx("keepNext")))
	for _, drawing := range drawings {
		paragraph.children = append(paragraph.children, dx("r", dx("drawing", drawing)))
	}
	return []docxNode{paragraph}, nil
}

func (w *docxWriter) coverOverlayText(pageW, pageH float64, colour string) ([]docxNode, error) {
	contentX := float64(w.margin) * 25.4 / 1440
	contentW := pageW - 2*contentX
	top := float64(w.top) * 25.4 / 1440
	bottom := float64(w.bottom) * 25.4 / 1440
	available := pageH - top - bottom
	ratio, err := coverPositionRatio(w.cfg.String("folio.title-page.title.position", "third"), false)
	if err != nil {
		return nil, err
	}
	titleSize, err := w.coverFontSize("Title")
	if err != nil {
		return nil, err
	}
	subtitleSize, err := w.coverFontSize("Subtitle")
	if err != nil {
		return nil, err
	}
	gap, err := docxPoints(w.cfg.String("folio.title-page.subtitle.space-before", "1em"), subtitleSize, available*72/25.4)
	if err != nil {
		return nil, fmt.Errorf("DOCX folio.title-page.subtitle.space-before: %w", err)
	}
	titleY := top + available*ratio
	// LibreOffice and Word position glyph ink approximately 2mm above a zero-inset
	// DrawingML text-box origin. Compensate so the configured position describes
	// the visible title top, matching the PDF renderer's contract.
	titleY += 2
	boxH := math.Max(55, titleSize*25.4/72+gap*25.4/72+subtitleSize*25.4/72+25)
	padding, err := docxPoints(w.cover.Config.OverlayPadding, w.baseSize, 0)
	if err != nil {
		return nil, fmt.Errorf("DOCX folio.cover-image.overlay.padding: %w", err)
	}
	paddingMM := padding * 25.4 / 72
	boxX, boxY := contentX-paddingMM, titleY-paddingMM
	boxW, boxHWithPadding := contentW+2*paddingMM, boxH+2*paddingMM
	var result []docxNode
	if len(w.cover.Blurred.Data) > 0 && !sameCoverAsset(w.cover.Sharp, w.cover.Blurred) {
		result = append(result, w.coverPicture("Folio title treatment", "coverBlurred", boxX, boxY, boxW, boxHWithPadding, w.cover.Blurred, true, true))
	} else {
		result = append(result, w.coverShape("Folio title treatment", boxX, boxY, boxW, boxHWithPadding, treatmentHex(w.cover.Config.OverlayWashTreatment), w.cover.Config.OverlayWashOpacity, true))
	}
	if w.cover.Config.OverlayWashOpacity > 0 && len(w.cover.Blurred.Data) > 0 && !sameCoverAsset(w.cover.Sharp, w.cover.Blurred) {
		result = append(result, w.coverShape("Folio title wash", boxX, boxY, boxW, boxHWithPadding, treatmentHex(w.cover.Config.OverlayWashTreatment), w.cover.Config.OverlayWashOpacity, true))
	}
	roles := []string{"Title", "Subtitle", "Author", "Date", "Version"}
	result = append(result, w.coverTextBox("Folio title text", contentX, titleY, contentW, boxH, roles, colour, gap))
	return result, nil
}

func (w *docxWriter) coverDodgeText(imageY, imageH, pageW, pageH float64, colour string) (docxNode, docxNode, error) {
	contentX := float64(w.margin) * 25.4 / 1440
	contentW := pageW - 2*contentX
	padding, err := docxPoints(w.cover.Config.DodgePadding, w.baseSize, 0)
	if err != nil {
		return docxNode{}, docxNode{}, fmt.Errorf("DOCX folio.cover-image.dodge.padding: %w", err)
	}
	paddingMM := padding * 25.4 / 72
	upperTop := float64(w.top) * 25.4 / 1440
	upperBottom := math.Max(upperTop, imageY-paddingMM)
	ratio, err := coverPositionRatio(w.cfg.String("folio.title-page.title.position", "third"), true)
	if err != nil {
		return docxNode{}, docxNode{}, err
	}
	titleSize, err := w.coverFontSize("Title")
	if err != nil {
		return docxNode{}, docxNode{}, err
	}
	subtitleSize, err := w.coverFontSize("Subtitle")
	if err != nil {
		return docxNode{}, docxNode{}, err
	}
	gap, err := docxPoints(w.cfg.String("folio.title-page.subtitle.space-before", "1em"), subtitleSize, (upperBottom-upperTop)*72/25.4)
	if err != nil {
		return docxNode{}, docxNode{}, fmt.Errorf("DOCX folio.title-page.subtitle.space-before: %w", err)
	}
	upperH := titleSize*25.4/72 + gap*25.4/72 + subtitleSize*25.4/72 + 8
	upperY := upperTop + (upperBottom-upperTop)*ratio - upperH/2
	upperY = math.Max(0, math.Min(upperY, math.Max(0, upperBottom-upperH)))
	upper := w.coverTextBox("Folio upper title text", contentX, upperY, contentW, upperH, []string{"Title", "Subtitle"}, colour, gap)

	lowerTop := imageY + imageH + paddingMM
	lowerBottom := pageH - float64(w.bottom)*25.4/1440
	lowerY := math.Min(lowerBottom-30, lowerTop+math.Max(0, lowerBottom-lowerTop)*0.5-15)
	lowerY = math.Max(lowerTop, lowerY)
	bottom := w.coverTextBox("Folio bottom metadata text", contentX, lowerY, contentW, math.Max(20, lowerBottom-lowerY), []string{"Author", "Date", "Version"}, colour, 0)
	return upper, bottom, nil
}

func coverPositionRatio(raw string, dodge bool) (float64, error) {
	raw = strings.TrimSpace(strings.ToLower(raw))
	switch raw {
	case "third", "one-third":
		if dodge {
			return 1.0 / 3, nil
		}
		return 0.3, nil
	case "center", "middle":
		if dodge {
			return 0.5, nil
		}
		return 0.4, nil
	case "top":
		return 0, nil
	}
	if strings.HasSuffix(raw, "%") {
		value, err := strconv.ParseFloat(strings.TrimSpace(strings.TrimSuffix(raw, "%")), 64)
		if err == nil && value >= 0 && value <= 100 {
			return value / 100, nil
		}
	}
	return 0, fmt.Errorf("DOCX folio.title-page.title.position %q is unsupported for a cover image", raw)
}

func (w *docxWriter) coverFontSize(role string) (float64, error) {
	_, size, err := w.font(docxRolePath(role) + ".font")
	return size, err
}

func (w *docxWriter) coverTextBox(name string, x, y, width, height float64, roles []string, colour string, subtitleGapPT float64) docxNode {
	var paragraphs []docxNode
	for _, role := range roles {
		text := w.doc.Metadata[strings.ToLower(role)]
		if text == "" {
			continue
		}
		if role == "Author" {
			text = w.cfg.String("folio.title-page.author.prefix", "") + text
		}
		style, _ := w.style(role)
		before := "0"
		if role == "Subtitle" && subtitleGapPT > 0 {
			before = dxInt(int(math.Round(subtitleGapPT * 20)))
		}
		pPr := dx("pPr", dxVal("pStyle", style), dx("spacing").attr("w:before", before).attr("w:after", "0"))
		paragraphs = append(paragraphs, dx("p", pPr, docxRun(text, dx("rPr", dxVal("color", colour)))))
	}
	if len(paragraphs) == 0 {
		paragraphs = append(paragraphs, dx("p"))
	}
	body := coverXML("wps:txbx", dx("txbxContent", paragraphs...))
	shape := coverXML("wps:wsp", coverXML("wps:cNvSpPr").attr("txBox", "1"), coverXML("wps:spPr", coverXML("a:xfrm", coverXML("a:off").attr("x", "0").attr("y", "0"), coverXML("a:ext").attr("cx", emu(width)).attr("cy", emu(height))), coverXML("a:prstGeom", coverXML("a:avLst")).attr("prst", "rect"), coverXML("a:noFill"), coverXML("a:ln", coverXML("a:noFill"))), body, coverXML("wps:bodyPr").attr("lIns", "0").attr("tIns", "0").attr("rIns", "0").attr("bIns", "0").attr("anchor", "t"))
	return w.coverAnchor(name, x, y, width, height, false, coverXML("a:graphic", coverXML("a:graphicData", shape).attr("uri", "http://schemas.microsoft.com/office/word/2010/wordprocessingShape")))
}

func (w *docxWriter) coverPicture(name, relationship string, x, y, width, height float64, asset ScriptCoverAsset, behind, crop bool) docxNode {
	fill := []docxNode{coverXML("a:blip").attr("r:embed", relationship)}
	if crop {
		left, top, right, bottom := coverCrop(asset, width, height)
		fill = append(fill, coverXML("a:srcRect").attr("l", left).attr("t", top).attr("r", right).attr("b", bottom))
	}
	fill = append(fill, coverXML("a:stretch", coverXML("a:fillRect")))
	picture := coverXML("pic:pic", coverXML("pic:nvPicPr", coverXML("pic:cNvPr").attr("id", "0").attr("name", name), coverXML("pic:cNvPicPr")), coverXML("pic:blipFill", fill...), coverXML("pic:spPr", coverXML("a:xfrm", coverXML("a:off").attr("x", "0").attr("y", "0"), coverXML("a:ext").attr("cx", emu(width)).attr("cy", emu(height))), coverXML("a:prstGeom", coverXML("a:avLst")).attr("prst", "rect")))
	graphic := coverXML("a:graphic", coverXML("a:graphicData", picture).attr("uri", "http://schemas.openxmlformats.org/drawingml/2006/picture"))
	return w.coverAnchor(name, x, y, width, height, behind, graphic)
}

func (w *docxWriter) coverShape(name string, x, y, width, height float64, colour string, opacity float64, behind bool) docxNode {
	alpha := int(math.Round(math.Max(0, math.Min(1, opacity)) * 100000))
	fill := coverXML("a:solidFill", coverXML("a:srgbClr", coverXML("a:alpha").attr("val", strconv.Itoa(alpha))).attr("val", colour))
	shape := coverXML("wps:wsp", coverXML("wps:cNvSpPr"), coverXML("wps:spPr", coverXML("a:xfrm", coverXML("a:off").attr("x", "0").attr("y", "0"), coverXML("a:ext").attr("cx", emu(width)).attr("cy", emu(height))), coverXML("a:prstGeom", coverXML("a:avLst")).attr("prst", "rect"), fill, coverXML("a:ln", coverXML("a:noFill"))), coverXML("wps:bodyPr"))
	return w.coverAnchor(name, x, y, width, height, behind, coverXML("a:graphic", coverXML("a:graphicData", shape).attr("uri", "http://schemas.microsoft.com/office/word/2010/wordprocessingShape")))
}

func (w *docxWriter) coverAnchor(name string, x, y, width, height float64, behind bool, graphic docxNode) docxNode {
	w.drawingID++
	return coverXML("wp:anchor", coverXML("wp:simplePos").attr("x", "0").attr("y", "0"), coverXML("wp:positionH", coverXML("wp:posOffset", docxNode{text: emu(x)})).attr("relativeFrom", "page"), coverXML("wp:positionV", coverXML("wp:posOffset", docxNode{text: emu(y)})).attr("relativeFrom", "page"), coverXML("wp:extent").attr("cx", emu(width)).attr("cy", emu(height)), coverXML("wp:effectExtent").attr("l", "0").attr("t", "0").attr("r", "0").attr("b", "0"), coverXML("wp:wrapNone"), coverXML("wp:docPr").attr("id", dxInt(w.drawingID)).attr("name", name), coverXML("wp:cNvGraphicFramePr"), graphic).attr("distT", "0").attr("distB", "0").attr("distL", "0").attr("distR", "0").attr("simplePos", "0").attr("relativeHeight", dxInt(1000+w.drawingID)).attr("behindDoc", strconv.FormatBool(behind)).attr("locked", "0").attr("layoutInCell", "1").attr("allowOverlap", "1")
}

func coverCrop(asset ScriptCoverAsset, targetW, targetH float64) (string, string, string, string) {
	sourceRatio := float64(asset.PixelWidth) / float64(asset.PixelHeight)
	targetRatio := targetW / targetH
	left, top, right, bottom := 0.0, 0.0, 0.0, 0.0
	if sourceRatio > targetRatio {
		crop := (1 - targetRatio/sourceRatio) * 50000
		left, right = crop, crop
	} else if sourceRatio < targetRatio {
		crop := (1 - sourceRatio/targetRatio) * 50000
		top, bottom = crop, crop
	}
	return strconv.Itoa(int(math.Round(left))), strconv.Itoa(int(math.Round(top))), strconv.Itoa(int(math.Round(right))), strconv.Itoa(int(math.Round(bottom)))
}

func emu(mm float64) string { return strconv.FormatInt(int64(math.Round(mm*docxEMUPerMM)), 10) }

func treatmentHex(treatment string) string {
	if treatment == "dark" {
		return "000000"
	}
	return "FFFFFF"
}

func coverHex(value, fallback string) string {
	value = strings.TrimPrefix(strings.TrimSpace(value), "#")
	if len(value) != 6 {
		return fallback
	}
	return strings.ToUpper(value)
}
