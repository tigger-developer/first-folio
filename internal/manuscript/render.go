// ABOUTME: Typst rendering for manuscript documents.
// ABOUTME: Uses a file-backed template and generated block markup.
package manuscript

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

type templateData struct {
	Config Config
	Meta   Metadata
	Header string
	Body   string
}

func RenderTypst(doc Document, cfg Config) (string, error) {
	body, err := renderBlocks(doc.Blocks, cfg)
	if err != nil {
		return "", err
	}
	safeMeta := escapedMetadata(doc.Metadata)
	data := templateData{
		Config: cfg,
		Meta:   safeMeta,
		Header: renderHeader(doc.Metadata, cfg),
		Body:   body,
	}
	root, err := projectRoot()
	if err != nil {
		return "", err
	}
	tmplPath := filepath.Join(root, "templates", "manuscript.typ")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return "", fmt.Errorf("parsing Typst template: %w", err)
	}
	var out bytes.Buffer
	if err := tmpl.Execute(&out, data); err != nil {
		return "", fmt.Errorf("executing Typst template: %w", err)
	}
	return out.String(), nil
}

func renderBlocks(blocks []Block, cfg Config) (string, error) {
	var lines []string
	firstPageBlock := true
	for _, block := range blocks {
		switch block.Kind {
		case "part":
			lines = append(lines, fmt.Sprintf("#folio-part(first: %t)[%s]",
				firstPageBlock,
				caseTransform(block.Text, cfg.Folio.Manuscript.Part.CaseTransform)))
			firstPageBlock = false
		case "chapter":
			lines = append(lines, fmt.Sprintf("#folio-chapter(first: %t)[%s]",
				firstPageBlock,
				typstInline(block.Text, cfg)))
			firstPageBlock = false
		case "section":
			lines = append(lines, "#folio-section["+typstInline(block.Text, cfg)+"]")
		case "paragraph":
			lines = append(lines, "#folio-para["+typstInline(block.Text, cfg)+"]")
		case "scene-break":
			lines = append(lines, "#folio-scene-break()")
		case "code":
			lines = append(lines, "#folio-code["+escapeTypst(block.Text)+"]")
		case "footnote":
			continue
		default:
			return "", fmt.Errorf("unknown manuscript block kind: %s", block.Kind)
		}
		lines = append(lines, "")
	}
	return strings.Join(lines, "\n"), nil
}

func renderHeader(meta Metadata, cfg Config) string {
	header := cfg.Folio.Manuscript.PageHeader.Format
	header = strings.ReplaceAll(header, "[author]", escapeTypst(meta.Author))
	header = strings.ReplaceAll(header, "[title]", escapeTypst(meta.Title))
	header = strings.ReplaceAll(header, "[page]", "#context counter(page).display()")
	return header
}

func typstInline(text string, cfg Config) string {
	monoFont := cfg.Folio.Manuscript.MonoFont
	escaped := escapeTypst(text)
	escaped = replaceInlineCode(escaped, monoFont)
	escaped = regexp.MustCompile(`\*\*([^*]+)\*\*`).ReplaceAllString(escaped, `*$1*`)
	escaped = regexp.MustCompile(`\*([^*]+)\*`).ReplaceAllString(escaped, `_$1_`)
	escaped = regexp.MustCompile(`/([^/]+?)/`).ReplaceAllString(escaped, `_$1_`)
	escaped = regexp.MustCompile(`\[fn:([^\]]+)\]`).ReplaceAllString(escaped, `#footnote[$1]`)
	escaped = regexp.MustCompile(`\[\^([^\]]+)\]`).ReplaceAllString(escaped, `#footnote[$1]`)
	return escaped
}

func replaceInlineCode(text string, monoFont string) string {
	re := regexp.MustCompile("`([^`]+)`")
	return re.ReplaceAllStringFunc(text, func(match string) string {
		content := strings.Trim(match, "`")
		return fmt.Sprintf(`#text(font: "%s")[%s]`, monoFont, content)
	})
}

func escapeTypst(text string) string {
	replacer := strings.NewReplacer(
		`\`, `\\`,
		`[`, `\[`,
		`]`, `\]`,
		`#`, `\#`,
		`$`, `\$`,
		`@`, `\@`,
	)
	return replacer.Replace(text)
}

func caseTransform(text string, mode string) string {
	if mode == "upper" {
		return escapeTypst(strings.ToUpper(text))
	}
	return typstInline(text, Config{})
}

func escapedMetadata(meta Metadata) Metadata {
	return Metadata{
		Title:             escapeTypst(meta.Title),
		Subtitle:          escapeTypst(meta.Subtitle),
		Author:            escapeTypst(meta.Author),
		AuthorAttribution: escapeTypst(meta.AuthorAttribution),
		Date:              escapeTypst(meta.Date),
		Version:           escapeTypst(meta.Version),
		WordCount:         escapeTypst(meta.WordCount),
		Address:           escapeTypst(meta.Address),
		Email:             escapeTypst(meta.Email),
		Website:           escapeTypst(meta.Website),
	}
}

func templateExistsForTests() bool {
	root, err := projectRoot()
	if err != nil {
		return false
	}
	_, err = os.Stat(filepath.Join(root, "templates", "manuscript.typ"))
	return err == nil
}
