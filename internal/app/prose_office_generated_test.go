// ABOUTME: Checks independent title alignment and copyright ordering through native paragraphs.
// ABOUTME: Uses labelled fields and contrasting settings without relying on generated style names.
package app

import (
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func proseAssertFile(t *testing.T, path, want string) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil || string(raw) != want {
		t.Errorf("%s changed: %v", path, err)
	}
}

func proseAssertAlignment(t *testing.T, pkg map[string]*officeXML, format, marker, want string) {
	t.Helper()
	layers := officeEffective(t, pkg, format, officeParagraph(t, officeBody(t, pkg, format), marker))
	element, attr := "jc", "val"
	if format == "odt" {
		element, attr = "paragraph-properties", "text-align"
	}
	got, exists := officeEffectiveAttr(layers, element, attr)
	if got == "start" {
		got = "left"
	}
	if got == "end" {
		got = "right"
	}
	if !exists || got != want {
		t.Errorf("%s alignment=%q present=%v; want %s", marker, got, exists, want)
	}
}

func TestRT063_5_OfficeIndependentTitleAlignment(t *testing.T) {
	metadata := "---\ntitle: TITLE-ALIGN\nsubtitle: SUBTITLE-ALIGN\nauthor: AUTHOR-ALIGN\ndate: DATE-ALIGN\nwordcount: WORDCOUNT-ALIGN\nversion: VERSION-ALIGN\ncontact-name: CONTACT-ALIGN\n---\n\nBODY-ALIGN\n"
	for _, format := range []string{"docx", "odt"} {
		for _, align := range []string{"left", "center", "right"} {
			t.Run(format+"/"+align, func(t *testing.T) {
				t.Setenv("HOME", t.TempDir())
				dir := t.TempDir()
				cfg := "folio:\n  manuscript:\n    toc:\n      enabled: false\n    title-page:\n      enabled: true\n      title-block-align: left\n      footer-align: right\n"
				roles := []string{"title", "subtitle", "author", "date", "wordcount", "version", "contact"}
				for _, role := range roles {
					cfg += "      " + role + ":\n        align: " + align + "\n"
				}
				proseSource(t, dir, "script.yaml", cfg)
				src := proseSource(t, dir, "source.md", metadata)
				pkg := proseConvert(t, src, filepath.Join(dir, "output."+format))
				for _, role := range roles {
					proseAssertAlignment(t, pkg, format, strings.ToUpper(role)+"-ALIGN", align)
				}
			})
		}
	}
}

func TestRT063_5_OfficeHeadingRichText(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		t.Run(format, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			dir := t.TempDir()
			proseSource(t, dir, "script.yaml", "folio:\n  manuscript:\n    title-page:\n      enabled: false\n    toc:\n      enabled: false\n    heading:\n      font:\n        weight: regular\n    chapter:\n      show-number: false\n      name-case: upper\n")
			src := proseSource(t, dir, "source.md", "## Chapter 7: *Italic-heading* and **Bold-heading** [brackets]\n\nBody.\n")
			pkg := proseConvert(t, src, filepath.Join(dir, "output."+format))
			body := officeBody(t, pkg, format)
			if !strings.Contains(body.text(), "ITALIC-HEADING AND BOLD-HEADING [BRACKETS]") || strings.Contains(body.text(), `\[`) {
				t.Errorf("composed heading escaped or lost literal text: %s", body.text())
			}
			for _, marker := range []string{"ITALIC-HEADING", "BOLD-HEADING"} {
				layers := proseInlineLayers(t, pkg, format, marker)
				element, attr, want := "i", "val", "true"
				if marker == "BOLD-HEADING" {
					element = "b"
				}
				if format == "odt" {
					element, attr, want = "text-properties", "font-style", "italic"
					if marker == "BOLD-HEADING" {
						attr, want = "font-weight", "bold"
					}
				}
				got, exists := officeEffectiveAttr(layers, element, attr)
				if format == "docx" {
					for _, layer := range layers {
						for _, property := range layer.all(element) {
							if property.attr("val") == "" {
								got, exists = "true", true
							}
						}
					}
				}
				if !exists || got != want && !(format == "docx" && got == "") {
					t.Errorf("%s emphasis=%q present=%v; want %s", marker, got, exists, want)
				}
			}
		})
	}
}

