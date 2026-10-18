// ABOUTME: Preserves the copyright configuration's small inline Markdown contract.
// ABOUTME: Reuses its bold and italic boundaries without interpreting block or link syntax.
package manuscript

import (
	"github.com/tigger-developer/first-folio/internal/office"
)

func officeCopyrightInlines(text string) []office.Node {
	text = applyMarkdownDashes(text)
	var result []office.Node
	start := 0
	for _, match := range markdownBoldRE.FindAllStringSubmatchIndex(text, -1) {
		result = append(result, officeCopyrightItalic(text[start:match[0]])...)
		result = append(result, office.N("Strong", officeTextLines(text[match[2]:match[3]])))
		start = match[1]
	}
	return append(result, officeCopyrightItalic(text[start:])...)
}

func officeCopyrightItalic(text string) []office.Node {
	var result []office.Node
	start := 0
	for i := 0; i < len(text); i++ {
		if text[i] != '*' || i > 0 && isWordChar(text[i-1]) || i+1 == len(text) || text[i+1] == ' ' {
			continue
		}
		end := -1
		for j := i + 1; j < len(text); j++ {
			if text[j] == '*' && text[j-1] != ' ' && (j+1 == len(text) || !isWordChar(text[j+1])) {
				end = j
				break
			}
		}
		if end < 0 {
			continue
		}
		result = append(result, officeTextLines(text[start:i])...)
		result = append(result, office.N("Emph", officeTextLines(text[i+1:end])))
		start, i = end+1, end
	}
	// Preserve newlines as native line breaks, including publication groups.
	return append(result, officeTextLines(text[start:])...)
}
