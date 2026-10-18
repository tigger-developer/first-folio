// ABOUTME: Checks continuous ODT wording and approved frontmatter/body page transitions.
// ABOUTME: Uses rendered page text so native field declarations alone cannot pass.
package app

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

func TestRT063_12_ODTContinuousRunningTemplates(t *testing.T) {
	proseCheckODTContinuousBoundary(t, "never", false)
}

func TestRT063_12_ODTContinuousNumberReset(t *testing.T) {
	proseCheckODTContinuousBoundary(t, "first-part-or-chapter", false)
}

func TestRT063_12_ODTRunningFieldStructureChanges(t *testing.T) {
	proseCheckODTContinuousBoundary(t, "never", true)
}

func proseCheckODTContinuousBoundary(t *testing.T, reset string, differentFields bool) {
	t.Helper()
	root := proseRenderBudget(t, 50<<20)
	dir := proseRenderCase(t, root, "continuous")
	t.Setenv("HOME", dir)
	configuration := fmt.Sprintf(`folio:
  page-header:
    enabled: true
    format: '[page] BODY-HEADER'
    alt-format: ''
    frontmatter-format: 'FRONT-HEADER [page]'
    alt-frontmatter-format: 'FRONT-HEADER [page]'
    font: {family: Iosevka Publish, size: 10pt, letter-spacing: 0pt}
  page-footer:
    enabled: false
  manuscript:
    page: a5
    margin: 20mm
    font: {family: Iosevka Publish, size: 12pt}
    title-page: {enabled: false}
    toc: {enabled: false}
    chapter:
      page-break-before: false
      blank-page-before: false
      blank-page-after: false
      position: 0mm
      skip-header: false
      skip-footer: false
    page-numbering:
      frontmatter-format: '1'
      body-format: '1'
      body-reset: %s
`, reset)
	if differentFields {
		configuration = strings.ReplaceAll(configuration, "FRONT-HEADER [page]", "FRONT-HEADER")
	}
	proseSource(t, dir, "script.yaml", configuration)
	source := proseSource(t, dir, "source.md", "INTRO-FIRST\n\n"+strings.Repeat("Introduction paragraph.\n\n", 30)+"INTRO-LAST\n\n## Body Chapter\n\n"+strings.Repeat("Body paragraph.\n\n", 30)+"BODY-LAST\n")
	target := filepath.Join(dir, "output.odt")
	proseConvert(t, source, target)
	pdf := proseRenderPDF(t, target)
	proseFailurePages(t, pdf, dir, 6, "continuous-failure")
	pages := strings.Split(proseTool(t, "pdftotext", "-layout", pdf, "-"), "\f")
	firstBody := -1
	found := false
	introLast := -1
	for pageIndex, page := range pages {
		if strings.Contains(page, "INTRO-LAST") {
			introLast = pageIndex
		}
		if firstBody < 0 && strings.Contains(page, "Body paragraph.") {
			firstBody = pageIndex
		}
		if strings.Contains(page, "INTRO-FIRST") && !strings.Contains(page, "FRONT-HEADER") {
			t.Errorf("frontmatter lost its running template: %s", page)
		}
		if strings.Contains(page, "BODY-LAST") {
			found = true
			if reset != "never" {
				want := fmt.Sprintf("%d BODY-HEADER", pageIndex-firstBody+1)
				if !strings.Contains(page, want) {
					t.Errorf("body reset wanted %q: %s", want, page)
				}
			}
			if !strings.Contains(page, "BODY-HEADER") || strings.Contains(page, "FRONT-HEADER") {
				t.Errorf("body continuation uses wrong running template: %s", page)
			}
		}
	}
	if reset != "never" || differentFields {
		if introLast < 0 || firstBody <= introLast {
			t.Error("approved ODT boundary did not start body on a new page")
		}
	} else if introLast < 0 || firstBody != introLast {
		t.Error("same-format wording-only ODT boundary did not remain on the introduction's last page")
	}
	if !found {
		t.Error("rendered body endpoint missing")
	}
}
