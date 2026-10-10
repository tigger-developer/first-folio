// ABOUTME: Verifies office styles are attached to content and native speech geometry.
// ABOUTME: Exercises active font roles and explicit page-break rules through the CLI.
package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func officeAssertStyleUsed(t *testing.T, pkg map[string]*officeXML, format, name string) *officeXML {
	t.Helper()
	style := officeStyle(t, pkg, format, name)
	id := style.attr("styleId")
	if format == "odt" {
		id = style.attr("name")
	}
	if id == "" {
		t.Fatalf("style %s has no native identifier", name)
	}
	for file, tree := range pkg {
		if strings.Contains(file, "styles.xml") {
			continue
		}
		for _, n := range tree.all("pStyle") {
			if n.attr("val") == id {
				return style
			}
		}
		for _, n := range tree.all("rStyle") {
			if n.attr("val") == id {
				return style
			}
		}
		for _, kind := range []string{"p", "span"} {
			for _, n := range tree.all(kind) {
				if n.attr("style-name") == id {
					return style
				}
			}
		}
	}
	t.Fatalf("native style %s is declared but never applied", name)
	return nil
}

func TestRT054_2OfficeActiveRoleFonts(t *testing.T) {
	binary := officeBuild(t)
	// Each family identifies its active role; this prevents an unrelated style
	// declaration from satisfying a role's property checks.
	const cfg = `folio:
  font:
    family: Body Family
    size: 10pt
  title-page:
    title:
      font: {family: Title Family, size: 14pt}
    subtitle:
      font: {family: Subtitle Family, size: 13pt}
    author:
      prefix: "Written by\nExample prefix\n"
      font: {family: Author Family, size: 12pt}
    date:
      font: {family: Date Family, size: 11pt}
    version:
      font: {family: Version Family, size: 10pt}
  page-header:
    enabled: true
    format: Header text
    font: {family: Header Family, size: 9pt}
  page-footer:
    font: {family: Footer Family, size: 8pt}
  positioning:
    speech:
      speaker:
        font: {family: Speaker Family, size: 11pt}
      speech-instruction:
        font: {family: Instruction Family, size: 12pt}
      dialogue:
        font: {family: Dialogue Family, size: 13pt}
    stage-direction:
      font: {family: Direction Family, size: 14pt}
    transition:
      font: {family: Transition Family, size: 15pt}
    frontmatter:
      header:
        font: {family: Intro Family, size: 16pt}
    act-header:
      font: {family: Act Family, size: 17pt}
    scene-header:
      font: {family: Scene Family, size: 18pt}
`
	roles := []struct{ name, family, size, halfPoints string }{
		{"Title", "Title", "14", "28"}, {"Subtitle", "Subtitle", "13", "26"},
		{"Author", "Author", "12", "24"}, {"Date", "Date", "11", "22"},
		{"Version", "Version", "10", "20"}, {"Header", "Header", "9", "18"},
		{"Footer", "Footer", "8", "16"}, {"Speaker", "Speaker", "11", "22"},
		{"Instruction", "Instruction", "12", "24"}, {"Dialogue", "Dialogue", "13", "26"},
		{"Direction", "Direction", "14", "28"}, {"Transition", "Transition", "15", "30"},
		{"Intro", "Intro", "16", "32"}, {"Act", "Act", "17", "34"}, {"Scene", "Scene", "18", "36"},
	}
	for _, format := range []string{"docx", "odt"} {
		t.Run(format, func(t *testing.T) {
			dir := t.TempDir()
			src := filepath.Join(dir, "roles.org")
			writeAppFile(t, src, "#+TITLE: Role title\n#+SUBTITLE: Role subtitle\n#+AUTHOR: Role author\n#+DATE: Role date\n#+VERSION: Role version\n* Introduction\nIntroductory paragraph.\n* ACT ONE\n** SCENE ONE\n*** Direction text.\n**** CÁIT (instruction text)\nDialogue text.[fn:n]\n***** Transition text\n[fn:n] Footnote text.\n")
			writeAppFile(t, filepath.Join(dir, "script.yaml"), cfg)
			raw, diagnostic, err := officeRun(t, binary, dir, officeEnv(t, t.TempDir(), nil), "convert", src, "--to", format)
			if err != nil {
				t.Fatalf("role export: %v %s", err, diagnostic)
			}
			pkg := officePackage(t, raw)
			for _, role := range roles {
				style := officeAssertStyleUsed(t, pkg, format, "Folio "+role.name)
				if format == "docx" {
					officeAssertAttr(t, style, "rFonts", "ascii", role.family+" Family")
					officeAssertAttr(t, style, "sz", "val", role.halfPoints)
				} else {
					officeAssertAttr(t, style, "text-properties", "font-name", role.family+" Family")
					officeAssertAttr(t, style, "text-properties", "font-size", role.size+"pt")
				}
			}
			if !strings.Contains(officeBody(t, pkg, format).text(), "Example prefix") {
				t.Error("multiline author prefix lost")
			}
		})
	}
}

