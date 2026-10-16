// ABOUTME: W039 regression for manuscript role-scoped font rendering.
// ABOUTME: Verifies all six properties apply without leaking from the body role.
package manuscript

import (
	"fmt"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

var manuscriptFontRolePaths = []string{
	"folio.manuscript.font",
	"folio.manuscript.heading.font",
	"folio.manuscript.mono.font",
	"folio.manuscript.quoted-block.font",
	"folio.manuscript.title-page.title.font",
	"folio.manuscript.title-page.subtitle.font",
	"folio.manuscript.title-page.author.font",
	"folio.manuscript.title-page.date.font",
	"folio.manuscript.title-page.wordcount.font",
	"folio.manuscript.title-page.version.font",
	"folio.manuscript.title-page.contact.font",
	"folio.manuscript.page-header.font",
	"folio.manuscript.page-footer.font",
	"folio.manuscript.toc.font",
	"folio.manuscript.toc.heading.font",
	"folio.manuscript.copyright.font",
	"folio.manuscript.copyright.label.font",
}

func TestRT039_1EveryManuscriptFontPropertyReachesItsRole(t *testing.T) {
	properties := []struct {
		name, value, marker string
	}{
		{"family", "W039 Family", `"W039 Family"`},
		{"size", "13.5pt", "13.5pt"},
		{"weight", "643", "643"},
		{"stretch", "117.5%", "117.5%"},
		{"style", "oblique", `"oblique"`},
		{"letter-spacing", "-0.03em", "-0.03em"},
	}
	for _, path := range manuscriptFontRolePaths {
		for _, property := range properties {
			t.Run(strings.TrimPrefix(path, "folio.manuscript.")+"/"+property.name, func(t *testing.T) {
				configYAML := manuscriptFontPropertyYAML(t, path, property.name, property.value)
				typst := renderIssue15Manuscript(t, configYAML)
				if !strings.Contains(typst, property.marker) {
					t.Fatalf("%s did not emit %s marker %q", path, property.name, property.marker)
				}
			})
		}
	}
}

func TestRT039_1ManuscriptFontBlockEmitsAllPropertiesForSelectedRole(t *testing.T) {
	typst := renderIssue15Manuscript(t, strings.Join([]string{
		"folio:",
		"  manuscript:",
		"    page-header:",
		"      font:",
		"        family: W039 Header",
		"        size: 13.5pt",
		"        weight: 650",
		"        stretch: 117.5%",
		"        style: oblique",
		"        letter-spacing: -0.03em",
		"",
	}, "\n"))
	header := extractHeaderBlock(t, typst)
	assertContains(t, header, `font: "W039 Header", size: 13.5pt, weight: 650, stretch: 117.5%, style: "oblique", tracking: -0.03em`)
	assertContains(t, typst, `font: "Libertinus Serif", size: 12pt, weight: "regular"`)
	assertNotContains(t, header, `font: "Libertinus Serif"`)
}

func manuscriptFontPropertyYAML(t *testing.T, path, property, value string) string {
	t.Helper()
	data := map[string]any{}
	setFontTestPath(data, path+"."+property, value)
	setFontTestPath(data, "folio.manuscript.copyright.enabled", true)
	setFontTestPath(data, "folio.manuscript.copyright.publisher", "W039 Publisher")
	setFontTestPath(data, "folio.manuscript.copyright.isbn", "978-1-23456-789-7")
	raw, err := yaml.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func setFontTestPath(data map[string]any, path string, value any) {
	parts := strings.Split(path, ".")
	node := data
	for _, part := range parts[:len(parts)-1] {
		child, ok := node[part].(map[string]any)
		if !ok {
			child = map[string]any{}
			node[part] = child
		}
		node = child
	}
	node[parts[len(parts)-1]] = value
}

func Example_manuscriptFontCoverage() {
	fmt.Println(len(manuscriptFontRolePaths) * 6)
	// Output: 102
}
