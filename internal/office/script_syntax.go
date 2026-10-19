// ABOUTME: Retains source-aware script inline tokens after exporter extraction.
// ABOUTME: Keeps script syntax independent of manuscript semantic nodes.
package office

import "regexp"

var (
	boldItalicRE = regexp.MustCompile(`\*\*\*([^*\n]+)\*\*\*`)
	boldRE       = regexp.MustCompile(`\*\*([^*\n]+)\*\*`)
	codeTickRE   = regexp.MustCompile("`([^`\\n]+)`")
	codeOrgRE    = regexp.MustCompile(`=([^=\n]+)=`)
	orgItalicRE  = regexp.MustCompile(`/([^/\n]+?)/`)
	underlineRE  = regexp.MustCompile(`_([^_\n]+)_`)
	starPairRE   = regexp.MustCompile(`\*([^*\n]+)\*`)
)
var scriptRunningTokenRE = regexp.MustCompile(`\[([^\[\]]+)\]`)

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
