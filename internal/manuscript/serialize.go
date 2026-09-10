// ABOUTME: Manuscript document serializers for canonical Markdown and org-mode.
// ABOUTME: Keeps Markdown and org contracts in sync through the shared manuscript AST.
package manuscript

import (
	"fmt"
	"strings"
)

func RenderMarkdown(doc Document) string {
	var lines []string
	appendMarkdownFrontmatter(&lines, doc.Metadata)
	appendMarkdownBlocks(&lines, doc.Blocks)
	return strings.TrimRight(strings.Join(lines, "\n"), "\n") + "\n"
}

func appendMarkdownFrontmatter(lines *[]string, meta Metadata) {
	*lines = append(*lines, "---")
	appendFrontmatterLine(lines, "schema", "https://github.com/tigger-developer/first-folio/blob/master/schema/manuscript.md")
	appendFrontmatterLine(lines, "title", meta.Title)
	appendFrontmatterLine(lines, "subtitle", meta.Subtitle)
	appendFrontmatterLine(lines, "author", meta.Author)
	appendFrontmatterLine(lines, "attribution", meta.AuthorAttribution)
	appendFrontmatterLine(lines, "date", meta.Date)
	appendFrontmatterLine(lines, "version", meta.Version)
	appendFrontmatterLine(lines, "wordcount", meta.WordCount)
	appendFrontmatterLine(lines, "contact-name", meta.ContactName)
	appendFrontmatterLine(lines, "address", meta.Address)
	appendFrontmatterLine(lines, "phone", meta.Phone)
	appendFrontmatterLine(lines, "email", meta.Email)
	appendFrontmatterLine(lines, "website", meta.Website)
	*lines = append(*lines, "---")
	*lines = append(*lines, "")
}

func appendFrontmatterLine(lines *[]string, key string, value string) {
	if value != "" {
		*lines = append(*lines, fmt.Sprintf("%s: %q", key, value))
	}
}

func appendMarkdownBlocks(lines *[]string, blocks []Block) {
	for _, block := range blocks {
		switch block.Kind {
		case "part":
			*lines = append(*lines, "# "+block.Text, "")
		case "chapter":
			*lines = append(*lines, "## "+block.Text, "")
		case "section":
			level := block.Level
			if level < 3 {
				level = 3
			}
			*lines = append(*lines, strings.Repeat("#", level)+" "+block.Text, "")
		case "paragraph":
			*lines = append(*lines, block.Text, "")
		case "scene-break":
			*lines = append(*lines, "***", "")
		case "code":
			*lines = append(*lines, "```"+block.Lang, block.Text, "```", "")
		case "raw-typst":
			*lines = append(*lines, block.Text, "")
		}
	}
}
