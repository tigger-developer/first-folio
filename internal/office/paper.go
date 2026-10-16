// ABOUTME: Resolves physical paper dimensions for native office page layouts.
// ABOUTME: Supports common named papers and explicit millimetre/inch dimensions.
package office

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var paperDimensions = regexp.MustCompile(`^(\d+(?:\.\d+)?)x(\d+(?:\.\d+)?)(mm|in)$`)

func Paper(value string) (float64, float64, error) {
	if m := paperDimensions.FindStringSubmatch(value); m != nil {
		w, err := Points(m[1]+m[3], 12, 0)
		if err != nil {
			return 0, 0, err
		}
		h, err := Points(m[2]+m[3], 12, 0)
		return w, h, err
	}
	sizes := map[string][2]float64{"a0": {841, 1189}, "a1": {594, 841}, "a2": {420, 594}, "a3": {297, 420}, "a4": {210, 297}, "a5": {148, 210}, "a6": {105, 148}, "a7": {74, 105}, "a8": {52, 74}, "a9": {37, 52}, "a10": {26, 37}, "a11": {18, 26}, "letter": {215.9, 279.4}, "us-letter": {215.9, 279.4}, "legal": {215.9, 355.6}, "us-legal": {215.9, 355.6}, "us-tabloid": {279.4, 431.8}, "us-executive": {184.15, 266.7}, "us-statement": {139.7, 215.9}, "us-trade": {152.4, 228.6}, "us-digest": {139.7, 215.9}, "us-mass-market": {107.95, 177.8}}
	if size, ok := sizes[value]; ok {
		return size[0] * 72 / 25.4, size[1] * 72 / 25.4, nil
	}
	for prefix, series := range map[string][][2]float64{"iso-b": {{1000, 1414}, {707, 1000}, {500, 707}, {353, 500}, {250, 353}, {176, 250}, {125, 176}, {88, 125}, {62, 88}}, "iso-c": {{917, 1297}, {648, 917}, {458, 648}, {324, 458}, {229, 324}, {162, 229}, {114, 162}, {81, 114}, {57, 81}}} {
		if strings.HasPrefix(value, prefix) {
			n, err := strconv.Atoi(strings.TrimPrefix(value, prefix))
			if err == nil && n >= 0 && n < len(series) {
				return series[n][0] * 72 / 25.4, series[n][1] * 72 / 25.4, nil
			}
		}
	}
	return 0, 0, fmt.Errorf("office paper dimensions unavailable for %q", value)
}
