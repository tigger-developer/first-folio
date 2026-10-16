// ABOUTME: Checks additive paragraph and block clearances through resolved native properties.
// ABOUTME: Compares independent configured deltas so font-box constants are not confused with spacing.
package app

import (
	"math"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func proseParagraphSpace(t *testing.T, pkg map[string]*officeXML, format, marker, side string) float64 {
	t.Helper()
	layers := officeEffective(t, pkg, format, officeParagraph(t, officeBody(t, pkg, format), marker))
	element, attr := "spacing", side
	if format == "odt" {
		element = "paragraph-properties"
		attr = "margin-top"
		if side == "after" {
			attr = "margin-bottom"
		}
	}
	raw, exists := officeEffectiveAttr(layers, element, attr)
	if !exists {
		return 0
	}
	if format == "odt" {
		return officeLengthPoints(t, raw)
	}
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		t.Fatal(err)
	}
	return value / 20
}

func TestRT063_4_OfficeAdditiveBlockClearance(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		for _, tc := range []struct {
			name, source, zero, changed, marker string
			before, after                       float64
		}{
			{"paragraph-spacing", "First.\n\nBODY-GAP\n\nLast.\n", "paragraph-spacing: 0pt", "paragraph-spacing: 6pt", "BODY-GAP", 0, 6},
			{"quote-clearance", "Before.\n\n> QUOTE-GAP\n\nAfter.\n", "quoted-block-spacing: 0pt", "quoted-block-spacing: 0.5em", "QUOTE-GAP", 6, 6},
			{"code-clearance", "Before.\n\n```\nCODE-GAP\n```\n\nAfter.\n", "code-block-spacing: 0pt", "code-block-spacing: 0.5em", "CODE-GAP", 6, 6},
			{"precise-code-clearance", "Before.\n\n```\nCODE-GAP\n```\n\nAfter.\n", "code-block-spacing: 2em\n    code-block:\n      space-before: 0pt\n      space-after: 0pt", "code-block-spacing: 2em\n    code-block:\n      space-before: 6pt\n      space-after: 9pt", "CODE-GAP", 6, 9},
			{"list-clearance", "Before.\n\n- LIST-GAP\n\nAfter.\n", "list:\n      space-before: 0pt\n      space-after: 0pt", "list:\n      space-before: 6pt\n      space-after: 9pt", "LIST-GAP", 6, 9},
		} {
			t.Run(format+"/"+tc.name, func(t *testing.T) {
				t.Setenv("HOME", t.TempDir())
				dir := t.TempDir()
				src := proseSource(t, dir, "source.md", tc.source)
				base := "folio:\n  manuscript:\n    title-page:\n      enabled: false\n    toc:\n      enabled: false\n    font:\n      family: Iosevka Publish\n      size: 12pt\n    mono:\n      font:\n        size: 9pt\n    quoted-block:\n      font:\n        size: 10pt\n    "
				var before, after [2]float64
				for i, config := range []string{tc.zero, tc.changed} {
					proseSource(t, dir, "script.yaml", base+config+"\n")
					pkg := proseConvert(t, src, filepath.Join(dir, "output."+format))
					before[i] = proseParagraphSpace(t, pkg, format, tc.marker, "before")
					after[i] = proseParagraphSpace(t, pkg, format, tc.marker, "after")
				}
				for _, check := range []struct {
					name      string
					got, want float64
				}{{"before", before[1] - before[0], tc.before}, {"after", after[1] - after[0], tc.after}} {
					if math.Abs(check.got-check.want) > 0.051 {
						t.Errorf("%s clearance changed %gpt; want %gpt", check.name, check.got, check.want)
					}
				}
			})
		}
	}
}

func proseTableClearance(t *testing.T, pkg map[string]*officeXML, format string) (float64, float64) {
	t.Helper()
	root := officeBody(t, pkg, format)
	flows := root.all("body")
	if format == "odt" {
		flows = root.all("text")
	}
	if len(flows) == 0 {
		t.Fatal("document has no text flow")
	}
	var blocks []*officeXML
	for _, node := range flows[0].Children {
		if node.Name.Local == "p" || node.Name.Local == "tbl" || node.Name.Local == "table" {
			blocks = append(blocks, node)
		}
	}
	before, table, after := -1, -1, -1
	for i, node := range blocks {
		if strings.Contains(node.text(), "BEFORE-TABLE") {
			before = i
		}
		if node.Name.Local == "tbl" || node.Name.Local == "table" {
			table = i
		}
		if strings.Contains(node.text(), "AFTER-TABLE") {
			after = i
		}
	}
	if before < 0 || table <= before || after <= table {
		t.Fatalf("table is not in ordinary flow between its markers: %d %d %d", before, table, after)
	}
	length := func(layers []*officeXML, wordElement, wordAttr, odtElement, odtAttr string) float64 {
		element, attr := wordElement, wordAttr
		if format == "odt" {
			element, attr = odtElement, odtAttr
		}
		raw, exists := officeEffectiveAttr(layers, element, attr)
		if !exists {
			return 0
		}
		if format == "odt" {
			return officeLengthPoints(t, raw)
		}
		value, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			t.Fatal(err)
		}
		return value / 20
	}
	space := func(start, end int, side string) float64 {
		var total float64
		for i := start; i <= end; i++ {
			node := blocks[i]
			layers := officeEffective(t, pkg, format, node)
			if i == table {
				attr := "margin-top"
				if side == "after" {
					attr = "margin-bottom"
				}
				if format == "odt" {
					total += length(layers, "", "", "table-properties", attr)
				}
				continue
			}
			if i != before {
				total += length(layers, "spacing", "before", "paragraph-properties", "margin-top")
			}
			if i != after {
				total += length(layers, "spacing", "after", "paragraph-properties", "margin-bottom")
			}
			if strings.TrimSpace(node.text()) == "" {
				total += length(layers, "spacing", "line", "paragraph-properties", "line-height")
			}
		}
		return total
	}
	return space(before, table, "before"), space(table, after, "after")
}

func TestRT063_4_OfficeTableClearance(t *testing.T) {
	for _, format := range []string{"docx", "odt"} {
		t.Run(format, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			dir := t.TempDir()
			src := proseSource(t, dir, "source.md", "BEFORE-TABLE\n\n| Head |\n|------|\n| Cell |\n\nAFTER-TABLE\n")
			var before, after [2]float64
			for i, spaces := range []struct{ before, after string }{{"0pt", "0pt"}, {"6pt", "9pt"}} {
				cfg := "folio:\n  manuscript:\n    title-page:\n      enabled: false\n    toc:\n      enabled: false\n    table:\n      space-before: " + spaces.before + "\n      space-after: " + spaces.after + "\n"
				proseSource(t, dir, "script.yaml", cfg)
				pkg := proseConvert(t, src, filepath.Join(dir, "output."+format))
				before[i], after[i] = proseTableClearance(t, pkg, format)
			}
			if math.Abs(before[1]-before[0]-6) > 0.051 || math.Abs(after[1]-after[0]-9) > 0.051 {
				t.Errorf("table clearance changes %g/%gpt; want 6/9pt", before[1]-before[0], after[1]-after[0])
			}
		})
	}
}
