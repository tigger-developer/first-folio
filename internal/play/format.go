// ABOUTME: Defines stage-play input and output formats and their aliases.
// ABOUTME: Centralizes extension deduction and readable-format validation.
package play

import (
	"fmt"
	"path/filepath"
	"strings"
)

type Format string

const (
	FormatOrg      Format = "org"
	FormatMarkdown Format = "markdown"
	FormatFountain Format = "fountain"
	FormatPDF      Format = "pdf"
	FormatDOCX     Format = "docx"
)

func FormatFromPath(path string) (Format, error) {
	ext := strings.ToLower(filepath.Ext(path))
	if ext == "" {
		return "", fmt.Errorf("No file extension on: %s", path)
	}
	switch ext {
	case ".org":
		return FormatOrg, nil
	case ".md", ".markdown":
		return FormatMarkdown, nil
	case ".fountain", ".ftn":
		return FormatFountain, nil
	case ".pdf", ".typ":
		return FormatPDF, nil
	case ".docx":
		return FormatDOCX, nil
	default:
		return "", fmt.Errorf("Unrecognised file extension: %s", ext)
	}
}

func ParseFormat(name string) (Format, error) {
	switch strings.ToLower(name) {
	case "org":
		return FormatOrg, nil
	case "md", "markdown":
		return FormatMarkdown, nil
	case "fountain", "ftn":
		return FormatFountain, nil
	case "pdf":
		return FormatPDF, nil
	case "docx":
		return FormatDOCX, nil
	default:
		return "", fmt.Errorf("Unrecognised format: %s", name)
	}
}

func (f Format) Readable() bool {
	return f == FormatOrg || f == FormatMarkdown || f == FormatFountain
}