func TestRT063_5_OfficeCopyrightOrder(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		for _, position := range []string{"after-title", "after-toc", "after-frontmatter"} {
			t.Run(format+"/"+position, func(t *testing.T) {
				t.Setenv("HOME", t.TempDir())
				dir := t.TempDir()
				cfg := "folio:\n  manuscript:\n    title-page:\n      enabled: true\n    toc:\n      enabled: true\n      title: CONTENTS-ORDER\n    copyright:\n      enabled: true\n      position: " + position + "\n      credits: [CREDIT-ORDER]\n      body: [LEGAL-ORDER]\n      separator: SEPARATOR-ORDER\n      publication: [PUBLICATION-ORDER]\n      publisher-preposition: Printed by\n      publisher: PUBLISHER-ORDER\n      isbn-label: NUMBER-ORDER\n      isbn: '978-0-000000-00-2'\n"
				proseSource(t, dir, "script.yaml", cfg)
				src := proseSource(t, dir, "source.md", "---\ntitle: TITLE-ORDER\n---\n\n## Chapter\n\nBODY-ORDER\n")
				pkg := proseConvert(t, src, filepath.Join(dir, "output."+format))
				text := officeBody(t, pkg, format).text()
				sequence := []string{"TITLE-ORDER", "CREDIT-ORDER", "LEGAL-ORDER", "SEPARATOR-ORDER", "PUBLICATION-ORDER", "Printed by", "PUBLISHER-ORDER", "NUMBER-ORDER", "978-0-000000-00-2", "BODY-ORDER"}
				previous := -1
				for _, marker := range sequence {
					at := strings.Index(text, marker)
					if at < 0 || at <= previous {
						t.Errorf("%s missing or out of order", marker)
					}
					previous = at
				}
				copyright, contents := strings.Index(text, "CREDIT-ORDER"), strings.Index(text, "CONTENTS-ORDER")
				if contents < 0 || (copyright < contents) != (position == "after-title") {
					t.Errorf("copyright position %s not retained relative to contents", position)
				}
			})
		}
	}
}

func TestRT063_5_OfficeInvalidISBNPreservesBothTargets(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		t.Run(format, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			dir := t.TempDir()
			proseSource(t, dir, "script.yaml", "folio:\n  manuscript:\n    copyright:\n      enabled: true\n      isbn: '978-0-000000-00-3'\n      isbn-barcode: render-and-file\n")
			src := proseSource(t, dir, "source.md", "BODY\n")
			target := proseSource(t, dir, "output."+format, "PRIOR-DOCUMENT")
			sidecar := proseSource(t, dir, "output.barcode.svg", "PRIOR-SIDECAR")
			status, _, stderr := runApp(t, "manuscript", src, target)
			if status == 0 || !strings.Contains(stderr, "isbn") {
				t.Errorf("invalid ISBN accepted or not identified: %d %s", status, stderr)
			}
			proseAssertFile(t, target, "PRIOR-DOCUMENT")
			proseAssertFile(t, sidecar, "PRIOR-SIDECAR")
		})
	}
}

func proseAssertVerticalAnchor(t *testing.T, pkg map[string]*officeXML, format, marker, want string) {
	t.Helper()
	var ancestors []*officeXML
	var locate func(*officeXML, []*officeXML) bool
	locate = func(node *officeXML, parents []*officeXML) bool {
		parents = append(append([]*officeXML{}, parents...), node)
		for _, child := range node.Children {
			if locate(child, parents) {
				return true
			}
		}
		if (node.Name.Local == "p" || node.Name.Local == "h") && strings.Contains(node.text(), marker) {
			ancestors = parents
			return true
		}
		return false
	}
	if !locate(officeBody(t, pkg, format), nil) {
		t.Fatalf("no anchored paragraph %s", marker)
	}
	for _, node := range ancestors {
		switch node.Name.Local {
		case "p":
			for _, frame := range node.all("framePr") {
				if frame.attr("yAlign") == want && frame.attr("vAnchor") == "margin" {
					return
				}
			}
		case "anchor":
			for _, position := range node.all("positionV") {
				if position.attr("relativeFrom") != "margin" {
					continue
				}
				for _, align := range position.all("align") {
					if align.text() == want {
						return
					}
				}
			}
		case "frame":
			layers := officeEffective(t, pkg, format, node)
			value, _ := officeEffectiveAttr(layers, "graphic-properties", "vertical-pos")
			reference, _ := officeEffectiveAttr(layers, "graphic-properties", "vertical-rel")
			if value == "middle" {
				value = "center"
			}
			if value == want && reference == "page-content" {
				return
			}
		}
	}
	t.Errorf("%s lacks independent %s positioning relative to page content", marker, want)
}