func TestRT054_2OfficeSeparateSpeechGeometry(t *testing.T) {
	binary := officeBuild(t)
	for _, format := range []string{"docx", "odt"} {
		for _, styleName := range []string{"us", "screenplay"} {
			t.Run(format+"/"+styleName, func(t *testing.T) {
				dir := t.TempDir()
				src := filepath.Join(dir, "speech.org")
				writeAppFile(t, src, officeSources[".org"])
				writeAppFile(t, filepath.Join(dir, "script.yaml"), `folio:
  positioning:
    speech:
      space-before: 14pt
      dialogue:
        indent: 18pt
        wrap-indent: 36pt
      speaker:
        indent: 9pt
        align: right
        prefix: '['
        suffix: ']'
        case-transform: lower
      speech-instruction:
        align: center
        prefix: '{'
        suffix: '}'
`)
				raw, diagnostic, err := officeRun(t, binary, dir, officeEnv(t, t.TempDir(), nil), "convert", src, "--to", format, "--style", styleName)
				if err != nil {
					t.Fatalf("speech geometry: %v %s", err, diagnostic)
				}
				pkg := officePackage(t, raw)
				body := officeBody(t, pkg, format)
				if !strings.Contains(body.text(), "[cáit]") || !strings.Contains(body.text(), "{softly}") {
					t.Error("speaker casing or cue delimiters lost", body.text())
				}
				for _, role := range []struct{ name, left, right, before, align string }{
					{"Speaker", "540", "720", "280", "right"},
					{"Instruction", "360", "720", "0", "center"},
					{"Dialogue", "360", "720", "0", "left"},
				} {
					style := officeAssertStyleUsed(t, pkg, format, "Folio "+role.name)
					if format == "docx" {
						officeAssertAttr(t, style, "ind", "left", role.left)
						officeAssertAttr(t, style, "ind", "right", role.right)
						officeAssertAttr(t, style, "jc", "val", role.align)
						if role.name == "Speaker" {
							officeAssertAttr(t, style, "spacing", "before", role.before)
						}
					} else {
						left := "18pt"
						if role.name == "Speaker" {
							left = "27pt"
						}
						officeAssertAttr(t, style, "paragraph-properties", "margin-left", left)
						officeAssertAttr(t, style, "paragraph-properties", "margin-right", "36pt")
						officeAssertAttr(t, style, "paragraph-properties", "text-align", role.align)
					}
				}
			})
		}
	}
}

func TestRT054_2OfficeActBreaks(t *testing.T) {
	binary := officeBuild(t)
	for _, format := range []string{"docx", "odt"} {
		for _, compact := range []bool{true, false} {
			t.Run(format+map[bool]string{true: "/compact", false: "/dedicated"}[compact], func(t *testing.T) {
				dir := t.TempDir()
				src := filepath.Join(dir, "acts.org")
				writeAppFile(t, src, "#+TITLE: Break rules\n* ACT ONE\n**** CÁIT\nFirst.\n* ACT TWO\n**** CÁIT\nSecond.\n")
				writeAppFile(t, filepath.Join(dir, "script.yaml"), "folio:\n  title-page:\n    enabled: "+map[bool]string{true: "false", false: "true"}[compact]+"\n")
				target := filepath.Join(dir, "acts."+format)
				_, diagnostic, err := officeRun(t, binary, dir, officeEnv(t, t.TempDir(), nil), "convert", src, target)
				if err != nil {
					t.Fatalf("act breaks: %v %s", err, diagnostic)
				}
				raw, err := os.ReadFile(target)
				if err != nil {
					t.Fatal(err)
				}
				pkg := officePackage(t, raw)
				body := officeBody(t, pkg, format)
				firstBreak := officeBreakBefore(t, pkg, format, "ACT ONE")
				if firstBreak == compact {
					t.Errorf("first act explicit break=%v; compact=%v", firstBreak, compact)
				}
				// Later acts need an explicit break even in compact mode.
				found := false
				for _, p := range body.all("p") {
					if !strings.Contains(p.text(), "ACT TWO") {
						continue
					}
					if format == "docx" {
						found = len(p.all("pageBreakBefore")) > 0
					} else {
						for _, style := range body.all("style") {
							if style.attr("name") != p.attr("style-name") {
								continue
							}
							for _, prop := range style.all("paragraph-properties") {
								found = prop.attr("break-before") == "page"
							}
						}
					}
				}
				if !found {
					t.Error("later act has no native page break")
				}
			})
		}
	}
}
