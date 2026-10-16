// ABOUTME: Covers script office composition, empty running variants and richer native content.
// ABOUTME: Resolves applied native properties and checks lengths independently of serialization units.
package app

import (
	"math"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func officeDocxPart(t *testing.T, pkg map[string]*officeXML, reference *officeXML) *officeXML {
	t.Helper()
	relations := pkg["word/_rels/document.xml.rels"]
	if relations == nil {
		t.Fatal("document relationship part missing")
	}
	for _, relation := range relations.all("Relationship") {
		if relation.attr("Id") == reference.attr("id") {
			name := path.Clean(path.Join("word", relation.attr("Target")))
			if pkg[name] == nil {
				t.Fatalf("running reference points to absent part %s", name)
			}
			return pkg[name]
		}
	}
	t.Fatalf("unresolved running relationship %s", reference.attr("id"))
	return nil
}

func officeBreakBefore(t *testing.T, pkg map[string]*officeXML, format, text string) bool {
	t.Helper()
	body := officeBody(t, pkg, format)
	broken, first := false, true
	for _, p := range body.all("p") {
		layers := officeEffective(t, pkg, format, p)
		if format == "docx" {
			for _, layer := range layers {
				for _, b := range layer.all("pageBreakBefore") {
					if b.attr("val") != "0" && b.attr("val") != "false" {
						broken = true
					}
				}
			}
			for _, b := range p.all("br") {
				if b.attr("type") == "page" {
					broken = true
				}
			}
		} else {
			if v, _ := officeEffectiveAttr(layers, "paragraph-properties", "break-before"); v == "page" {
				broken = true
			}
			if !first {
				if v, _ := officeEffectiveAttr(layers, "style", "master-page-name"); v != "" {
					broken = true
				}
			}
		}
		if strings.Contains(p.text(), text) {
			return broken
		}
		if format == "docx" {
			for _, section := range p.all("sectPr") {
				kind := "nextPage"
				for _, value := range section.all("type") {
					kind = value.attr("val")
				}
				if kind != "continuous" {
					broken = true
				}
			}
		} else {
			if v, _ := officeEffectiveAttr(layers, "paragraph-properties", "break-after"); v == "page" {
				broken = true
			}
		}
		first = false
	}
	t.Fatalf("no paragraph %q", text)
	return false
}

func officeAssertUntitledFirstPage(t *testing.T, pkg map[string]*officeXML, format string) {
	t.Helper()
	if officeBreakBefore(t, pkg, format, "ACT ONE") {
		t.Error("untitled script gained an extra first-act/title page break")
	}
	if format == "docx" {
		sections := officeBody(t, pkg, format).all("sectPr")
		if len(sections) != 1 {
			t.Errorf("untitled script has %d sections; expected only the body", len(sections))
		}
		section := sections[0]
		for _, pair := range []struct{ kind, text string }{{"headerReference", "UNTITLED HEADER"}, {"footerReference", "UNTITLED FOOTER"}} {
			wantedType := "default"
			if len(section.all("titlePg")) != 0 {
				wantedType = "first"
			}
			found := false
			for _, ref := range section.all(pair.kind) {
				if ref.attr("type") == wantedType && strings.Contains(officeDocxPart(t, pkg, ref).text(), pair.text) {
					found = true
				}
			}
			if !found {
				t.Errorf("untitled physical first-page %s suppressed", pair.kind)
			}
		}
		return
	}
	masters := pkg["styles.xml"].all("master-page")
	if len(masters) == 0 {
		t.Fatal("no native master page")
	}
	firstPara := officeParagraph(t, officeBody(t, pkg, format), "ACT ONE")
	masterName, hasMaster := officeEffectiveAttr(officeEffective(t, pkg, format, firstPara), "style", "master-page-name")
	master := masters[0]
	if hasMaster {
		for _, candidate := range masters {
			if candidate.attr("name") == masterName {
				master = candidate
			}
		}
	}
	for _, pair := range []struct{ kind, text string }{{"header", "UNTITLED HEADER"}, {"footer", "UNTITLED FOOTER"}} {
		definitions := master.all(pair.kind + "-first")
		if len(definitions) == 0 {
			definitions = master.all(pair.kind)
		}
		if len(definitions) != 1 || !strings.Contains(definitions[0].text(), pair.text) {
			t.Errorf("untitled physical first-page %s suppressed", pair.kind)
		}
	}
}

func officeAssertTitleSlot(t *testing.T, pkg map[string]*officeXML, format, text, align string) {
	t.Helper()
	// Accepted office approximation: date/version remain editable body text.
	roots := []*officeXML{officeBody(t, pkg, format)}
	for _, root := range roots {
		for _, p := range root.all("p") {
			if !strings.Contains(p.text(), text) {
				continue
			}
			layers := officeEffective(t, pkg, format, p)
			if format == "docx" {
				officeAssertEffective(t, layers, "jc", "val", align)
			} else {
				officeAssertEffective(t, layers, "paragraph-properties", "text-align", align)
			}
			return
		}
	}
	t.Fatalf("title metadata %q is not in the document body", text)
}

func officeAssertInlineNoteFormatting(t *testing.T, pkg map[string]*officeXML, format string, notes *officeXML) {
	t.Helper()
	for _, pair := range []struct{ text, property, attribute, value string }{
		{"Bold note", "b", "font-weight", "bold"}, {"italic note", "i", "font-style", "italic"},
	} {
		found := false
		kind := "r"
		if format == "odt" {
			kind = "span"
		}
		for _, n := range notes.all(kind) {
			if !strings.Contains(n.text(), pair.text) {
				continue
			}
			if format == "docx" {
				for _, prop := range n.all(pair.property) {
					if prop.attr("val") != "0" && prop.attr("val") != "false" {
						found = true
					}
				}
			} else {
				got, exists := officeEffectiveAttr(officeEffective(t, pkg, format, n), "text-properties", pair.attribute)
				found = exists && got == pair.value
			}
		}
		if !found {
			t.Errorf("native note text %q has lost %s formatting", pair.text, pair.value)
		}
	}
}

func officeAssertProp(t *testing.T, pkg map[string]*officeXML, format string, body *officeXML) {
	t.Helper()
	p := officeParagraph(t, body, "Prop card")
	layers := officeEffective(t, pkg, format, p)
	if format == "docx" {
		officeAssertEffective(t, layers, "jc", "val", "center")
		for _, property := range []string{"b", "i"} {
			found := false
			for _, layer := range layers {
				for _, n := range layer.all(property) {
					if n.attr("val") != "0" && n.attr("val") != "false" {
						found = true
					}
				}
			}
			if !found {
				t.Errorf("prop missing effective %s", property)
			}
		}
	} else {
		officeAssertEffective(t, layers, "paragraph-properties", "text-align", "center")
		officeAssertEffective(t, layers, "text-properties", "font-weight", "bold")
		officeAssertEffective(t, layers, "text-properties", "font-style", "italic")
	}
	// British scene rows use the accepted half-spacing: 0.8em/0.3em at 12pt.
	officeAssertLength(t, layers, format, "spacing", "before", "paragraph-properties", "margin-top", 9.6)
	officeAssertLength(t, layers, format, "spacing", "after", "paragraph-properties", "margin-bottom", 3.6)
}

func officeAssertTableRatio(t *testing.T, pkg map[string]*officeXML, format string, body *officeXML, marker string, want float64) {
	t.Helper()
	kind := "tbl"
	if format == "odt" {
		kind = "table"
	}
	for _, table := range body.all(kind) {
		if !strings.Contains(table.text(), marker) {
			continue
		}
		var widths []float64
		if format == "docx" {
			for _, col := range table.all("gridCol") {
				n, err := strconv.ParseFloat(col.attr("w"), 64)
				if err != nil {
					t.Fatal(err)
				}
				widths = append(widths, n)
			}
			bordersList := table.all("tblBorders")
			if len(bordersList) != 1 {
				t.Error("borderless table has no explicit native border definition")
			}
			for _, borders := range bordersList {
				for _, border := range borders.Children {
					if border.Name.Local != "" && border.attr("val") != "nil" && border.attr("val") != "none" {
						t.Errorf("visible table border %s", border.Name.Local)
					}
				}
			}
		} else {
			for _, col := range table.all("table-column") {
				layers := officeEffective(t, pkg, format, col)
				width, exists := officeEffectiveAttr(layers, "table-column-properties", "column-width")
				if exists {
					widths = append(widths, officeLengthPoints(t, width))
				} else {
					relative, present := officeEffectiveAttr(layers, "table-column-properties", "rel-column-width")
					n, err := strconv.ParseFloat(strings.TrimSuffix(relative, "*"), 64)
					if !present || err != nil || n <= 0 {
						t.Fatalf("invalid relative column width %q", relative)
					}
					widths = append(widths, n)
				}
			}
			for _, cell := range table.all("table-cell") {
				officeAssertEffective(t, officeEffective(t, pkg, format, cell), "table-cell-properties", "border", "none")
			}
		}
		if len(widths) != 2 || math.Abs(widths[0]/(widths[0]+widths[1])-want) > 0.0002 {
			t.Errorf("native column widths %v do not preserve ratio %g", widths, want)
		}
		return
	}
	t.Fatalf("no native table containing %q", marker)
}

func TestRT054_2OfficeBodyLayoutLengths(t *testing.T) {
	binary := officeBuild(t)
	const source = "#+TITLE: Geometry\n* Introduction\nIntro paragraph.\n* ACT ONE\n** SCENE ONE\n*** Direction text.\n**** CÁIT\nDialogue text.\n*** Prop card :prop:\n***** Transition text\n"
	const cfg = `folio:
  font: {size: 12pt}
  title-page: {enabled: false}
  page-header: {enabled: true, distance-from-edge: 12pt, content-padding-after: 15pt}
  page-footer: {enabled: true, distance-from-edge: 14pt, content-padding-after: 16pt}
  positioning:
    speech:
      dialogue: {wrap-indent: 5em}
    stage-direction:
      space-before: 2em
      indent: 1cm
      align: right
      font: {size: 12pt}
    transition:
      space-before: 1em
      indent: 0.5in
      align: left
      case-transform: lower
      font: {size: 12pt}
    act-header:
      space-before: 1em
      align: right
      case-transform: lower
      font: {size: 12pt}
    scene-header:
      space-before: 2em
      space-after: 0.5em
      align: center
      case-transform: lower
      font: {size: 12pt}
    frontmatter:
      header:
        space-before: 1em
        align: center
        font: {size: 12pt}
`
	for _, format := range []string{"docx", "odt"} {
		t.Run(format, func(t *testing.T) {
			dir := t.TempDir()
			src := filepath.Join(dir, "geometry.org")
			writeAppFile(t, src, source)
			writeAppFile(t, filepath.Join(dir, "script.yaml"), cfg)
			raw, diagnostic, err := officeRun(t, binary, dir, officeEnv(t, t.TempDir(), nil), "convert", src, "--to", format)
			if err != nil {
				t.Fatalf("body geometry: %v %s", err, diagnostic)
			}
			pkg := officePackage(t, raw)
			body := officeBody(t, pkg, format)
			for _, role := range []struct {
				text, align         string
				before, after, left float64
			}{
				{"Direction text.", "right", 12, 3.6, 720 / 25.4},
				{"transition text", "left", 6, 0, 36},
				{"act one", "right", 12, 3.6, 0},
				{"scene one", "center", 24, 6, 0},
				{"Introduction", "center", 12, 0, 0},
			} {
				layers := officeEffective(t, pkg, format, officeParagraph(t, body, role.text))
				if format == "docx" {
					officeAssertEffective(t, layers, "jc", "val", role.align)
				} else {
					officeAssertEffective(t, layers, "paragraph-properties", "text-align", role.align)
				}
				officeAssertLength(t, layers, format, "spacing", "before", "paragraph-properties", "margin-top", role.before)
				officeAssertLength(t, layers, format, "spacing", "after", "paragraph-properties", "margin-bottom", role.after)
				officeAssertLength(t, layers, format, "ind", "left", "paragraph-properties", "margin-left", role.left)
			}
			officeAssertProp(t, pkg, format, body)
			page := body
			if format == "odt" {
				page = pkg["styles.xml"]
			}
			if format == "docx" {
				officeAssertLength(t, []*officeXML{page}, format, "pgMar", "top", "", "", 27)
				officeAssertLength(t, []*officeXML{page}, format, "pgMar", "bottom", "", "", 30)
			} else {
				// ODF running areas occupy the page margin plus font height and clearance.
				for _, layout := range page.all("page-layout") {
					for _, edge := range []struct {
						role, margin, gap string
						want              float64
					}{
						{"header", "margin-top", "margin-bottom", 27},
						{"footer", "margin-bottom", "margin-top", 30},
					} {
						raw, _ := officeEffectiveAttr([]*officeXML{layout}, "page-layout-properties", edge.margin)
						total := officeLengthPoints(t, raw)
						for _, running := range layout.all(edge.role + "-style") {
							for _, attr := range []string{"min-height", edge.gap} {
								value, exists := officeEffectiveAttr([]*officeXML{running}, "header-footer-properties", attr)
								if !exists {
									t.Fatalf("missing running area %s", attr)
								}
								total += officeLengthPoints(t, value)
							}
						}
						if math.Abs(total-edge.want) > .026 {
							t.Errorf("%s body boundary = %gpt, want %g", edge.role, total, edge.want)
						}
					}
				}
			}
			// British cue width resolves 5em against the 12pt root layout font.
			width := (210 - 50) * 72 / 25.4
			officeAssertTableRatio(t, pkg, format, body, "Dialogue text.", 60/width)
		})
	}
}

func TestRT054_2OfficeTitleCompositionAndMissingTitle(t *testing.T) {
	binary := officeBuild(t)
	for _, format := range []string{"docx", "odt"} {
		for _, position := range []string{"third", "center", "missing"} {
			t.Run(format+"/"+position, func(t *testing.T) {
				dir := t.TempDir()
				src := filepath.Join(dir, "title.org")
				source := "#+TITLE: Composition title\n#+SUBTITLE: Composition subtitle\n#+AUTHOR: Composition author\n#+DATE: Composition date\n#+VERSION: Composition version\n* ACT ONE\n**** CÁIT\nHello.\n"
				if position == "missing" {
					source = "* ACT ONE\n**** CÁIT\nHello.\n"
				}
				writeAppFile(t, src, source)
				writeAppFile(t, filepath.Join(dir, "script.yaml"), "folio:\n  page: us-letter\n  font: {size: 12pt}\n  title-page:\n    enabled: true\n    skip-header: true\n    skip-footer: true\n    title: {position: "+position+", align: right}\n    subtitle: {space-before: 2em, font: {size: 12pt}}\n    author: {space-before: 3em, font: {size: 12pt}}\n    date: {position: bottom-right}\n    version: {position: bottom-left}\n  page-header: {enabled: true, format: 'UNTITLED HEADER', alt-format: '', distance-from-edge: 10pt, content-padding-after: 20pt}\n  page-footer: {enabled: true, format: 'UNTITLED FOOTER', distance-from-edge: 10pt, content-padding-after: 20pt}\n")
				raw, diagnostic, err := officeRun(t, binary, dir, officeEnv(t, t.TempDir(), nil), "convert", src, "--to", format)
				if err != nil {
					t.Fatalf("title composition: %v %s", err, diagnostic)
				}
				pkg := officePackage(t, raw)
				body := officeBody(t, pkg, format)
				if position == "missing" {
					officeAssertUntitledFirstPage(t, pkg, format)
					return
				}
				title := officeEffective(t, pkg, format, officeParagraph(t, body, "Composition title"))
				fraction := .3
				if position == "center" {
					fraction = .4
				}
				// Letter is 792pt high, with configured 30pt running areas at each edge.
				officeAssertLength(t, title, format, "spacing", "before", "paragraph-properties", "margin-top", fraction*732)
				if format == "docx" {
					officeAssertEffective(t, title, "jc", "val", "right")
				} else {
					officeAssertEffective(t, title, "paragraph-properties", "text-align", "right")
				}
				for _, item := range []struct {
					text   string
					before float64
				}{{"Composition subtitle", 24}, {"Composition author", 36}} {
					layers := officeEffective(t, pkg, format, officeParagraph(t, body, item.text))
					officeAssertLength(t, layers, format, "spacing", "before", "paragraph-properties", "margin-top", item.before)
				}
				officeAssertTitleSlot(t, pkg, format, "Composition version", "left")
				officeAssertTitleSlot(t, pkg, format, "Composition date", "right")
			})
		}
	}
}