func TestRT063_5_OfficeIndependentTitlePlacement(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		t.Run(format, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			dir := t.TempDir()
			cfg := "folio:\n  manuscript:\n    toc:\n      enabled: false\n    title-page:\n      enabled: true\n      title:\n        align: top-left\n      author:\n        align: bottom-right\n      contact:\n        align: center-left\n"
			proseSource(t, dir, "script.yaml", cfg)
			src := proseSource(t, dir, "source.md", "---\ntitle: INDEPENDENT-TOP\nauthor: INDEPENDENT-BOTTOM\ncontact-name: INDEPENDENT-CENTER\n---\n\nBODY-SENTINEL\n")
			pkg := proseConvert(t, src, filepath.Join(dir, "output."+format))
			for _, item := range []struct{ marker, vertical, horizontal string }{
				{"INDEPENDENT-TOP", "top", "left"}, {"INDEPENDENT-BOTTOM", "bottom", "right"}, {"INDEPENDENT-CENTER", "center", "left"},
			} {
				proseAssertVerticalAnchor(t, pkg, format, item.marker, item.vertical)
				proseAssertAlignment(t, pkg, format, item.marker, item.horizontal)
			}
		})
	}
}

func TestRT063_5_OfficeCopyrightBottomGroup(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		t.Run(format, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			dir := t.TempDir()
			proseSource(t, dir, "script.yaml", "folio:\n  manuscript:\n    title-page:\n      enabled: false\n    toc:\n      enabled: false\n    copyright:\n      enabled: true\n      credits: [TOP-LEGAL-CREDIT]\n      body: [TOP-LEGAL-BODY]\n      publication: [BOTTOM-PUBLICATION]\n      publisher: BOTTOM-PUBLISHER\n      isbn: '978-0-000000-00-2'\n")
			src := proseSource(t, dir, "source.md", "## Chapter\n\nBODY-SENTINEL\n")
			pkg := proseConvert(t, src, filepath.Join(dir, "output."+format))
			proseAssertVerticalAnchor(t, pkg, format, "BOTTOM-PUBLICATION", "bottom")
			proseAssertVerticalAnchor(t, pkg, format, "BOTTOM-PUBLISHER", "bottom")
			if format == "odt" {
				flow := pkg["content.xml"]
				groups := 0
				for _, box := range flow.all("text-box") {
					if strings.Contains(box.text(), "BOTTOM-PUBLICATION") && strings.Contains(box.text(), "BOTTOM-PUBLISHER") {
						groups++
					}
				}
				if groups != 1 {
					t.Errorf("copyright publication and publisher occupy %d common text boxes; want one non-overlapping group", groups)
				}
			}
		})
	}
}

func TestRT063_5_OfficeODTTitleFloatsShareVerticalFlow(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	proseSource(t, dir, "script.yaml", "folio:\n  manuscript:\n    toc:\n      enabled: false\n    title-page:\n      enabled: true\n      contact:\n        align: top-left\n      wordcount:\n        align: top-right\n      date:\n        align: bottom-left\n      version:\n        align: bottom-right\n")
	source := proseSource(t, dir, "source.md", "---\ntitle: CENTER-TITLE\ncontact-name: TOP-CONTACT\nwordcount: TOP-WORDCOUNT\ndate: BOTTOM-DATE\nversion: BOTTOM-VERSION\n---\n\nBODY-SENTINEL\n")
	pkg := proseConvert(t, source, filepath.Join(dir, "output.odt"))
	for _, pair := range [][2]string{{"TOP-CONTACT", "TOP-WORDCOUNT"}, {"BOTTOM-DATE", "BOTTOM-VERSION"}} {
		found := false
		for _, box := range pkg["content.xml"].all("text-box") {
			text := box.text()
			if strings.Contains(text, pair[0]) && strings.Contains(text, pair[1]) {
				found = true
				if strings.Index(text, pair[0]) > strings.Index(text, pair[1]) {
					t.Errorf("positioned fields have reversed flow: %s", text)
				}
			}
		}
		if !found {
			t.Errorf("%s and %s occupy overlapping independent frames instead of one vertical flow", pair[0], pair[1])
		}
	}
}

