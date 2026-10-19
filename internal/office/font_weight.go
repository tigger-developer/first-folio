// ABOUTME: Resolves configured weight names without collapsing them to a bold flag.
// ABOUTME: Keeps explicit generated-role bold overrides independent of source font weight.
package office

import "strconv"

func fontWeight(value string) int {
	weights := map[string]int{"thin": 100, "extralight": 200, "light": 300, "regular": 400, "normal": 400, "medium": 500, "semibold": 600, "bold": 700, "extrabold": 800, "black": 900}
	if weight, ok := weights[value]; ok {
		return weight
	}
	if weight, err := strconv.Atoi(value); err == nil && weight >= 100 && weight <= 900 {
		return weight
	}
	return 400
}

func (s Style) effectiveWeight() int {
	weight := fontWeight(s.Font.Weight)
	if s.Bold != (weight >= 600) {
		if s.Bold {
			return 700
		}
		return 400
	}
	return weight
}

func (s Style) odtWeight() string {
	weight := s.effectiveWeight()
	// ODF admits weight values in steps of 100 only. Preserve the existing
	// binary fallback for intermediate values until distinct-face support exists.
	if weight%100 != 0 {
		if s.Bold {
			return "bold"
		}
		return "normal"
	}
	if weight == 400 {
		return "normal"
	}
	if weight == 700 {
		return "bold"
	}
	return strconv.Itoa(weight)
}
