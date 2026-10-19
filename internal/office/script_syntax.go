// ABOUTME: Retains source-aware script inline tokens after exporter extraction.
// ABOUTME: Keeps script syntax independent of manuscript semantic nodes.
package office

import "regexp"

var scriptRunningTokenRE = regexp.MustCompile(`\[([^\[\]]+)\]`)

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