func TestRT063_5_OfficeCopyrightSpacing(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		t.Run(format, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			dir := t.TempDir()
			src := proseSource(t, dir, "source.md", "## Chapter\n\nBODY-SENTINEL\n")
			var gap, separatorTop, separatorBottom [2]float64
			for i, spacing := range []string{"0pt", "6pt"} {
				before, after := "0pt", "0pt"
				if i == 1 {
					before, after = "8pt", "9pt"
				}
				cfg := "folio:\n  manuscript:\n    title-page:\n      enabled: false\n    toc:\n      enabled: false\n    line-spacing: 2\n    copyright:\n      enabled: true\n      font:\n        family: Iosevka Publish\n        size: 12pt\n      line-spacing: 1.4\n      block-spacing: " + spacing + "\n      credits: [CREDIT-SPACING]\n      body: [LEGAL-SPACING]\n      separator: SEPARATOR-SPACING\n      separator-space-before: " + before + "\n      separator-space-after: " + after + "\n"
				proseSource(t, dir, "script.yaml", cfg)
				pkg := proseConvert(t, src, filepath.Join(dir, "output."+format))
				gap[i] = proseParagraphSpace(t, pkg, format, "CREDIT-SPACING", "after") + proseParagraphSpace(t, pkg, format, "LEGAL-SPACING", "before")
				legal := officeEffective(t, pkg, format, officeParagraph(t, officeBody(t, pkg, format), "LEGAL-SPACING"))
				// Copyright retains its existing (multiplier-1)em leading,
				// distinct from body/TOC's multiplier of the normal interval.
				officeAssertLength(t, legal, format, "spacing", "line", "paragraph-properties", "line-height", 13.62)
				separatorTop[i] = proseParagraphSpace(t, pkg, format, "LEGAL-SPACING", "after") + proseParagraphSpace(t, pkg, format, "SEPARATOR-SPACING", "before")
				separatorBottom[i] = proseParagraphSpace(t, pkg, format, "SEPARATOR-SPACING", "after")
			}
			if math.Abs(gap[1]-gap[0]-6) > 0.051 {
				t.Errorf("copyright block gap changed %gpt; want 6pt", gap[1]-gap[0])
			}
			if math.Abs(separatorTop[1]-separatorTop[0]-14) > 0.051 {
				t.Errorf("separator top gap changed %gpt; want additive 6+8pt", separatorTop[1]-separatorTop[0])
			}
			if math.Abs(separatorBottom[1]-separatorBottom[0]-9) > 0.051 {
				t.Errorf("separator bottom gap changed %gpt; want 9pt", separatorBottom[1]-separatorBottom[0])
			}
		})
	}
}

func TestRT063_4_OfficePartVerticalPlacement(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		for _, vertical := range []string{"center", "bottom"} {
			t.Run(format+"/"+vertical, func(t *testing.T) {
				t.Setenv("HOME", t.TempDir())
				dir := t.TempDir()
				cfg := "folio:\n  manuscript:\n    title-page:\n      enabled: false\n    toc:\n      enabled: false\n    part:\n      vertical-align: " + vertical + "\n    chapter:\n      page-break-before: true\n"
				proseSource(t, dir, "script.yaml", cfg)
				src := proseSource(t, dir, "source.md", "# VERTICAL-PART\n\n## Chapter\n\nBody.\n")
				pkg := proseConvert(t, src, filepath.Join(dir, "output."+format))
				if format == "docx" {
					section := proseSectionForMarker(t, pkg, "VERTICAL-PART")
					value, _ := officeEffectiveAttr([]*officeXML{section}, "vAlign", "val")
					if value == vertical {
						return
					}
				}
				proseAssertVerticalAnchor(t, pkg, format, "VERTICAL-PART", vertical)
			})
		}
	}
}

