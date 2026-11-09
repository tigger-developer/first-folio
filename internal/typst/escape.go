// ABOUTME: Escapes plain content and string literals for Typst templates.
// ABOUTME: Centralizes the base contexts shared by script and letter renderers.
package typst

import "strings"

// HighlightOpen starts the shared pale-yellow inline highlight used for Org underline semantics.
const HighlightOpen = `#highlight(fill: rgb("#FFF59D"))[`

func EscapeContent(value string) string {
	return strings.NewReplacer(
		`\`, `\\`,
		`$`, `\$`,
		`#`, `\#`,
		`@`, `\@`,
		`[`, `\[`,
		`]`, `\]`,
		`*`, `\*`,
		`_`, `\_`,
	).Replace(value)
}

func EscapeString(value string) string {
	return strings.NewReplacer(
		`\`, `\\`,
		`"`, `\"`,
		"\n", `\n`,
		"\r", `\r`,
	).Replace(value)
}