func TestRT063_4_OfficeChapterPosition(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		for _, position := range []struct {
			value  string
			points float64
		}{
			{"20mm", 20 * 72 / 25.4}, {"2em", 24}, {"30%", 257 * 0.3 * 72 / 25.4}, {"one-third", 257 * 0.3 * 72 / 25.4}, {"center", 257 * 0.5 * 72 / 25.4}, {"top", 0},
		} {
			t.Run(format+"/"+position.value, func(t *testing.T) {
				t.Setenv("HOME", t.TempDir())
				dir := t.TempDir()
				cfg := "folio:\n  page-header:\n    enabled: false\n  page-footer:\n    enabled: false\n  manuscript:\n    page: a4\n    margin: 20mm\n    font:\n      size: 12pt\n    title-page:\n      enabled: false\n    toc:\n      enabled: false\n    chapter:\n      position: " + position.value + "\n"
				proseSource(t, dir, "script.yaml", cfg)
				src := proseSource(t, dir, "source.md", "## POSITIONED-CHAPTER\n\nBody.\n")
				pkg := proseConvert(t, src, filepath.Join(dir, "output."+format))
				got := proseParagraphSpace(t, pkg, format, "POSITIONED-CHAPTER", "before")
				if math.Abs(got-position.points) > 0.026 {
					t.Errorf("chapter preceding space %gpt; want %gpt for %s", got, position.points, position.value)
				}
			})
		}
	}
}

func TestRT063_5_OfficeGroupedTitleFooter(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		t.Run(format, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			dir := t.TempDir()
			proseSource(t, dir, "script.yaml", "folio:\n  manuscript:\n    font: {size: 12pt}\n    toc: {enabled: false}\n")
			src := proseSource(t, dir, "source.md", "---\ntitle: GROUP-TITLE\nsubtitle: GROUP-SUBTITLE\nauthor: GROUP-AUTHOR\ndate: GROUP-DATE\nversion: GROUP-VERSION\nwordcount: GROUP-WORDCOUNT\n---\n\nBody.\n")
			pkg := proseConvert(t, src, filepath.Join(dir, "output."+format))
			for _, item := range []struct{ marker, align string }{{"GROUP-DATE", "right"}, {"GROUP-VERSION", "left"}, {"GROUP-WORDCOUNT", "center"}} {
				proseAssertVerticalAnchor(t, pkg, format, item.marker, "bottom")
				proseAssertAlignment(t, pkg, format, item.marker, item.align)
			}
			for _, item := range []struct {
				marker string
				before float64
			}{{"GROUP-SUBTITLE", 12}, {"GROUP-AUTHOR", 24}} {
				layers := officeEffective(t, pkg, format, officeParagraph(t, officeBody(t, pkg, format), item.marker))
				officeAssertLength(t, layers, format, "spacing", "before", "paragraph-properties", "margin-top", item.before)
			}
		})
	}
}

func TestRT063_5_OfficeWordTitleFloatsShareVerticalFlow(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	proseSource(t, dir, "script.yaml", "folio:\n  manuscript:\n    toc: {enabled: false}\n    title-page:\n      contact: {align: top-left}\n      wordcount: {align: top-right}\n      date: {align: bottom-left}\n      version: {align: bottom-right}\n")
	source := proseSource(t, dir, "source.md", "---\ntitle: CENTER-TITLE\ncontact-name: TOP-CONTACT\nwordcount: TOP-WORDCOUNT\ndate: BOTTOM-DATE\nversion: BOTTOM-VERSION\n---\n\nBODY-SENTINEL\n")
	pkg := proseConvert(t, source, filepath.Join(dir, "output.docx"))
	paragraphs := officeBody(t, pkg, "docx").all("p")
	for _, pair := range [][2]string{{"TOP-CONTACT", "TOP-WORDCOUNT"}, {"BOTTOM-DATE", "BOTTOM-VERSION"}} {
		first, second := -1, -1
		for i, paragraph := range paragraphs {
			if strings.Contains(paragraph.text(), pair[0]) {
				first = i
			}
			if strings.Contains(paragraph.text(), pair[1]) {
				second = i
			}
		}
		if first < 0 || second != first+1 {
			t.Errorf("%s and %s do not share adjacent frame paragraphs: %d, %d", pair[0], pair[1], first, second)
			continue
		}
		for _, attribute := range []string{"xAlign", "yAlign", "w", "hAnchor", "vAnchor", "wrap"} {
			a, b := paragraphs[first].all("framePr"), paragraphs[second].all("framePr")
			if len(a) != 1 || len(b) != 1 || a[0].attr(attribute) != b[0].attr(attribute) {
				t.Errorf("%s and %s cannot form one frame: differing %s", pair[0], pair[1], attribute)
			}
		}
	}
}
